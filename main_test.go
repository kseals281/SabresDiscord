package main

import (
	"fmt"
	"github.com/dghubble/go-twitter/twitter"
	"testing"
	"time"
)

type SpyTrackedTime struct {
	time time.Time
}

func (t SpyTrackedTime) Now() time.Time {
	l := time.Location{}
	t.time = time.Date(2000, 1, 1, 0, 0, 0, 0, &l)
	return t.time
}

func Test_getTweet(t *testing.T) {
	type args struct {
		screenName string
		c          chan twitter.Tweet
		t          SpyTrackedTime
	}
	tests := []struct {
		name string
		args args
	}{
		{
			"receives tweet",
			args{
				screenName: "TwitterAPI",
				c:          make(chan twitter.Tweet, 1),
				t:          SpyTrackedTime{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			go getTweets(tt.args.screenName, tt.args.c, tt.args.t)
			tweet := <-tt.args.c
			if tweet.User.ScreenName != tt.args.screenName {
				t.Errorf("got tweet from user %s, wanted user %s",
					tt.args.screenName, tweet.User.ScreenName)
			}
			fmt.Println(tweet.Text)
			close(tt.args.c)
		})
	}
}
