package main

import (
	"encoding/json"
	"github.com/pkg/errors"
	"net/http"
	"net/url"
)

const (
	SLACK_ATTACHMENT_START = "good"
	SLACK_ATTACHMENT_END   = "warning"
	SLACK_ATTACHMENT_ERR   = "danger"
)

type Slack struct {
	Channel     string       `json:"channel"`
	UserName    string       `json:"username"`
	IconUrl     string       `json:"icon_url"`
	Attachments []Attachment `json:"attachments"`
}

type Attachment struct {
	Fallback string            `json:"fallback"`
	Color    string            `json:"color"`
	Fields   []AttachmentField `json:"fields"`
}

type AttachmentField struct {
	Title string `json:"title"`
	Value string `json:"value"`
}

func SetAttachmentField(title, value string) AttachmentField {
	return AttachmentField{
		Title: title,
		Value: value,
	}
}

func SetAttachment(fallback, color string, fields []AttachmentField) Attachment {
	return Attachment{
		Fallback: fallback,
		Color:    color,
		Fields:   fields,
	}
}

func setSlack(channel string, attachments []Attachment) Slack {
	return Slack{
		Channel:     channel,
		UserName:    "Wildfly State Monitor",
		IconUrl:     "http://design.jboss.org/wildfly/logo/final/wildfly_icon_64px.png",
		Attachments: attachments,
	}
}

var (
	SlackRequestErr = errors.New("Request Error")
)

func SlackNotification(api_url string, msg Slack) error {
	params, _ := json.Marshal(msg)

	resp, err := http.PostForm(
		api_url,
		url.Values{
			"payload": {
				string(params),
			},
		},
	)
	if err != nil {
		return errors.Wrapf(SlackRequestErr, "Http Status is ", resp.Status)
	}

	return nil
}
