package twitter

import (
	"context"
	"database/sql"
	"strconv"
	"strings"

	"github.com/dghubble/go-twitter/twitter"
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
func db_saveRequest(sc *smClient, ctx context.Context, consumerkey string, consumersecret string, accesstoken string, accesssecret string) (int, error) {
	var idVal int
	sc.log.Printf("Opening Database with %s", sc.dbconnString)

	db, teardown, err := sc.OpenConnection()
	if err != nil {
		return 0, err
	}
	defer teardown()

	_, err = db.ExecContext(ctx, `INSERT INTO REQUEST (consumerkey,consumersecret,accesstoken,accesssecret) VALUES(?,?,?,?) 
		ON CONFLICT(consumerkey,consumersecret,accesstoken,accesssecret) DO NOTHING`, consumerkey, consumersecret, accesstoken, accesssecret, consumerkey)

	if err != nil {
		sc.log.Printf("Error while inserting request %+v", err)
		return 0, err
	}
	row := db.QueryRowContext(ctx, "SELECT id FROM REQUEST WHERE consumerkey=? and consumersecret=? and accesstoken=? and accesssecret=?", consumerkey, consumersecret, accesstoken, accesssecret)

	if err := row.Scan(&idVal); err != nil {
		// Check for a scan error.
		// Query rows will be closed with defer.
		sc.log.Print("Error while retrieving id")
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

	tx, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		sc.log.Fatal("error starting transaction")
		return err
	}
	stmt, err := tx.Prepare(`INSERT INTO followers (request_id, email, id, followerscount, location)
										values (?,?,?,?,?) 
										on conflict(request_id, email) Do update set 
											id=excluded.id, 
											followerscount=excluded.followerscount,
											location=excluded.location`)

	if err != nil {
		sc.log.Print("Error while preparing insert statement")
		return err
	}
	// ref: https://golang.org/ref/spec#For_statements
	for index, _ := range userList {
		user := userList[index]
		//ref:https://www.sqlite.org/lang_UPSERT.html
		_, err := stmt.ExecContext(ctx, idVal, user.Email, user.IDStr, user.FollowersCount, user.Location)
		if err != nil {
			break
		}
		// ref: https://github.com/dghubble/go-twitter/blob/093ee2cf4e32ef608cded0ace84d5d82a52da4fa/twitter/users.go
		// ref: https://stackoverflow.com/questions/21108084/how-to-insert-multiple-data-at-once
	}

	if err != nil {
		sc.log.Print("Error while inserting followers")
		return err
	} else {
		if err = tx.Commit(); err != nil {
			sc.log.Fatal("Error while committing data")
			return err
		}
	}

	return nil
}

func db_getFollowersbyLocation(sc *smClient, ctx context.Context, idVal int, maxCount int) ([]twFollower, error) {
	db, teardown, err := sc.OpenConnection()
	if err != nil {
		return nil, err
	}
	defer teardown()

	stmt := `Select followerid, id 
						  from followers 
						  where request_id= ? and dmsent = 0
						  order by location, followerscount limit ` + strconv.Itoa(maxCount)

	rows, err := db.QueryContext(ctx, stmt, idVal)

	if err != nil {
		sc.log.Print("Error fetching Followers data")
		return nil, err
	}
	defer rows.Close()

	var followersList []twFollower
	for rows.Next() {
		var twF twFollower
		if err := rows.Scan(&twF.followerid, &twF.Id); err != nil {
			sc.log.Print(err)
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
				set dmsent = 1
				where followerid = 0`)

	for index, _ := range list {
		sb.WriteString(" or followerid = ? ")
		vals = append(vals, list[index].followerid)
	}

	_, err = db.ExecContext(ctx, sb.String(), vals...)

	if err != nil {
		sc.log.Print("Error while updating DM sent Status")
		return err
	}

	return nil
}
