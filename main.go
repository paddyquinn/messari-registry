package main

import (
	"github.com/paddyquinn/messari/database"
	"github.com/paddyquinn/messari/server"
	log "github.com/sirupsen/logrus"
)

const errorKey = "error"

func main() {
	// Establish connection to SQLite database.
	sqlite, err := database.NewSQLite()
	if err != nil {
		log.WithField(errorKey, err.Error()).Fatal("could not establish connection to sqlite")
	}
	defer sqlite.Close()

	// Start the server.
	srv := server.NewServer(sqlite)
	if err = srv.Start(); err != nil {
		log.WithField(errorKey, err.Error()).Fatal("server failed to start")
	}
}
