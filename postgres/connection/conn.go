package connection

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/nuveo/prest/config"
)

var db *sqlx.DB
var cfg config.Prest

func init() {
	cfg := config.Prest{}
	config.Parse(&cfg)
}

// MustGet get postgre cpnnection
func MustGet() *sqlx.DB {
	if db == nil {
		var err error
		dbURI := fmt.Sprintf("user=%s dbname=%s host=%s port=%v sslmode=disable", cfg.PGUser, cfg.PGDatabase, cfg.PGHost, cfg.PGPort)
		if cfg.PGPass != "" {
			dbURI += " password=" + cfg.PGPass
		}
		db, err = sqlx.Connect("postgres", dbURI)
		if err != nil {
			panic(fmt.Sprintf("Unable to connection to database: %v\n", err))
		}
		db.SetMaxIdleConns(cfg.PGMaxIdleConn)
		db.SetMaxOpenConns(cfg.PGMAxOpenConn)
	}
	return db
}
