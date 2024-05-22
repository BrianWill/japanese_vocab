package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/asticode/go-astisub"
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

func importSources(dbPath string) error {
	fmt.Println("importing sources...")

	sqldb, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		panic(err)
	}
	defer sqldb.Close()

	const SOURCES_PATH = "../static/sources/"

	entries, err := os.ReadDir(SOURCES_PATH)
	if err != nil {
		log.Fatal(err)
	}

	for _, e := range entries {
		if e.IsDir() {
			err = importSource(SOURCES_PATH, e.Name(), sqldb)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func importSource(sourcePath string, source string, sqldb *sql.DB) error {
	entries, err := os.ReadDir(sourcePath + source)
	if err != nil {
		log.Fatal(err)
	}

	storyMap := make(map[int]*StoryImport) // episode number to

	for _, entry := range entries {
		name := entry.Name()
		components := strings.Split(name, ".")
		if len(components) < 2 {
			return fmt.Errorf("malformed file name in source: %s", name)
		}

		epNumber, err := strconv.Atoi(components[len(components)-2])
		if err != nil {
			return fmt.Errorf("malformed episode number in file: %s", name)
		}

		story, ok := storyMap[epNumber]
		if !ok {
			story = &StoryImport{}
			storyMap[epNumber] = story
		}
		numStr := strconv.Itoa(epNumber)
		story.EpisodeNumber = numStr
		story.Title = source + " - ep " + numStr
		story.Source = source
		story.Level = "medium"
		story.ContentFormat = "text"

		extension := components[len(components)-1]
		isVideo := extension == "mp4"
		if isVideo {
			story.Video = name
		}
		isSubtitle := extension == "vtt" || extension == "ass" || extension == "srt"
		if isSubtitle {
			if len(components) < 3 {
				return fmt.Errorf("subtitle file does not specify language")
			}
			lang := components[len(components)-3]
			switch lang {
			case "en":
				story.TranscriptEN, story.Content, err = getSubtitles(sourcePath + source + "/" + name)
				if err != nil {
					return err
				}
			case "ja":
				story.TranscriptJA, story.Content, err = getSubtitles(sourcePath + source + "/" + name)
				if err != nil {
					return err
				}
			default:
				return fmt.Errorf("subtitle file language is invalid: %s", name)
			}
		}
	}

	for _, s := range storyMap {
		if s.Video == "" {
			continue
		}

		err = importStory(*s, sqldb)
		if err != nil {
			return err
		}
	}

	return nil
}

func getSubtitles(path string) (newSubtitles string, content string, err error) {
	subs, err := astisub.OpenFile(path)
	if err != nil {
		return "", "", err
	}

	var sb strings.Builder
	for _, item := range subs.Items {
		for _, line := range item.Lines {
			for _, lineItem := range line.Items {
				sb.WriteString(lineItem.Text)
			}
			sb.WriteString("\n")
		}
	}

	var buf = &bytes.Buffer{}
	err = subs.WriteToWebVTT(buf)
	if err != nil {
		return "", "", err
	}
	return buf.String(), sb.String(), nil
}

func storyExists(story StoryImport, sqldb *sql.DB) bool {
	var id int64

	err := sqldb.QueryRow(`SELECT id FROM stories WHERE title = $1 and source = $2;`,
		story.Title, story.Source).Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			return false
		}
		log.Fatal(err)
	}

	return true
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

		_, err := sqldb.Exec(`UPDATE stories SET 
				date = $1, link = $2, episode_number = $3, audio = $4, video = $5, 
				content = $6, content_format = $7, 
				transcript_en = CASE WHEN transcript_en = '' THEN $8 ELSE transcript_en END,
				transcript_ja = CASE WHEN transcript_ja = '' THEN $8 ELSE transcript_ja END,
				words = $10 
				WHERE title = $11 and source = $12;`,
			story.Date, story.Link, epNum, story.Audio, story.Video,
			story.Content, story.ContentFormat, story.TranscriptEN,
			story.TranscriptJA, wordIdsJson,
			story.Title, story.Source)
		return err
	}

	fmt.Printf("importing story: %s, has %d new words \n", story.Title, newWordCount)

	_, err = sqldb.Exec(`INSERT INTO stories (title, source, date, link, episode_number, audio, video, 
				content, content_format, archived, transcript_en, transcript_ja, 
				words, repetitions, level) 
				VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15);`,
		story.Title, story.Source, story.Date, story.Link, epNum,
		story.Audio, story.Video, story.Content, story.ContentFormat, 0,
		story.TranscriptEN, story.TranscriptJA, wordIdsJson, 0, story.Level)
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
		return 0, "", err
	}

	wordIdsJsonBytes, err := json.Marshal(newWordIds)
	if err != nil {
		return 0, "", fmt.Errorf("failure to jsonify word ids: " + err.Error())
	}

	return newWordCount, string(wordIdsJsonBytes), nil
}
