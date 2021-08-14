package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/ashwanthkumar/slack-go-webhook"
	"github.com/davecgh/go-spew/spew"
	"github.com/robfig/cron/v3"
	"github.com/valyala/fasthttp"
)

var (
	webhookURL = getEnvVar("slack", "")
)

func main() {
	c := cron.New()
	fmt.Println("set webhook url is:")
	fmt.Println(webhookURL)
	c.AddJob("@every 30s", diff{})
	c.Start()
	for {
		time.Sleep(time.Second)
	}
}

type diff struct {
}

func (f diff) Run() {
	res := doRequest("https://app-api6.podbbang.com/channels/1771386/comments?limit=1&sort=desc&next=0&playlist_id=0")
	s, _ := UnmarshalComment(res)
	p := s.Summary.TotalCount
	if *p == int64(0) {
		return
	}

	spew.Printf("pre res: %#v\n", s)
	fmt.Println("pre reply count: ", p)
	time.Sleep(time.Second * 31)
	res = doRequest("https://app-api6.podbbang.com/channels/1771386/comments?limit=1&sort=desc&next=0&playlist_id=0")
	s, _ = UnmarshalComment(res)
	n := s.Summary.TotalCount
	if *n == int64(0) {
		return
	}
	spew.Printf("now res: %#v\n", s)
	fmt.Println("now reply count: ", n)
	if p != n {
		Slack("댓글에 변경이 발생했습니다.")
		fmt.Println("diff!")
	} else {
		fmt.Println("no diff")
	}
}

func doRequest(url string) []byte {
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)

	req.SetRequestURI(url)

	fasthttp.Do(req, resp)
	return resp.Body()
}

// This file was generated from JSON Schema using quicktype, do not modify it directly.
// To parse and unparse this JSON data, add this code to your project and do:
//
//    comment, err := UnmarshalComment(bytes)
//    bytes, err = comment.Marshal()

func UnmarshalComment(data []byte) (Comment, error) {
	var r Comment
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *Comment) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

type Comment struct {
	Data    []Datum  `json:"data,omitempty"`
	Next    *string  `json:"next,omitempty"`
	Summary *Summary `json:"summary,omitempty"`
}

type Datum struct {
	ID         *int64      `json:"id,omitempty"`
	Channel    *Channel    `json:"channel,omitempty"`
	User       *User       `json:"user,omitempty"`
	Support    interface{} `json:"support"`
	Parent     interface{} `json:"parent"`
	State      *string     `json:"state,omitempty"`
	Message    *string     `json:"message,omitempty"`
	Image      *string     `json:"image,omitempty"`
	BgColor    *string     `json:"bgColor,omitempty"`
	CreatedAt  *string     `json:"createdAt,omitempty"`
	ReplyCount *int64      `json:"replyCount,omitempty"`
	Episode    *Channel    `json:"episode,omitempty"`
	CanBlind   *bool       `json:"canBlind,omitempty"`
	CanDelete  *bool       `json:"canDelete,omitempty"`
	CanEdit    *bool       `json:"canEdit,omitempty"`
	CanReport  *bool       `json:"canReport,omitempty"`
}

type Channel struct {
	ID *int64 `json:"id,omitempty"`
}

type User struct {
	ID           *int64  `json:"id,omitempty"`
	Picture      *string `json:"picture,omitempty"`
	PictureColor *string `json:"pictureColor,omitempty"`
	Nickname     *string `json:"nickname,omitempty"`
	Role         *string `json:"role,omitempty"`
}

type Summary struct {
	TotalCount *int64 `json:"totalCount,omitempty"`
}

type Report struct {
	Text       string
	Attachment slack.Attachment
}

func (r *Report) Send() {
	payload := slack.Payload{
		Text:        r.Text,
		Attachments: []slack.Attachment{r.Attachment},
	}

	err := slack.Send(webhookURL, "", payload)
	if len(err) > 0 {
		log.Printf("error: %s\n", err)
	}
}

func Slack(text string) {
	nw := Report{Text: text}
	nw.Send()
}

func getEnvVar(key, fallbackValue string) string {
	if val, ok := os.LookupEnv(key); ok {
		return strings.TrimSpace(val)
	}
	return fallbackValue
}
