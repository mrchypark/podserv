package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/disgoorg/disgo/webhook"
	"github.com/disgoorg/snowflake/v2"
	"github.com/robfig/cron"
	"github.com/valyala/fasthttp"
)

var (
	msgID    = snowflake.GetEnv("message_webhook_id")
	msgToken = getEnvVar("message_webhook_token", "")
)

func main() {
	c := cron.New()
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
	s, err := UnmarshalChannelInfo(res)
	if err != nil {
		fmt.Printf("err!!: %s\n", err)
	}
	p := s.Summary.TotalCount
	if *p == int(0) {
		return
	}

	fmt.Println("pre reply count: ", *p)
	time.Sleep(time.Second * 30)
	res = doRequest("https://app-api6.podbbang.com/channels/1771386/comments?limit=10000&sort=desc&with=replies,votes,playlist,episode&next=0")
	s, err = UnmarshalChannelInfo(res)
	if err != nil {
		fmt.Printf("err!!: %s\n", err)
	}
	n := s.Summary.TotalCount
	if *n == int(0) {
		return
	}

	fmt.Println("now reply count: ", *n)
	if *p != *n {
		Send("댓글에 변경이 발생했습니다.\n<https://www.podbbang.com/creatorstudio/1771386/broadcast/comment_list>")
		fmt.Println("diff!")
	} else {
		fmt.Println("no diff")
	}
}

func Send(msg string) {
	dnclt := webhook.NewClient(msgID, msgToken)
	defer dnclt.Close(context.TODO())

	dnclt.CreateContent(msg)
}

func doRequest(url string) []byte {
	try := 0
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)

	req.SetRequestURI(url)
	resp.Header.SetStatusCode(502)

	for resp.Header.StatusCode() == 502 && try < 5 {
		time.Sleep(time.Second * 2)
		fasthttp.Do(req, resp)
		try += 1
	}
	return resp.Body()
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
	Next    string  `json:"next"`
	Summary Summary `json:"summary"`
}

type Summary struct {
	TotalCount int64 `json:"totalCount"`
}

func getEnvVar(key, fallbackValue string) string {
	if val, ok := os.LookupEnv(key); ok {
		return strings.TrimSpace(val)
	}
	return fallbackValue
}
