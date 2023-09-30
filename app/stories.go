package main

import (
	"compress/gzip"
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
	//"unicode/utf8"

	"github.com/gorilla/mux"
	"github.com/ikawaha/kagome/v2/tokenizer"
	_ "github.com/mattn/go-sqlite3"
	"go.mongodb.org/mongo-driver/bson"
)

const INITIAL_RANK = 1
const INITIAL_STORY_COUNTDOWN = 3

const DRILL_FILTER_ON_COOLDOWN = "on"
const DRILL_FILTER_OFF_COOLDOWN = "off"
const DRILL_FILTER_ALL = "all"

const STORY_STATUS_CURRENT = 2
const STORY_STATUS_NEVER_READ = 1
const STORY_STATUS_ARCHIVE = 0
const STORY_INITIAL_STATUS = STORY_STATUS_NEVER_READ

const STORY_LOG_COOLDOWN = 60 * 60 * 8 // 8 hour cooldown (in seconds)

func CreateStory(response http.ResponseWriter, request *http.Request) {
	dbPath, redirect, err := GetUserDb(response, request)
	if redirect || err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}

	response.Header().Set("Content-Type", "application/json")

	var story Story
	json.NewDecoder(request.Body).Decode(&story)

	sqldb, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `"}`))
	}
	defer sqldb.Close()

	_, newWordCount, err := addStory(story, sqldb, false)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}
	fmt.Println("total new words added:", newWordCount)
	json.NewEncoder(response).Encode("Success adding story")
}

func RetokenizeStory(response http.ResponseWriter, request *http.Request) {
	dbPath, redirect, err := GetUserDb(response, request)
	if redirect || err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}

	response.Header().Set("Content-Type", "application/json")

	var story Story
	json.NewDecoder(request.Body).Decode(&story)

	sqldb, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `"}`))
	}
	defer sqldb.Close()

	_, newWordCount, err := addStory(story, sqldb, true)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}
	fmt.Println("total new words added:", newWordCount)
	json.NewEncoder(response).Encode("Success retokenizing story")
}

func tokenize(content string) ([]*JpToken, []string, error) {
	analyzerTokens := tok.Analyze(content, tokenizer.Normal)
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
		if tokens[i].BaseForm == "" {
			tokens[i].BaseForm = tokens[i].Surface
		}
	}

	kanji, err := extractKanji(tokens)
	if err != nil {
		return nil, nil, fmt.Errorf(`failure to extract kanji` + err.Error())
	}

	return tokens, kanji, nil
}

func addStory(story Story, sqldb *sql.DB, retokenize bool) (id int64, newWordCount int, err error) {
	if retokenize {
		var linesJSON string
		row := sqldb.QueryRow(`SELECT title, link, lines, date_added, audio 
			FROM stories WHERE id = $1;`, story.ID)
		if err := row.Scan(&story.Title, &story.Link, &linesJSON, &story.DateAdded, &story.Audio); err != nil {
			return 0, 0, fmt.Errorf("failure to read story: " + err.Error())
		}
		err := json.Unmarshal([]byte(linesJSON), &story.Lines)
		if err != nil {
			return 0, 0, fmt.Errorf("failure to unmarshall story lines: " + err.Error())
		}
		for _, line := range story.Lines {
			story.Content += line.Timestamp + "\n"
			for _, word := range line.Words {
				story.Content += word.Surface
			}
		}
	} else {
		row := sqldb.QueryRow(`SELECT id FROM stories WHERE title = $1;`, story.Title)
		if err := row.Scan(&story.ID); err != nil && err != sql.ErrNoRows {
			return 0, 0, fmt.Errorf("story with same title already exists: " + err.Error())
		}
	}

	// if text has timestamps, split on timestamps,
	// otherwise split on blank lines
	timestampRegex := regexp.MustCompile(`(?m)^\s*\d*:\d*\s*$`) // match timestamp line
	timestamps := timestampRegex.FindAllString(story.Content, -1)
	lineContents := timestampRegex.Split(story.Content, -1)

	if len(timestamps) > 0 {
		// todo: check that the timestamps increase in value
		lineContents = lineContents[1:]
	} else {
		blanklinesRegex := regexp.MustCompile(`(?m)^\s*$\n`) // match timestamp line
		lineContents = blanklinesRegex.Split(story.Content, -1)
		if len(lineContents) > 0 && lineContents[0] == "" {
			lineContents = lineContents[1:]
		}
	}

	lines := make([]Line, len(lineContents))

	newWordCount = 0

	for i, content := range lineContents {
		timestamp := "0:00"
		if len(timestamps) > 0 {
			timestamp = timestamps[i]
		}
		timestamp = strings.TrimSpace(timestamp)
		content = strings.TrimSpace(content)

		//fmt.Println(timestamp, content)

		tokens, kanjiSet, err := tokenize(content)
		if err != nil {
			return 0, 0, fmt.Errorf("failure to tokenize story: " + err.Error())
		}

		wordsOfLine, lineKanji, addedWordCount, err := addWords(tokens, kanjiSet, sqldb)
		if err != nil {
			return 0, 0, fmt.Errorf("failure to add words: " + err.Error())
		}
		newWordCount += addedWordCount

		lines[i] = Line{
			Timestamp: timestamp,
			Words:     wordsOfLine,
			Kanji:     lineKanji,
		}
	}

	linesJson, err := json.Marshal(lines)
	if err != nil {
		return 0, 0, fmt.Errorf("failure to lines: " + err.Error())
	}

	if retokenize {
		_, err = sqldb.Exec(`UPDATE stories SET lines = $1 WHERE id = $2;`,
			linesJson, story.ID)
		if err != nil {
			return 0, 0, fmt.Errorf("failure to update story: " + err.Error())
		}
		return story.ID, newWordCount, nil
	} else {
		date := time.Now().Unix()
		result, err := sqldb.Exec(`INSERT INTO stories (lines, title, link, audio, date_added, status, countdown, read_count, date_last_read) 
				VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9);`,
			linesJson, story.Title, story.Link, story.Audio, date, STORY_INITIAL_STATUS, INITIAL_STORY_COUNTDOWN, 0, 0)
		if err != nil {
			return 0, 0, fmt.Errorf("failure to insert story: " + err.Error())
		}
		id, err = result.LastInsertId()
		if err != nil {
			return 0, 0, fmt.Errorf("failure to insert story: " + err.Error())
		}
		return id, newWordCount, nil
	}
}

func getTokenPOS(token *JpToken, priorToken *JpToken) string {
	if token.Surface == "。" {
		return ""
	} else if token.Surface == "\n\n" {
		return ""
	} else if token.Surface == "\n" {
		return ""
	} else if token.Surface == " " {
		return ""
	}

	if (token.POS == "動詞" && token.POS_1 == "接尾") ||
		(token.POS == "助動詞") ||
		(token.Surface == "で" && token.POS == "助詞" && token.POS_1 == "接続助詞") ||
		(token.Surface == "て" && token.POS == "助詞" && token.POS_1 == "接続助詞") ||
		(token.Surface == "じゃ" && token.POS == "助詞" && token.POS_1 == "副助詞") ||
		(token.Surface == "し" && token.POS == "動詞" && token.POS_1 == "自立") { // auxilliary verb
		return "verb_auxiliary"
	} else if token.POS == "動詞" && token.POS_1 == "非自立" { // auxilliary verb
		return "verb_auxiliary"
	} else if (token.POS == "助詞" && token.POS_1 == "格助詞") || // case particle
		(token.POS == "助詞" && token.POS_1 == "接続助詞") || // conjunction particle
		(token.POS == "助詞" && token.POS_1 == "係助詞") || // binding particle (も　は)
		(token.POS == "助詞" && token.POS_1 == "副助詞") { // auxiliary particle
		return "particle"
	} else if token.POS == "副詞" {
		return "adverb"
	} else if token.POS == "接続詞" && token.POS_1 == "*" { // conjunction
		return "conjunction"
	} else if (token.POS == "助詞" && token.POS_1 == "連体化") || // connecting particle　(の)
		(token.POS == "助詞" && token.POS_1 == "並立助詞") { // connecting particle (や)
		return "connecting_particle"
	} else if token.POS == "形容詞" { // i-adj
		return "i_adjective pad_left"
	} else if token.POS == "名詞" && token.POS_1 == "代名詞" { // pronoun
		return "pronoun pad_left"
	} else if token.POS == "連体詞" { // adnominal adjective
		return "admoninal_adjective pad_left"
	} else if token.POS == "動詞" { //　verb
		return "verb pad_left"
	} else if token.POS == "名詞" && token.POS_1 == "接尾" { // noun suffix
		return "noun"
	} else if (priorToken.POS == "助詞" && (priorToken.POS_1 == "連体化" || priorToken.POS_1 == "並立助詞")) || // preceded by connective particle
		(priorToken.POS == "接頭詞" && priorToken.POS_1 == "名詞接続") { // preceded by prefix
		return "noun"
	} else if token.POS == "名詞" { // noun
		return "noun"
	} else if token.POS == "記号" { // symbol
		return ""
	} else if token.POS == "号" { // counter
		return "counter"
	} else {
		return ""
	}
}

func extractKanji(tokens []*JpToken) ([]string, error) {
	kanjiMap := make(map[string]bool)

	for _, t := range tokens {
		var re = regexp.MustCompile(`[\x{4E00}-\x{9FAF}]`)
		kanji := re.FindAllString(t.Surface+t.BaseForm, -1)

		for _, s := range kanji {
			kanjiMap[s] = true
		}
	}

	kanji := make([]string, len(kanjiMap))

	i := 0
	for k := range kanjiMap {
		kanji[i] = k
		i++
	}

	return kanji, nil
}

func addWords(tokens []*JpToken, kanjiSet []string, sqldb *sql.DB) ([]LineWord, []LineKanji, int, error) {
	var reHasKanji = regexp.MustCompile(`[\x{4E00}-\x{9FAF}]`)
	var reHasKatakana = regexp.MustCompile(`[ア-ン]`)
	var reHasKana = regexp.MustCompile(`[ア-ンァ-ヴぁ-ゔ]`)

	newWordCount := 0
	unixtime := time.Now().Unix()

	lineWords := make([]LineWord, len(tokens))
	for i, token := range tokens {
		priorToken := &JpToken{}
		if i > 0 {
			priorToken = tokens[i-1]
		}

		lineWord := &lineWords[i]
		lineWord.Surface = token.Surface
		lineWord.BaseForm = token.BaseForm
		lineWord.POS = getTokenPOS(token, priorToken)

		category := 0

		if lineWord.POS == "" { // not a vocab word
			continue
		}

		hasKatakana := len(reHasKatakana.FindStringIndex(token.BaseForm)) > 0
		hasKana := len(reHasKana.FindStringIndex(token.BaseForm)) > 0
		hasKanji := len(reHasKanji.FindStringIndex(token.BaseForm)) > 0

		if !hasKana && !hasKanji {
			continue
		}

		// has katakana
		if hasKatakana {
			category |= DRILL_CATEGORY_KATAKANA
		}

		entries := getDefinitions(token.BaseForm)
		for _, entry := range entries {
			for _, sense := range entry.Senses {
				category |= getVerbCategory(sense)
			}
		}

		lineWord.Category = category

		var id int64
		err := sqldb.QueryRow(`SELECT id FROM words WHERE base_form = $1`, token.BaseForm).Scan(&id)
		if err != nil && err != sql.ErrNoRows {
			return nil, nil, 0, err
		}
		if err == nil {
			lineWord.ID = id // word already exists
			continue
		}

		insertResult, err := sqldb.Exec(`INSERT INTO words (base_form, date_marked,
			date_added, category, rank, drill_count) 
			VALUES($1, $2, $3, $4, $5, $6);`,
			lineWord.BaseForm, 0, unixtime, lineWord.Category, INITIAL_RANK, 0)
		if err != nil {
			return nil, nil, 0, fmt.Errorf("failure to insert word: " + err.Error())
		}

		id, err = insertResult.LastInsertId()
		if err != nil {
			return nil, nil, 0, fmt.Errorf("failure to get id of inserted word: " + err.Error())
		}
		newWordCount++
		lineWord.ID = id
	}

	kanjiToInsert := make(map[string]*LineKanji)

	lineKanji := make([]LineKanji, len(kanjiSet))
	for i, kanji := range kanjiSet {
		if _, ok := kanjiToInsert[kanji]; ok {
			continue // early out
		}

		lk := &lineKanji[i]
		lk.Character = kanji

		var id int64
		err := sqldb.QueryRow(`SELECT id FROM words WHERE base_form = $1;`, lk.Character).Scan(&id)
		if err != nil && err != sql.ErrNoRows {
			return nil, nil, 0, err
		}
		if err == nil {
			lk.ID = id // kanji already exists
			continue
		}

		kanjiToInsert[kanji] = lk
	}

	for _, lk := range kanjiToInsert {
		insertResult, err := sqldb.Exec(`INSERT INTO words (base_form, date_marked,
			date_added, category, rank, drill_count) 
			VALUES($1, $2, $3, $4, $5, $6);`,
			lk.Character, 0, unixtime, DRILL_CATEGORY_KANJI, INITIAL_RANK, 0)
		if err != nil {
			return nil, nil, 0, fmt.Errorf("failure to insert kanji: " + err.Error())
		}

		id, err := insertResult.LastInsertId()
		if err != nil {
			return nil, nil, 0, fmt.Errorf("failure to get id of inserted kanji: " + err.Error())
		}
		newWordCount++
		lk.ID = id
	}

	return lineWords, lineKanji, newWordCount, nil
}

func getDefinitions(baseForm string) []JMDictEntry {
	if entries, ok := definitionsCache[baseForm]; ok {
		return entries
	}

	entries := make([]JMDictEntry, 0)

	hasKanji := len(reHasKanji.FindStringIndex(baseForm)) > 0
	if hasKanji {
		for _, e := range allEntriesByKanjiSpellings[baseForm] {
			entries = append(entries, *e)
		}
	} else {
		for _, e := range allEntriesByReading[baseForm] {
			entries = append(entries, *e)
		}
	}

	//fmt.Println("get definitions", baseForm, len(entries))

	definitionsCache[baseForm] = entries
	return entries
}

// func getDefinitionsJSON(baseForm string) (string, error) {
// 	if json, ok := definitionsJSONCache[baseForm]; ok {
// 		return json, nil
// 	}

// 	entries := getDefinitions(baseForm)

// 	entriesJSON, err := json.Marshal(entries)
// 	if err != nil {
// 		return "", fmt.Errorf("failure marshalling definitions for word: %s", baseForm)
// 	}

// 	definitionsJSONCache[baseForm] = string(entriesJSON)
// 	return string(entriesJSON), nil
// }

func getVerbCategory(sense JMDictSense) int {
	category := 0
	for _, pos := range sense.Pos {
		switch pos {
		case "verb-ichidan":
			category |= DRILL_CATEGORY_ICHIDAN
		case "verb-godan-su":
			category |= DRILL_CATEGORY_GODAN_SU
		case "verb-godan-ku":
			category |= DRILL_CATEGORY_GODAN_KU
		case "verb-godan-gu":
			category |= DRILL_CATEGORY_GODAN_GU
		case "verb-godan-ru":
			category |= DRILL_CATEGORY_GODAN_RU
		case "verb-godan-u":
			category |= DRILL_CATEGORY_GODAN_U
		case "verb-godan-tsu":
			category |= DRILL_CATEGORY_GODAN_TSU
		case "verb-godan-mu":
			category |= DRILL_CATEGORY_GODAN_MU
		case "verb-godan-nu":
			category |= DRILL_CATEGORY_GODAN_NU
		case "verb-godan-bu":
			category |= DRILL_CATEGORY_GODAN_BU
		}
	}
	return category
}

func GetStoriesList(response http.ResponseWriter, request *http.Request) {
	dbPath, redirect, err := GetUserDb(response, request)
	if redirect {
		return
	}
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `"}`))
	}

	response.Header().Set("Content-Type", "application/json")

	sqldb, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}
	defer sqldb.Close()

	rows, err := sqldb.Query(`SELECT id, title, link, status, date_added, countdown, read_count, date_last_read FROM stories;`)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + "failure to get story: " + err.Error() + `"}`))
		return
	}
	defer rows.Close()

	var stories []Story
	for rows.Next() {
		var story Story
		if err := rows.Scan(&story.ID, &story.Title, &story.Link, &story.Status,
			&story.DateAdded, &story.Countdown, &story.ReadCount, &story.DateLastRead); err != nil {
			response.WriteHeader(http.StatusInternalServerError)
			response.Write([]byte(`{ "message": "` + "failure to read story list: " + err.Error() + `"}`))
			return
		}
		stories = append(stories, story)
	}

	json.NewEncoder(response).Encode(stories)
}

func GetStory(w http.ResponseWriter, r *http.Request) {
	dbPath, redirect, err := GetUserDb(w, r)
	if redirect || err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("content-encoding", "gzip")

	gw := gzip.NewWriter(w)
	defer gw.Close()

	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		gw.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}

	sqldb, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		gw.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}
	defer sqldb.Close()

	story, err := getStory(int64(id), sqldb)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		gw.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}

	err = json.NewEncoder(gw).Encode(story)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		gw.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}
}

func getStory(id int64, sqldb *sql.DB) (Story, error) {
	row := sqldb.QueryRow(`SELECT title, link, lines, date_added, audio, countdown, read_count, date_last_read FROM stories WHERE id = $1;`, id)

	var linesJSON string
	story := Story{ID: id}
	if err := row.Scan(&story.Title, &story.Link, &linesJSON, &story.DateAdded,
		&story.Audio, &story.Countdown, &story.ReadCount, &story.DateLastRead); err != nil {
		return Story{}, fmt.Errorf("failure to scan story row: " + err.Error())
	}

	err := json.Unmarshal([]byte(linesJSON), &story.Lines)
	if err != nil {
		return Story{}, fmt.Errorf("failure to unmarshall story lines: " + err.Error())
	}

	wordInfo := make(map[string]WordInfo)

	for _, line := range story.Lines {
		for _, word := range line.Words {
			wordInfo[word.BaseForm] = WordInfo{
				Definitions: getDefinitions(word.BaseForm),
			}
		}
	}

	story.WordInfo = wordInfo

	for baseForm := range story.WordInfo {
		wordInfo := story.WordInfo[baseForm]

		row := sqldb.QueryRow(`SELECT rank, date_marked FROM words WHERE base_form = $1;`, baseForm)

		err = row.Scan(&wordInfo.Rank, &wordInfo.DateMarked)
		if err != nil && err != sql.ErrNoRows {
			return Story{}, fmt.Errorf("failure to get word info: " + err.Error())
		}

		story.WordInfo[baseForm] = wordInfo
	}

	return story, nil
}

func UpdateStoryCounts(w http.ResponseWriter, r *http.Request) {
	dbPath, redirect, err := GetUserDb(w, r)
	if redirect || err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}

	w.Header().Set("Content-Type", "application/json")

	var story Story
	err = json.NewDecoder(r.Body).Decode(&story)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}

	sqldb, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}
	defer sqldb.Close()

	// make sure the story actually exists
	rows, err := sqldb.Query(`SELECT id FROM stories WHERE id = $1;`, story.ID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + "failure to get story: " + err.Error() + `"}`))
		return
	}

	if !rows.Next() {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + "story with ID does not exist: " + strconv.FormatInt(story.ID, 10) + `"}`))
		rows.Close()
		return
	}
	rows.Close()

	_, err = sqldb.Exec(`UPDATE stories SET countdown = $1, read_count = $2, date_last_read = $3 WHERE id = $4;`,
		story.Countdown, story.ReadCount, story.DateLastRead, story.ID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + "failure to update story: " + err.Error() + `"}`))
		return
	}

	json.NewEncoder(w).Encode(bson.M{"status": "success"})
}

// remove a line and combine its content with the previous line
func ConsolidateLine(w http.ResponseWriter, r *http.Request) {
	dbPath, redirect, err := GetUserDb(w, r)
	if redirect || err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}

	sqldb, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}
	defer sqldb.Close()

	w.Header().Set("Content-Type", "application/json")

	var consolidateLine ConsolidateLineRequest
	err = json.NewDecoder(r.Body).Decode(&consolidateLine)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}

	if consolidateLine.LineToRemove < 1 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{ "message: line to conslidate must have index of 1 or greater}`))
		return
	}

	row := sqldb.QueryRow(`SELECT lines FROM stories WHERE id = $1;`, consolidateLine.StoryID)

	var linesJSON string
	if err := row.Scan(&linesJSON); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message: failure to get story lines": "` + err.Error() + `"}`))
		return
	}

	var lines []Line
	err = json.Unmarshal([]byte(linesJSON), &lines)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message: failure to unmarhsal story lines JSON": "` + err.Error() + `"}`))
		return
	}

	if consolidateLine.LineToRemove >= len(lines) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{ "message: line to conslidate must have index of 1 or greater}`))
		return
	}

	idx := consolidateLine.LineToRemove
	prevLine := &lines[idx-1]
	line := lines[idx]

	//prevLine.Content += line.Content
	prevLine.Words = append(prevLine.Words, line.Words...)

	kanjiMap := make(map[string]LineKanji)
	for _, v := range prevLine.Kanji {
		kanjiMap[v.Character] = v
	}
	for _, v := range line.Kanji {
		kanjiMap[v.Character] = v
	}

	prevLine.Kanji = make([]LineKanji, len(kanjiMap))
	i := 0
	for _, v := range kanjiMap {
		prevLine.Kanji[i] = v
		i++
	}

	// remove the line
	lines = append(lines[:idx], lines[idx+1:]...)

	linesBytes, err := json.Marshal(lines)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message: failure to JSONify story lines": "` + err.Error() + `"}`))
		return
	}

	_, err = sqldb.Exec(`UPDATE stories SET lines = $1 WHERE id = $2;`, linesBytes, consolidateLine.StoryID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + "failure to update lines in story: " + err.Error() + `"}`))
		return
	}

	w.Write(linesBytes)
}

func SplitLine(w http.ResponseWriter, r *http.Request) {
	dbPath, redirect, err := GetUserDb(w, r)
	if redirect || err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}

	sqldb, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}
	defer sqldb.Close()

	w.Header().Set("Content-Type", "application/json")

	var splitLine SplitLineRequest
	err = json.NewDecoder(r.Body).Decode(&splitLine)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}

	row := sqldb.QueryRow(`SELECT lines FROM stories WHERE id = $1;`, splitLine.StoryID)

	var linesJSON string
	if err := row.Scan(&linesJSON); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message: failure to get story lines": "` + err.Error() + `"}`))
		return
	}

	var lines []Line
	err = json.Unmarshal([]byte(linesJSON), &lines)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message: failure to unmarhsal story lines JSON": "` + err.Error() + `"}`))
		return
	}

	if splitLine.LineToSplit < 0 || splitLine.LineToSplit >= len(lines) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{ "message: line to split must have index of 1 or greater}`))
		return
	}

	idx := splitLine.LineToSplit
	origLine := &lines[idx]

	if splitLine.WordIdx < 1 || splitLine.WordIdx >= len(origLine.Words) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{ "message: line to split word index must be in range}`))
		return
	}

	timestamp := secondsToTimestamp(splitLine.Timestamp)
	if splitLine.Timestamp == 0 {
		timestamp = origLine.Timestamp
	}

	newLine := Line{
		Words:     origLine.Words[splitLine.WordIdx:],
		Timestamp: timestamp,
	}
	origLine.Words = origLine.Words[:splitLine.WordIdx]

	kanjiMap := make(map[string]LineKanji)
	for _, v := range origLine.Kanji {
		kanjiMap[v.Character] = v
	}
	for _, v := range newLine.Kanji {
		kanjiMap[v.Character] = v
	}

	origKanjiMap := make(map[string]bool)
	for _, word := range origLine.Words {
		for _, rune := range word.Surface {
			s := string(rune)
			if _, ok := kanjiMap[s]; ok {
				origKanjiMap[string(rune)] = true
			}
		}
	}
	newKanjiMap := make(map[string]bool)
	for _, word := range newLine.Words {
		for _, rune := range word.Surface {
			s := string(rune)
			if _, ok := kanjiMap[s]; ok {
				newKanjiMap[string(rune)] = true
			}
		}
	}

	origLine.Kanji = make([]LineKanji, 0)
	for ch := range origKanjiMap {
		origLine.Kanji = append(origLine.Kanji, kanjiMap[ch])
	}
	newLine.Kanji = make([]LineKanji, 0)
	for ch := range newKanjiMap {
		newLine.Kanji = append(newLine.Kanji, kanjiMap[ch])
	}

	// insert the line
	lines = append(lines[:idx+1], lines[idx:]...)
	lines[idx+1] = newLine

	linesBytes, err := json.Marshal(lines)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message: failure to JSONify story lines": "` + err.Error() + `"}`))
		return
	}

	_, err = sqldb.Exec(`UPDATE stories SET lines = $1 WHERE id = $2;`, linesBytes, splitLine.StoryID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + "failure to update lines in story: " + err.Error() + `"}`))
		return
	}

	w.Write(linesBytes)
}

func SetTimestamp(w http.ResponseWriter, r *http.Request) {
	dbPath, redirect, err := GetUserDb(w, r)
	if redirect || err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}

	sqldb, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}
	defer sqldb.Close()

	w.Header().Set("Content-Type", "application/json")

	var setTimestamp SetTimestampRequest
	err = json.NewDecoder(r.Body).Decode(&setTimestamp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}

	row := sqldb.QueryRow(`SELECT lines FROM stories WHERE id = $1;`, setTimestamp.StoryID)

	var linesJSON string
	if err := row.Scan(&linesJSON); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message: failure to get story lines": "` + err.Error() + `"}`))
		return
	}

	var lines []Line
	err = json.Unmarshal([]byte(linesJSON), &lines)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message: failure to unmarshal story lines JSON": "` + err.Error() + `"}`))
		return
	}

	if setTimestamp.LineIdx < 0 || setTimestamp.LineIdx >= len(lines) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{ "message: line to set timestamp must have index of 1 or greater}`))
		return
	}

	line := &lines[setTimestamp.LineIdx]

	if setTimestamp.Timestamp < 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{ "message: cannot set timestamp to value less than zero}`))
		return
	}

	line.Timestamp = secondsToTimestamp(setTimestamp.Timestamp)

	linesBytes, err := json.Marshal(lines)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message: failure to JSONify story lines": "` + err.Error() + `"}`))
		return
	}

	_, err = sqldb.Exec(`UPDATE stories SET lines = $1 WHERE id = $2;`, linesBytes, setTimestamp.StoryID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + "failure to update lines in story: " + err.Error() + `"}`))
		return
	}

	w.Write(linesBytes)
}

func secondsToTimestamp(seconds float64) string {
	minutes := math.Floor(seconds / 60)
	seconds -= minutes * 60
	wholeSeconds := math.Floor(seconds)
	fracSeconds := seconds - wholeSeconds
	s := fmt.Sprintf("%d:%02d", int(minutes), int(seconds))
	if fracSeconds > 0 {
		s += ".5"
	}
	return s
}

func SetLineMark(w http.ResponseWriter, r *http.Request) {
	dbPath, redirect, err := GetUserDb(w, r)
	if redirect || err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}

	sqldb, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}
	defer sqldb.Close()

	w.Header().Set("Content-Type", "application/json")

	var setLineMark SetLineMarkRequest
	err = json.NewDecoder(r.Body).Decode(&setLineMark)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}

	row := sqldb.QueryRow(`SELECT lines FROM stories WHERE id = $1;`, setLineMark.StoryID)

	var linesJSON string
	if err := row.Scan(&linesJSON); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message: failure to get story lines": "` + err.Error() + `"}`))
		return
	}

	var lines []Line
	err = json.Unmarshal([]byte(linesJSON), &lines)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message: failure to unmarshal story lines JSON": "` + err.Error() + `"}`))
		return
	}

	if setLineMark.LineIdx < 0 || setLineMark.LineIdx >= len(lines) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{ "message: line to set mark must have index of 1 or greater}`))
		return
	}

	line := &lines[setLineMark.LineIdx]
	line.Marked = setLineMark.Marked

	linesBytes, err := json.Marshal(lines)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message: failure to JSONify story lines": "` + err.Error() + `"}`))
		return
	}

	_, err = sqldb.Exec(`UPDATE stories SET lines = $1 WHERE id = $2;`, linesBytes, setLineMark.StoryID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + "failure to update lines in story: " + err.Error() + `"}`))
		return
	}

	w.Write(linesBytes)
}
