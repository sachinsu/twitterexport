package twitter

import (
	"database/sql"
	"fmt"
	// _ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
)

func (s *smClient) OpenConnection() (*sql.DB, func() error, error) {
	//todo: Use Server.Configuration to get connection string, Pooling settings etc.
	dbConn, err := sql.Open("sqlite3", s.dbconnString)
	if err != nil {
		s.log.Printf("Error opening DB Connection, %+v", err)
		return nil, nil, fmt.Errorf("Error while opening DB Connection %w", err)
	}

	_, err = dbConn.Exec("CREATE TABLE IF NOT EXISTS REQUEST (id INTEGER PRIMARY KEY AUTOINCREMENT,consumerkey text,consumersecret text,accesstoken text,accesssecret text );")
	if err != nil {
		s.log.Printf("Error creating table, %+v", err)
		return nil, nil, fmt.Errorf("Error creating table %w", err)
	}

	_, err = dbConn.Exec("CREATE UNIQUE INDEX IF NOT EXISTS requnique on REQUEST (consumerkey,consumersecret,accesstoken,accesssecret);")
	if err != nil {
		s.log.Printf("Error creating index, %+v", err)
		return nil, nil, fmt.Errorf("Error creating table %w", err)
	}

	_, err = dbConn.Exec("CREATE TABLE IF NOT EXISTS FOLLOWERS (followerid INTEGER PRIMARY KEY AUTOINCREMENT,request_id int, email text, id text,followerscount int, location text, dmsent int default 0 );")
	if err != nil {
		s.log.Printf("Error creating table, %+v", err)
		return nil, nil, fmt.Errorf("Error creating table %w", err)
	}

	_, err = dbConn.Exec("CREATE UNIQUE INDEX IF NOT EXISTS funique on FOLLOWERS (request_id, email);")
	if err != nil {
		s.log.Printf("Error creating index, %+v", err)
		return nil, nil, fmt.Errorf("Error creating table %w", err)
	}
	// db.SetMaxIdleConns(2)
	// db.SetMaxOpenConns(5)

	return dbConn, dbConn.Close, nil
}
