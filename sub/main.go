package main

import (
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/ashwanthkumar/slack-go-webhook"
	"github.com/robfig/cron"
	"github.com/valyala/fasthttp"
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

	doc, err := goquery.NewDocument("http://www.podbbang.com/ch/1771386")
	if err != nil {
		log.Fatal(err)
	}
	pl := doc.Find("dl.likes dd").Text()
	ps := doc.Find("dl.subscribes dd").Text()
	println("pre like:", pl)
	println("pre sub:", ps)
	time.Sleep(time.Second * 60)
	doc, err = goquery.NewDocument("http://www.podbbang.com/ch/1771386")
	if err != nil {
		log.Fatal(err)
	}
	nl := doc.Find("dl.likes dd").Text()
	ns := doc.Find("dl.subscribes dd").Text()
	println("now like:", nl)
	println("now sub:", ns)
	if pl != nl {
		SlackLike("좋아요 수가 달라졌습니다.", nl)
		println("diff!")
	} else {
		println("no diff like")
	}
	if ps != ns {
		SlackLike("구독 수가 달라졌습니다.", ns)
		println("diff!")
	} else {
		println("no diff sub")
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

func SlackLike(text string, like string) {
	nw := Report{Text: text}
	nw.Attachment.
		AddField(slack.Field{Title: "좋아요", Value: like})
	nw.Send()
}

func SlackSub(text string, sub string) {
	nw := Report{Text: text}
	nw.Attachment.
		AddField(slack.Field{Title: "구독", Value: sub})
	nw.Send()
}
