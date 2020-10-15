// Copyright 2015 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build ignore

package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/fasthttp/websocket"
)

func main() {
	flag.Parse()
	log.SetFlags(0)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	key := os.Getenv("pb_key")

	u := url.URL{Scheme: "wss", Host: "stream.pushbullet.com", Path: "/websocket/" + key}
	log.Printf("connecting !")

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
				log.Println("read:", err)
				return
			}
			n, _ := UnmarshalNoti(message)
			if n.Type == "push" {
				log.Printf("recv: %s", n.Push.PackageName)
				log.Printf("recv: %s", n.Push.ApplicationName)
				log.Printf("recv: %s", n.Push.Title)
				log.Printf("recv: %s", n.Push.Body)
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
				log.Println("write:", err)
				return
			}
		case <-interrupt:
			log.Println("interrupt")

			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close:", err)
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
