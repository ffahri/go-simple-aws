package processor

import (
	"context"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"go.opentelemetry.io/otel/api/global"
	"go.opentelemetry.io/otel/api/trace"
	"simpleawsgo/pkg/client"
	"simpleawsgo/pkg/model"
	"strconv"
	"sync"
)

type Service struct {
	SQS         *sqs.SQS
	DynamoDB    *dynamodb.DynamoDB
	QueueURL    *string
	ThreadCount int
}

const QUEUE_NAME = "queueName"
const TABLE_NAME = "tableName"
const THREAD_COUNT = "threadCount"

func (s *Service) Init() error {
	s.SQS = client.InitSQS()
	out, err := s.SQS.GetQueueUrl(&sqs.GetQueueUrlInput{QueueName: aws.String(viper.Get(QUEUE_NAME).(string))})
	if err != nil {
		return err
	}
	s.QueueURL = out.QueueUrl
	s.ThreadCount = viper.GetInt(THREAD_COUNT)
	log.Info().Timestamp().Msg("Queue found ! : " + *s.QueueURL)
	s.DynamoDB = client.InitDynamoDB()
	return nil
}

func (s *Service) StartPoller() {
	wg := &sync.WaitGroup{}
	wg.Add(s.ThreadCount)
	for i := 0; i < s.ThreadCount; i++ {
		go s.poll() //didn't added wg.done since pollers need to run always - todo is that a good practice ? - how to stop ?
	}
	wg.Wait()
}

func (s *Service) poll() {
	log.Info().Msg("Poller started")
	for true {
		out, err := s.SQS.ReceiveMessage(&sqs.ReceiveMessageInput{
			MaxNumberOfMessages:   aws.Int64(10),
			QueueUrl:              s.QueueURL,
			WaitTimeSeconds:       aws.Int64(1),
			MessageAttributeNames: aws.StringSlice([]string{"traceID", "spanID", "traceFlag"}),
		})

		if err != nil {
			log.Err(err).Timestamp().Msg("Error happened while receiving message")
		}
		if len(out.Messages) > 0 {
			go func() { //todo check maybe potentially goroutine leak?
				err, entries := s.writeDynamoDB(out.Messages)
				if err != nil {
					log.Err(err).Timestamp().Msg("error")
				} else {
					_, err = s.SQS.DeleteMessageBatch(&sqs.DeleteMessageBatchInput{
						Entries:  entries,
						QueueUrl: s.QueueURL,
					})
					if err != nil {
						log.Err(err).Msg("Could not deleted message batch from sqs")
					}
				}
			}()
		}
	}
}

func (s *Service) writeDynamoDB(messageList []*sqs.Message) (error, []*sqs.DeleteMessageBatchRequestEntry) {
	var wrArr []*dynamodb.WriteRequest
	var reqEntryArr []*sqs.DeleteMessageBatchRequestEntry
	for _, message := range messageList {
		traceID := message.MessageAttributes["traceID"].StringValue
		spanID := message.MessageAttributes["spanID"].StringValue

		traceId, err := trace.IDFromHex(*traceID)
		spanId, err := trace.SpanIDFromHex(*spanID)

		sc2 := trace.SpanContext{
			TraceID:    traceId,
			SpanID:     spanId,
			TraceFlags: byte(0x01),
		}

		ctx := trace.ContextWithRemoteSpanContext(context.Background(), sc2)
		_, span := global.Tracer("").Start(ctx, "writeMessageToDynamoDb")
		defer span.End()
		mdl, err := model.UnmarshallModel(message.Body)
		if err != nil {
			span.SetAttribute("err", err.Error())
			//collect and send err in end
		} else {
			wrArr = append(wrArr, &dynamodb.WriteRequest{
				PutRequest: &dynamodb.PutRequest{
					Item: map[string]*dynamodb.AttributeValue{
						"modelId": {
							S: aws.String(mdl.Id),
						},
						"name": {
							S: aws.String(mdl.Name),
						},
						"value": {
							N: aws.String(strconv.Itoa(mdl.Value)),
						},
					},
				},
			})
			reqEntryArr = append(reqEntryArr, &sqs.DeleteMessageBatchRequestEntry{
				Id:            message.MessageId,
				ReceiptHandle: message.ReceiptHandle,
			})
		}
	}
	_, err := s.DynamoDB.BatchWriteItem(&dynamodb.BatchWriteItemInput{
		RequestItems: map[string][]*dynamodb.WriteRequest{
			viper.GetString(TABLE_NAME): wrArr,
		},
	})
	if err != nil {
		return err, nil
	}

	return nil, reqEntryArr
}
