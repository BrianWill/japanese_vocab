package main

import (
	"compress/gzip"
	"database/sql"
	"encoding/json"

	// "math"
	"net/http"
	//"time"
	"go.mongodb.org/mongo-driver/bson"
)

func WordDrill(w http.ResponseWriter, r *http.Request) {
	dbPath := GetUserDb()

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

	var story_title string
	var wordIdsJson string
	var wordIds []int64

	row := sqldb.QueryRow(`SELECT title, words FROM catalog_stories WHERE id = $1;`, drillRequest.StoryId)
	err = row.Scan(&story_title, &wordIdsJson)
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

		row := sqldb.QueryRow(`SELECT base_form, date_marked, 
				audio, audio_start, audio_end FROM words WHERE id = $1;`, id)

		err = row.Scan(&word.BaseForm, &word.DateMarked, &word.Audio, &word.AudioStart, &word.AudioEnd)
		if err != nil && err != sql.ErrNoRows {
			w.WriteHeader(http.StatusInternalServerError)
			gw.Write([]byte(`{ "message": "` + "failure to get word info: " + err.Error() + `"}`))
			return
		}
	}

	json.NewEncoder(gw).Encode(bson.M{"words": words, "story_title": story_title})
}

func UpdateWord(w http.ResponseWriter, r *http.Request) {
	dbPath := GetUserDb()

	w.Header().Set("Content-Type", "application/json")

	var word WordUpdate
	json.NewDecoder(r.Body).Decode(&word)

	sqldb, err := sql.Open("sqlite3", dbPath)
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

	_, err = sqldb.Exec(`UPDATE words SET rank = $1, date_marked = $2, audio = $3, audio_start = $4, audio_end = $5, WHERE base_form = $6;`,
		word.Rank, word.DateMarked, word.Audio, word.AudioStart, word.AudioEnd, word.BaseForm)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + "failure to update word: " + err.Error() + `"}`))
		return
	}

	json.NewEncoder(w).Encode(word)
}
