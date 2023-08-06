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
	"sort"
	"strconv"
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

	err = addStory(story, response, dbPath)
	if err != nil {
		return
	}
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

	err = retokenizeStory(story, response, dbPath)
	if err != nil {
		return
	}
}

func tokenize(story Story, response http.ResponseWriter, sqldb *sql.DB) ([]byte, []byte, error) {
	analyzerTokens := tok.Analyze(story.Content, tokenizer.Normal)
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

	tokens, err := extractKanji(tokens, response)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": failure to extract kanji"` + err.Error() + `"}`))
		return nil, nil, err
	}

	err = getDefinitions(tokens, response)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": failure to get definitions"` + err.Error() + `"}`))
		return nil, nil, err
	}

	wordIds, err := addDrillWords(tokens, response, sqldb)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + "failure to add words: " + err.Error() + `"}`))
		return nil, nil, err
	}

	wordsJson, err := json.Marshal(wordIds)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + "failure to marshall wordIds: " + err.Error() + `"}`))
		return nil, nil, err
	}

	tokensJson, err := json.Marshal(tokens)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + "failure to marshall tokens: " + err.Error() + `"}`))
		return nil, nil, err
	}

	return wordsJson, tokensJson, nil
}

func addStory(story Story, response http.ResponseWriter, dbPath string) error {
	sqldb, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return err
	}
	defer sqldb.Close()

	rows, err := sqldb.Query(`SELECT id FROM stories WHERE title = $1;`, story.Title)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + "failure to get story: " + err.Error() + `"}`))
		return err
	}
	defer rows.Close()

	for rows.Next() {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "story with same title already exists"}`))
		return fmt.Errorf("story with same title already exists")
	}

	wordsJson, tokensJson, err := tokenize(story, response, sqldb)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + "failure to tokenize story: " + err.Error() + `"}`))
		return err
	}

	date := time.Now().Unix()
	_, err = sqldb.Exec(`INSERT INTO stories (words, content, title, link, tokens, date_last_read, date_added, status) 
			VALUES($1, $2, $3, $4, $5, $6, $7, $8);`,
		wordsJson, story.Content, story.Title, story.Link, tokensJson, date, date, INITIAL_STATUS)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + "failure to insert story: " + err.Error() + `"}`))
		return err
	}
	json.NewEncoder(response).Encode("Success adding story")
	return nil
}

func extractKanji(tokens []*JpToken, response http.ResponseWriter) ([]*JpToken, error) {
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
		fmt.Println("FOUND KANJI: ", k)
		tok := JpToken{
			Surface:  k,
			BaseForm: k,
		}
		newTokens = append(newTokens, &tok)
	}

	return newTokens, nil
}

func retokenizeStory(story Story, response http.ResponseWriter, dbPath string) error {
	sqldb, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return err
	}
	defer sqldb.Close()

	rows, err := sqldb.Query(`SELECT title, link, tokens, content, date_added, 
		date_last_read FROM stories WHERE id = $1;`, story.ID)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + "failure to get story: " + err.Error() + `"}`))
		return err
	}
	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(&story.Title, &story.Link, &story.Tokens, &story.Content,
			&story.DateAdded, &story.DateLastRead); err != nil {
			response.WriteHeader(http.StatusInternalServerError)
			response.Write([]byte(`{ "message": "` + "failure to read story: " + err.Error() + `"}`))
			return err
		}
	}

	wordsJson, tokensJson, err := tokenize(story, response, sqldb)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + "failure to tokenize story: " + err.Error() + `"}`))
		return err
	}

	_, err = sqldb.Exec(`UPDATE stories SET tokens = $1, words = $2 WHERE id = $3;`,
		tokensJson, wordsJson, story.ID)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + "failure to update story: " + err.Error() + `"}`))
		return err
	}
	json.NewEncoder(response).Encode("Success retokenizing story")
	return nil
}

func filterPartsOfSpeech(tokens []*JpToken) []*JpToken {
	filteredTokens := make([]*JpToken, 0)
	prior := &JpToken{}
	for _, t := range tokens {
		if t.Surface == "。" {
			continue
		} else if t.Surface == "、" {
			continue
		} else if t.Surface == " " {
			continue
		} else if t.POS == "" { // kanji added in extractKanji()
			filteredTokens = append(filteredTokens, t)
		} else if t.POS == "動詞" && t.POS_1 == "非自立" { // auxilliary verb
			filteredTokens = append(filteredTokens, t)
		} else if t.POS == "副詞" { // adverb
			filteredTokens = append(filteredTokens, t)
		} else if t.POS == "接続詞" && t.POS_1 == "*" { // conjunction
			filteredTokens = append(filteredTokens, t)
		} else if t.POS == "形容詞" { // i-adj
			filteredTokens = append(filteredTokens, t)
		} else if t.POS == "名詞" && t.POS_1 == "代名詞" { // pronoun
			filteredTokens = append(filteredTokens, t)
		} else if t.POS == "連体詞" { // adnominal adjective
			filteredTokens = append(filteredTokens, t)
		} else if t.POS == "動詞" { //　verb
			filteredTokens = append(filteredTokens, t)
		} else if t.POS == "名詞" && t.POS_1 == "接尾" { // noun suffix
			filteredTokens = append(filteredTokens, t)
		} else if (prior.POS == "助詞" && (prior.POS_1 == "連体化" || prior.POS_1 == "並立助詞")) || // preceded by connective particle
			(prior.POS == "接頭詞" && prior.POS_1 == "名詞接続") { // preceded by prefix
			filteredTokens = append(filteredTokens, t)
		} else if t.POS == "名詞" { // noun
			filteredTokens = append(filteredTokens, t)
		} else if t.POS == "号" { // counter
			filteredTokens = append(filteredTokens, t)
		}
		prior = t
	}
	return filteredTokens
}

func addDrillWords(tokens []*JpToken, response http.ResponseWriter, sqldb *sql.DB) ([]int64, error) {
	var reHasKanji = regexp.MustCompile(`[\x{4E00}-\x{9FAF}]`)
	var reHasKana = regexp.MustCompile(`[あ-んア-ン]`)
	var reHasKatakana = regexp.MustCompile(`[ア-ン]`)

	fmt.Println("prefiltered tokens: ", len(tokens))
	filteredTokens := filterPartsOfSpeech(tokens)
	fmt.Println("filtered tokens: ", len(filteredTokens))

	// deduplicate
	tokenSet := make(map[string]*JpToken)
	for _, token := range filteredTokens {
		tokenSet[token.BaseForm] = token
	}
	filteredTokens = nil
	for k := range tokenSet {
		filteredTokens = append(filteredTokens, tokenSet[k])
	}

	unixtime := time.Now().Unix()

	idsByBaseForm := make(map[string]int64)

	wordIds := make([]int64, 0)
	for _, token := range filteredTokens {
		hasKanji := len(reHasKanji.FindStringIndex(token.BaseForm)) > 0
		hasKana := len(reHasKana.FindStringIndex(token.BaseForm)) > 0
		if !hasKanji && !hasKana {
			continue
		}

		rows, err := sqldb.Query(`SELECT id FROM words WHERE base_form = $1;`, token.BaseForm)
		if err != nil {
			response.WriteHeader(http.StatusInternalServerError)
			response.Write([]byte(`{ "message": "` + "error while looking up word: " + err.Error() + `"}`))
			return nil, err
		}
		exists := rows.Next()

		var id int64
		if exists {
			rows.Scan(&id)
			idsByBaseForm[token.BaseForm] = id
			wordIds = append(wordIds, id)

			// todo update
			//fmt.Printf("existing word: %s \t %d\n", token.BaseForm, id)
		} else {
			drillType := 0
			hasKatakana := len(reHasKatakana.FindStringIndex(token.BaseForm)) > 0
			if hasKatakana {
				drillType |= DRILL_TYPE_KATAKANA
			}
			if hasKanji && utf8.RuneCountInString(token.BaseForm) == 1 {
				drillType |= DRILL_TYPE_KANJI
			}

			for _, entry := range token.Entries {
				for _, sense := range entry.Senses {
					drillType |= getVerbDrillType(sense)
				}
			}

			entriesJson, err := json.Marshal(token.Entries)
			if err != nil {
				response.WriteHeader(http.StatusInternalServerError)
				response.Write([]byte(`{ "message": "` + "failure to json encode entry: " + err.Error() + `"}`))
				rows.Close()
				return nil, err
			}

			fmt.Printf("\nadding word: %s %d \t %d\n", token.BaseForm, len(token.Entries), id)

			insertResult, err := sqldb.Exec(`INSERT INTO words (base_form,  
					date_last_read, date_last_drill, date_added, date_last_wrong, definitions, drill_type, rank, drill_count) VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9);`,
				token.BaseForm, unixtime, 0, unixtime, 0, entriesJson, drillType, INITIAL_RANK, 0)
			if err != nil {
				response.WriteHeader(http.StatusInternalServerError)
				response.Write([]byte(`{ "message": "` + "failure to insert word: " + err.Error() + `"}`))
				rows.Close()
				return nil, err
			}

			id, err := insertResult.LastInsertId()
			if err != nil {
				response.WriteHeader(http.StatusInternalServerError)
				response.Write([]byte(`{ "message": "` + "failure to get id of inserted word: " + err.Error() + `"}`))
				rows.Close()
				return nil, err
			}
			fmt.Printf("new word: %s \t %d\n", token.BaseForm, id)
			idsByBaseForm[token.BaseForm] = id
			wordIds = append(wordIds, id)

		}
		rows.Close()
	}
	sort.Slice(wordIds, func(a, b int) bool {
		return wordIds[a] < wordIds[b]
	})

	for _, token := range tokens {
		token.Entries = nil
		if id, ok := idsByBaseForm[token.BaseForm]; ok {
			token.WordId = id
		}
	}

	return wordIds, nil
}

func getDefinitions(tokens []*JpToken, response http.ResponseWriter) error {
	start := time.Now()
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
	duration := time.Since(start)
	fmt.Printf("time to get definitions of %d tokens: %s \n ", len(tokens), duration)
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

	rows, err := sqldb.Query(`SELECT id, title, link, status, date_last_read, date_added FROM stories;`)
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
			&story.DateLastRead, &story.DateAdded); err != nil {
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

	rows, err := sqldb.Query(`SELECT title, link, tokens, content, date_added, 
		date_last_read, words FROM stories WHERE id = $1;`, id)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + "failure to get story: " + err.Error() + `"}`))
		return
	}
	defer rows.Close()

	var story Story
	for rows.Next() {
		if err := rows.Scan(&story.Title, &story.Link, &story.Tokens, &story.Content,
			&story.DateAdded, &story.DateLastRead, &story.Words); err != nil {
			response.WriteHeader(http.StatusInternalServerError)
			response.Write([]byte(`{ "message": "` + "failure to read story: " + err.Error() + `"}`))
			return
		}
	}
	story.ID = int64(id)

	var wordIds []int64
	err = json.Unmarshal([]byte(story.Words), &wordIds)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + "failure to read story: " + err.Error() + `"}`))
		return
	}

	// Helper function to generate the placeholders for the SQL query
	placeholders := func(n int) string {
		args := make([]byte, 2*n-1)
		for i := range args {
			if i%2 == 0 {
				args[i] = '?'
			} else {
				args[i] = ','
			}
		}
		return string(args)
	}

	query := fmt.Sprintf(`SELECT id, base_form, rank, drill_count,
		date_last_read, date_last_drill, definitions, drill_type, date_last_wrong, 
		date_added FROM words WHERE id IN (%s);`, placeholders(len(wordIds)))
	args := make([]interface{}, len(wordIds))
	for i, id := range wordIds {
		args[i] = id
	}

	rows, err = sqldb.Query(query, args...)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + "failure to get words: " + err.Error() + `"}`))
		return
	}
	defer rows.Close()

	words := make(map[int64]DrillWord)
	for rows.Next() {
		var word DrillWord
		err = rows.Scan(&word.ID, &word.BaseForm, &word.Rank,
			&word.DrillCount,
			&word.DateLastRead, &word.DateLastDrill,
			&word.Definitions, &word.DrillType, &word.DateLastWrong, &word.DateAdded)
		if err != nil {
			response.WriteHeader(http.StatusInternalServerError)
			response.Write([]byte(`{ "message": "` + "failure to scan word: " + err.Error() + `"}`))
			return
		}
		words[word.ID] = word
	}

	wordsJson, err := json.Marshal(words)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + "failure to marshall words: " + err.Error() + `"}`))
		return
	}
	story.Words = string(wordsJson)

	json.NewEncoder(response).Encode(story)
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

	_, err = sqldb.Exec(`UPDATE stories SET date_last_read = $1, status = $2 WHERE id = $3;`,
		story.DateLastRead, story.Status, story.ID)
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
