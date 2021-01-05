// Copyright 2015 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build ignore

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/ashwanthkumar/slack-go-webhook"
	"github.com/fasthttp/websocket"
)

func main() {

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	key := os.Getenv("pb_key")
	str := os.Getenv("filter_str")

	u := url.URL{Scheme: "wss", Host: "stream.pushbullet.com", Path: "/websocket/" + key}
	fmt.Println("connecting !")

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				fmt.Println("read:", err)
				return
			}
			n, _ := UnmarshalNoti(message)
			if n.Type == "push" && strings.Contains(n.Push.Body, str) {
				fmt.Println("app: ", n.Push.ApplicationName)
				fmt.Println("title: ", n.Push.Title)
				fmt.Println("body: ", n.Push.Body)
				Slack("알람이 왔습니다.", n)
			}
		}
	}()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case t := <-ticker.C:
			err := c.WriteMessage(websocket.TextMessage, []byte(t.String()))
			if err != nil {
				fmt.Println("write:", err)
				return
			}
		case <-interrupt:
			fmt.Println("interrupt")

			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				fmt.Println("write close:", err)
				return
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}
}

func UnmarshalNoti(data []byte) (Noti, error) {
	var n Noti
	err := json.Unmarshal(data, &n)
	return n, err
}

type Noti struct {
	Type    string   `json:"type"`
	Targets []string `json:"targets"`
	Push    Push     `json:"push"`
}

type Push struct {
	Type             string `json:"type"`
	SourceDeviceIden string `json:"source_device_iden"`
	SourceUserIden   string `json:"source_user_iden"`
	ClientVersion    int64  `json:"client_version"`
	Dismissible      bool   `json:"dismissible"`
	Icon             string `json:"icon"`
	Title            string `json:"title"`
	Body             string `json:"body"`
	ApplicationName  string `json:"application_name"`
	PackageName      string `json:"package_name"`
	NotificationID   string `json:"notification_id"`
	NotificationTag  string `json:"notification_tag"`
}

type Report struct {
	Text       string
	Attachment slack.Attachment
}

func (r *Report) Send() {
	webhookURL := os.Getenv("slack_money")
	payload := slack.Payload{
		Text:        r.Text,
		Attachments: []slack.Attachment{r.Attachment},
	}

	err := slack.Send(webhookURL, "", payload)
	if len(err) > 0 {
		log.Printf("error: %s\n", err)
	}
}

func Slack(text string, n Noti) {
	nw := Report{Text: text}
	nw.Attachment.
		AddField(slack.Field{Title: "Title", Value: n.Push.Title}).
		AddField(slack.Field{Title: "App", Value: n.Push.ApplicationName}).
		AddField(slack.Field{Title: "Body", Value: n.Push.Body})
	nw.Send()
}
