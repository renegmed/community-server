package main

import (
	"astoria-serverless-event-multi/db"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func Handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	/*
		Headers                         map[string]string             `json:"headers"`
		MultiValueHeaders               map[string][]string           `json:"multiValueHeaders"`
		QueryStringParameters           map[string]string             `json:"queryStringParameters"`
		MultiValueQueryStringParameters map[string][]string           `json:"multiValueQueryStringParameters"`
		PathParameters                  map[string]string             `json:"pathParameters"`
		StageVariables                  map[string]string             `json:"stageVariables"`

	*/

	log.Printf("Headers:\n %v\n", request.Headers)
	log.Printf("MultiValueHeaders:\n %v\n", request.MultiValueHeaders)
	log.Printf("QueryStringParameters:\n %v\n", request.QueryStringParameters)
	log.Printf("MultiValueQueryStringParameters:\n %v\n", request.MultiValueQueryStringParameters)
	log.Printf("PathParameters:\n %v\n", request.PathParameters)
	log.Printf("StageVariables:\n %v\n", request.StageVariables)

	// log.Println("Received request body for events by category: ", request.StageVariables)

	// var slackRequest model.SlackRequest

	// err := util.ParseSlackRequest(request.Body, &slackRequest)
	// if err != nil {
	// 	log.Printf("--- error during parsing of request body: %v", err)
	// 	r := slackRequest.EventResponse(
	// 		"Your request cannot be processed. Possible malformed request text.",
	// 		[]error{err}, true).Text
	// 	return events.APIGatewayProxyResponse{Body: r, StatusCode: 200}, nil
	// }

	// var event db.Event

	// fieldsTable, errs := util.ParseToFields(&event, slackRequest.Text)
	// if len(errs) > 0 {
	// 	fmt.Printf("Problem on parsing slack request: %v\n", errs)
	// 	r := slackRequest.EventResponseWithFields(
	// 		"Your request cannot be processed. Possibly the request format is not valid.",
	// 		errs, slackRequest.Debug, fieldsTable).Text
	// 	return events.APIGatewayProxyResponse{Body: r, StatusCode: 200}, nil
	// }

	fieldsTable := request.QueryStringParameters

	log.Printf("Fields table category: %s\n", fieldsTable["category"])

	if _, ok := fieldsTable["category"]; !ok {
		r := fmt.Sprintf("No passed parameter 'category' found.")
		return events.APIGatewayProxyResponse{Body: r, StatusCode: 200}, nil
	}

	dbService, err := db.NewEventService()
	if err != nil {
		r := fmt.Sprintf("Problem on creating event db service: %v", err)
		return events.APIGatewayProxyResponse{Body: r, StatusCode: 200}, nil
	}

	evts, err := dbService.GetEventsByCategory(fieldsTable["category"])
	if err != nil {
		r := fmt.Sprintf("Problem getting events by category: %v", err)
		return events.APIGatewayProxyResponse{Body: r, StatusCode: 200}, nil
	}
	if len(evts) == 0 {
		return events.APIGatewayProxyResponse{Body: "", StatusCode: 200}, nil
	}

	sb := strings.Builder{}
	sb.WriteString("{ \"events\": [")
	for i := 0; i < len(evts); i++ {
		jsonItem, _ := json.Marshal(evts[i])

		sb.WriteString(string(jsonItem))
		if i != len(evts)-1 { // if not the last element of array
			sb.WriteString(",\n")
		}
	}
	sb.WriteString("]}\n")

	log.Println("--- string events by category: \n", sb.String())

	resp := events.APIGatewayProxyResponse{Body: sb.String(), Headers: make(map[string]string), StatusCode: 200}
	resp.Headers["Access-Control-Allow-Origin"] = "*"

	return resp, nil
}

func main() {
	lambda.Start(Handler)
}
