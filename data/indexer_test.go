package data

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"os"
	"testing"
	"time"
)

func TestIndex(t *testing.T) {
	now := time.Now()
	fileName := fmt.Sprintf("filed_%v.db", now.Unix())

	db, err := sql.Open("sqlite3", fileName)
	if err != nil {
		log.Fatal(err)
	}

	fileRepository := NewSQLiteRepository(db)
	if err := fileRepository.Migrate(); err != nil {
		log.Fatal(err)
	}

	indexer := NewIndexer(fileRepository)

	// TODO
	indexPath := ""

	if err := indexer.Index(&indexPath); err != nil {
		log.Fatal(err)
	}

	if err := os.Remove(fileName); err != nil {
		log.Fatal(err)
	}
}
