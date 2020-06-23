package main

import (
	"github.com/bwmarrin/discordgo"
	"github.com/dghubble/go-twitter/twitter"
	"os"
	"strings"
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

func (t SpyTrackedTime) Epoch() time.Time {
	return time.Time{}
}

func Test_getTweet(t *testing.T) {
	type args struct {
		screenNames []string
		c           chan twitter.Tweet
		t           SpyTrackedTime
	}
	tests := []struct {
		name string
		args args
	}{
		{
			"a new tweet",
			args{
				screenNames: []string{"TwitterAPI"},
				c:           make(chan twitter.Tweet),
				t:           SpyTrackedTime{},
			},
		}, {
			"multiple accounts new tweet",
			args{
				screenNames: []string{"twitter", "TwitterAPI"},
				c:           make(chan twitter.Tweet),
				t:           SpyTrackedTime{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authors := make(map[string]bool)
			// getting the tweets from all the names given
			for _, name := range tt.args.screenNames {
				// explicitly setting all author values to false for clarity later in the test
				name = strings.ToLower(name)
				authors[name] = false
				go getTweets(tt.args.c, name, tt.args.t)
			}

			// checking one tweet per author
			for i := 0; i < len(tt.args.screenNames); i++ {
				tweet := <-tt.args.c
				name := strings.ToLower(tweet.User.ScreenName)

				if _, ok := authors[name]; !ok {
					t.Errorf("got tweet from unexpected user %s", name)
				}
				authors[name] = true
			}
			close(tt.args.c)
		})
	}
}

func Test_postTweetToChannel(t *testing.T) {
	// TODO: Mock out discordgo ChannelMessageSend
	type args struct {
		tweet twitter.Tweet
		ch    *discordgo.Channel
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"successful post",
			args{
				tweet: twitter.Tweet{User: &twitter.User{ScreenName: "jack"}, ID: 20},
				ch:    &discordgo.Channel{ID: os.Getenv("BOT_TESTING_GROUND")},
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := postTweetToChannel(tt.args.tweet, tt.args.ch); (err != nil) != tt.wantErr {
				t.Errorf("postTweetToChannel() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
