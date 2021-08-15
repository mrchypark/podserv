package main

import (
	"testing"

	cp "github.com/r3labs/diff/v2"
	"github.com/rs/zerolog/log"

	"github.com/stretchr/testify/assert"
)

// func Test_difftest(t *testing.T) {
// 	res := doRequest("https://app-api6.podbbang.com/channels/1771386/comments?limit=4&sort=desc&with=replies,votes,playlist,episode&next=0")
// 	s, err := UnmarshalComment(res)
// 	if err != nil {
// 		log.Error().Err(err).Send()
// 	}

// 	changelog, err := cp.Diff(s.Data, s.Data)
// 	if err != nil {
// 		log.Error().Err(err).Send()
// 	}

// 	assert.Equal(t, diff.Changelog(diff.Changelog{}), changelog, "they should be equal")
// }

func Test_diffMessage(t *testing.T) {
	res := `{"data":[{"id":10401556,"channel":{"id":1771386},"user":{"id":7136148,"picture":"https:\/\/dimg.podbbang.com\/poduser\/7136148.jpg","pictureColor":"#9acee6","nickname":"\ub370\uc774\ud130\ud640\ub9ad\ubc15\ubc15\uc0ac","role":""},"support":null,"parent":null,"state":"O","message":"\ud6c4\uc6d0 \uac10\uc0ac\ub4dc\ub9bd\ub2c8\ub2e4!!","image":"","bgColor":"","createdAt":"2021-08-14 12:57:25","replyCount":0,"episode":{"id":24119138,"title":"Ep(120) \uc5b8\ub860\uc740 \ub370\uc774\ud130\ub97c \uc5b4\ub5bb\uac8c \ub2e4\ub8e8\ub294\uac00?"},"canBlind":false,"canDelete":false,"canEdit":false,"canReport":true,"replies":[],"upvoteCount":0,"downvoteCount":0}],"next":"10401556","summary":{"totalCount":810}}`
	s, err := UnmarshalComment([]byte(res))
	if err != nil {
		log.Error().Err(err).Send()
	}

	changelog, err := cp.Diff(s.Data, s.Data)
	if err != nil {
		log.Error().Err(err).Send()
	}

	log.Debug().Interface("diff", changelog).Send()

	assert.Equal(t, "", changelog, "they should be equal")
}
