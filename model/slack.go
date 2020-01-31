package model

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

// Constants for where the response should go
// By default it is set to ResponseInChannel
const (
	ResponseInChannel = "in_channel"
	ResponseEphemeral = "ephemeral"
)

// SlackRequest represents an incoming slash command request
type SlackRequest struct {
	Token       string
	TeamID      string
	TeamDomain  string
	ChannelID   string
	ChannelName string
	UserID      string
	UserName    string
	Command     string
	Text        string
	ResponseURL string
	TriggerURL  string
	Debug       bool
}

// SlackResponse represents a response to slash command
type SlackResponse struct {
	ResponseType string       `json:"response_type"`
	Text         string       `json:"text"`
	Attachments  []Attachment `json:"attachments"`
}

// Attachment is Slack attachment for slash Response
type Attachment struct {
	Fallback      string   `json:"fallback"`
	Text          string   `json:"text"`
	MarkdownIn    []string `json:"mrkdwn_in,omitempty"`
	Color         string   `json:"color,omitempty"`
	AuthorName    string   `json:"author_name,omitempty"`
	AuthorSubname string   `json:"author_subname,omitempty"`
	AuthorLink    string   `json:"author_link,omitempty"`
	AuthorIcon    string   `json:"author_icon,omitempty"`
	Title         string   `json:"title,omitempty"`
	TitleLink     string   `json:"title_link,omitempty"`
	Pretext       string   `json:"pretext,omitempty"`
	ImageURL      string   `json:"image_url,omitempty"`
	ThumbURL      string   `json:"thumb_url,omitempty"`
	Fields        []Field  `json:"fields,omitempty"`
	Footer        string   `json:"footer,omitempty"`
	FooterIcon    string   `json:"footer_icon,omitempty"`
	Timestamp     int64    `json:"ts,omitempty"`
}

// Field is a field attachment
type Field struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}

func createTimestamp(t time.Time) int64 {
	return t.UTC().Unix()
}

func (req *SlackRequest) Update(r *http.Request) {
	req.Token = r.FormValue("token")              // 4lnUutTDuMrXfnYnNOcdhM1E
	req.TeamID = r.FormValue("team_id")           // TP0N9QXC1
	req.TeamDomain = r.FormValue("team_domain")   // golang-slack-dev
	req.ChannelID = r.FormValue("channel_id")     // CP0N9S5CZ
	req.ChannelName = r.FormValue("channel_name") // business-processes-automation
	req.UserID = r.FormValue("user_id")           // UNVLVE8E7
	req.UserName = r.FormValue("user_name")       // renegmed
	req.Command = r.FormValue("command")          // /new/event
	req.Text = r.FormValue("text")                // 39 58th Street. Apt 2, Woodside NY 11377
	req.ResponseURL = r.FormValue("response_url")
	if r.FormValue("debug") == "true" {
		req.Debug = true
	} else {
		req.Debug = false
	}

}

func (req *SlackRequest) EventResponseWithFields(text string,
	errs []error, debug bool, table map[string]string) SlackResponse {

	fields := responseFieldsToString(table)
	errorsText := errorsToString(errs)
	return formatResponse(text, fields, errorsText, debug)
}

func responseFieldsToString(table map[string]string) string {
	var sb strings.Builder
	for k, v := range table {
		sb.WriteString(k + ": " + v + "\n")
	}
	return sb.String()
}
func errorsToString(errs []error) string {
	var sb strings.Builder
	for _, err := range errs {
		sb.WriteString(fmt.Sprintf("%v\n", err))
	}
	return sb.String()
}
func (req *SlackRequest) EventResponse(text string, errs []error, debug bool) SlackResponse {
	errorsText := errorsToString(errs)
	return formatResponse(text, "", errorsText, debug)
}

func formatResponse(text, fields, errors string, debug bool) SlackResponse {
	slackResponse := SlackResponse{}
	if errors != "" {
		slackResponse.ResponseType = ResponseEphemeral
	} else {
		slackResponse.ResponseType = ResponseInChannel
	}

	if debug {
		slackResponse.Text = fmt.Sprintf("%s\n%s\n%s\n", text, fields, errors)
	} else {
		slackResponse.Text = fmt.Sprintf("%s\n%s\n", text, fields)
	}

	var attachments []Attachment
	attachment := Attachment{}
	attachment.Color = "RED"
	attachment.Timestamp = createTimestamp(time.Now())
	slackResponse.Attachments = append(attachments, attachment)

	return slackResponse
}
