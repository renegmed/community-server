package db

const (
	LayoutUS = "Jan. 2, 2006 3:04pm"
)

type Event struct {
	ID             string  `json:"id" dynamo:"ID,hash"`
	TeamID         string  `json:"teamid" dynamo:"TeamID"`
	TeamDomain     string  `json:"teamdomain" dynamo:"TeamDomain"`
	ChannelId      string  `json:"channelid" dynamo:"ChannelID"` // channel id
	Channel        string  `json:"channel" dynamo:"Channel"`
	Title          string  `json:"title" dynamo:"Title"`
	Description    string  `json:"description" dynamo:"Description"`
	Street         string  `json:"street" dynamo:"Street"`
	City           string  `json:"city" dynamo:"City"`
	County         string  `json:"county" dynamo:"County"`
	State          string  `json:"state" dynamo:"State"`
	PostalCode     string  `json:"postalcode" dynamo:"PostalCode"`
	Country        string  `json:"country" dynamo:"Country"`
	Lat            float64 `json:"lat" dynamo:"Lat"`
	Lon            float64 `json:"lon" dynamo:"Lon"`
	Contact        string  `json:"contact" dynamo:"Contact"`
	Email          string  `json:"email" dynamo:"Email"`
	StartDatetime  string  `json:"startdatetime" dynamo:"StartDatetime"`
	EndDatetime    string  `json:"enddatetime" dynamo:"EndDatetime"`
	EvtStatus      string  `json:"evtstatus" dynamo:"EvtStatus"`
	EvtCategory    string  `json:"evtcategory" dynamo:"EvtCategory"`
	EvtSubCategory string  `json:"evtsubcategory" dynamo:"EvtSubCategory"`
	Link           string  `json:"link" dynamo:"Link"`
	LinkLabel      string  `json:"linklabel" dynamo:"LinkLabel"`
}

type EventDb interface {
	CreateNewEvent(e Event) (string, error)
	GetAllEvents() ([]Event, error)
	GetActiveEvents() ([]Event, error)
	GetEventByTitle(title string) (Event, error)
	GetEventsByCategory(category string) ([]Event, error)
	UdpateEvent(newEvent Event) (Event, error)
	DeleteEvent(event Event) (Event, error)
}
