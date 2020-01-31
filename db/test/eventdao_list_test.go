package test

import (
	util "astoria-serverless-event-multi/test_utils"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEventService_ListEvents(t *testing.T) {
	service, err := util.NewEventService()
	if err != nil {
		t.Logf(" ---- Error: %v", err)
		t.Fatal(err)
	}
	t.Log("--- POINT 1")

	events, err := service.GetAllEvents()
	if err != nil {
		t.Logf(" ---- Error: %v", err)
		t.Fatal(err)
	}
	t.Logf("--- POINT 2 \n%v\n", events)

	assert.Equal(t, 0, len(events))

}
