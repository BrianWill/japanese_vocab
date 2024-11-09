package main

import (
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
	"github.com/ikawaha/kagome/v2/tokenizer"
	_ "github.com/mattn/go-sqlite3"
	"go.mongodb.org/mongo-driver/bson"
)

const SOURCES_PATH = "../static/sources/"
const INITTIAL_EXCERPT = `[ {"reps_todo": 0, "reps_logged": [], "hash" : 1 }]`

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
			_, err := importSource(e.Name(), sqldb)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func importSource(sourceName string, sqldb *sql.DB) ([]string, error) {
	sourceDir := SOURCES_PATH + sourceName
	entries, err := os.ReadDir(sourceDir)
	if err != nil {
		return nil, err
	}

	storiesByTitle := make(map[string]Story)
	unusedPaths := make([]string, 0)

	// find stories
	for _, entry := range entries {
		fileName := entry.Name()
		components := strings.Split(fileName, ".")
		extension := components[len(components)-1]

		var title string

		if extension == "mp4" || extension == "m4a" || extension == "mp3" {
			title = strings.Join(components[:len(components)-1], ".")
			story := storiesByTitle[title]
			story.Title = title
			story.Source = sourceName
			story.Video = fileName

			storyBasePath := sourceDir + "/" + title

			story.SubtitlesJA, err = readWriteSubtitleFiles(storyBasePath, "ja")
			if err != nil {
				return nil, err
			}
			story.SubtitlesEN, err = readWriteSubtitleFiles(storyBasePath, "en")
			if err != nil {
				return nil, err
			}

			storiesByTitle[title] = story
		}
	}

	// gather unusedPaths
	for _, entry := range entries {
		fileName := entry.Name()
		components := strings.Split(fileName, ".")

		if len(components) > 1 {
			extension := components[len(components)-1]

			if extension == "mp4" || extension == "m4a" || extension == "mp3" || extension == "json" {
				title := strings.Join(components[:len(components)-1], ".")
				if _, ok := storiesByTitle[title]; ok {
					continue
				}
			}

			if extension == "vtt" || extension == "ass" || extension == "srt" {
				if len(components) > 2 {
					lang := components[len(components)-2]
					validLang := lang == "ja" || lang == "en"
					title := strings.Join(components[:len(components)-2], ".")
					if _, ok := storiesByTitle[title]; ok && validLang {
						continue
					}
				}
			}
		}

		unusedPaths = append(unusedPaths, fileName)
	}

	for _, s := range storiesByTitle {
		if s.Video == "" {
			continue
		}

		err = storeStory(s, sqldb)
		if err != nil {
			return nil, err
		}
	}

	return unusedPaths, nil
}

// pathExists returns the first path that exists; return "" if none match
func firstExtensionThatExists(path string, extensions []string) (string, error) {
	for _, extension := range extensions {
		_, err := os.Stat(path + extension)
		if err == nil {
			return extension, nil
		}
		if !os.IsNotExist(err) {
			return "", err // some kind of error other than the file not existing
		}
	}
	return "", nil // none of the paths exist, but no error occurred
}

func importStory(sourceName string, storyTitle string, sqldb *sql.DB) error {
	sourceDir := SOURCES_PATH + sourceName

	storyBasePath := sourceDir + "/" + storyTitle

	var story Story
	story.Title = storyTitle
	story.Source = sourceName

	// find video or audio
	var err error
	mediaExtension, err := firstExtensionThatExists(storyBasePath, []string{".mp4", ".m4a", ".mp3"})
	if err != nil {
		return err
	}
	if mediaExtension == "" {
		return fmt.Errorf("no audio or video file found for story: " + storyTitle + ", in source: " + sourceName)
	}

	story.Video = storyTitle + mediaExtension

	story.SubtitlesJA, err = readWriteSubtitleFiles(storyBasePath, "ja")
	if err != nil {
		return err
	}
	story.SubtitlesEN, err = readWriteSubtitleFiles(storyBasePath, "en")
	if err != nil {
		return err
	}

	err = storeStory(story, sqldb)
	if err != nil {
		return err
	}

	return nil
}

func tokenizeSubtitle(text string) ([]Word, error) {
	analyzerTokens := tok.Analyze(text, tokenizer.Normal)
	tokens := make([]*JpToken, len(analyzerTokens))

	for i, t := range analyzerTokens {
		features := t.Features()
		if len(features) < 9 {
			tokens[i] = &JpToken{
				Surface: t.Surface,
				POS:     features[0],
				POS_1:   features[1],
			}
		} else {
			tokens[i] = &JpToken{
				Surface:          t.Surface,
				POS:              features[0],
				POS_1:            features[1],
				POS_2:            features[2],
				POS_3:            features[3],
				InflectionalType: features[4],
				InflectionalForm: features[5],
				BaseForm:         features[6],
				Reading:          features[7],
				Pronunciation:    features[8],
			}
		}
	}

	words := make([]Word, 0)

	// group tokens into words
	for len(tokens) > 0 {
		word, numTokensConsumed := nextWord(tokens)
		words = append(words, word)
		tokens = tokens[numTokensConsumed:]
	}

	return words, nil
}

// expects tokens len to be > 0
// return next word and number of tokens consumedv
func nextWord(tokens []*JpToken) (Word, int) {
	t := tokens[0]
	word := Word{Display: t.Surface, BaseForm: t.BaseForm, POS: t.POS}

	if t.POS == "記号" { // is puncuation
		return nextPunctuation(tokens)
	} else if t.POS == "感動詞" { // interjection
		return word, 1
	} else if t.POS == "名詞" { // noun
		return word, 1
	} else if t.POS == "助詞" { // particle
		return nextParticle(tokens)
	} else if t.POS == "副詞" { // adverb
		return word, 1
	} else if t.POS == "動詞" { // verb
		return nextVerb(tokens)
	} else if t.POS == "フィラー" { // filler
		return word, 1
	} else if t.POS == "接頭詞" { // prefix
		return word, 1
	} else if t.POS == "連体詞" { // na adjective
		return word, 1
	} else if t.POS == "接続詞" { // conjunction
		return word, 1
	} else if t.POS == "形容詞" { // i-adjective
		return word, 1
	} else if t.POS == "助動詞" { // auxillary verb
		//if t.Surface == "です" || t.Surface == "だ" || t.Surface == "な" || t.Surface == "でしょ" || t.Surface == "う" {
		return word, 1
		//}
		//fmt.Println("Auxillary verb that doesn't follow a verb: ", t.POS)
		//panic("Auxillary verb that doesn't follow a verb: " + t.POS + " : " + t.POS_1 + " : " + t.Surface)
		//return Word{Display: t.Surface, BaseForm: t.BaseForm}, 1
	} else {
		panic("POS that is not currently accounted for: " + t.POS + " : " + t.Surface)
	}

	return Word{Display: t.Surface}, 1 // for types of tokens we haven't accounted for
}

// assumes that the first token is a punctuation mark
func nextPunctuation(tokens []*JpToken) (Word, int) {
	token := tokens[0]
	word := Word{Display: token.Surface}
	numTokensConsumed := 1

	for _, token := range tokens[1:] {
		if token.POS == "記号" {
			word.Display += token.Surface
			numTokensConsumed++
		} else {
			break
		}
	}

	return word, numTokensConsumed
}

// assumes that the first token is a particle
func nextParticle(tokens []*JpToken) (Word, int) {
	token := tokens[0]
	word := Word{Display: token.Surface, POS: "助詞"}
	numTokensConsumed := 1

	for _, token := range tokens[1:] {
		if token.POS == "助詞" {
			word.Display += token.Surface
			numTokensConsumed++
		} else {
			break
		}
	}

	word.BaseForm = word.Display
	return word, numTokensConsumed
}

// assumes that the first token is a verb
func nextVerb(tokens []*JpToken) (Word, int) {
	token := tokens[0]
	word := Word{Display: token.Surface, BaseForm: token.BaseForm}
	numTokensConsumed := 1

	for _, token := range tokens[1:] {
		if token.POS == "助動詞" || // auxillary verb
			(token.POS == "助詞" && token.POS_1 == "接続助詞") || // conjungtive particle て (maybe other things too?)
			(token.POS == "動詞" && token.POS_1 == "非自立") { // dependent verb e.g. てる
			word.Display += token.Surface
			numTokensConsumed++
		} else {
			break
		}
	}

	return word, numTokensConsumed
}

// Read original subtitle file, write subtitles to json file, and return the json.
func readWriteSubtitleFiles(storyBasePath string, lang string) (string, error) {
	subtitlesExtension, err := firstExtensionThatExists(storyBasePath,
		[]string{"." + lang + ".vtt", "." + lang + ".ass", "." + lang + ".srt"})
	if err != nil {
		return "", nil
	}

	subtitles := make([]Subtitle, 0)

	// read subtitle file if it exists (otherwise we'll write an empty json file)
	if subtitlesExtension != "" {
		subs, err := astisub.OpenFile(storyBasePath + subtitlesExtension)
		if err != nil {
			return "", err
		}

		var sb strings.Builder
		for _, item := range subs.Items {
			text := ""
			for _, line := range item.Lines {

				for _, lineItem := range line.Items {
					text += lineItem.Text
					sb.WriteString(lineItem.Text)
				}
				sb.WriteString("\n")
			}

			var words []Word = nil
			if lang == "ja" {

				words, err = tokenizeSubtitle(text)
				if err != nil {
					return "", err
				}

				// fmt.Println("LINE: ", text)
				// for _, word := range words {
				// 	fmt.Println("WORD: ", word.BaseForm, " : ", word.Display)
				// }
			}

			subtitles = append(subtitles, Subtitle{
				StartTime: float64(item.StartAt) / (1000 * 1000 * 1000), // convert from nanoseconds to seconds
				EndTime:   float64(item.EndAt) / (1000 * 1000 * 1000),
				Text:      text,
				Words:     words,
			})
		}
	}

	jsonBytes, err := json.MarshalIndent(subtitles, "", "    ")
	if err != nil {
		return "", err
	}

	// write out subtitle json
	err = os.WriteFile(storyBasePath+"."+lang+".json", jsonBytes, 0644)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func getSubtitlesContentInTimeRange(subtitlesJSON string, startTime float64, endTime float64) (string, error) {
	var subtitles []Subtitle
	err := json.Unmarshal([]byte(subtitlesJSON), &subtitles)
	if err != nil {
		return "", err
	}

	if endTime == 0 {
		endTime = math.MaxFloat64
	}

	var sb strings.Builder
	for _, subtitle := range subtitles {
		if endTime != 0 && (subtitle.EndTime < startTime || subtitle.StartTime > endTime) {
			continue
		}
		sb.WriteString(subtitle.Text)
		sb.WriteString("\n")
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

func storeStory(story Story, sqldb *sql.DB) error {
	newWordCount, _, err := processStoryWords(story, sqldb)
	if err != nil {
		return err
	}

	if storyExists(story, sqldb) {
		fmt.Printf(`updating story: "%s"`+"\n", story.Title)

		_, err := sqldb.Exec(`UPDATE stories SET 
				date = $1, link = $2, video = $3, content = $4,  
				subtitles_en = $5, subtitles_ja = $6
				WHERE title = $7 and source = $8;`,
			story.Date, story.Link, story.Video, story.Content,
			story.SubtitlesEN, story.SubtitlesJA,
			story.Title, story.Source)
		return err
	}

	fmt.Printf("importing story: %s, has %d new words \n", story.Title, newWordCount)

	_, err = sqldb.Exec(`INSERT INTO stories (title, source, date, link, video, 
				content, subtitles_en, subtitles_ja, excerpts, date_last_rep, has_reps_todo) 
				VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11);`,
		story.Title, story.Source, story.Date, story.Link,
		story.Video, story.Content, story.SubtitlesEN, story.SubtitlesJA, INITTIAL_EXCERPT, 0, 0)

	return err
}

func updateStorySubtitleFiles(story Story) error {
	path := SOURCES_PATH + story.Source + "/" + story.Title
	if story.SubtitlesEN != "" {
		err := os.WriteFile(path+".en.json", []byte(story.SubtitlesEN), os.ModePerm)
		if err != nil {
			return err
		}
	}
	if story.SubtitlesJA != "" {
		err := os.WriteFile(path+".ja.json", []byte(story.SubtitlesJA), os.ModePerm)
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

func ImportSource(w http.ResponseWriter, r *http.Request) {
	if importLock.TryLock() {
		defer importLock.Unlock()
	} else {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + `Aborting: an import is already in progress` + `"}`))
		return
	}

	dbPath := MAIN_USER_DB_PATH

	w.Header().Set("Content-Type", "application/json")

	sqldb, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}
	defer sqldb.Close()

	loadDictionary()

	var body ImportSourceRequest
	err = json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}

	_, err = importSource(body.Source, sqldb)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}

	json.NewEncoder(w).Encode(bson.M{"message": "imported source: " + body.Source})
}

func ImportStory(w http.ResponseWriter, r *http.Request) {
	if importLock.TryLock() {
		defer importLock.Unlock()
	} else {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + `Aborting: an import is already in progress` + `"}`))
		return
	}

	dbPath := MAIN_USER_DB_PATH

	w.Header().Set("Content-Type", "application/json")

	sqldb, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}
	defer sqldb.Close()

	loadDictionary()

	var body ImportStoryRequest
	err = json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}

	err = importStory(body.Source, body.StoryTitle, sqldb)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}

	json.NewEncoder(w).Encode(bson.M{"message": "imported story: " + body.StoryTitle + ", from source: " + body.Source})
}
