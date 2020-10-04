package client

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/sqs"
)

type AWSClient struct {
	SQS      *sqs.SQS
	DynamoDB *dynamodb.DynamoDB
}

func InitSQS() *sqs.SQS {
	mySession := session.Must(session.NewSession(&aws.Config{
		Region: aws.String("eu-central-1")}))
	return sqs.New(mySession)
}
func InitDynamoDB() *dynamodb.DynamoDB {
	mySession := session.Must(session.NewSession(&aws.Config{
		Region: aws.String("eu-central-1")}))
	return dynamodb.New(mySession)

}
