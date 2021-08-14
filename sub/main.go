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
		pd := Slack{
			Text:    "파싱에 문제가 발생했습니다.",
			Rawbody: string(res),
		}
		pd.errReport()
	}
	plc := s.Channels.Data[0].LikeCount
	psc := s.Channels.Data[0].SubscribeCount

	fmt.Println("pre like count: ", *plc)
	fmt.Println("pre subscribe count: ", *psc)
	time.Sleep(time.Second * 31)

	res = doRequest("https://app-api6.podbbang.com/search-content?keyword=%EB%8D%B0%EC%9D%B4%ED%84%B0%ED%99%80%EB%A6%AD&offset=0&limit=3")
	s, err = UnmarshalChannelInfo(res)
	if err != nil {
		pd := Slack{
			Text:    "파싱에 문제가 발생했습니다.",
			Rawbody: string(res),
		}
		pd.errReport()
	}
	lc := s.Channels.Data[0].LikeCount
	sc := s.Channels.Data[0].SubscribeCount

	fmt.Println("pre like count: ", *lc)
	fmt.Println("pre subscribe count: ", *sc)

	pd := Slack{
		Text:         "팟빵에 변경이 발생했습니다.",
		PreLike:      plc,
		Like:         lc,
		PreSubscribe: psc,
		Subscribe:    sc,
	}

	if *plc != *lc || *psc != *sc {
		pd.report()
		fmt.Println("count diff!")
	} else {
		fmt.Println("count no diff")
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

func (s *Slack) report() {
	nw := Report{Text: s.Text}
	nw.Attachment.
		AddField(slack.Field{Title: "이전구독", Value: fmt.Sprintf("%v", *s.PreSubscribe)}).
		AddField(slack.Field{Title: "구독", Value: fmt.Sprintf("%v", *s.Subscribe)}).
		AddField(slack.Field{Title: "이전좋아요", Value: fmt.Sprintf("%v", *s.PreLike)}).
		AddField(slack.Field{Title: "좋아요", Value: fmt.Sprintf("%v", *s.Like)})
	nw.Send()
}

func (s *Slack) errReport() {
	nw := Report{Text: s.Text}
	nw.Attachment.
		AddField(slack.Field{Title: "api 응답", Value: s.Rawbody})
	nw.Send()
}

type Slack struct {
	Text         string
	PreLike      *int
	Like         *int
	PreSubscribe *int
	Subscribe    *int
	Rawbody      string
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
