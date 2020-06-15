package twitter

import (
	"reflect"
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
			"Twitter Consumer key is empty", nil},
		{[]string{"verbose", "-twittertoken=12"},
			"Twitter Consumer key is empty", nil},
		{[]string{"verbose", "-twitterkey=12"},
			"Twitter Consumer token is empty", nil},
		{[]string{"verbose", "-twitterkey=12", "-twittertoken=34"},
			"", &twitterClient{Consumerkey: "12", ConsumerToken: "34",MaxDMPerDay: 1000,MaxFollowerRetrievalLimit: 200}},
	}

	for _, tt := range tests {
		t.Run(strings.Join(tt.args, " "), func(t *testing.T) {
			t.Logf("%s\n", tt.args)
			tC, err := newtwitterClient(tt.args)

			if err != nil && tt.err != err.Error() {
				t.Errorf("expected %+v and got %+v", tt.err, err)
			}

			if !reflect.DeepEqual(tC, tt.tc) {
				t.Errorf("expected %+v and got %+v", tt.tc, tC)
			}
		})
	}
}


func TestRetrieveFollowers(t *testing.T) {
	tw := &twitterClient{Consumerkey: "12", ConsumerToken: "34"}

	err := tw.getFollowers()

	if err != nil {
		t.Errorf("Follower retrival failed with %+v", err)
	}

	
}