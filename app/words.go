package main

import (
	"compress/gzip"
	"database/sql"
	"encoding/json"
	"regexp"

	// "math"
	"net/http"
	//"time"
	"go.mongodb.org/mongo-driver/bson"
)

func WordDrill(w http.ResponseWriter, r *http.Request) {
	dbPath := MAIN_USER_DB_PATH

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("content-encoding", "gzip")

	gw := gzip.NewWriter(w)
	defer gw.Close()

	var drillRequest DrillRequest
	json.NewDecoder(r.Body).Decode(&drillRequest)

	sqldb, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		gw.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}
	defer sqldb.Close()

	var wordIds []int64

	if drillRequest.Set == "in_progress" {

		rows, err := sqldb.Query(`SELECT base_form, date_marked, status, audio, audio_start, 
				audio_end, drill_countdown, definitions FROM words WHERE status = $1;`, "in progress")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			gw.Write([]byte(`{ "message": "` + err.Error() + `"}`))
			return
		}
		defer rows.Close()

		words := make([]DrillWord, 0)
		for rows.Next() {
			word := DrillWord{}
			if err := rows.Scan(&word.BaseForm, &word.DateMarked, &word.Status, &word.Audio,
				&word.AudioStart, &word.AudioEnd, &word.DrillCountdown, &word.Definitions); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				gw.Write([]byte(`{ "message": "` + err.Error() + `"}`))
			}

			words = append(words, word)
		}

		json.NewEncoder(gw).Encode(bson.M{"words": words, "story_link": "", "story_title": "", "story_source": "All Stories In Progress"})
		return
	}

	var story_title string
	var story_source string
	var story_link string
	var wordIdsJson string

	row := sqldb.QueryRow(`SELECT title, source, link, words FROM catalog_stories WHERE id = $1;`, drillRequest.StoryId)
	err = row.Scan(&story_title, &story_source, &story_link, &wordIdsJson)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		gw.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}

	err = json.Unmarshal([]byte(wordIdsJson), &wordIds)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		gw.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}

	words := make([]DrillWord, len(wordIds))

	for i, id := range wordIds {
		word := &words[i]

		row := sqldb.QueryRow(`SELECT base_form, date_marked, status,
				audio, audio_start, audio_end, drill_countdown, definitions FROM words WHERE id = $1;`, id)

		err = row.Scan(&word.BaseForm, &word.DateMarked, &word.Status, &word.Audio,
			&word.AudioStart, &word.AudioEnd, &word.DrillCountdown, &word.Definitions)
		if err != nil && err != sql.ErrNoRows {
			w.WriteHeader(http.StatusInternalServerError)
			gw.Write([]byte(`{ "message": "` + "failure to get word info: " + err.Error() + `"}`))
			return
		}
	}

	json.NewEncoder(gw).Encode(bson.M{"words": words, "story_link": story_link, "story_title": story_title, "story_source": story_source})
}

func UpdateWord(w http.ResponseWriter, r *http.Request) {
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

	_, err = sqldb.Exec(`UPDATE words SET status = $1, date_marked = $2, audio = $3, audio_start = $4, audio_end = $5 WHERE base_form = $6;`,
		word.Status, word.DateMarked, word.Audio, word.AudioStart, word.AudioEnd, word.BaseForm)
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

	kanjiSet := make(map[string]bool)
	for _, ch := range kanjiStrings {
		kanjiSet[ch] = true
	}

	kanjiDefinitions := make(map[string]string)
	for ch := range kanjiSet {
		row := sqldb.QueryRow(`SELECT definitions FROM words WHERE base_form = $1;`, ch)
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
		kanjiDefinitions[ch] = def
	}

	json.NewEncoder(w).Encode(kanjiDefinitions)
}
