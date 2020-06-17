package twitter

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
)

func (s *smClient) OpenConnection() (*sql.DB, func() error, error) {
	//todo: Use Server.Configuration to get connection string, Pooling settings etc.
	dbConn, err := sql.Open("mysql", s.dbconnString)
	if err != nil {
		s.log.Printf("Error opening DB Connection, %+v", err)
		return nil, nil, fmt.Errorf("Error while opening DB Connection %w", err)
	}

	// db.SetMaxIdleConns(2)
	// db.SetMaxOpenConns(5)

	return dbConn, dbConn.Close, nil
}
