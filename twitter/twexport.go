package twitter

import (
	"context"
	"errors"
	"flag"
	"io"
	"log"
	"os"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	"golang.org/x/sync/errgroup"
)

type smClient struct {
	log          *log.Logger
	dbconnString string
}

type twitterClient struct {
	smClient
	Consumerkey               string
	ConsumerSecret            string
	AccessToken               string
	AccessSecret              string
	DMMessage                 string
	MaxDMPerDay               int
	MaxFollowerRetrievalLimit int
}

func SendMessages(ctx context.Context, args []string, stdout io.Writer) error {
	_, err := newtwitterClient(args, stdout)

	return err
}

func newtwitterClient(args []string, stdout io.Writer) (*twitterClient, error) {
	var twClient twitterClient

	flags := flag.NewFlagSet(args[0], flag.ExitOnError)

	flags.StringVar(&twClient.Consumerkey, "twitterkey", "", "Twitter Consumer Key to use")
	flags.StringVar(&twClient.AccessToken, "twittertoken", "", "Twitter Consumer Token to use")
	flags.StringVar(&twClient.ConsumerSecret, "twitterconsumersecret", "", "Twitter Consumer secret to use")
	flags.StringVar(&twClient.AccessSecret, "twitteraccesssecret", "", "Twitter Access Secret to use")
	flags.StringVar(&twClient.DMMessage, "Message", "", "Direct Message to Send")

	flags.IntVar(&twClient.MaxDMPerDay, "twitterdmlimit", 1000, "Twitter Max DM Per day")
	flags.IntVar(&twClient.MaxFollowerRetrievalLimit, "twitterfollowerlimit", 200, "Twitter Max follower limit")

	if err := flags.Parse(args[1:]); err != nil {
		return nil, err
	}

	if twClient.Consumerkey == "" || twClient.ConsumerSecret == "" || twClient.AccessToken == "" || twClient.AccessSecret == "" {
		return nil, errors.New("Twitter Consumer key/secret and Access token/secret required")
	}

	if twClient.DMMessage == "" {
		return nil, errors.New("Direct Message to be sent is required")
	}

	twClient.log = log.New(stdout, "TWEXPORT: ", log.Lshortfile)

	return &twClient, nil
}

func getenv(name string) string {
	v := os.Getenv(name)
	if v == "" {
		panic("missing required environment variable " + name)
	}
	return v
}

func (t *twitterClient) SendDM(requestId int, ctx context.Context) error {
	// this should read max followers upto t.MaxDMPerDayfrom DB using some ranking say by location
	requestid, err := db_saveRequest(&t.smClient, ctx, t.Consumerkey, t.ConsumerSecret, t.AccessToken, t.AccessSecret)
	// send DM using API  & update status for each record

	followerList, err := db_getFollowersbyLocation(&t.smClient, ctx, requestid, t.MaxDMPerDay)
	if err != nil {
		return err
	}

	config := oauth1.NewConfig(t.Consumerkey, t.ConsumerSecret)
	token := oauth1.NewToken(t.AccessToken, t.AccessSecret)

	httpClient := config.Client(oauth1.NoContext, token)

	client := twitter.NewClient(httpClient)

	for index := 0; index < len(followerList); index++ {
		follower := followerList[index]
		_, _, err = client.DirectMessages.EventsNew(&twitter.DirectMessageEventsNewParams{
			Event: &twitter.DirectMessageEvent{
				Type: "message_create",
				Message: &twitter.DirectMessageEventMessage{
					Target: &twitter.DirectMessageTarget{
						RecipientID: follower.Id,
					},
					Data: &twitter.DirectMessageData{
						Text: t.DMMessage,
					},
				},
			},
		})
		if err != nil {
			t.log.Fatalf("Error sending DM for %s", follower.Id)
			return err
		} else {
			err = db_updateFollowerDMStatus(&t.smClient, ctx, follower)
			if err != nil {
				t.log.Fatalf("Error Updating DM Status for %s", follower.Id)
				return err
			}
		}
	}

	return nil
}

func (t *twitterClient) getFollowers(ctx context.Context) error {
	g, ctx := errgroup.WithContext(ctx)
	users := make(chan []twitter.User)
	/*
		Add/update request table and retrieve Id
	*/
	requestid, err := db_saveRequest(&t.smClient, ctx, t.Consumerkey, t.ConsumerSecret, t.AccessToken, t.AccessSecret)
	if err != nil {
		t.log.Fatal("Error saving Request")
		return err
	}

	// refer https://godoc.org/golang.org/x/sync/errgroup#ex-Group--Pipeline
	g.Go(func() error {
		defer close(users)

		config := oauth1.NewConfig(t.Consumerkey, t.ConsumerSecret)
		token := oauth1.NewToken(t.AccessToken, t.AccessSecret)

		httpClient := config.Client(oauth1.NoContext, token)

		client := twitter.NewClient(httpClient)

		params := twitter.FollowerListParams{Count: t.MaxFollowerRetrievalLimit}

		for {

			followers, _, err := client.Followers.List(&params)

			if err != nil {
				t.log.Fatal("Error while retrieving followers ")
				return err
			}

			//ref: https://stackoverflow.com/questions/24703943/passing-a-slice-into-a-channel
			newUsers := make([]twitter.User, len(followers.Users))
			copy(newUsers, followers.Users)

			select {
			case users <- newUsers:
			case <-ctx.Done():
				return ctx.Err()
			}

			if len(followers.Users) < t.MaxFollowerRetrievalLimit {
				return nil
			} else {
				// below is  set to get next set of followers
				params.Cursor = followers.NextCursor
			}

		}

	})

	// Below is routine to insert followers in database
	g.Go(func() error {
		for elem := range users {

			err := db_saveUsers(&t.smClient, ctx, requestid, elem)
			if err != nil {
				return err
			}

			select {
			case <-ctx.Done():
				return ctx.Err()
			}
		}
		return nil
	})

	go func() {
		g.Wait()
	}()

	if err := g.Wait(); err != nil {
		return err
	} else {
		return nil
	}
}
