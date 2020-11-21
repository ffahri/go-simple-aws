package api

import (
	"context"
	"encoding/json"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"go.opentelemetry.io/otel/api/global"
	"io/ioutil"
	"simpleawsgo/pkg/client"
	"simpleawsgo/pkg/model"
)

type Service struct {
	SQS      *sqs.SQS
	QueueURL *string
}

const QUEUE_NAME = "queueName"

func (s *Service) Init() error {
	s.SQS = client.InitSQS()
	out, err := s.SQS.GetQueueUrl(&sqs.GetQueueUrlInput{QueueName: aws.String(viper.Get(QUEUE_NAME).(string))})
	if err != nil {
		return err
	}
	s.QueueURL = out.QueueUrl
	log.Info().Timestamp().Msg("Queue found ! : " + *s.QueueURL)
	return nil
}

type scBasic struct {
	TraceID    []byte
	SpanID     []byte
	TraceFlags byte
}

func (s *Service) sendMessage(mdl model.TestModel, ctx context.Context) error {
	_, span := global.Tracer("").Start(ctx, "sendMessageToQueue")
	defer span.End()
	arr, err := model.MarshallModel(mdl)
	if err != nil {
		return err
	}
	sc := span.SpanContext()
	tid := sc.TraceID.String()
	sid := sc.SpanID.String()

	out, err := s.SQS.SendMessage(&sqs.SendMessageInput{
		MessageAttributes: map[string]*sqs.MessageAttributeValue{
			"contentType": {
				DataType:    aws.String("String"),
				StringValue: aws.String("application/json"),
			},
			"traceID": {
				DataType:    aws.String("String"),
				StringValue: aws.String(tid),
			},
			"spanID": {
				DataType:    aws.String("String"),
				StringValue: aws.String(sid),
			}},

		MessageBody: aws.String(string(arr)),
		QueueUrl:    s.QueueURL,
	})
	if err != nil {
		return err
	}
	span.SetAttribute("messageId", out.MessageId)
	return nil
}

func (s *Service) SendHandler(c *gin.Context) {
	ctx, span := global.Tracer("").Start(c.Request.Context(), "sendHandler")
	defer span.End()
	arr, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(400, "bad json")
		return
	}
	mdl := model.TestModel{}
	if err := json.Unmarshal(arr, &mdl); err != nil {
		c.JSON(400, "bad request")
		return
	}
	if err := model.CheckModelFields(mdl); err != nil {
		c.JSON(400, "bad request")
		return
	}
	span.SetAttribute("id", mdl.Id)
	go func() {
		err := s.sendMessage(mdl, ctx)
		log.Err(err).Msg("Error happened while sending message to sqs") //todo alert - retry based on exception
	}()

	c.JSON(201, "success")
	return

}
