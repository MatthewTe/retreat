package rssparcer

import (
	"database/sql"
	"fmt"
	"log"
	"testing"

	"github.com/MatthewTe/retreat/database"
)

var testFeeds = []string{
	"https://www.google.ca/alerts/feeds/04364436683859664111/15221188731977553381",
}

func TestRssFeedCoreContentExtraction(t *testing.T) {

	existingDB, _ := sql.Open("sqlite", "file:memdb1?mode=memory&cache=shared")
	db, err := database.InitDatabase(existingDB)
	if err != nil {
		log.Fatalln(err)
	}

	changedFeeds, err := CheckFeedsForChanges(testFeeds, db)
	if err != nil {
		log.Fatal(err)
	}

	updatedArticles, err := CheckArticlesForChanges(changedFeeds, db)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(updatedArticles)

}
