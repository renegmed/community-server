package main

import (
	"astoria-serverless-event-multi/db"
	"astoria-serverless-event-multi/model"
	"astoria-serverless-event-multi/util"
	"fmt"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func Handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	log.Println("Received request body for deletion: ", request.Body)

	var slackRequest model.SlackRequest

	err := util.ParseSlackRequest(request.Body, &slackRequest)
	if err != nil {
		log.Printf("--- error during parsing of request body: %v", err)
		r := slackRequest.EventResponse(
			"Your request cannot be processed. Possible malformed request text.",
			[]error{err}, true).Text
		return events.APIGatewayProxyResponse{Body: r, StatusCode: 200}, nil
	}

	var event db.Event

	fieldsTable, errs := util.ParseToFields(&event, slackRequest.Text)
	if len(errs) > 0 {
		fmt.Printf("Problem on parsing slack request: %v\n", errs)
		r := slackRequest.EventResponseWithFields(
			"Your request cannot be processed. Possibly the request format is not valid.",
			errs, slackRequest.Debug, fieldsTable).Text
		return events.APIGatewayProxyResponse{Body: r, StatusCode: 200}, nil
	}

	if event.Title == "" {
		log.Printf("Invalid. Title is empty.")
		r := slackRequest.EventResponseWithFields(
			"Your request cannot be processed. Title is missing",
			[]error{}, slackRequest.Debug, fieldsTable).Text
		return events.APIGatewayProxyResponse{Body: r, StatusCode: 200}, nil
	}

	dbService, err := db.NewEventService()
	if err != nil {
		log.Printf("Problem on creating event db service: %v\n", err)
		r := slackRequest.EventResponseWithFields(
			"Your request cannot be processed. Problem with event service for database.",
			[]error{err}, slackRequest.Debug, fieldsTable).Text
		return events.APIGatewayProxyResponse{Body: r, StatusCode: 200}, nil
	}

	ev, err := dbService.GetEventByTitle(event.Title)
	if err != nil {
		log.Printf("Db service problem on calling GetEventByTitle() function: %v\n", err)
		r := slackRequest.EventResponseWithFields(
			"Your request cannot be processed. Problem with database service function.",
			[]error{err}, slackRequest.Debug, fieldsTable).Text
		return events.APIGatewayProxyResponse{Body: r, StatusCode: 200}, nil
	}

	if ev.Title == "" {
		log.Printf("Title '%s' doesn't exist.\n", ev.Title)
		r := slackRequest.EventResponseWithFields(
			"Your request cannot be processed. Record based on title not found.",
			[]error{err}, slackRequest.Debug, fieldsTable).Text
		return events.APIGatewayProxyResponse{Body: r, StatusCode: 200}, nil
	}

	log.Printf("--- event gathered after calling GetEventByTitle '%s' \n RESULT: %v\n", event.Title, ev)

	_, err = dbService.DeleteEvent(ev.ID)
	if err != nil {
		log.Printf("Problem on deleting an event: %v\n", err)
		r := slackRequest.EventResponseWithFields(
			"Your request cannot be processed. Problem deleting an event.",
			[]error{err}, slackRequest.Debug, fieldsTable).Text
		return events.APIGatewayProxyResponse{Body: r, StatusCode: 200}, nil
	}
	r := slackRequest.EventResponseWithFields(
		"Your request for event deletion was processed successfully.",
		[]error{}, false, fieldsTable).Text
	return events.APIGatewayProxyResponse{Body: r, StatusCode: 200}, nil
}

func main() {
	log.Println("... Started delete event.... ")
	lambda.Start(Handler)
}
