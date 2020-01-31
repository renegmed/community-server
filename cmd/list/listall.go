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
	dbService, err := db.NewEventService()
	if err != nil {
		r := fmt.Sprintf("Problem on creating event db service: %v", err)
		return events.APIGatewayProxyResponse{Body: r, StatusCode: 200}, nil
	}

	evts, err := dbService.GetAllEvents()
	if err != nil {
		r := fmt.Sprintf("Problem getting events: %v", err)
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

	log.Println("--- string events: \n", sb.String())

	resp := events.APIGatewayProxyResponse{Body: sb.String(), Headers: make(map[string]string), StatusCode: 200}
	resp.Headers["Access-Control-Allow-Origin"] = "*"

	return resp, nil
}

func main() {
	lambda.Start(Handler)
}
