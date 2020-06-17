package twitter

import (
	"bytes"
	"strings"
	"testing"
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

// func TestRetrieveFollowers(t *testing.T) {
// 	tw := &twitterClient{Consumerkey: "12", ConsumerToken: "34"}

// 	err := tw.getFollowers(context.Background())

// 	if err != nil {
// 		t.Errorf("Follower retrival failed with %+v", err)
// 	}

// }
