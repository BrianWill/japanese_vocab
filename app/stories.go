package main

import (
	"compress/gzip"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
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

func addWords(tokens []*JpToken, kanjiSet []string, sqldb *sql.DB) (wordIds []int64, newWordCount int, err error) {
	var reHasKanji = regexp.MustCompile(`[\x{4E00}-\x{9FAF}]`)
	var reHasKatakana = regexp.MustCompile(`[ア-ン]`)
	var reHasKana = regexp.MustCompile(`[ア-ンァ-ヴぁ-ゔ]`)

	newWordCount = 0
	unixtime := time.Now().Unix()

	wordIds = make([]int64, 0)

	keptTokens := make([]*JpToken, 0)

	// filter out tokens that have no part of speech
	for i, token := range tokens {
		priorToken := &JpToken{}
		if i > 0 {
			priorToken = tokens[i-1]
		}

		pos := getTokenPOS(token, priorToken)
		if pos == "" {
			continue
		}

		keptTokens = append(keptTokens, token)
	}

	words := make(map[string]bool)
	for _, t := range keptTokens {
		words[t.BaseForm] = true
	}

	for _, k := range kanjiSet {
		words[k] = true
	}

	wordIdsMap := make(map[int64]bool)

	for baseForm := range words {
		hasKatakana := len(reHasKatakana.FindStringIndex(baseForm)) > 0
		hasKana := len(reHasKana.FindStringIndex(baseForm)) > 0
		hasKanji := len(reHasKanji.FindStringIndex(baseForm)) > 0
		isKanji := len([]rune(baseForm)) == 1 && reHasKanji.FindStringIndex(baseForm) != nil

		if !hasKana && !hasKanji { // not a valid word
			continue
		}

		category := 0

		// has katakana
		if hasKatakana {
			category |= DRILL_CATEGORY_KATAKANA
		}

		entries := getDefinitions(baseForm)
		for _, entry := range entries {
			for _, sense := range entry.Senses {
				category |= getVerbCategory(sense)
			}
		}

		entriesJSON, err := json.Marshal(entries)
		if err != nil {
			return nil, 0, err
		}

		kanjiDef := KanjiCharacter{}

		if isKanji {
			for _, ch := range allKanji.Characters {
				if ch.Literal == baseForm {
					kanjiDef = ch
					break
				}
			}

			category |= DRILL_CATEGORY_KANJI
		}

		kanjiDefJSON, err := json.Marshal(kanjiDef)
		if err != nil {
			return nil, 0, err
		}

		var id int64
		err = sqldb.QueryRow(`SELECT id FROM words WHERE base_form = $1`, baseForm).Scan(&id)
		if err != nil && err != sql.ErrNoRows {
			return nil, 0, err
		}
		if err == nil {
			if _, ok := wordIdsMap[id]; !ok {
				wordIdsMap[id] = true
				wordIds = append(wordIds, id)
			}
			continue
		}

		insertResult, err := sqldb.Exec(`INSERT INTO words (base_form, date_marked,
			date_added, category, lifetime_repetitions, status, definitions, kanji) 
			VALUES($1, $2, $3, $4, $5, $6, $7, $8);`,
			baseForm, 0, unixtime, category, 0, "catalog", entriesJSON, kanjiDefJSON)
		if err != nil {
			return nil, 0, fmt.Errorf("failure to insert word: " + err.Error())
		}

		id, err = insertResult.LastInsertId()
		if err != nil {
			return nil, 0, fmt.Errorf("failure to get id of inserted word: " + err.Error())
		}

		fmt.Println("inserted word: ", baseForm, id)

		newWordCount++
		wordIds = append(wordIds, id)
		wordIdsMap[id] = true
	}

	return wordIds, newWordCount, nil
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

	definitionsCache[baseForm] = entries
	return entries
}

type BaseFormCategoryPair struct {
	BaseForm string
	Category int
}

func GetCatalogStories(response http.ResponseWriter, request *http.Request) {
	dbPath := MAIN_USER_DB_PATH

	response.Header().Set("Content-Type", "application/json")

	sqldb, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}
	defer sqldb.Close()

	rows, err := sqldb.Query(`SELECT id, title, source, link, episode_number, audio, video, 
			status, level, date, date_marked, lifetime_repetitions FROM catalog_stories;`)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + "failure to get story: " + err.Error() + `"}`))
		return
	}
	defer rows.Close()

	var stories []CatalogStory
	for rows.Next() {
		var story CatalogStory
		if err := rows.Scan(&story.ID, &story.Title, &story.Source, &story.Link, &story.EpisodeNumber, &story.Audio, &story.Video,
			&story.Status, &story.Level, &story.Date, &story.DateMarked, &story.LifetimeRepetitions); err != nil {
			response.WriteHeader(http.StatusInternalServerError)
			response.Write([]byte(`{ "message": "` + "failure to read story list: " + err.Error() + `"}`))
			return
		}
		stories = append(stories, story)
	}

	json.NewEncoder(response).Encode(stories)
}

func GetStory(w http.ResponseWriter, r *http.Request) {
	dbPath := MAIN_USER_DB_PATH

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

	row := sqldb.QueryRow(`SELECT title, source, link, content, date, audio, video, 
		level, words, date_marked, status, transcript_en, transcript_ja FROM catalog_stories WHERE id = $1;`, id)

	var words string
	story := CatalogStory{ID: int64(id)}
	if err := row.Scan(&story.Title, &story.Source, &story.Link, &story.Content, &story.Date,
		&story.Audio, &story.Video,
		&story.Level, &words, &story.DateMarked, &story.Status,
		&story.TranscriptEN, &story.TranscriptJA); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		gw.Write([]byte(`{ "message": failure to scan story row:"` + err.Error() + `"}`))
		return
	}

	err = json.Unmarshal([]byte(words), &story.Words)
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

func UpdateStoryInfo(w http.ResponseWriter, r *http.Request) {
	dbPath := MAIN_USER_DB_PATH

	w.Header().Set("Content-Type", "application/json")

	var story CatalogStory
	err := json.NewDecoder(r.Body).Decode(&story)
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
	rows, err := sqldb.Query(`SELECT id FROM catalog_stories WHERE id = $1;`, story.ID)
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

	_, err = sqldb.Exec(`UPDATE catalog_stories SET 
			date_marked = $1, level = $2, lifetime_repetitions = $3, status = $4, 
			transcript_en = CASE WHEN $5 = '' THEN transcript_en ELSE $5 END,
			transcript_ja = CASE WHEN $6 = '' THEN transcript_ja ELSE $6 END
			WHERE id = $7;`,
		story.DateMarked, story.Level, story.LifetimeRepetitions, story.Status,
		story.TranscriptEN, story.TranscriptJA, story.ID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + "failure to update story: " + err.Error() + `"}`))
		return
	}

	json.NewEncoder(w).Encode(bson.M{"status": "success"})
}

func UnscheduleStory(w http.ResponseWriter, r *http.Request) {
	dbPath := MAIN_USER_DB_PATH

	var scheduleRequest ScheduleStoryRequest
	err := json.NewDecoder(r.Body).Decode(&scheduleRequest)
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

	if scheduleRequest.ID > 0 {
		_, err = sqldb.Exec(`DELETE FROM schedule_entries WHERE id = $1;`, scheduleRequest.ID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{ "message": "` + "failure to add schedule entry: " + err.Error() + `"}`))
			return
		}
	} else {
		_, err = sqldb.Exec(`DELETE FROM schedule_entries WHERE story = $1;`, scheduleRequest.Story)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{ "message": "` + "failure to add schedule entry: " + err.Error() + `"}`))
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(bson.M{"status": "success"})
}

func ScheduleStory(w http.ResponseWriter, r *http.Request) {
	dbPath := MAIN_USER_DB_PATH

	var scheduleRequest ScheduleStoryRequest
	err := json.NewDecoder(r.Body).Decode(&scheduleRequest)
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

	for i := 0; i < NUM_SCHEDULED_REPETITIONS; i++ {
		day := i * 3

		// add rep
		_, err := sqldb.Exec(`INSERT INTO schedule_entries (story, day_offset, type) VALUES($1, $2, $3);`,
			scheduleRequest.Story, day, LISTENING)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{ "message": "` + "failure to add schedule entry: " + err.Error() + `"}`))
			return
		}

		// add drill except after last rep
		if i < NUM_SCHEDULED_REPETITIONS-1 {
			_, err = sqldb.Exec(`INSERT INTO schedule_entries (story, day_offset, type) VALUES($1, $2, $3);`,
				scheduleRequest.Story, day+1, DRILLING)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{ "message": "` + "failure to add schedule entry: " + err.Error() + `"}`))
				return
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(bson.M{"status": "success"})
}

func GetSchedule(w http.ResponseWriter, r *http.Request) {
	dbPath := MAIN_USER_DB_PATH

	w.Header().Set("Content-Type", "application/json")

	sqldb, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}
	defer sqldb.Close()

	// make sure the story actually exists
	rows, err := sqldb.Query(`SELECT e.id, story, day_offset, type, 
		title, source, lifetime_repetitions, level 
		FROM schedule_entries as e INNER JOIN catalog_stories as s 
		ON e.id = s.id;`)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + "failure to read schedule entry: " + err.Error() + `"}`))
		return
	}
	defer rows.Close()

	entries := make([]ScheduleEntry, 0)

	for rows.Next() {
		var entry ScheduleEntry
		if err := rows.Scan(&entry.ID, &entry.Story, &entry.DayOffset, &entry.Type,
			&entry.Title, &entry.Source, &entry.LifetimeRepetitions, &entry.Level); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{ "message": "` + "failure to read schedule entry: " + err.Error() + `"}`))
			return
		}
		entries = append(entries, entry)
	}

	json.NewEncoder(w).Encode(entries)
}

func GetLog(w http.ResponseWriter, r *http.Request) {
	dbPath := MAIN_USER_DB_PATH

	w.Header().Set("Content-Type", "application/json")

	sqldb, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}
	defer sqldb.Close()

	// make sure the story actually exists
	rows, err := sqldb.Query(`SELECT id, story, date, type FROM log_entries;`)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + "failure to get story: " + err.Error() + `"}`))
		return
	}
	defer rows.Close()

	entries := make([]LogEntry, 0)

	for rows.Next() {
		var entry LogEntry
		if err := rows.Scan(&entry.ID, &entry.Story, &entry.Date, &entry.Type); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{ "message": "` + "failure to read story list: " + err.Error() + `"}`))
			return
		}
		entries = append(entries, entry)
	}

	json.NewEncoder(w).Encode(entries)
}
