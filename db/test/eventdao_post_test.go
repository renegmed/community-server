package test

import (
	"astoria-serverless-event-multi/db"
	util "astoria-serverless-event-multi/test_utils"
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEventService_CreateNewEvent(t *testing.T) {
	str, err := ioutil.ReadFile("../../testdata/post1.json")
	if err != nil {
		t.Fatal(err)
	}

	content := string(str)

	//t.Log("---POINT 1 content: ", content)

	if len(content) == 0 {
		t.Fatal("File is empty.")
	}

	service, err := util.NewEventService()
	if err != nil {
		t.Logf(" ---- Error: %v", err)
		t.Fatal(err)
	}

	var event db.Event

	err = json.Unmarshal([]byte(content), &event)
	if err != nil {
		t.Fatal(err)
	}
	//t.Logf("--- POINT 4 event:\n%v\n", event)

	id, err := service.CreateNewEvent(event)
	if err != nil {
		t.Fatal(err)
	}

	assert.NotEqual(t, id, "")

	events, err := service.GetAllEvents()
	if err != nil {
		t.Logf(" ---- Error: %v", err)
		t.Fatal(err)
	}
	//t.Logf("--- POINT 2 \n%v\n", events)

	assert.Equal(t, 1, len(events))
	assert.NotEqual(t, 0, len(events))
	assert.NotEqual(t, 2, len(events))

	const (
		title  = "Food Fair 2019"
		street = "123 Main St."
	)

	evt, err := service.GetEventByTitle(title)
	if err != nil {
		t.Logf(" ----GetEventByTitle Error: %v", err)
		t.Fatal(err)
	}
	assert.Equal(t, evt.Title, title)
	assert.Equal(t, evt.Street, street)
}

func TestEventService_GetActiveEvents(t *testing.T) {

	events := []db.Event{
		{
			ID:         "CH0001",
			Channel:    "Experimental",
			Title:      "Food Fair 2019",
			Street:     "123 Main St.",
			City:       "Flushing",
			County:     "Queens",
			State:      "NY",
			PostalCode: "11012",
			Country:    "US",
			Lat:        40.757121,
			Lon:        -73.970382,
			Contact:    "Joe Doe",
			EvtStatus:  "inactive",
		},
		{
			ID:         "CH0002",
			Channel:    "Experimental2",
			Title:      "In Concert 2019",
			Street:     "John Dewey Park 556 Reddy St.",
			City:       "Flushing",
			County:     "Queens",
			State:      "NY",
			PostalCode: "11012",
			Country:    "US",
			Lat:        40.767121,
			Lon:        -73.980382,
			Contact:    "Jane Jones",
			EvtStatus:  "active",
		},
	}

	service, err := util.NewEventService()
	if err != nil {
		t.Logf(" ---- NewEventService Error: %v", err)
		t.Fatal(err)
	}

	for _, e := range events {
		t.Run("save event", func(t *testing.T) {
			_, err := service.CreateNewEvent(e)
			if err != nil {
				t.Fatal(err)
			}
		})
	}

	evts, err := service.GetActiveEvents()
	if err != nil {
		t.Logf(" ---- Error: %v", err)
		t.Fatal(err)
	}
	//t.Logf("--- POINT 2 \n%v\n", events)

	assert.Equal(t, 1, len(evts))
	assert.NotEqual(t, 0, len(evts))
	assert.NotEqual(t, 2, len(evts))

	const (
		title  = "In Concert 2019"
		street = "John Dewey Park 556 Reddy St."
	)

	assert.Equal(t, evts[0].Title, title)
	assert.Equal(t, evts[0].Street, street)

}
