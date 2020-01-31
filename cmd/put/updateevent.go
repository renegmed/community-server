package main

import (
	"astoria-serverless-event-multi/db"
	"astoria-serverless-event-multi/geolocation"
	"astoria-serverless-event-multi/model"
	"astoria-serverless-event-multi/util"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/kelvins/geocoder"
)

func Handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	log.Println("Received request body: ", request.Body)

	var slackRequest model.SlackRequest

	// parse slack request fields/values

	// get event from storage based on provided title

	// validate event title if already existed

	// update event with field/value entries

	// validate values startdate/enddate

	// update event storage

	err := util.ParseSlackRequest(request.Body, &slackRequest)
	if err != nil {
		log.Printf("--- error during parsing of request body: %v", err)
		r := slackRequest.EventResponse(
			"Your request cannot be processed. Possible malformed request text.",
			[]error{err}, true).Text
		return events.APIGatewayProxyResponse{Body: r, StatusCode: 200}, nil
	}

	var newEvent db.Event

	fieldsTable, err := util.ParseToFieldsForUpdate(&newEvent, slackRequest.Text)
	if err != nil {
		log.Printf("Problem on parsing slack request to fields: %v\n", err)
		r := slackRequest.EventResponseWithFields(
			"Your request cannot be processed. Possibly an invalid request text structure.",
			[]error{err}, slackRequest.Debug, fieldsTable).Text
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

	ev, err := dbService.GetEventByTitle(newEvent.Title)
	if err != nil {
		log.Printf("Db service problem on calling GetEventByTitle() function: %v\n", err)
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

	// if title is being changed, update the newEvent title to be copied over the current/stored event title
	if util.FieldExists("new-title", fieldsTable) {
		log.Printf("---new-title: ---%s---", fieldsTable["new-title"])
		newEvent.Title = fieldsTable["new-title"]
	}

	updateCurrentEvent(&ev, newEvent, fieldsTable) //update ev with newEvent fields listed in fieldsTable

	if addressChange(fieldsTable) { // if there is address fields change, recompute geo location
		apiKey := aws.String(os.Getenv("GOOGLEGEO_API_KEY"))
		geoLocCoder, err := geolocation.NewGoogleGeoCoder(*apiKey, "")
		if err != nil {
			log.Printf("Problem creating geo coder service: %v\n", err)
			r := slackRequest.EventResponseWithFields(
				"Your request cannot be processed. Problem with creating geographical coder service.",
				[]error{err}, slackRequest.Debug, fieldsTable).Text
			return events.APIGatewayProxyResponse{Body: r, StatusCode: 200}, nil
		}

		log.Println("--- geoLocCoder is ready for coordinates determination ---.")

		address := fmt.Sprintf("%s %s %s %s", ev.Street, ev.City, ev.State, ev.Country)
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
		ev.Lat = location.Latitude
		ev.Lon = location.Longitude
	}

	// update event from storage
	log.Printf("--- Updated event for storage: %v\n", ev)

	_, err = dbService.UdpateEvent(ev)
	if err != nil {
		log.Printf("Problem on storing new event: %v\n", err)
		r := slackRequest.EventResponseWithFields(
			"Your request cannot be processed. Problem updating event.",
			[]error{err}, slackRequest.Debug, fieldsTable).Text
		return events.APIGatewayProxyResponse{Body: r, StatusCode: 200}, nil
	}

	r := slackRequest.EventResponseWithFields(
		"Your request for event update was processed successfully.",
		[]error{}, false, fieldsTable).Text
	return events.APIGatewayProxyResponse{Body: r, StatusCode: 200}, nil
}

func updateCurrentEvent(ev *db.Event, e db.Event, table map[string]string) {
	if util.FieldExists("new-title", table) { // case on title is being changed
		ev.Title = e.Title
	}
	if util.FieldExists("description", table) {
		ev.Description = e.Description
	}

	if util.FieldExists("street", table) {
		ev.Street = e.Street
	}

	if util.FieldExists("city", table) {
		ev.City = e.City
	}

	if util.FieldExists("county", table) {
		ev.County = e.County
	}
	if util.FieldExists("state", table) {
		ev.State = e.State
	}
	if util.FieldExists("postalcode", table) {
		ev.PostalCode = e.PostalCode
	}
	if util.FieldExists("country", table) {
		ev.Country = e.Country
	}
	if util.FieldExists("start", table) {
		ev.StartDatetime = e.StartDatetime
	}
	if util.FieldExists("end", table) {
		ev.EndDatetime = e.EndDatetime
	}
	if util.FieldExists("contact", table) {
		ev.Contact = e.Contact
	}
	if util.FieldExists("email", table) {
		ev.Email = e.Email
	}
	if util.FieldExists("status", table) {
		ev.EvtStatus = e.EvtStatus
	}
	if util.FieldExists("category", table) {
		ev.EvtCategory = e.EvtCategory
	}
	if util.FieldExists("subcategory", table) {
		ev.EvtSubCategory = e.EvtSubCategory
	}
	if util.FieldExists("link", table) {
		ev.Link = e.Link
	}
	if util.FieldExists("linklabel", table) {
		ev.LinkLabel = e.LinkLabel
	}
}

func addressChange(table map[string]string) bool {
	if util.FieldExists("street", table) || util.FieldExists("city", table) || util.FieldExists("country", table) || util.FieldExists("county", table) {
		return true
	}
	return false
}
func main() {
	log.Println("... Started event update.... ")
	lambda.Start(Handler)
}
