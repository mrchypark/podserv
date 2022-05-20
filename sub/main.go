package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/disgoorg/disgo/discord"
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
	res := doRequest("https://app-api6.podbbang.com/search-content?keyword=%EB%8D%B0%EC%9D%B4%ED%84%B0%ED%99%80%EB%A6%AD&offset=0&limit=3")
	s, err := UnmarshalChannelInfo(res)
	if err != nil {
		fmt.Printf("err!!: %s\n", err)
		return
	}
	plc := s.Channels.Data[0].LikeCount
	psc := s.Channels.Data[0].SubscribeCount

	fmt.Println("pre like count: ", *plc)
	fmt.Println("pre subscribe count: ", *psc)
	time.Sleep(time.Second * 30)

	res = doRequest("https://app-api6.podbbang.com/search-content?keyword=%EB%8D%B0%EC%9D%B4%ED%84%B0%ED%99%80%EB%A6%AD&offset=0&limit=3")
	s, err = UnmarshalChannelInfo(res)
	if err != nil {
		fmt.Printf("err!!: %s\n", err)
		return
	}
	lc := s.Channels.Data[0].LikeCount
	sc := s.Channels.Data[0].SubscribeCount

	fmt.Println("pre like count: ", *lc)
	fmt.Println("pre subscribe count: ", *sc)

	pd := rpClient{
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
	try := 0
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)

	req.Header.Add("User-Agent", "podserv")
	req.SetRequestURI(url)
	resp.Header.SetStatusCode(502)

	for resp.Header.StatusCode() == 502 && try < 5 {
		time.Sleep(time.Second * 2)
		fasthttp.Do(req, resp)
		try += 1
	}
	return resp.Body()
}

func (s *rpClient) report() {
	msgclt := webhook.NewClient(msgID, msgToken)
	defer msgclt.Close(context.TODO())

	subr := *s.Subscribe - *s.PreSubscribe
	liker := *s.Like - *s.PreLike
	sub := fmt.Sprintf("%v", *s.Subscribe) + "(" + fmt.Sprintf("%v", subr) + ")"
	like := fmt.Sprintf("%v", *s.Like) + "(" + fmt.Sprintf("%v", liker) + ")"

	msgclt.CreateContent(s.Text)
	var ems = []discord.Embed{
		discord.NewEmbedBuilder().
			AddField("구독", sub, true).
			AddField("좋아요", like, true).
			Build(),
	}
	msgclt.CreateEmbeds(ems)
}

type rpClient struct {
	Text         string
	PreLike      *int
	Like         *int
	PreSubscribe *int
	Subscribe    *int
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
