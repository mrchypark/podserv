package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/ashwanthkumar/slack-go-webhook"
	"github.com/robfig/cron"
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
	res := doRequest("https://app-api6.podbbang.com/search-content?keyword=%EB%8D%B0%EC%9D%B4%ED%84%B0%ED%99%80%EB%A6%AD&offset=0&limit=3")
	s, err := UnmarshalChannelInfo(res)
	if err != nil {
		Slack("파싱에 문제가 발생했습니다.")
	}
	plc := s.Channels.Data[0].LikeCount
	psc := s.Channels.Data[0].SubscribeCount

	fmt.Println("pre like count: ", *plc)
	fmt.Println("pre subscribe count: ", *psc)
	time.Sleep(time.Second * 31)

	res = doRequest("https://app-api6.podbbang.com/search-content?keyword=%EB%8D%B0%EC%9D%B4%ED%84%B0%ED%99%80%EB%A6%AD&offset=0&limit=3")
	s, err = UnmarshalChannelInfo(res)
	if err != nil {
		Slack("파싱에 문제가 발생했습니다.")
	}
	lc := s.Channels.Data[0].LikeCount
	sc := s.Channels.Data[0].SubscribeCount

	fmt.Println("pre like count: ", *lc)
	fmt.Println("pre subscribe count: ", *sc)

	if *plc != *lc {
		SlackLike("좋아요 수에 변경이 발생했습니다.", lc)
		fmt.Println("like count diff!")
	} else {
		fmt.Println("like count no diff")
	}

	if *psc != *sc {
		SlackSub("구독 수에 변경이 발생했습니다.", sc)
		fmt.Println("like count diff!")
	} else {
		fmt.Println("like count no diff")
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

type Report struct {
	Text       string
	Attachment slack.Attachment
}

func (r *Report) Send() {
	webhookURL := os.Getenv("slack")
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

func SlackLike(text string, like *int) {
	nw := Report{Text: text}
	nw.Attachment.
		AddField(slack.Field{Title: "좋아요", Value: fmt.Sprintf("%v", *like)})
	nw.Send()
}

func SlackSub(text string, sub *int) {
	nw := Report{Text: text}
	nw.Attachment.
		AddField(slack.Field{Title: "구독", Value: fmt.Sprintf("%v", *sub)})
	nw.Send()
}

func UnmarshalChannelInfo(data []byte) (ChannelInfo, error) {
	var r ChannelInfo
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *ChannelInfo) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

type ChannelInfo struct {
	Channels *Channels `json:"channels,omitempty"`
}

type Channels struct {
	Data    []Datum  `json:"data,omitempty"`
	Summary *Summary `json:"summary,omitempty"`
}

type Datum struct {
	SubscribeCount *int `json:"subscribeCount,omitempty"`
	LikeCount      *int `json:"likeCount,omitempty"`
}

type Summary struct {
	TotalCount *int `json:"totalCount,omitempty"`
}

func getEnvVar(key, fallbackValue string) string {
	if val, ok := os.LookupEnv(key); ok {
		return strings.TrimSpace(val)
	}
	return fallbackValue
}
