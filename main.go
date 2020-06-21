package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/dghubble/go-twitter/twitter"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/oauth2/clientcredentials"

	"github.com/bwmarrin/discordgo"
)

// Session is declared in the global space so it can be easily used throughout this program.
// In this use case, there is no error that would be returned.
// We also declare the twitter Client for easy use as well.
var Session, _ = discordgo.New()
var Client *twitter.Client

type TrackedTime interface {
	Now() time.Time
	Second() time.Duration
	Epoch() time.Time
}

type DefaultTime struct{}

func (t *DefaultTime) Now() time.Time {
	return time.Now()
}

func (t *DefaultTime) Second() time.Duration {
	return time.Second
}

func (t *DefaultTime) Epoch() time.Time {
	return time.Date(1970, 1, 1, 0, 0, 0, 0, &time.Location{})
}

// Read in all options from environment variables and command line arguments.
func init() {
	rand.Seed(time.Now().Unix())

	// Discord Authentication Token
	Session.Token = os.Getenv("DISCORD_SABRES_TWITTER_BOT")
	if Session.Token == "" {
		// Pointer, flag, default, description
		flag.StringVar(&Session.Token, "t", "", "Discord Authentication Token")
	}

	// Twitter Access Token
	config := &clientcredentials.Config{
		ClientID:     os.Getenv("SABRES_TWITTER_API_KEY"),
		ClientSecret: os.Getenv("SABRES_TWITTER_API_SECRET"),
		TokenURL:     "https://api.twitter.com/oauth2/token",
	}
	if config.ClientID == "" || config.ClientSecret == "" {
		log.Println("You must provide a Twitter API key and API secret.")
	}
	httpClient := config.Client(context.Background())
	Client = twitter.NewClient(httpClient)
}

func main() {
	// Declare any variables needed later.
	var err error

	// Setup interrupt
	interrupt := make(chan os.Signal)
	signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)

	// Parse command line arguments
	flag.Parse()

	// Verify a Token was provided
	if Session.Token == "" {
		log.Println("You must provide a Discord authentication token.")
		return
	}

	// Verify the Token is valid and grab user information
	Session.State.User, err = Session.User("@me")
	errCheck("error retrieving account", err)

	Session.AddHandler(func(discord *discordgo.Session, ready *discordgo.Ready) {
		err = discord.UpdateStatus(0, "Let's go Buffalo!")
		if err != nil {
			fmt.Println("Error attempting to set my status")
		}
		servers := discord.State.Guilds
		fmt.Printf("SabresPRTwitter has started on %d servers\n", len(servers))
	})

	// Open a websocket connection to Discord
	err = Session.Open()
	defer Session.Close()
	errCheck("Error opening connection to Discord", err)

	go twitterHandler(Session)

	<-interrupt
}

func twitterHandler(discord *discordgo.Session) {
	ch, _ := discord.Channel(os.Getenv("BOT_TESTING_GROUND"))
	acts := []string{"NHL", "NFL", "NBA", "MLB"}
	feed := make(chan twitter.Tweet)

	d := DefaultTime{}

	for _, s := range acts {
		go getTweets(feed, s, &d)
	}

	for t := range feed {

		err := postTweetToChannel(t, ch)

		if err != nil {
			log.Println(fmt.Sprintf("Error posting tweet to channel: %+v", err))
		}
	}
}

func getTweets(c chan twitter.Tweet, screenName string, tt TrackedTime) {
	// Keep track of the time of the last tweet
	var prevTweetTime time.Time
	prevTweetTime = tt.Epoch()

	for {
		tweets, _, err := Client.Timelines.UserTimeline(&twitter.UserTimelineParams{
			ScreenName:     screenName,
			Count:          1,
			ExcludeReplies: newTrue(),
		})
		if err != nil {
			log.Printf("Error retrieving tweets from timeline: %+v", err)
			continue
		}

		// Check to see if tweet is new
		tweet := tweets[0]
		if curr, _ := tweet.CreatedAtTime(); curr.After(prevTweetTime) {
			c <- tweet
			prevTweetTime = curr
		}

		// Twitter rate limits requests to 100,000 per day OR 1500 per min so we check every second
		time.Sleep(tt.Second())
	}
}

func postTweetToChannel(tweet twitter.Tweet, ch *discordgo.Channel) error {
	// Posting link crudely to chat so that the user can trust the embed
	url := fmt.Sprintf("https://twitter.com/%s/status/%d", tweet.User.ScreenName, tweet.ID)

	_, err := Session.ChannelMessageSend(ch.ID, url)
	if err != nil {
		log.Println(fmt.Sprintf("Error sending tweet to channel %s", ch.Name), err)
		return err
	}
	return nil
}

func errCheck(msg string, err error) {
	if err != nil {
		fmt.Printf("%s: %+v", msg, err)
		panic(err)
	}
}

func newTrue() *bool {
	b := true
	return &b
}
