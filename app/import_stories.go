package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/asticode/go-astisub"
	_ "github.com/mattn/go-sqlite3"
	"go.mongodb.org/mongo-driver/bson"
)

const SOURCES_PATH = "../static/sources/"

var newlineRegEx *regexp.Regexp

func scanSourceDirForStoryFiles(sourceName string) ([]StoryFilePaths, []string, error) {
	entries, err := os.ReadDir(SOURCES_PATH + sourceName)
	if err != nil {
		return nil, nil, err
	}

	malformedPaths := make([]string, 0)
	pathsByTitle := make(map[string]StoryFilePaths, 0)

	for _, entry := range entries {
		path := entry.Name()
		components := strings.Split(path, ".")

		extension := components[len(components)-1]

		var title string
		var paths StoryFilePaths

		if extension == "mp4" || extension == "m4a" {
			paths.Video = path
			if len(components) < 2 {
				malformedPaths = append(malformedPaths, sourceName+"/"+path)
				continue
			}

			title = strings.Join(components[:len(components)-1], ".")
			paths = pathsByTitle[title]

		} else if extension == "vtt" || extension == "ass" || extension == "srt" {
			if len(components) < 3 {
				malformedPaths = append(malformedPaths, sourceName+"/"+path)
				continue
			}
			title = strings.Join(components[:len(components)-2], ".")
			paths = pathsByTitle[title]

			lang := components[1]

			if lang == "en" {
				paths.EnglishSubtitles = path
			} else if lang == "ja" {
				paths.JapaneseSubtitles = path
			} else {
				malformedPaths = append(malformedPaths, sourceName+"/"+path)
				continue
			}
		} else {
			//malformedPaths = append(malformedPaths, sourceName+"/"+path)
			continue
		}

		pathsByTitle[title] = paths
	}

	storyFilePaths := make([]StoryFilePaths, 0)

	for _, paths := range pathsByTitle {
		storyFilePaths = append(storyFilePaths, paths)
	}

	return storyFilePaths, malformedPaths, nil
}

// get the paths for all story files in all source dirs
func scanSources() (map[string][]StoryFilePaths, []string, error) {
	sourceMap := make(map[string][]StoryFilePaths)

	entries, err := os.ReadDir(SOURCES_PATH)
	if err != nil {
		return nil, nil, err
	}

	malformedPaths := make([]string, 0)

	for _, e := range entries {
		if e.IsDir() {
			storyFilePaths, malPaths, err := scanSourceDirForStoryFiles(e.Name())
			malformedPaths = append(malformedPaths, malPaths...)
			sourceMap[e.Name()] = storyFilePaths
			if err != nil {
				return nil, nil, err
			}
		}
	}

	return sourceMap, malformedPaths, nil
}

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
			err = importSource(e.Name(), sqldb)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func importSource(source string, sqldb *sql.DB) error {
	sourceDir := SOURCES_PATH + source
	entries, err := os.ReadDir(sourceDir)
	if err != nil {
		log.Fatal(err)
	}

	storiesByName := make(map[string]*Story)

	for _, entry := range entries {
		name := entry.Name()
		components := strings.Split(name, ".")
		if len(components) < 2 {
			return fmt.Errorf("malformed file name in source: %s", name)
		}

		title := components[0]
		extension := components[len(components)-1]
		isVideo := extension == "mp4"
		isSubtitle := extension == "vtt" || extension == "ass" || extension == "srt"
		if !isVideo && !isSubtitle {
			continue
		}

		story, ok := storiesByName[title]
		if !ok {
			story = &Story{}
			storiesByName[title] = story
		}
		story.Title = title
		story.Source = source

		if isVideo {
			story.Video = name
		}

		if isSubtitle {
			path := sourceDir + "/" + name

			if len(components) != 3 {
				return fmt.Errorf("subtitle file does not specify language")
			}
			lang := components[len(components)-3]
			switch lang {
			case "en":
				story.TranscriptEN, story.Content, err = getSubtitles(path)
				if err != nil {
					return err
				}

				if extension == "vtt" {
					err := os.WriteFile(path, []byte(story.TranscriptEN), os.ModePerm)
					if err != nil {
						return err
					}
				}
			case "ja":
				story.TranscriptJA, story.Content, err = getSubtitles(path)
				if err != nil {
					return err
				}

				if extension == "vtt" {
					err := os.WriteFile(path, []byte(story.TranscriptJA), os.ModePerm)
					if err != nil {
						return err
					}
				}
			default:
				return fmt.Errorf("subtitle file language is invalid: %s", name)
			}

			// delete .ass or .srt files
			if extension != "vtt" {
				err = os.Remove(path)
				if err != nil {
					return err
				}
			}
		}
	}

	for _, s := range storiesByName {
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
		if item.EndAt.Seconds() < startTime || item.StartAt.Seconds() > endTime {
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

	if storyExists(story, sqldb) {
		fmt.Printf(`updating story: "%s"`+"\n", story.Title)

		_, err := sqldb.Exec(`UPDATE stories SET 
				date = $1, link = $2, video = $3, content = $4,  
				transcript_en = CASE WHEN transcript_en = '' THEN $5 ELSE transcript_en END,
				transcript_ja = CASE WHEN transcript_ja = '' THEN $6 ELSE transcript_ja END
				WHERE title = $7 and source = $8;`,
			story.Date, story.Link, story.Video,
			story.Content, story.TranscriptEN,
			story.TranscriptJA, story.Title, story.Source)
		return err
	}

	fmt.Printf("importing story: %s, has %d new words \n", story.Title, newWordCount)

	initialExcerpt := `[ {"reps_todo": [], "reps_logged": [] }]`

	_, err = sqldb.Exec(`INSERT INTO stories (title, source, date, link, episode_number, video, 
				content, transcript_en, transcript_ja, words, start_time, end_time, excerpts, date_last_rep, has_reps_todo) 
				VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11);`,
		story.Title, story.Source, story.Date, story.Link,
		story.Video, story.Content, story.TranscriptEN,
		story.TranscriptJA, initialExcerpt, 0, 0)

	return err
}

func updateStorySubtitleFiles(story Story) error {
	path := SOURCES_PATH + story.Source + "/" + story.Title
	if story.TranscriptEN != "" {
		err := os.WriteFile(path+".en.vtt", []byte(story.TranscriptEN), os.ModePerm)
		if err != nil {
			return err
		}
	}
	if story.TranscriptJA != "" {
		err := os.WriteFile(path+".ja.vtt", []byte(story.TranscriptJA), os.ModePerm)
		if err != nil {
			return err
		}
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

func GetSources(response http.ResponseWriter, request *http.Request) {
	dbPath := MAIN_USER_DB_PATH

	response.Header().Set("Content-Type", "application/json")

	sqldb, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}
	defer sqldb.Close()

	ips, err := GetOutboundIP()
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + "failure to get ip: " + err.Error() + `"}`))
		fmt.Println(err)
		return
	}

	for _, ip := range ips {
		fmt.Println("ip: ", ip)
	}

	rows, err := sqldb.Query(`SELECT id, title, source, link, video, 
			date, date_last_rep, has_reps_todo FROM stories;`)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + "failure to get story: " + err.Error() + `"}`))
		return
	}
	defer rows.Close()

	var stories []Story
	for rows.Next() {
		var story Story
		if err := rows.Scan(&story.ID, &story.Title, &story.Source, &story.Link,
			&story.Video, &story.Date, &story.DateLastRep, &story.HasRepsTodo); err != nil {
			response.WriteHeader(http.StatusInternalServerError)
			response.Write([]byte(`{ "message": "` + "failure to read story list: " + err.Error() + `"}`))
			return
		}
		stories = append(stories, story)
	}

	storiesBySource := make(map[string][]Story)
	for _, s := range stories {
		if _, ok := storiesBySource[s.Source]; !ok {
			storiesBySource[s.Source] = make([]Story, 0)
		}
		sourceStories := storiesBySource[s.Source]
		storiesBySource[s.Source] = append(sourceStories, s)
	}

	storyFilePathsBySource, malformedPaths, err := scanSources()
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + "failure to scan sources: " + err.Error() + `"}`))
		return
	}

	json.NewEncoder(response).Encode(bson.M{"storiesBySource": storiesBySource,
		"storyFilePathsBySource": storyFilePathsBySource, "malformedPaths": malformedPaths})
}
