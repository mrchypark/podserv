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
	res := doRequest("http://www.podbbang.com/_m_api/podcasts/1771386/comments?with=summary&offset=0&next=0")
	s, _ := UnmarshalReply(res)
	p := s.Summary.TotalCount

	if p == 0 {
		return
	}

	spew.Printf("pre res: %#v\n", s)
	fmt.Println("pre reply count: ", p)
	time.Sleep(time.Second * 31)
	res = doRequest("http://www.podbbang.com/_m_api/podcasts/1771386/comments?with=summary&offset=0&next=0")
	s, _ = UnmarshalReply(res)
	n := s.Summary.TotalCount
	if n == 0 {
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

func UnmarshalReply(data []byte) (Reply, error) {
	var r Reply
	err := json.Unmarshal(data, &r)
	return r, err
}

func doRequest(url string) []byte {
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)   // <- do not forget to release
	defer fasthttp.ReleaseResponse(resp) // <- do not forget to release

	req.SetRequestURI(url)

	fasthttp.Do(req, resp)
	return resp.Body()
}

type Reply struct {
	Summary Summary `json:"summary"`
}

type Summary struct {
	TotalCount int64 `json:"total_count"`
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
