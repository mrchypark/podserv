package main

import (
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/ashwanthkumar/slack-go-webhook"
	"github.com/valyala/fasthttp"
)

func main() {
	res := doRequest("http://www.podbbang.com/_m_api/podcasts/1771386/comments?with=replies,votes,summary&offset=0&next=0")
	s, _ := UnmarshalReply(res)
	p := s.Summary.TotalCount
	println("pre reply count: ", p)
	time.Sleep(time.Second * 150)
	res = doRequest("http://www.podbbang.com/_m_api/podcasts/1771386/comments?with=replies,votes,summary&offset=0&next=0")
	s, _ = UnmarshalReply(res)
	n := s.Summary.TotalCount
	println("now reply count: ", n)
	if p != n {
		Slack("댓글에 변경이 발생했습니다.")
		println("diff!")
	} else {
		println("no diff")
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
