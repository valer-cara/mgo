package slack

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/valer-cara/mgo/pkg/notification"
)

const (
	defaultUsername = "Juno"
	defaultIcon     = ":rocket:"
)

// Webhook implements the Notification interface and contains other Slack related properties
type Webhook struct {
	Webhook   string
	Channel   string
	Username  string
	IconEmoji string
}

// NewWebhook returns a Webhook that meets the Notification interface
func NewWebhook(webhook string, channel string, username string, icon string) notification.Notification {
	n := Webhook{
		Webhook:   webhook,
		Channel:   channel,
		Username:  getValueOrDefault(username, defaultUsername),
		IconEmoji: getValueOrDefault(icon, defaultIcon),
	}

	return &n
}

// Deployed sends a notification that a deployed was attempted
func (w *Webhook) Deployed(repo string, imageRepo string, tag string, cluster string, author string, err error) error {
	return w.sendMessage(
		w.generateDeployedMessage(
			repo,
			imageRepo,
			tag,
			cluster,
			author,
			err,
		),
	)
}

type message struct {
	Attachments []attachment `json:"attachments,omitempty"`
	Text        string       `json:"text,omitempty"`
	Username    string       `json:"username,omitempty"`
	IconEmoji   string       `json:"icon_emoji,omitempty"`
	Mrkdwn      bool         `json:"mrkdwn,omitempty"`
	Channel     string       `json:"channel,omitempty"`
}

type attachment struct {
	Fallback   string  `json:"fallback,omitempty"`
	Color      string  `json:"color,omitempty"`
	Pretext    string  `json:"pretext,omitempty"`
	AuthorName string  `json:"author_name,omitempty"`
	AuthorLink string  `json:"author_link,omitempty"`
	AuthorIcon string  `json:"author_icon,omitempty"`
	Title      string  `json:"title,omitempty"`
	TitleLink  string  `json:"title_link,omitempty"`
	Text       string  `json:"text,omitempty"`
	Fields     []field `json:"fields,omitempty"`
	ImageURL   string  `json:"image_url,omitempty"`
	ThumbURL   string  `json:"thumb_url,omitempty"`
	Footer     string  `json:"footer,omitempty"`
	FooterIcon string  `json:"footer_icon,omitempty"`
	Ts         int64   `json:"ts,omitempty"`
}

type field struct {
	Title string `json:"title,omitempty"`
	Value string `json:"value,omitempty"`
	Short bool   `json:"short,omitempty,omitempty"`
}

func (w *Webhook) generateDeployedMessage(repo string, imageRepo string, tag string, cluster string, author string, err error) message {
	m := message{
		Username:  w.Username,
		IconEmoji: w.IconEmoji,
		Channel:   w.Channel,
	}

	if err == nil {
		m.Attachments = []attachment{
			{
				Color:   "#36a64f",
				Pretext: fmt.Sprintf("Application %s was deployed successfully", repo),
				Fields: []field{
					{
						Title: "Repository",
						Value: repo,
					},
					{
						Title: "Image",
						Value: fmt.Sprintf("%s:%s", imageRepo, tag),
					},
					{
						Title: "Author",
						Value: author,
					},
					{
						Title: "Cluster",
						Value: cluster,
					},
				},
				Ts: time.Now().Unix(),
			},
		}
	} else {
		m.Attachments = []attachment{
			{
				Color:   "#ff0000",
				Pretext: fmt.Sprintf("Application %s was not deployed successfully", repo),
				Fields: []field{
					{
						Title: "Repository",
						Value: repo,
					},
					{
						Title: "Image",
						Value: fmt.Sprintf("%s:%s", imageRepo, tag),
					},
					{
						Title: "Author",
						Value: author,
					},
					{
						Title: "Cluster",
						Value: cluster,
					},
					{
						Title: "Error",
						Value: err.Error(),
					},
				},
				Ts: time.Now().Unix(),
			},
		}
	}

	return m
}

func getValueOrDefault(value string, defaultValue string) string {
	if len(value) > 0 {
		return value
	}
	return defaultValue
}

func (w *Webhook) sendMessage(msg message) error {
	if len(w.Webhook) == 0 {
		return fmt.Errorf("There was no webhook url specified")
	}

	msgBody, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	_, err = http.Post(w.Webhook, "application/json", strings.NewReader(string(msgBody)))
	if err != nil {
		return err
	}

	return nil
}
