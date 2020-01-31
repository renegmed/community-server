package db

import (
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/guregu/dynamo"
	uuid "github.com/satori/go.uuid"
)

// holds the dynamo client
type EventService struct {
	Table dynamo.Table
}

func NewEventService() (*EventService, error) {
	dynamoTable, err := NewDynamoTable(*aws.String(os.Getenv("DYNAMODB_EVENT_TABLE")), "")
	if err != nil {
		log.Printf("Not able to create dynamo table: %v\n", err)
		return nil, err
	}
	return &EventService{
		Table: dynamoTable,
	}, nil
}

func NewDynamoTable(tableName, endpoint string) (dynamo.Table, error) {
	if tableName == "" {
		log.Println("Table name is empty")
		return dynamo.Table{}, fmt.Errorf("you must supply a table name")
	}
	cfg := aws.Config{}
	cfg.Region = aws.String(os.Getenv("AWS_REGION"))
	if endpoint != "" {
		cfg.Endpoint = aws.String(endpoint)
	}

	sess := session.Must(session.NewSession())
	db := dynamo.New(sess, &cfg)
	table := db.Table(tableName)
	return table, nil
}

func (e *EventService) CreateNewEvent(event Event) (string, error) {

	id := uuid.NewV4()
	//if err != nil {
	//	return thisItem, err
	//}
	event.ID = id.String()
	// log.Println("--- event.ID: ", event.ID)
	// log.Println("--- event.ChannelId: ", event.ChannelId)
	// log.Println("--- event.Channel: ", event.Channel)
	// log.Println("--- event.Title: ", event.Title)
	// log.Println("--- event.PostalCode: ", event.PostalCode)

	err := e.Table.Put(event).Run()
	return event.ID, err
}

func (e *EventService) GetAllEvents() ([]Event, error) {
	events := []Event{}
	err := e.Table.Scan().All(&events)
	return events, err
}

func (e *EventService) GetActiveEvents() ([]Event, error) {
	events := []Event{}
	err := e.Table.Scan().Filter("EvtStatus <> ?", "inactive").All(&events)
	return events, err
}

func (e *EventService) GetEventByTitle(title string) (Event, error) {
	evnts := []Event{}
	err := e.Table.Scan().Filter("Title = ?", title).All(&evnts)
	if err != nil {
		log.Printf("Problem getting events with title '%s'. Error: %v\n", title, err)
		return Event{}, err
	}
	if len(evnts) == 0 {
		return Event{}, nil
	}
	return evnts[0], nil
}

// func (e *EventService) GetEventsByFields(rawfields map[string]string) ([]Event, error) {

// 	// log.Printf("+++++ GetEventsByFields rawfields: %v\n", rawfields)

// 	fields := convertToDBFields(rawfields)
// 	// log.Printf("+++++  GetEventsByFieldsquery converted fields: %v\n", fields)

// 	query := strings.Builder{}
// 	for k, v := range fields {
// 		query.WriteString(fmt.Sprintf(" %s = '%s' AND", k, v))
// 	}
// 	s := query.String()

// 	// log.Printf("+++++ GetEventsByFieldsquery query string: %s\n", s)

// 	s = s[:len(s)-4]

// 	log.Printf("+++++ GetEventsByFieldsquery truncated query string: %s\n", s)

// 	events := []Event{}
// 	err := e.Table.Scan().Filter(s).All(&events)
// 	if err == nil {
// 		log.Printf("+++++ GetEventsByFieldsquery events: %v\n", events)
// 	}
// 	return events, err
// }

func (e *EventService) GetEventsByCategory(category string) ([]Event, error) {

	// // log.Printf("+++++ GetEventsByFields rawfields: %v\n", rawfields)

	// fields := convertToDBFields(rawfields)
	// // log.Printf("+++++  GetEventsByFieldsquery converted fields: %v\n", fields)

	// query := strings.Builder{}
	// for k, v := range fields {
	// 	query.WriteString(fmt.Sprintf(" %s = '%s' AND", k, v))
	// }
	// s := query.String()

	// // log.Printf("+++++ GetEventsByFieldsquery query string: %s\n", s)

	// s = s[:len(s)-4]

	// log.Printf("+++++ GetEventsByFieldsquery truncated query string: %s\n", s)

	// events := []Event{}
	// err := e.Table.Scan().Filter(s).All(&events)
	// if err == nil {
	// 	log.Printf("+++++ GetEventsByFieldsquery events: %v\n", events)
	// }
	// return events, err
	events := []Event{}
	err := e.Table.Scan().Filter("EvtCategory = ?", category).All(&events)
	return events, err

}

func (e *EventService) UdpateEvent(newEvent Event) (Event, error) {
	var oldEvent Event
	var cc dynamo.ConsumedCapacity
	err := e.Table.Put(newEvent).ConsumedCapacity(&cc).OldValue(&oldEvent)
	if err != nil {
		log.Println("Problem with update event: ", err.Error())
		return Event{}, err
	}
	if cc.Total != 1 || cc.Table != 1 { // || cc.TableName != testTable {
		return oldEvent, fmt.Errorf("Bad consumed capacity: %v", cc)
	}

	return oldEvent, nil
}

func (e *EventService) DeleteEvent(id string) (Event, error) {
	var oldEvent Event

	log.Printf("ID to delete: %s", id)

	//err := e.Table.Delete("ID", "*").If("Title = ?", title).OldValue(&oldEvent)
	err := e.Table.Delete("ID", id).OldValue(&oldEvent)

	if err != nil {
		log.Printf("Problem while deleting record with ID '%s'.\n %v.\n", id, err.Error())
		return oldEvent, err
	}
	return oldEvent, nil
}

func convertToDBFields(flds map[string]string) map[string]string {

	// log.Printf("+++++ convertToDBFields flds: %v\n", flds)

	var fields = make(map[string]string)
	for k, v := range flds {
		switch k {
		case "title":
			fields["Title"] = v
		case "city":
			fields["City"] = v
		case "county":
			fields["County"] = v
		case "country":
			fields["Country"] = v
		case "postalcode":
			fields["PostalCode"] = v
		case "status":
			fields["EvtStatus"] = v
		case "category":
			// log.Printf("+++++ convertToDBFields category: %s\n", v)
			fields["EvtCategory"] = v
		case "subcategory":
			fields["EvtSubCategory"] = v
		}
	}
	return fields
}
