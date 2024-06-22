package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/asticode/go-astisub"
	_ "github.com/mattn/go-sqlite3"
)

const SOURCES_PATH = "../static/sources/"

var newlineRegEx *regexp.Regexp

func importSources(dbPath string) error {
	fmt.Println("importing sources...")

	sqldb, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		panic(err)
	}
	defer sqldb.Close()

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

	storyMap := make(map[int]*Story) // episode number to

	for _, entry := range entries {
		name := entry.Name()
		components := strings.Split(name, ".")
		if len(components) < 2 {
			return fmt.Errorf("malformed file name in source: %s", name)
		}

		extension := components[len(components)-1]
		isVideo := extension == "mp4"
		isSubtitle := extension == "vtt" || extension == "ass" || extension == "srt"
		if !isVideo && !isSubtitle {
			continue
		}

		epNumber, err := strconv.Atoi(components[len(components)-2])
		if err != nil {
			return fmt.Errorf("malformed episode number in file: %s", name)
		}

		story, ok := storyMap[epNumber]
		if !ok {
			story = &Story{}
			storyMap[epNumber] = story
		}
		numStr := strconv.Itoa(epNumber)
		story.EpisodeNumber = epNumber
		story.Title = source + " - ep " + numStr
		story.Source = source
		story.ContentFormat = "text"

		if isVideo {
			story.Video = name
		}

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

func getSubtitlesContent(vtt string) (string, error) {
	subs, err := astisub.ReadFromWebVTT(bytes.NewReader([]byte(vtt)))
	if err != nil {
		return "", err
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

	return sb.String(), nil
}

func getSubtitlesContentInTimeRange(vtt string, startTime float64, endTime float64) (string, error) {
	subs, err := astisub.ReadFromWebVTT(bytes.NewReader([]byte(vtt)))
	if err != nil {
		return "", err
	}

	if endTime == 0 {
		endTime = math.MaxFloat64
	}

	var sb strings.Builder
	for _, item := range subs.Items {
		if float64(item.EndAt) < startTime || float64(item.StartAt) > endTime {
			continue
		}
		for _, line := range item.Lines {
			for _, lineItem := range line.Items {
				sb.WriteString(lineItem.Text)
			}
			sb.WriteString("\n")
		}
	}

	return sb.String(), nil
}

func storyExists(story Story, sqldb *sql.DB) bool {
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
func importStory(story Story, sqldb *sql.DB) error {
	newWordCount, _, err := processStoryWords(story, sqldb)
	if err != nil {
		return err
	}

	var jsonPath string = SOURCES_PATH + story.Source + "/" + story.Title + ".json"

	jsonExisted := true
	if _, err := os.Stat(jsonPath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			jsonExisted = false
		} else {
			return err
		}
	}

	if jsonExisted {
		content, err := os.ReadFile(jsonPath)
		if err != nil {
			log.Fatal("Error when opening file: ", err)
		}

		var jsonStory Story
		err = json.Unmarshal(content, &jsonStory)
		if err != nil {
			log.Fatal("Error during Unmarshal(): ", err)
		}

		if jsonStory.TranscriptEN != "" {
			story.TranscriptEN = jsonStory.TranscriptEN
		}
		if jsonStory.TranscriptJA != "" {
			story.TranscriptJA = jsonStory.TranscriptJA
		}
		if jsonStory.Content != "" {
			story.Content = jsonStory.Content
		}
	} else {
		err = writeStoryJson(jsonPath, story)
		if err != nil {
			return err
		}
	}

	epNumStr := strconv.Itoa(story.EpisodeNumber)

	if storyExists(story, sqldb) {
		fmt.Printf(`updating story: "%s"`+"\n", story.Title)

		_, err := sqldb.Exec(`UPDATE stories SET 
				date = $1, link = $2, episode_number = $3, video = $4, 
				content = $5, content_format = $6, 
				transcript_en = CASE WHEN transcript_en = '' THEN $7 ELSE transcript_en END,
				transcript_ja = CASE WHEN transcript_ja = '' THEN $8 ELSE transcript_ja END
				WHERE title = $10 and source = $11;`,
			story.Date, story.Link, epNumStr, story.Video,
			story.Content, story.ContentFormat, story.TranscriptEN,
			story.TranscriptJA, story.Title, story.Source)
		return err
	}

	fmt.Printf("importing story: %s, has %d new words \n", story.Title, newWordCount)

	initialExcerpt := `[ {"reps_todo": [], "reps_logged": [] }]`

	_, err = sqldb.Exec(`INSERT INTO stories (title, source, date, link, episode_number, video, 
				content, content_format, transcript_en, transcript_ja, words, start_time, end_time, excerpts, date_last_rep) 
				VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13);`,
		story.Title, story.Source, story.Date, story.Link, epNumStr,
		story.Video, story.Content, story.ContentFormat, story.TranscriptEN,
		story.TranscriptJA, initialExcerpt, 0)

	return err
}

func writeStoryJson(filePath string, story Story) error {
	jsonString, err := json.Marshal(&story)
	if err != nil {
		return err
	}

	err = os.WriteFile(filePath, jsonString, os.ModePerm)
	if err != nil {
		return err
	}

	return nil
}

func updateStoryJson(story Story) error {
	var jsonPath string = SOURCES_PATH + story.Source + "/" + story.Title + ".json"
	fmt.Println("jsonpPath", jsonPath)

	content, err := os.ReadFile(jsonPath)
	if err != nil {
		return err
	}

	var jsonStory Story
	err = json.Unmarshal(content, &jsonStory)
	if err != nil {
		return err
	}

	if story.TranscriptEN != "" {
		jsonStory.TranscriptEN = story.TranscriptEN
	}
	if story.TranscriptJA != "" {
		jsonStory.TranscriptJA = story.TranscriptJA
	}
	if story.Content != "" {
		jsonStory.Content = story.Content
	}

	err = writeStoryJson(jsonPath, jsonStory)
	if err != nil {
		return err
	}

	return nil
}

func processStoryWords(story Story, sqldb *sql.DB) (newWordCount int, wordIdsJson string, err error) {

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
