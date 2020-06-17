package twitter

import (
	"context"
	"log"
	"strings"

	"github.com/dghubble/go-twitter/twitter"
	_ "github.com/go-sql-driver/mysql"
)

type twFollower struct {
	followerid int
	Id         string
	dmsent     int
}

// todo: add db class to interact with SQLITE DB
// Table structures ,
// Request:
//	id
//  consumerkey
//  consumertoken
func db_saveRequest(sc *smClient, ctx context.Context, consumerkey string, consumertoken string, accesstoken string, accesssecret string) (int, error) {
	var idVal int
	db, teardown, err := sc.OpenConnection()
	if err != nil {
		return 0, err
	}
	defer teardown()

	_, err = db.ExecContext(ctx, `INSERT INTO REQUEST (consumerkey,consumertoken,accesstoken,accesssecret) VALUES(?,?,?,?) 
		ON DUPLICATE KEY UPDATE consumerkey = ?`, consumerkey, consumertoken, accesstoken, accesssecret, consumerkey)

	if err != nil {
		sc.log.Fatal("Error while inserting request")
		return 0, err
	}
	row := db.QueryRowContext(ctx, "SELECT id FROM REQUEST WHERE consumerkey=? and consumertoken=?", consumerkey, consumertoken)

	if err := row.Scan(&idVal); err != nil {
		// Check for a scan error.
		// Query rows will be closed with defer.
		log.Fatal(err)
		return 0, err
	}

	return idVal, nil
}

// followers:
//  request_id
//  Email
//  ID
//  FollowersCount
//  Location
//	DM Sent(Y/N)
//  TimeStamp
func db_saveUsers(sc *smClient, ctx context.Context, idVal int, userList []twitter.User) error {

	db, teardown, err := sc.OpenConnection()
	if err != nil {
		return err
	}
	defer teardown()

	// tx, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	// if err != nil {
	// 	sc.log.Fatal("error starting transaction")
	// 	return err
	// }

	var sb strings.Builder
	var vals []interface{}

	sb.WriteString(`INSERT INTO followers (request_id, email, id, followerscount, location)`)

	// ref: https://golang.org/ref/spec#For_statements
	for index, _ := range userList {
		user := userList[index]
		sb.WriteString(`values(?,?,?,?,?,?)`)
		if index < len(userList)-1 {
			sb.WriteString(",")
		}
		// ref: https://github.com/dghubble/go-twitter/blob/093ee2cf4e32ef608cded0ace84d5d82a52da4fa/twitter/users.go
		// ref: https://stackoverflow.com/questions/21108084/how-to-insert-multiple-data-at-once
		vals = append(vals, idVal, user.Email, user.IDStr, user.FollowersCount, user.Location)
	}

	sb.WriteString(` ON DUPLICATE KEY UPDATE followerscount = values(followerscount), location = values(location), dmsent = values(dmsent)`)

	_, err = db.ExecContext(ctx, sb.String(), vals)

	if err != nil {
		sc.log.Fatal("Error while inserting request")
		return err
	}

	return nil
}

func db_getFollowersbyLocation(sc *smClient, ctx context.Context, idVal int, maxCount int) ([]twFollower, error) {
	db, teardown, err := sc.OpenConnection()
	if err != nil {
		return nil, err
	}
	defer teardown()

	rows, err := db.QueryContext(ctx, `Select followerid, id 
						  from followers 
						  where request_id= ? and dmsent = 0 
						  order by location, followercount limit `+string(maxCount), idVal)

	if err != nil {
		sc.log.Fatal("Error fetching Followers data")
		return nil, err
	}
	defer rows.Close()

	var followersList []twFollower
	for rows.Next() {
		var twF twFollower
		if err := rows.Scan(&twF.followerid, &twF.Id); err != nil {
			sc.log.Fatal(err)
			return nil, err
		}
		followersList = append(followersList, twF)
	}

	return followersList, nil
}

func db_updateFollowerDMStatus(sc *smClient, ctx context.Context, list ...twFollower) error {
	var (
		vals []interface{}
		sb   strings.Builder
	)

	db, teardown, err := sc.OpenConnection()
	if err != nil {
		return err
	}
	defer teardown()

	sb.WriteString(` update followers 
				set dmset = 1 
				where `)

	for index, _ := range list {
		sb.WriteString("  followerid = ? ")
		if index < len(list) {
			sb.WriteString(` or `)
		}
		vals = append(vals, list[index].followerid)
	}

	_, err = db.ExecContext(ctx, sb.String(), vals...)

	if err != nil {
		sc.log.Fatal("Error while updating DM sent Status")
		return err
	}

	return nil
}
