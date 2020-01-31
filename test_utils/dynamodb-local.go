package test_utils

import (
	"astoria-serverless-event-multi/db"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/guregu/dynamo"
	"github.com/rs/xid"
)

func NewEventService() (*db.EventService, error) {
	dynamoTable, err := CreateItemTable(db.Event{})
	if err != nil {
		log.Printf("Not able to create dynamo table: %v\n", err)
		return nil, err
	}

	service := &db.EventService{
		Table: dynamoTable,
	}

	return service, nil
}

func CreateItemTable(table interface{}) (dynamo.Table, error) {
	cfg := aws.Config{
		Endpoint:                      aws.String("http://localhost:9000"),
		Region:                        aws.String("us-east-1"),
		CredentialsChainVerboseErrors: aws.Bool(false),
	}

	sess := session.Must(session.NewSession())
	ddb := dynamo.New(sess, &cfg)
	tableName := xid.New().String()

	log.Println("--- Table name: ", tableName)

	err := ddb.CreateTable(tableName, table).Run()
	if err != nil {
		log.Printf("..... Problem creating table:\n%v\n", err)
		return dynamo.Table{}, err
	}

	dbtable := ddb.Table(tableName)
	return dbtable, err

}
