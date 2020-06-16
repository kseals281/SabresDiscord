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

	twitterHandler()

	<-interrupt
}

func twitterHandler() {
	// Twitter rate limits requests to 100,000 per day OR 1500 per min so we check every 5 sec
	for {
		if time.Now().Second()%5 == 0 {
			//tweets := getTweets()
			//fmt.Println("********************START********************")
			//for _, t := range tweets {
			//	fmt.Printf("%+v\n", t.Text)
			//}
			//fmt.Println("********************END********************")
		}
		time.Sleep(time.Second)
	}

}

func getTweets(screenName string, c int) []twitter.Tweet {
	tweets, _, err := Client.Timelines.UserTimeline(&twitter.UserTimelineParams{
		ScreenName:     screenName,
		Count:          c,
		ExcludeReplies: newTrue(),
	})
	tweets = tweets[:c] // Count does not actually reduce the number of tweets received.
	if err != nil {
		log.Printf("Error retrieving tweets from timeline: %+v", err)
	}
	return tweets
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
