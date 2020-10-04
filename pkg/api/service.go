package api

import (
	"encoding/json"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
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

func (s *Service) sendMessage(mdl model.TestModel) error {
	arr, err := model.MarshallModel(mdl)
	if err != nil {
		return err
	}
	_, err = s.SQS.SendMessage(&sqs.SendMessageInput{
		MessageAttributes: map[string]*sqs.MessageAttributeValue{
			"contentType": {
				DataType:    aws.String("String"),
				StringValue: aws.String("application/json")},
		},
		MessageBody: aws.String(string(arr)),
		QueueUrl:    s.QueueURL,
	})
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) SendHandler(c *gin.Context) {
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
	if err := s.sendMessage(mdl); err != nil {
		c.JSON(500, "internal server error")
		return
	} else {
		c.JSON(201, "success")
	}
}
