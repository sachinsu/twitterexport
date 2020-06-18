package twitter

import (
	"bytes"
	"context"
	"os"
	"strings"
	"testing"

	"github.com/dghubble/go-twitter/twitter"
)

func TestFlagsCorrectness(t *testing.T) {
	var tests = []struct {
		args []string
		err  string
		tc   *twitterClient
	}{
		{[]string{"-verbose"},
			"Twitter Consumer key/secret and Access token/secret required", nil},
		{[]string{"-verbose", "-twittertoken=12"},
			"Twitter Consumer key/secret and Access token/secret required", nil},
		{[]string{"-verbose", "-twitterkey=12"},
			"Twitter Consumer key/secret and Access token/secret required", nil},
		{[]string{"-verbose", "-twitteraccesssecret=12", "-twitterkey=34", "-twittertoken=12", "-twitterconsumersecret=12"},
			"Direct Message to be sent is required", nil},
		{[]string{"-verbose", "-Message=high", "-twitteraccesssecret=12", "-twitterkey=34", "-twittertoken=1", "-twitterconsumersecret=2"},
			"", &twitterClient{DMMessage: "high", AccessSecret: "12", Consumerkey: "34", AccessToken: "1", ConsumerSecret: "2", MaxDMPerDay: 1000, MaxFollowerRetrievalLimit: 200}},
	}

	for _, tt := range tests {
		t.Run(strings.Join(tt.args, " "), func(t *testing.T) {
			var buf bytes.Buffer
			t.Logf("%s\n", tt.args)
			tC, err := newtwitterClient(tt.args, &buf)

			if err != nil && tt.err != err.Error() {
				t.Errorf("expected %+v and got %+v", tt.err, err)
			}

			if tt.tc != nil && !(tC.DMMessage == tt.tc.DMMessage &&
				tC.Consumerkey == tt.tc.Consumerkey &&
				tC.ConsumerSecret == tt.tc.ConsumerSecret &&
				tC.AccessToken == tt.tc.AccessToken &&
				tC.AccessSecret == tt.tc.AccessSecret) {
				t.Errorf("expected %+v and got %+v", tt.tc, tC)
			}
		})
	}
}

func cleanupDB(tw *twitterClient, t *testing.T) {
	t.Cleanup(func() {
		db, teardown, err := tw.smClient.OpenConnection()
		if err != nil {
			t.Logf("Error while cleanup %v", err)
			return
		}
		defer teardown()

		_, _ = db.Exec("delete  from REQUEST")

		_, _ = db.Exec("delete  from Followers")

	})
}

func TestAddRequest(t *testing.T) {
	ctx := context.Background()
	t.Log("starting test ...")
	args := []string{"-verbose", "-Message=high", "-twitteraccesssecret=12", "-twitterkey=34", "-twittertoken=1", "-twitterconsumersecret=2"}

	tw, err := newtwitterClient(args, os.Stdout)

	cleanupDB(tw, t)

	if err != nil {
		t.Errorf("expected no error but got %+v", err)
	}
	t.Log("calling save request ...")

	id, err := db_saveRequest(&tw.smClient, ctx, tw.Consumerkey, tw.ConsumerSecret, tw.AccessToken, tw.AccessSecret)

	if err != nil {
		t.Errorf("expected no error but got %+v", err)
	}

	if id == 0 {
		t.Errorf("expected non zero id but got %d", id)
	}

	newid, err := db_saveRequest(&tw.smClient, ctx, tw.Consumerkey, tw.ConsumerSecret, tw.AccessToken, tw.AccessSecret)

	if err != nil {
		t.Errorf("expected no error but got %+v", err)
	}

	if id != newid {
		t.Errorf("expected same id %d but got %d", id, newid)
	}

}

func TestAddFollowers(t *testing.T) {

	ctx := context.Background()
	t.Log("starting test ...")
	args := []string{"-verbose", "-Message=high", "-twitteraccesssecret=12", "-twitterkey=34", "-twittertoken=1", "-twitterconsumersecret=2"}

	tw, err := newtwitterClient(args, os.Stdout)

	cleanupDB(tw, t)

	if err != nil {
		t.Errorf("expected no error but got %+v", err)
	}
	t.Log("calling save request ...")

	id, err := db_saveRequest(&tw.smClient, ctx, tw.Consumerkey, tw.ConsumerSecret, tw.AccessToken, tw.AccessSecret)

	if err != nil {
		t.Errorf("expected no error but got %+v", err)
	}

	if id == 0 {
		t.Errorf("expected non zero id but got %d", id)
	}

	var ulist []twitter.User

	for i := 1; i < 6; i++ {
		var t twitter.User

		t.Email = "abc" + string(i) + "@mail.com"
		t.IDStr = "abc" + string(i*10)
		t.IDStr = "abc" + string(i*10)
		t.FollowersCount = i * 20
		t.Location = "abc" + string(i)

		ulist = append(ulist, t)
	}

	err = db_saveUsers(&tw.smClient, ctx, id, ulist)

	if err != nil {
		t.Errorf("expected no error but got %+v", err)
	}

	userlist, err := db_getFollowersbyLocation(&tw.smClient, ctx, id, tw.MaxDMPerDay)

	if err != nil {
		t.Errorf("expected no error but got %+v", err)
	}

	if len(ulist) != len(userlist) {
		t.Errorf("expected count %d but got %d", len(ulist), len(userlist))
	}

}

func TestUpdateDMStatus(t *testing.T) {
	ctx := context.Background()
	t.Log("starting test ...")
	args := []string{"-verbose", "-Message=high", "-twitteraccesssecret=112", "-twitterkey=34", "-twittertoken=1", "-twitterconsumersecret=2"}

	tw, err := newtwitterClient(args, os.Stdout)

	cleanupDB(tw, t)

	if err != nil {
		t.Errorf("expected no error but got %+v", err)
	}
	t.Log("calling save request ...")

	id, err := db_saveRequest(&tw.smClient, ctx, tw.Consumerkey, tw.ConsumerSecret, tw.AccessToken, tw.AccessSecret)

	if err != nil {
		t.Errorf("expected no error but got %+v", err)
	}

	if id == 0 {
		t.Errorf("expected non zero id but got %d", id)
	}

	var ulist []twitter.User

	for i := 1; i < 6; i++ {
		var t twitter.User

		t.Email = "abc" + string(i) + "@mail.com"
		t.IDStr = "abc" + string(i*10)
		t.IDStr = "abc" + string(i*10)
		t.FollowersCount = i * 20
		t.Location = "abc" + string(i)

		ulist = append(ulist, t)
	}

	err = db_saveUsers(&tw.smClient, ctx, id, ulist)

	if err != nil {
		t.Errorf("expected no error but got %+v", err)
	}

	userlist, err := db_getFollowersbyLocation(&tw.smClient, ctx, id, tw.MaxDMPerDay)

	if err != nil {
		t.Errorf("expected no error but got %+v", err)
	}

	if len(ulist) != len(userlist) {
		t.Errorf("expected count %d but got %d", len(ulist), len(userlist))
	}

	err = db_updateFollowerDMStatus(&tw.smClient, ctx, userlist...)

	if err != nil {
		t.Errorf("expected no error but got %+v", err)
	}

}
