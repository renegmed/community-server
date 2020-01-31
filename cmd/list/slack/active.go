package main

import (
	"astoria-serverless-event-multi/db"
	"astoria-serverless-event-multi/model"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func Handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var slackRequest model.SlackRequest

	dbService, err := db.NewEventService()
	if err != nil {
		log.Printf("Problem on creating event db service: %v", err)
		r := slackRequest.EventResponse(
			"Your request cannot be processed. Problem with database service.", []error{err}, true).Text
		return events.APIGatewayProxyResponse{Body: r, StatusCode: 200}, nil
	}

	evts, err := dbService.GetActiveEvents()
	if err != nil {
		log.Printf("Problem getting access to data with db service: %v", err)
		r := slackRequest.EventResponse(
			"Your request cannot be processed. Problem getting access to data.", []error{err}, true).Text
		return events.APIGatewayProxyResponse{Body: r, StatusCode: 200}, nil
	}
	if len(evts) == 0 {
		r := slackRequest.EventResponse(
			"Sorry. No record found.", []error{}, false).Text
		return events.APIGatewayProxyResponse{Body: r, StatusCode: 200}, nil
	}

	sb := strings.Builder{}
	sb.WriteString("Title\t\tStatus\tStreet\n")
	sb.WriteString("------------------------------------------\n")
	for _, evt := range evts {
		sb.WriteString(fmt.Sprintf("%s\t\t%s\t%s\n", evt.Title, evt.EvtStatus, evt.Street))
	}
	log.Println("--- string events: \n", sb.String())

	return events.APIGatewayProxyResponse{Body: sb.String(), StatusCode: 200}, nil
}

func main() {
	lambda.Start(Handler)
}
