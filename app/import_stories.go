package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

func importStories(dbPath string, jsonPath string) error {
	fmt.Println("importing stories...")
	jsonBytes, err := os.ReadFile(jsonPath)
	if err != nil {
		panic(err)
	}

	// parse the json file
	storyJSON := StoryImportJSON{}
	storyJSON.Stories = make([]StoryImport, 0)
	err = json.Unmarshal(jsonBytes, &storyJSON)
	if err != nil {
		panic(err)
	}

	// open the db
	sqldb, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		panic(err)
	}
	defer sqldb.Close()

	fmt.Println("Defaults source: ", storyJSON.Defaults.Source)
	fmt.Println("NUM STORIES: ", len(storyJSON.Stories))

	removeAllStories(sqldb)

	for _, s := range storyJSON.Stories {

		if s.Source == "" {
			s.Source = storyJSON.Defaults.Source
		}

		err = importStory(s, sqldb)
		if err != nil {
			log.Fatal(err)
			return err
		}
	}
	return nil
}

func isUrl(s string) bool {
	return strings.HasPrefix(s, "http")
}

func storyExists(story StoryImport, sqldb *sql.DB) bool {
	var id int64

	err := sqldb.QueryRow(`SELECT id FROM catalog_stories WHERE title = $1 and source = $2;`,
		story.Title, story.Source).Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			return false
		}
		log.Fatal(err)
	}

	return true
}

func removeAllStories(sqldb *sql.DB) {
	// atlernatively, `DELETE FROM catalog_stories;`
	_, err := sqldb.Exec(`DELETE FROM catalog_stories;`)
	if err != nil {
		log.Fatal(err)
	}
}

// (the story will be parsed into lines and words only when its added from the catalog to the main story table,
// so in this importer, we just check that the data is valid)
// check that the content can be parsed as the specified format
func importStory(s StoryImport, sqldb *sql.DB) error {
	if storyExists(s, sqldb) {
		fmt.Printf(`story "%s" already exists\n`, s.Title)
		return nil
	}

	epNum, err := strconv.Atoi(s.EpisodeNumber)
	if err != nil {
		fmt.Printf(`story "%s" has malformed episode number: %s \n`, s.Title, s.EpisodeNumber)
		return err
	}

	fmt.Println("importing: ", s.Title)

	_, err = sqldb.Exec(`INSERT INTO catalog_stories (title, source, date, link, episode_number, audio, video, 
				content, content_format, status, transcript_en, transcript_en_format, transcript_jp, transcript_jp_format) 
				VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14);`,
		s.Title, s.Source, s.Date, s.Link, epNum,
		s.Audio, s.Video, s.Content, s.ContentFormat, "catalog",
		s.TranscriptEN, s.TranscriptENFormat, s.TranscriptJP, s.TranscriptJPFormat)
	return err
}
