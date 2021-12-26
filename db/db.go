package db

import (
	"github.com/boltdb/bolt"
	log "github.com/sirupsen/logrus"
)

var db *bolt.DB

func Init() {
	boltDB, err := bolt.Open("my.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	db = boltDB
}

func Destroy() {
	db.Close()
}

func Insert(data map[string]interface{}) {
	
}
