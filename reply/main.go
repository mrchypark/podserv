package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/ashwanthkumar/slack-go-webhook"
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
	res := doRequest("https://app-api6.podbbang.com/channels/1771386/comments?limit=10000&sort=desc&with=replies,votes,playlist,episode&next=0")
	s, err := UnmarshalComment(res)
	if err != nil {
		Slack("파싱에 문제가 발생했습니다.")
	}
	p := s.Summary.TotalCount
	if *p == int(0) {
		return
	}

	fmt.Println("pre reply count: ", *p)
	time.Sleep(time.Second * 30)
	res = doRequest("https://app-api6.podbbang.com/channels/1771386/comments?limit=10000&sort=desc&with=replies,votes,playlist,episode&next=0")
	s, err = UnmarshalComment(res)
	if err != nil {
		Slack("파싱에 문제가 발생했습니다.")
	}
	n := s.Summary.TotalCount
	if *n == int(0) {
		return
	}

	fmt.Println("now reply count: ", *n)
	if *p != *n {
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
	Summary *Summary `json:"summary,omitempty"`
}

type Datum struct {
	ID            *int     `json:"id"`
	User          *User    `json:"user"`
	Support       *Support `json:"support,omitempty"`
	Parent        *Parent  `json:"parent,omitempty"`
	Message       *string  `json:"message,omitempty"`
	CreatedAt     *string  `json:"createdAt,omitempty"`
	ReplyCount    *int     `json:"replyCount,omitempty"`
	Episode       *Channel `json:"episode,omitempty"`
	Replies       []Datum  `json:"replies,omitempty"`
	UpvoteCount   *int     `json:"upvoteCount,omitempty"`
	DownvoteCount *int     `json:"downvoteCount,omitempty"`
}

type Channel struct {
	ID *int64 `json:"id,omitempty"`
}

type Parent struct {
	ID *int `json:"id,omitempty"`
}

type Support struct {
	Type *string `json:"type,omitempty"`
	Cash *int    `json:"cash,omitempty"`
}

type User struct {
	ID       *int    `json:"id,omitempty"`
	Nickname *string `json:"nickname,omitempty"`
	Role     *string `json:"role,omitempty"`
}

type Summary struct {
	TotalCount *int `json:"totalCount,omitempty"`
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
