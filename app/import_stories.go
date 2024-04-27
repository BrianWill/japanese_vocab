package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

var newlineRegEx *regexp.Regexp

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

	fmt.Println("Defaults source: ", storyJSON.Source)
	fmt.Println("NUM STORIES: ", len(storyJSON.Stories))

	//removeAllStories(sqldb)

	for _, s := range storyJSON.Stories {

		s.Title = strings.TrimSpace(s.Title)

		if s.Source == "" {
			s.Source = storyJSON.Source
		}

		if s.Level == "" {
			s.Level = "Intermediate"
		}

		if s.ContentFormat == "" {
			s.ContentFormat = storyJSON.ContentFormat
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
func importStory(story StoryImport, sqldb *sql.DB) error {
	epNum, err := strconv.Atoi(story.EpisodeNumber)
	if err != nil {
		fmt.Printf(`story "%s" has malformed episode number: %s`+"\n", story.Title, story.EpisodeNumber)
		return err
	}

	newWordCount, wordIdsJson, err := processStoryWords(story, sqldb)
	if err != nil {
		return err
	}

	if storyExists(story, sqldb) {
		fmt.Printf(`updating story: "%s"`+"\n", story.Title)

		_, err := sqldb.Exec(`UPDATE catalog_stories SET 
				date = $1, link = $2, episode_number = $3, audio = $4, video = $5, 
				content = $6, content_format = $7, transcript_en = $8, 
				transcript_en_format = $9, transcript_jp = $10, transcript_jp_format = $11, words = $12 
				WHERE title = $13 and source = $14;`,
			story.Date, story.Link, epNum, story.Audio, story.Video,
			story.Content, story.ContentFormat, story.TranscriptEN,
			story.TranscriptENFormat, story.TranscriptJP, story.TranscriptJPFormat, wordIdsJson,
			story.Title, story.Source)
		return err
	}

	fmt.Printf("importing story: %s, has %d new words \n", story.Title, newWordCount)

	_, err = sqldb.Exec(`INSERT INTO catalog_stories (title, source, date, link, episode_number, audio, video, 
				content, content_format, status, transcript_en, transcript_en_format, transcript_jp, transcript_jp_format, 
				words, repetitions_remaining, date_marked, level) 
				VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18);`,
		story.Title, story.Source, story.Date, story.Link, epNum,
		story.Audio, story.Video, story.Content, story.ContentFormat, "catalog",
		story.TranscriptEN, story.TranscriptENFormat, story.TranscriptJP, story.TranscriptJPFormat, wordIdsJson, 0, 0, story.Level)
	return err
}

func processStoryWords(story StoryImport, sqldb *sql.DB) (newWordCount int, wordIdsJson string, err error) {

	// remove newlines from the string in case words are split across lines
	if newlineRegEx == nil {
		newlineRegEx = regexp.MustCompile(`\x{000D}\x{000A}|[\x{000A}\x{000B}\x{000C}\x{000D}\x{0085}\x{2028}\x{2029}]`)
	}
	content := newlineRegEx.ReplaceAllString(story.Content, ``)

	tokens, kanjiSet, err := tokenize(content)
	if err != nil {
		return 0, "", fmt.Errorf("failure to tokenize story: " + err.Error())
	}

	newWordIds, newWordCount, err := addWords(tokens, kanjiSet, sqldb)
	if err != nil {
		return 0, "", fmt.Errorf("failure to add words: " + err.Error())
	}

	wordIdsJsonBytes, err := json.Marshal(newWordIds)
	if err != nil {
		return 0, "", fmt.Errorf("failure to jsonify word ids: " + err.Error())
	}

	return newWordCount, string(wordIdsJsonBytes), nil
}
