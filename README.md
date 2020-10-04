# go-simple-aws
Simple go application which receives message from API then send it to SQS, another application will consume and write to DynamoDB

## API 
Responsible for receiving the request from HTTP then sends payload to Amazon SQS.


## Processor
Polls the SQS then writes to Amazon DynamoDB
