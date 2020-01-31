package util

import (
	"astoria-serverless-event-multi/db"
	"astoria-serverless-event-multi/model"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func Timestamp(t time.Time) int64 {
	return t.UTC().Unix()
}

// datetime format is "Oct. 28, 2019 6:15pm"
func FormatDatetime(datetime, layout string) (string, error) {

	log.Println("---FormatDateitme: ", datetime, "  Layout: ", layout)

	t, err := time.Parse(layout, datetime)
	if err != nil {
		log.Printf("Problem parsing datetime: %v\n", err)
		return datetime, err
	}
	year := t.Year()
	month := t.Month()
	day := t.Day()
	hour := t.Hour()
	minute := t.Minute()

	t = time.Now() // to get the UTC offset time
	now := t.Format(time.RFC3339)
	log.Println("NOW Format(time.RFC3339): ", now) // aws result:  NOW: 2019-10-29T02:35:13Z expecting format like 2019-10-28T20:38:13-04:00
	//parts := strings.Split(now, "-")

	hourMinute := fmt.Sprintf("%d-%02d-%02dT%02d:%02d:00-%s", year, month, day, hour, minute, "04:00")
	log.Println("hourMinute:", hourMinute)

	t2, err := time.Parse(time.RFC3339, hourMinute)
	if err != nil {
		log.Printf("Problem parsing hour-minute %v\n", err)
		log.Print(err)
	}
	log.Println("Verified - ", t2)

	return hourMinute, nil
}

// Utility function to response with JSON and setting header
func RespondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	// Even if you enable CORS in API Gateway, the integration response aka the API response from Lambda
	// needs to return the headers for it to work
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

	// Set response status code
	w.WriteHeader(code)

	// Encode and reply
	json.NewEncoder(w).Encode(payload)
}

func ParseSlackRequest(s string, req *model.SlackRequest) error {
	if len(s) == 0 || !strings.Contains(s, "&") {
		return fmt.Errorf("Error: request body is invalid")
	}
	str, err := url.QueryUnescape(s)
	if err != nil {
		log.Printf("Problem unescaping string: %v\n", err)
		return err
	}

	parts := strings.Split(str, "&")

	for _, c := range parts {

		//fmt.Println(c)
		part := strings.Split(c, "=")
		//fmt.Println("FIELD:", part[0], " CONTENT:", part[1])

		switch strings.Trim(part[0], " ") {
		case "token":
			req.Token = part[1]
		case "team_id":
			req.TeamID = part[1]
		case "team_domain":
			req.TeamDomain = part[1]
		case "channel_id":
			req.ChannelID = part[1]
		case "channel_name":
			req.ChannelName = part[1]
		case "user_id":
			req.UserID = part[1]
		case "user_name":
			req.UserName = part[1]
		case "command":
			req.Command = part[1]
		case "text":
			req.Text = part[1]
		case "response_url":
			req.ResponseURL = part[1]
		case "trigger_url":
			req.TriggerURL = part[1]
		}
	}

	return nil
}

// --title Event Title --street 3957 58th St. --city Woodside --county Queens --state New York  --country USA --postalcode 11377
func ParseToFields(e *db.Event, str string) (map[string]string, []error) {

	var errs []error

	fmt.Printf("String for parsing: \n %s\n", str)

	fieldsTable, err := parseNewEvent(e, str)
	if err != nil {
		errs = append(errs, err)
		return fieldsTable, errs
	}
	// validation should be moved
	// if e.Title == "" {
	// 	errs = append(errs, errors.New("Title is missing."))
	// }
	// if e.Street == "" {
	// 	errs = append(errs, errors.New("Street is missing."))
	// }
	// if e.City == "" {
	// 	errs = append(errs, errors.New("City is missing."))
	// }
	// if e.County == "" {
	// 	errs = append(errs, errors.New("County is missing."))
	// }
	// if e.Country == "" {
	// 	errs = append(errs, errors.New("Country is missing."))
	// }
	// Postal code is optional

	return fieldsTable, errs

}

func ParseToFieldsForUpdate(e *db.Event, str string) (map[string]string, error) {

	fmt.Printf("String for parsing for edit: \n %s\n", str)

	return parseNewEvent(e, str)
}

func parseNewEvent(event *db.Event, str string) (map[string]string, error) {
	var runtimeError error
	var table = make(map[string]string)
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
			//panic(err)
			runtimeError = fmt.Errorf("Parsing error on new event request.")

		}
	}()

	newStr := strings.ReplaceAll(str, "+", " ")
	for {
		if strings.Index(newStr[1:], "--") > -1 {
			i2 := strings.Index(newStr[1:], "--") + 2
			//fmt.Println(i2)
			//fmt.Println(newStr[:i2-1]) // + "+")  // --country USA +
			temp := strings.TrimSpace(newStr[:i2-1])
			//parts := strings.Split(temp)

			idx := strings.Index(temp, " ")
			field := strings.TrimPrefix(temp[:idx], "--")
			content := strings.Trim(temp[idx+1:], " ")

			updateEvent(event, field, content)
			table[field] = content

			newStr = newStr[i2-1:]
			//fmt.Println(newStr)
		} else { // to handle the last field
			idx := strings.Index(newStr, " ")
			field := strings.TrimPrefix(newStr[:idx], "--")
			content := strings.Trim(newStr[idx+1:], " ")

			updateEvent(event, field, content)
			table[field] = content
			break
		}
	}

	return table, runtimeError
}

func updateEvent(e *db.Event, field string, content string) {
	log.Printf("Field: %s   Content: %s \n", field, content)
	switch field {
	case "title":
		e.Title = content
	// case "new-title": // for changing current title
	// 	e.Title = content
	case "description":
		e.Description = content
	case "street":
		e.Street = content
	case "city":
		e.City = content
	case "state":
		e.State = content
	case "county":
		e.County = content
	case "country":
		e.Country = content
	case "postalcode":
		e.PostalCode = content
	case "start":
		e.StartDatetime = content
	case "end":
		e.EndDatetime = content
	case "contact":
		e.Contact = content
	case "email":
		e.Email = content
	case "status":
		e.EvtStatus = content
	case "category":
		e.EvtCategory = content
	case "subcategory":
		e.EvtSubCategory = content
	case "link":
		e.Link = content
	case "linklabel":
		e.LinkLabel = content
	}
}

// iterate and compare a given field name if exist in a given list of fields
func FieldExists(field string, table map[string]string) bool {
	if _, ok := table[field]; ok {
		return true
	}
	return false
}
