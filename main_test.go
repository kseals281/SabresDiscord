package main

import (
	"github.com/dghubble/go-twitter/twitter"
	"log"
	"reflect"
	"testing"
)

func Test_twitterHandler(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			"test",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			twitterHandler()
		})
	}
}

func Test_getTweets(t *testing.T) {
	type args struct {
		screenName string
		count      int
	}
	tests := []struct {
		name string
		args args
		want []twitter.Tweet
	}{
		{
			"zero tweets",
			args{
				screenName: "TwitterAPI",
				count:      0,
			},
			[]twitter.Tweet{},
		}, {
			"correct user",
			args{
				screenName: "TwitterAPI",
				count:      1,
			},
			[]twitter.Tweet{{User: &twitter.User{ScreenName: "TwitterAPI"}}},
		}, {
			"multiple tweets",
			args{
				screenName: "TwitterAPI",
				count:      2,
			},
			[]twitter.Tweet{
				{User: &twitter.User{ScreenName: "TwitterAPI"}}, {User: &twitter.User{ScreenName: "TwitterAPI"}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getTweets(tt.args.screenName, tt.args.count)
			log.Printf("len of tweets %d\n", len(got))
			for i, tweet := range got {
				if tweet.User.ScreenName != tt.want[i].User.ScreenName {
					t.Errorf("got tweet from user %s, wanted user %s",
						tt.want[i].User.ScreenName, tweet.User.ScreenName)
				}
			}
			if !reflect.DeepEqual(len(got), len(tt.want)) {
				t.Errorf("getTweets() = %v, want %v", got, tt.want)
			}
		})
	}
}
