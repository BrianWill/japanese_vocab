package main

// [START import]
import (
	// "context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gorilla/mux"
	"github.com/ikawaha/kagome/v2/tokenizer"
	_ "github.com/mattn/go-sqlite3"
	"go.mongodb.org/mongo-driver/bson"
	//"go.mongodb.org/mongo-driver/bson/primitive"
)

const INITIAL_STATUS = 1
const INITIAL_RANK = 4

const DRILL_FILTER_ON_COOLDOWN = "on"
const DRILL_FILTER_OFF_COOLDOWN = "off"
const DRILL_FILTER_ALL = "all"

const STORY_STATUS_CURRENT = 3
const STORY_STATUS_READ = 2
const STORY_STATUS_NEVER_READ = 1
const STORY_STATUS_ARCHIVE = 0

func CreateStory(response http.ResponseWriter, request *http.Request) {
	dbPath, redirect, err := GetUserDb(response, request)
	if redirect || err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `"}`))
	}

	response.Header().Add("content-type", "application/json")

	var story Story
	json.NewDecoder(request.Body).Decode(&story)

	sqldb, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `"}`))
	}
	defer sqldb.Close()

	_, err = addStory(story, sqldb, false)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}
	json.NewEncoder(response).Encode("Success adding story")
}

func RetokenizeStory(response http.ResponseWriter, request *http.Request) {
	dbPath, redirect, err := GetUserDb(response, request)
	if redirect || err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `"}`))
	}

	response.Header().Add("content-type", "application/json")

	var story Story
	json.NewDecoder(request.Body).Decode(&story)

	sqldb, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `"}`))
	}
	defer sqldb.Close()

	_, err = addStory(story, sqldb, true)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}
	json.NewEncoder(response).Encode("Success retokenizing story")
}

func tokenize(content string) ([]*JpToken, error) {
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

	tokens, err := extractKanji(tokens)
	if err != nil {
		return nil, fmt.Errorf(`failure to extract kanji` + err.Error())
	}

	// err = getDefinitions(tokens)
	// if err != nil {
	// 	return nil, fmt.Errorf(`failure to get definitions` + err.Error())
	// }

	return tokens, nil
}

func addStory(story Story, sqldb *sql.DB, retokenize bool) (id int64, err error) {
	if retokenize {
		var linesJSON string
		row := sqldb.QueryRow(`SELECT title, link, lines, date_added 
			FROM stories WHERE id = $1;`, story.ID)
		if err := row.Scan(&story.Title, &story.Link, &linesJSON, &story.DateAdded); err != nil {
			return 0, fmt.Errorf("failure to read story: " + err.Error())
		}
		err := json.Unmarshal([]byte(linesJSON), &story.Lines)
		if err != nil {
			return 0, fmt.Errorf("failure to unmarshall story lines: " + err.Error())
		}
		for _, line := range story.Lines {
			story.Content += line.Timestamp + "\n" + line.Content + "\n"
		}
	} else {
		row := sqldb.QueryRow(`SELECT id FROM stories WHERE title = $1;`, story.Title)
		if err := row.Scan(&story.ID); err != nil && err != sql.ErrNoRows {
			return 0, fmt.Errorf("story with same title already exists: " + err.Error())
		}
	}

	timestampRegex := regexp.MustCompile(`(?m)^\s*\d*:\d*\s*$`) // match timestamp line
	timestamps := timestampRegex.FindAllString(story.Content, -1)
	lineContents := timestampRegex.Split(story.Content, -1)

	lines := make([]Line, 0)

	// todo: check that the timestamps increase in value

	for i, timestamp := range timestamps {
		fmt.Println(timestamp, strings.TrimSpace(lineContents[i+1]))

		tokens, err := tokenize(strings.TrimSpace(lineContents[i+1]))
		if err != nil {
			return 0, fmt.Errorf("failure to tokenize story: " + err.Error())
		}

		wordsOfLine, err := addWords(tokens, sqldb)
		if err != nil {
			return 0, fmt.Errorf("failure to add words: " + err.Error())
		}

		lines = append(lines, Line{
			Content:   lineContents[i+1],
			Timestamp: timestamp,
			Words:     wordsOfLine,
		})
	}

	if len(timestamps) == 0 {
		blanklinesRegex := regexp.MustCompile(`(?m)^\s*$\n`) // match timestamp line
		lineContents = blanklinesRegex.Split(story.Content, -1)

		for _, content := range lineContents {
			fmt.Println("line without timestamp", content)

			tokens, err := tokenize(strings.TrimSpace(content))
			if err != nil {
				return 0, fmt.Errorf("failure to tokenize story: " + err.Error())
			}

			wordsOfLine, err := addWords(tokens, sqldb)
			if err != nil {
				return 0, fmt.Errorf("failure to add words: " + err.Error())
			}

			lines = append(lines, Line{
				Content:   content,
				Timestamp: ":",
				Words:     wordsOfLine,
			})
		}
	}

	linesJson, err := json.Marshal(lines)
	if err != nil {
		return 0, fmt.Errorf("failure to lines: " + err.Error())
	}

	if retokenize {
		_, err = sqldb.Exec(`UPDATE stories SET lines = $1 WHERE id = $2;`,
			linesJson, story.ID)
		if err != nil {
			return 0, fmt.Errorf("failure to update story: " + err.Error())
		}
		return story.ID, nil
	} else {
		date := time.Now().Unix()
		result, err := sqldb.Exec(`INSERT INTO stories (lines, title, link, date_added, status) 
				VALUES($1, $2, $3, $4, $5);`,
			linesJson, story.Title, story.Link, date, INITIAL_STATUS)
		if err != nil {
			return 0, fmt.Errorf("failure to insert story: " + err.Error())
		}
		id, err = result.LastInsertId()
		if err != nil {
			return 0, fmt.Errorf("failure to insert story: " + err.Error())
		}
		return id, nil
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

func extractKanji(tokens []*JpToken) ([]*JpToken, error) {
	newTokens := tokens
	kanjiMap := make(map[string]bool)

	for _, t := range tokens {
		var re = regexp.MustCompile(`[\x{4E00}-\x{9FAF}]`)
		kanji := re.FindAllString(t.Surface+t.BaseForm, -1)

		for _, s := range kanji {
			kanjiMap[s] = true
		}
	}

	for k := range kanjiMap {
		tok := JpToken{
			Surface:  k,
			BaseForm: k,
		}
		newTokens = append(newTokens, &tok)
	}

	return newTokens, nil
}

func addWords(tokens []*JpToken, sqldb *sql.DB) ([]LineWord, error) {
	var reHasKanji = regexp.MustCompile(`[\x{4E00}-\x{9FAF}]`)
	var reHasKatakana = regexp.MustCompile(`[ア-ン]`)

	lineWords := make([]LineWord, len(tokens))
	for i, token := range tokens {
		priorToken := &JpToken{}
		if i > 0 {
			priorToken = tokens[i-1]
		}

		lineWord := &lineWords[i]
		lineWord.Surface = token.Surface
		lineWord.POS = getTokenPOS(token, priorToken)

		if lineWord.POS == "" { // not a vocab word
			continue
		}

		row := sqldb.QueryRow(`SELECT id FROM words WHERE base_form = $1;`, token.BaseForm)

		var id int64
		err := row.Scan(&id)
		if err != nil { // word already exists in word set
			lineWord.ID = id
		} else {
			drillType := 0

			// has katakana
			if len(reHasKatakana.FindStringIndex(token.BaseForm)) > 0 {
				drillType |= DRILL_TYPE_KATAKANA
			}

			// is a single kanji
			hasKanji := len(reHasKanji.FindStringIndex(token.BaseForm)) > 0
			if hasKanji && utf8.RuneCountInString(token.BaseForm) == 1 {
				drillType |= DRILL_TYPE_KANJI
			}

			for _, entry := range token.Entries {
				for _, sense := range entry.Senses {
					drillType |= getVerbDrillType(sense)
				}
			}

			unixtime := time.Now().Unix()

			insertResult, err := sqldb.Exec(`INSERT INTO words (base_form,  date_last_read, date_last_drill,
					date_added, date_last_wrong,  drill_type, rank, drill_count) 
					VALUES($1, $2, $3, $4, $5, $6, $7, $8);`,
				token.BaseForm, unixtime, 0, unixtime, 0, drillType, INITIAL_RANK, 0)
			if err != nil {
				return nil, fmt.Errorf("failure to insert word: " + err.Error())
			}

			id, err := insertResult.LastInsertId()
			if err != nil {
				return nil, fmt.Errorf("failure to get id of inserted word: " + err.Error())
			}
			lineWord.ID = id
		}
	}

	return lineWords, nil
}

func getDefinitions(tokens []*JpToken) error {
	reHasKanji := regexp.MustCompile(`[\x{4E00}-\x{9FAF}]`)
	for i, token := range tokens {
		entries := make([]JMDictEntry, 0)

		searchTerm := token.BaseForm
		hasKanji := len(reHasKanji.FindStringIndex(searchTerm)) > 0

		if hasKanji {
			for _, e := range allEntriesByKanjiSpellings[searchTerm] {
				entries = append(entries, *e)
			}
		} else {
			for _, e := range allEntriesByReading[searchTerm] {
				entries = append(entries, *e)
			}
		}

		// too many matching entries is just noise
		if len(entries) > 10 {
			entries = entries[:10]
		}

		tokens[i].Entries = entries
	}
	return nil
}

func getVerbDrillType(sense JMDictSense) int {
	drillType := 0
	for _, pos := range sense.Pos {
		switch pos {
		case "verb-ichidan":
			drillType |= DRILL_TYPE_ICHIDAN
		case "verb-godan-su":
			drillType |= DRILL_TYPE_GODAN_SU
		case "verb-godan-ku":
			drillType |= DRILL_TYPE_GODAN_KU
		case "verb-godan-gu":
			drillType |= DRILL_TYPE_GODAN_GU
		case "verb-godan-ru":
			drillType |= DRILL_TYPE_GODAN_RU
		case "verb-godan-u":
			drillType |= DRILL_TYPE_GODAN_U
		case "verb-godan-tsu":
			drillType |= DRILL_TYPE_GODAN_TSU
		case "verb-godan-mu":
			drillType |= DRILL_TYPE_GODAN_MU
		case "verb-godan-nu":
			drillType |= DRILL_TYPE_GODAN_NU
		case "verb-godan-bu":
			drillType |= DRILL_TYPE_GODAN_BU
		}
	}
	return drillType
}

func GetUserDb(response http.ResponseWriter, request *http.Request) (string, bool, error) {
	session, err := sessionStore.Get(request, "session")
	if err != nil {
		return "", false, err
	}

	if session.IsNew {
		http.Redirect(response, request, "/login.html", http.StatusSeeOther)
		return "", true, err
	}

	dbPath, ok := session.Values["user_db_path"].(string)
	if !ok {
		return "", false, errors.New("session missing db path")
	}

	return dbPath, false, nil
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

	response.Header().Add("content-type", "application/json")

	sqldb, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}
	defer sqldb.Close()

	rows, err := sqldb.Query(`SELECT id, title, link, status, date_added FROM stories;`)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + "failure to get story: " + err.Error() + `"}`))
		return
	}
	defer rows.Close()

	var stories []Story
	for rows.Next() {
		var story Story
		if err := rows.Scan(&story.ID, &story.Title, &story.Link, &story.Status, &story.DateAdded); err != nil {
			response.WriteHeader(http.StatusInternalServerError)
			response.Write([]byte(`{ "message": "` + "failure to read story list: " + err.Error() + `"}`))
			return
		}
		stories = append(stories, story)
	}

	json.NewEncoder(response).Encode(stories)
}

func GetStory(response http.ResponseWriter, request *http.Request) {
	dbPath, redirect, err := GetUserDb(response, request)
	if redirect || err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `"}`))
	}

	response.Header().Add("content-type", "application/json")
	params := mux.Vars(request)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}

	sqldb, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}
	defer sqldb.Close()

	story, err := getStory(int64(id), sqldb)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}

	json.NewEncoder(response).Encode(story)
}

func getStory(id int64, sqldb *sql.DB) (Story, error) {
	row := sqldb.QueryRow(`SELECT title, link, lines, date_added FROM stories WHERE id = $1;`, id)

	var linesJSON string
	story := Story{ID: id}
	if err := row.Scan(&story.Title, &story.Link, &linesJSON, &story.DateAdded); err != nil {
		return Story{}, fmt.Errorf("failure to scan story row: " + err.Error())
	}

	err := json.Unmarshal([]byte(linesJSON), &story.Lines)
	if err != nil {
		return Story{}, fmt.Errorf("failure to unmarshall story lines: " + err.Error())
	}

	return story, nil
}

func UpdateStory(response http.ResponseWriter, request *http.Request) {
	dbPath, redirect, err := GetUserDb(response, request)
	if redirect || err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `"}`))
	}

	response.Header().Add("content-type", "application/json")

	var story Story
	err = json.NewDecoder(request.Body).Decode(&story)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}

	sqldb, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}
	defer sqldb.Close()

	// make sure the story actually exists
	rows, err := sqldb.Query(`SELECT id FROM stories WHERE id = $1;`, story.ID)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + "failure to get story: " + err.Error() + `"}`))
		return
	}

	if !rows.Next() {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + "story with ID does not exist: " + strconv.FormatInt(story.ID, 10) + `"}`))
		rows.Close()
		return
	}
	rows.Close()

	_, err = sqldb.Exec(`UPDATE stories SET status = $1 WHERE id = $2;`,
		story.Status, story.ID)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + "failure to update story: " + err.Error() + `"}`))
		return
	}

	json.NewEncoder(response).Encode(bson.M{"status": "success"})
}

func AddLogEvent(response http.ResponseWriter, request *http.Request) {
	dbPath, redirect, err := GetUserDb(response, request)
	if redirect || err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `"}`))
	}

	response.Header().Add("content-type", "application/json")

	params := mux.Vars(request)
	var storyId int64
	id, err := strconv.Atoi(params["id"])
	storyId = int64(id)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}

	sqldb, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}
	defer sqldb.Close()

	date := time.Now().Unix()

	if storyId < 0 {
		// get random story from set of current stories

		rows, err := sqldb.Query(`SELECT id, status FROM stories WHERE status = $1;`, STORY_STATUS_CURRENT)
		if err != nil {
			response.WriteHeader(http.StatusInternalServerError)
			response.Write([]byte(`{ "message": "` + "failure to get story: " + err.Error() + `"}`))
			return
		}
		defer rows.Close()

		var stories = make([]Story, 0)
		var story Story
		for rows.Next() {
			if err := rows.Scan(&story.ID, &story.Status); err != nil {
				response.WriteHeader(http.StatusInternalServerError)
				response.Write([]byte(`{ "message": "` + "failure to read story: " + err.Error() + `"}`))
				return
			}
			if story.Status == STORY_STATUS_CURRENT {
				stories = append(stories, story)
			}
		}

		if len(stories) == 0 {
			json.NewEncoder(response).Encode("Did not add a log even. No current stories.")
			return
		}

		rand.Seed(date)
		storyId = stories[rand.Intn(len(stories))].ID
	}

	_, err = sqldb.Exec(`INSERT INTO log_events (date, story) 
			VALUES($1, $2);`,
		date, storyId)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + "failure to insert log event: " + err.Error() + `"}`))
	}
	json.NewEncoder(response).Encode("Success adding log event")
}

func RemoveLogEvent(response http.ResponseWriter, request *http.Request) {
	dbPath, redirect, err := GetUserDb(response, request)
	if redirect || err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `"}`))
	}

	response.Header().Add("content-type", "application/json")

	params := mux.Vars(request)
	var logId int64
	id, err := strconv.Atoi(params["id"])
	logId = int64(id)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}

	sqldb, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}
	defer sqldb.Close()

	_, err = sqldb.Exec(`DELETE FROM log_events WHERE id = $1;`, logId)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + "failure to delete log event: " + err.Error() + `"}`))
		return
	}

	json.NewEncoder(response).Encode(bson.M{"status": "success"})
}

func GetLogEvents(response http.ResponseWriter, request *http.Request) {
	dbPath, redirect, err := GetUserDb(response, request)
	if redirect {
		return
	}
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `"}`))
	}

	response.Header().Add("content-type", "application/json")

	sqldb, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}
	defer sqldb.Close()

	rows, err := sqldb.Query(`SELECT id, date, story FROM log_events ORDER BY date DESC;`)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + "failure to get story: " + err.Error() + `"}`))
		return
	}
	defer rows.Close()

	var logEvents = make([]LogEvent, 0)
	for rows.Next() {
		var logEvent LogEvent
		if err := rows.Scan(&logEvent.ID, &logEvent.Date, &logEvent.StoryID); err != nil {
			response.WriteHeader(http.StatusInternalServerError)
			response.Write([]byte(`{ "message": "` + "failure to read story: " + err.Error() + `"}`))
			return
		}
		logEvents = append(logEvents, logEvent)
	}

	json.NewEncoder(response).Encode(bson.M{"logEvents": logEvents})
}

// [END indexHandler]
// [END gae_go111_app]
