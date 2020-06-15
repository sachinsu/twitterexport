package twitter

import (
	"errors"
	"flag"
	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	// "io"
)

type twitterClient struct {
	Consumerkey   string
	ConsumerToken string
	MaxDMPerDay	int
	MaxFollowerRetrievalLimit int
}

func SendMessages(args []string) error {
	_, err := newtwitterClient(args)
	return err

}

func newtwitterClient(args []string) (*twitterClient, error) {
	var twClient twitterClient

	flags := flag.NewFlagSet(args[0], flag.ExitOnError)

	flags.StringVar(&twClient.Consumerkey, "twitterkey", "", "Twitter Consumer Key to use")
	flags.StringVar(&twClient.ConsumerToken, "twittertoken", "", "Twitter Consumer Token to use")
	flags.IntVar(&twClient.MaxDMPerDay, "twitterdmlimit", 1000, "Twitter Max DM Per day")
	flags.IntVar(&twClient.MaxFollowerRetrievalLimit, "twitterfollowerlimit", 200, "Twitter Max follower limit")

	if err := flags.Parse(args[1:]); err != nil {
		return nil, err
	}

	if twClient.Consumerkey == "" {
		return nil, errors.New("Twitter Consumer key is empty")
	}

	if twClient.ConsumerToken == "" {
		return nil, errors.New("Twitter Consumer token is empty")
	}

	return &twClient, nil
}

func (t *twitterClient) getFollowers() error {
	config := oauth1.NewConfig("consumerKey", t.Consumerkey)
	token := oauth1.NewToken("accessToken", t.ConsumerToken)

	httpClient := config.Client(oauth1.NoContext, token)

	// Twitter client
	client := twitter.NewClient(httpClient)

	// Followers
	// followers, resp, err := client.Followers.List(&twitter.FollowerListParams{})
	_, _, err := client.Followers.List(&twitter.FollowerListParams{})

	// todo: add db class to interact with SQLITE DB 
	// Table structures , 
	// Request: 
	//	id
	//  consumerkey 
	//  consumertoken 
	// followers:
	//  request_id
	//  Email
	//  FollowersCount                 
	//  Location
	//	DM Sent(Y/N)
	//  TimeStamp
	// todo: Get the count of followers and if greater than per day limit then loop thru in batches of max limit per day
	// todo: start a goroutine to save the follower details to DB (this should be upsert)

	return err
}
