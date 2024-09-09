package main

import (
	"compress/gzip"
	"database/sql"
	"encoding/json"
	"fmt"
	"regexp"
	"time"

	// "math"
	"net/http"
	//"time"
	"go.mongodb.org/mongo-driver/bson"
)

func GetWords(w http.ResponseWriter, r *http.Request) {
	dbPath := MAIN_USER_DB_PATH

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("content-encoding", "gzip")

	gw := gzip.NewWriter(w)
	defer gw.Close()

	var body DrillRequest
	json.NewDecoder(r.Body).Decode(&body)

	sqldb, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		gw.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}
	defer sqldb.Close()

	var story_title string
	var story_source string
	var story_link string

	row := sqldb.QueryRow(`SELECT title, source, link FROM stories WHERE id = $1;`, body.StoryId)
	err = row.Scan(&story_title, &story_source, &story_link)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		gw.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}

	words, err := getWordsFromExcerpt(sqldb, body.StoryId, body.ExcerptHash)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		gw.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}

	json.NewEncoder(gw).Encode(bson.M{"words": words, "story_link": story_link, "story_title": story_title, "story_source": story_source})
}

func getExcerpt(sqldb *sql.DB, storyId int64, excerptHash int64) (Excerpt, error) {
	var excerptsJSON string
	row := sqldb.QueryRow(`SELECT excerpts FROM stories WHERE id = $1;`, storyId)
	err := row.Scan(&excerptsJSON)
	if err != nil {
		return Excerpt{}, err
	}

	var excerpts []Excerpt
	err = json.Unmarshal([]byte(excerptsJSON), &excerpts)
	if err != nil {
		return Excerpt{}, err
	}

	for _, ex := range excerpts {
		if ex.Hash == excerptHash {
			return ex, nil
		}
	}

	return Excerpt{}, fmt.Errorf("excerpt with matching hash not found")
}

func getWordsFromExcerpt(sqldb *sql.DB, storyId int64, excerptHash int64) ([]DrillWord, error) {
	excerpt, err := getExcerpt(sqldb, storyId, excerptHash)
	if err != nil {
		return nil, err
	}

	var subtitlesJA string
	var content string
	row := sqldb.QueryRow(`SELECT content, subtitles_ja FROM stories WHERE id = $1;`, storyId)
	err = row.Scan(&content, &subtitlesJA)
	if err != nil {
		return nil, err
	}

	excerptText := content
	if subtitlesJA != "" {
		excerptText, err = getSubtitlesContentInTimeRange(subtitlesJA, excerpt.StartTime, excerpt.EndTime)
		if err != nil {
			return nil, err
		}
	}

	tokens, _, err := tokenize(excerptText)
	if err != nil {
		return nil, err
	}

	// get word info
	wordMap := make(map[string]DrillWord)

	for _, token := range tokens {
		word := DrillWord{}
		word.BaseForm = token.BaseForm
		row := sqldb.QueryRow(`SELECT id, archived, category,
				repetitions, definitions, date_last_rep FROM words WHERE base_form = $1;`, token.BaseForm)
		err = row.Scan(&word.ID, &word.Archived, &word.Category,
			&word.Repetitions, &word.Definitions, &word.DateLastRep)
		if err == sql.ErrNoRows {
			continue
		}
		if err != nil {
			return nil, err
		}
		wordMap[word.BaseForm] = word
	}

	words := make([]DrillWord, len(wordMap))
	i := 0
	for _, word := range wordMap {
		words[i] = word
		i++
	}

	return words, nil
}

func UpdateWordArchiveState(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var word WordUpdate
	json.NewDecoder(r.Body).Decode(&word)

	sqldb, err := sql.Open("sqlite3", MAIN_USER_DB_PATH)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}
	defer sqldb.Close()

	row := sqldb.QueryRow(`SELECT id FROM words WHERE base_form = $1;`, word.BaseForm)
	var id int64

	err = row.Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{ "message": "` + "cannot update word; word not found" + err.Error() + `"}`))
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + "error looking up word " + err.Error() + `"}`))
		return
	}

	_, err = sqldb.Exec(`UPDATE words SET archived = $1 WHERE base_form = $2;`,
		word.Archived, word.BaseForm)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + "failure to update word: " + err.Error() + `"}`))
		return
	}

	json.NewEncoder(w).Encode(word)
}

func GetKanji(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var str string
	json.NewDecoder(r.Body).Decode(&str)

	var re = regexp.MustCompile(`[\x{4E00}-\x{9FAF}]`)
	kanjiStrings := re.FindAllString(str, -1)

	sqldb, err := sql.Open("sqlite3", MAIN_USER_DB_PATH)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}
	defer sqldb.Close()

	// remove duplicate kanji but maintain original order
	kanjiSet := make([]string, 0)
outer:
	for _, ch := range kanjiStrings {
		for _, k := range kanjiSet {
			if k == ch {
				continue outer
			}
		}
		kanjiSet = append(kanjiSet, ch)
	}

	kanjiDefs := make([]string, 0)
	for _, ch := range kanjiSet {
		row := sqldb.QueryRow(`SELECT kanji FROM words WHERE base_form = $1;`, ch)
		var def string
		err = row.Scan(&def)
		if err != nil {
			if err == sql.ErrNoRows {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(`{ "message": "` + "kanji not found: " + err.Error() + `"}`))
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{ "message": "` + "error looking up kanji: " + err.Error() + `"}`))
			return
		}
		kanjiDefs = append(kanjiDefs, def)
	}

	json.NewEncoder(w).Encode(kanjiDefs)
}

func IncWords(w http.ResponseWriter, r *http.Request) {
	dbPath := MAIN_USER_DB_PATH

	var body IncWordsRequest
	err := json.NewDecoder(r.Body).Decode(&body)
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

	unixtime := time.Now().Unix()

	for _, wordId := range body.Words {
		_, err := sqldb.Exec(`UPDATE words SET repetitions = repetitions + 1, date_last_rep = $1 WHERE id = $2;`, unixtime, wordId)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{ "message": "` + "failure to update word repetition counts: " + err.Error() + `"}`))
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(bson.M{"status": "success"})
}
