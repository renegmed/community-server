package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"astoria-serverless-event-multi/db"
	"astoria-serverless-event-multi/geolocation"
	"astoria-serverless-event-multi/model"
	"astoria-serverless-event-multi/util"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/kelvins/geocoder"
)

func Handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	log.Println("Received request body: ", request.Body)
	/*
		token=4lnUutTDuMrXfnYnNOcdhM1E&team_id=TP0N9QXC1&team_domain=golang-slack-dev&channel_id=CP0N9S5CZ&channel_name=business-processes-automation&user_id=UNVLVE8E7&user_name=renegmed&command=%2Fnew%2Fevent&text=--title+Street+Fair+November+2019+--street+3959+58th+St.+--city+Woodside+--county+Queens+--state+NY+--country+US+--postalcode+11388&response_url=https%3A%2F%2Fhooks.slack.com%2Fcommands%2FTP0N9QXC1%2F810455995556%2F8K3na1ulJoyVbMRBQEPt6vic&trigger_id=812776680071.782757847409.6d14f9a35f7a7ccf0b4198f3891741a6
	*/

	apiKey := aws.String(os.Getenv("GOOGLEGEO_API_KEY"))

	//fmt.Printf("------google geo api key: %s\n", *apiKey)

	var slackRequest model.SlackRequest

	err := util.ParseSlackRequest(request.Body, &slackRequest)
	if err != nil {
		log.Printf("--- error during parsing of request body: %v", err)
		r := slackRequest.EventResponse(
			"Your request cannot be processed because of possible malformed slack request text",
			[]error{}, true).Text
		return events.APIGatewayProxyResponse{Body: r, StatusCode: 200}, nil
	}

	log.Printf("------slack request:\n %v\n", slackRequest)

	var event db.Event

	event.TeamID = slackRequest.TeamID
	event.TeamDomain = slackRequest.TeamDomain
	event.ChannelId = slackRequest.ChannelID
	event.Channel = slackRequest.ChannelName

	fieldsTable, errs := util.ParseToFields(&event, slackRequest.Text)
	if errs != nil {
		log.Printf("Problem on parsing title and address: %v\n", errs)
		r := slackRequest.EventResponseWithFields(
			"Your request cannot be processed. Possible malformed request text",
			errs, slackRequest.Debug, fieldsTable).Text
		return events.APIGatewayProxyResponse{Body: r, StatusCode: 200}, nil
	}

	if default_, ok := fieldsTable["default"]; ok {
		switch strings.ToLower(default_) {
		case "astoria-mesh":
			event.City = "Astoria"
			event.County = "Queens"
			event.State = "NY"
			event.Country = "US"
			event.EvtCategory = "mesh"
			fieldsTable["city"] = event.City
			fieldsTable["county"] = event.County
			fieldsTable["state"] = event.State
			fieldsTable["country"] = event.Country
			fieldsTable["category"] = event.EvtCategory
		case "astoria-community":
			event.City = "Astoria"
			event.County = "Queens"
			event.State = "NY"
			event.Country = "US"
			event.EvtCategory = "community"
			fieldsTable["city"] = event.City
			fieldsTable["county"] = event.County
			fieldsTable["state"] = event.State
			fieldsTable["country"] = event.Country
			fieldsTable["category"] = event.EvtCategory
		default:
			log.Printf("Default parameter has invalid value: %s\n", strings.ToLower(default_))
			r := slackRequest.EventResponseWithFields(
				"Default parameter has invalid value.",
				[]error{err}, slackRequest.Debug, fieldsTable).Text
			return events.APIGatewayProxyResponse{Body: r, StatusCode: 200}, nil
		}
	}

	log.Printf("------event after parsing :\n %v\n", event)

	// validate fields requirement for creating new event
	errs = validateFields(fieldsTable)
	if errs != nil {
		log.Printf("Problem on fields for creating new event: %v\n", errs)
		r := slackRequest.EventResponseWithFields(
			"Your request cannot be processed. Some fields are required",
			errs, slackRequest.Debug, fieldsTable).Text
		return events.APIGatewayProxyResponse{Body: r, StatusCode: 200}, nil
	}

	geoLocCoder, err := geolocation.NewGoogleGeoCoder(*apiKey, "")
	if err != nil {
		log.Printf("Problem creating geo coder service: %v\n", err)
		r := slackRequest.EventResponseWithFields(
			"Your request cannot be processed. Problem with creating geographical coder service.",
			[]error{err}, slackRequest.Debug, fieldsTable).Text
		return events.APIGatewayProxyResponse{Body: r, StatusCode: 200}, nil
	}

	log.Println("--- geoLocCoder is ready for coordinates determination ---.")

	address := fmt.Sprintf("%s %s %s %s", event.Street, event.City, event.State, event.Country)
	geoLoc, err := geoLocCoder.GeoLocationRequest(address)
	if err != nil {
		log.Printf("Problem processing addres for geo location %v\n", err)
		r := slackRequest.EventResponseWithFields(
			"Your request cannot be processed. Possible invalid address information.",
			[]error{err}, slackRequest.Debug, fieldsTable).Text
		return events.APIGatewayProxyResponse{Body: r, StatusCode: 200}, nil
	}
	// Get the results (latitude and longitude)
	var location geocoder.Location
	location.Latitude = geoLoc.Results[0].Geometry.Location.Lat
	location.Longitude = geoLoc.Results[0].Geometry.Location.Lng
	event.Lat = location.Latitude
	event.Lon = location.Longitude

	log.Printf("------event to be stored :\n %v\n", event)
	dbService, err := db.NewEventService()
	if err != nil {
		log.Printf("Problem on creating event db service: %v\n", err)
		r := slackRequest.EventResponseWithFields(
			"Your request cannot be processed. Problem with event service for database.",
			[]error{err}, slackRequest.Debug, fieldsTable).Text
		return events.APIGatewayProxyResponse{Body: r, StatusCode: 200}, nil
	}
	e, err := dbService.GetEventByTitle(event.Title)
	if err != nil {
		log.Printf("Db service problem on calling GetEventByTitle() function: %v\n", err)
		r := slackRequest.EventResponseWithFields(
			"Your request cannot be processed. Problem with database service function.",
			[]error{err}, slackRequest.Debug, fieldsTable).Text
		return events.APIGatewayProxyResponse{Body: r, StatusCode: 200}, nil
	}
	if e.Title != "" {
		log.Printf("Title '%s' already exists.\n", e.Title)
		r := slackRequest.EventResponseWithFields(
			"Your request cannot be processed. Title already existed. Try to use different title.",
			[]error{}, slackRequest.Debug, fieldsTable).Text
		return events.APIGatewayProxyResponse{Body: r, StatusCode: 200}, nil
	}

	_, err = dbService.CreateNewEvent(event)
	if err != nil {
		log.Printf("Problem on storing new event: %v\n", err)
		r := slackRequest.EventResponseWithFields(
			"Your request cannot be processed. Problem creating new event.",
			[]error{err}, slackRequest.Debug, fieldsTable).Text
		return events.APIGatewayProxyResponse{Body: r, StatusCode: 200}, nil
	}

	r := slackRequest.EventResponseWithFields(
		"Your request for create new event was successfully processed.",
		[]error{}, false, fieldsTable).Text
	return events.APIGatewayProxyResponse{Body: r, StatusCode: 200}, nil
}

func main() {
	lambda.Start(Handler)
}

func validateFields(flds map[string]string) []error {

	var errors []error
	if flds == nil || len(flds) == 0 {
		errors = append(errors, fmt.Errorf("No valid field found."))
	}
	if _, ok := flds["title"]; !ok {
		errors = append(errors, fmt.Errorf("title is missing"))
	}

	if _, ok := flds["street"]; !ok {
		errors = append(errors, fmt.Errorf("street is missing"))
	}

	if _, ok := flds["city"]; !ok {
		errors = append(errors, fmt.Errorf("city is missing."))
	}

	if _, ok := flds["county"]; !ok {
		errors = append(errors, fmt.Errorf("county is missing."))
	}

	if _, ok := flds["country"]; !ok {
		errors = append(errors, fmt.Errorf("country is missing."))
	}

	return errors
}
