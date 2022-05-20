package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/webhook"
	"github.com/disgoorg/snowflake/v2"
	"github.com/fasthttp/websocket"
)

var (
	dnID    = snowflake.GetEnv("donation_webhook_id")
	dnToken = getEnvVar("donation_webhook_token", "")
)

func main() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	dnclt := webhook.NewClient(dnID, dnToken)
	defer dnclt.Close(context.TODO())

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
				fmt.Println("title: ", n.Push.Title)
				fmt.Println("app: ", n.Push.ApplicationName)
				fmt.Println("body: ", n.Push.Body)
				var ems = []discord.Embed{
					discord.NewEmbedBuilder().
						SetTitle(n.Push.Title).
						SetField(0, n.Push.ApplicationName, n.Push.Body, true).
						Build(),
				}
				dnclt.CreateEmbeds(ems)
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

func getEnvVar(key, fallbackValue string) string {
	if val, ok := os.LookupEnv(key); ok {
		return strings.TrimSpace(val)
	}
	return fallbackValue
}
