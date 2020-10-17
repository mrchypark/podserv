package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
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
		return
	}
	pl := doc.Find("dl.likes dd").Text()
	ps := doc.Find("dl.subscribes dd").Text()
	pls := strings.Split(pl, "{{")
	pss := strings.Split(ps, "{{")
	fmt.Println("pre like:", pls[0])
	fmt.Println("pre sub:", pss[0])
	time.Sleep(time.Second * 35)
	doc, err = goquery.NewDocument("http://www.podbbang.com/ch/1771386")
	if err != nil {
		log.Fatal(err)
		return
	}
	nl := doc.Find("dl.likes dd").Text()
	ns := doc.Find("dl.subscribes dd").Text()
	nls := strings.Split(nl, "{{")
	nss := strings.Split(ns, "{{")
	fmt.Println("pre like:", nls[0])
	fmt.Println("pre sub:", nss[0])
	if pls[0] != nls[0] {
		SlackLike("좋아요 수가 달라졌습니다.", pls[0]+"->"+nls[0])
		fmt.Println("diff!")
	} else {
		fmt.Println("no diff like")
	}
	if pss[0] != nss[0] {
		SlackLike("구독 수가 달라졌습니다.", pss[0]+"->"+nss[0])
		fmt.Println("diff!")
	} else {
		fmt.Println("no diff sub")
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
