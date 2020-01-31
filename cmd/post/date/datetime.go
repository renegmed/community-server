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
	log.Println("Received request body for date times : ", request.Body)

	var slackRequest model.SlackRequest

	err := util.ParseSlackRequest(request.Body, &slackRequest)
	if err != nil {
		log.Printf("--- error during parsing of request body: %v", err)
		r := fmt.Sprintf("--- error during parsing of request body: %v", err)
		return events.APIGatewayProxyResponse{Body: r, StatusCode: 200}, nil
	}

	var event db.Event

	fieldsTable, err := util.ParseToFieldsForUpdate(&event, slackRequest.Text)
	if err != nil {
		log.Printf("--- error during parsing of request body: %v", err)
		r := slackRequest.EventResponse(
			"Your request cannot be processed. Possible malformed request text.",
			[]error{err}, true).Text
		return events.APIGatewayProxyResponse{Body: r, StatusCode: 200}, nil
	}

	log.Printf("Event based on slack request: %v\n", event)

	// validate datetimes here

	// convert date Oct. 28, 2019 6:15pm to 2019-10-28T18:15:00-04:00
	formattedDatetime, err := util.FormatDatetime(event.StartDatetime, db.LayoutUS)
	if err != nil {
		log.Printf("Problem converting start date time: %v\n", err)
		s := fmt.Sprintf("Your request cannot be processed. Problem converting start date time for '%s'\n", event.Title)
		r := slackRequest.EventResponse(s, []error{err}, true).Text
		return events.APIGatewayProxyResponse{Body: r, StatusCode: 200}, nil
	}
	event.StartDatetime = formattedDatetime

	formattedDatetime, err = util.FormatDatetime(event.EndDatetime, db.LayoutUS)
	if err != nil {
		log.Printf("Problem converting end date time: %v\n", err)
		s := fmt.Sprintf("Your request cannot be processed. Problem converting end date time for '%s'\n", event.Title)
		r := slackRequest.EventResponse(s, []error{err}, true).Text
		return events.APIGatewayProxyResponse{Body: r, StatusCode: 200}, nil
	}
	event.EndDatetime = formattedDatetime

	dbService, err := db.NewEventService()
	if err != nil {
		log.Printf("Problem on creating event db service: %v\n", err)
		r := slackRequest.EventResponse("Your request cannot be processed. Problem creating db service\n",
			[]error{err}, true).Text
		return events.APIGatewayProxyResponse{Body: r, StatusCode: 200}, nil
	}

	ev, err := dbService.GetEventByTitle(event.Title)
	if err != nil {
		log.Printf("Problem on getting event by title: %s - %v\n", event.Title, err)
		r := slackRequest.EventResponseWithFields(
			"Your request cannot be processed. Problem with database service function.",
			[]error{err}, slackRequest.Debug, fieldsTable).Text
		return events.APIGatewayProxyResponse{Body: r, StatusCode: 200}, nil
	}

	if ev.Title == "" { // event title already existed
		log.Printf("Title '%s' doesn't exist.\n", ev.Title)
		r := slackRequest.EventResponseWithFields(
			"Your request cannot be processed. Record based on title not found.",
			[]error{err}, slackRequest.Debug, fieldsTable).Text
		return events.APIGatewayProxyResponse{Body: r, StatusCode: 200}, nil
	}

	updateCurrentEvent(&ev, event, fieldsTable)

	log.Printf("Event to be updated to db:\n %v\n", ev)

	_, err = dbService.UdpateEvent(ev)
	if err != nil {
		log.Printf("Problem db service updating date/times %s: %v\n", ev.Title, err)
		r := slackRequest.EventResponseWithFields(
			"Your request cannot be processed. Problem db service updating data/times.",
			[]error{err}, slackRequest.Debug, fieldsTable).Text
		return events.APIGatewayProxyResponse{Body: r, StatusCode: 200}, nil
	}
	r := slackRequest.EventResponseWithFields(
		"Your request for date/time update was processed successfully.",
		[]error{}, false, fieldsTable).Text
	return events.APIGatewayProxyResponse{Body: r, StatusCode: 200}, nil
}

func updateCurrentEvent(ev *db.Event, e db.Event, table map[string]string) {
	if util.FieldExists("start", table) {
		ev.StartDatetime = e.StartDatetime // or table["start"]
	}

	if util.FieldExists("end", table) {
		ev.EndDatetime = e.EndDatetime // or table["end"]
	}
}

func main() {
	lambda.Start(Handler)
}
