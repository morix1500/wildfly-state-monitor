package main

import (
	"encoding/json"
	"github.com/pkg/errors"
	"net/http"
	"net/url"
)

const (
	// SlackAttachementStart -- color of slack message "green"
	SlackAttachementStart = "good"
	// SlackAttachementEnd -- color of slack message "yellow"
	SlackAttachementEnd   = "warning"
	// SlackAttachementErr -- color of slack message "red"
	SlackAttachementErr   = "danger"
)

// ErrSlackRequest -- slack api request error
var ErrSlackRequest = errors.New("Request Error")


// Slack -- parts of slack message
type Slack struct {
	Channel     string       `json:"channel"`
	UserName    string       `json:"username"`
	IconURL     string       `json:"icon_url"`
	Attachments []Attachment `json:"attachments"`
}

// Attachment -- parts of slack message
type Attachment struct {
	Fallback string            `json:"fallback"`
	Color    string            `json:"color"`
	Fields   []AttachmentField `json:"fields"`
}

// AttachmentField -- parts of slack message
type AttachmentField struct {
	Title string `json:"title"`
	Value string `json:"value"`
}

// SetAttachmentField -- set parameter
func SetAttachmentField(title, value string) AttachmentField {
	return AttachmentField{
		Title: title,
		Value: value,
	}
}

// SetAttachment -- set parameter
func SetAttachment(fallback, color string, fields []AttachmentField) Attachment {
	return Attachment{
		Fallback: fallback,
		Color:    color,
		Fields:   fields,
	}
}

// SetSlack -- set parameter
func SetSlack(channel string, attachments []Attachment) Slack {
	return Slack{
		Channel:     channel,
		UserName:    "Wildfly State Monitor",
		IconURL:     "http://design.jboss.org/wildfly/logo/final/wildfly_icon_64px.png",
		Attachments: attachments,
	}
}

// SlackNotification -- send request to slack api
func SlackNotification(apiURL string, msg Slack) error {
	params, _ := json.Marshal(msg)

	resp, err := http.PostForm(
		apiURL,
		url.Values{
			"payload": {
				string(params),
			},
		},
	)
	if err != nil {
		return errors.Wrapf(ErrSlackRequest, "Http Status is ", resp.Status)
	}

	return nil
}
