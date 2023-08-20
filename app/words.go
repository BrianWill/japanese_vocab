// Copyright 2019 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// [START gae_go111_app]

package main

// [START import]
import (
	"compress/gzip"
	"database/sql"
	"encoding/json"
	"fmt"

	// "math"
	"net/http"
	//"time"
	"go.mongodb.org/mongo-driver/bson"
)

func WordDrill(w http.ResponseWriter, r *http.Request) {
	dbPath, redirect, err := GetUserDb(w, r)
	if redirect {
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("content-encoding", "gzip")

	gw := gzip.NewWriter(w)
	defer gw.Close()

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		gw.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}

	var drillRequest DrillRequest
	json.NewDecoder(r.Body).Decode(&drillRequest)

	sqldb, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		gw.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}
	defer sqldb.Close()

	baseForms, err := getStoryWords(drillRequest.StoryIds, sqldb)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		gw.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}

	rows, err := sqldb.Query(`SELECT id, base_form, rank, date_marked, category FROM words;`)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		gw.Write([]byte(`{ "message": "` + "failure to get word: " + err.Error() + `"}`))
		return
	}
	defer rows.Close()

	words := make([]DrillWord, 0)
	for rows.Next() {
		var word DrillWord
		err = rows.Scan(&word.ID, &word.BaseForm,
			&word.Rank, &word.DateMarked,
			&word.Category)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			gw.Write([]byte(`{ "message": "` + "failure to scan word: " + err.Error() + `"}`))
			return
		}
		if len(drillRequest.StoryIds) == 0 {
			words = append(words, word)
		} else if _, ok := baseForms[word.BaseForm]; ok {
			words = append(words, word)
		}
	}

	wordInfoMap := make(map[string]WordInfo)

	for _, word := range words {
		wordInfo := WordInfo{
			Definitions: getDefinitions(word.BaseForm),
		}

		row := sqldb.QueryRow(`SELECT rank, date_marked FROM words WHERE base_form = $1;`, word.BaseForm)

		err = row.Scan(&wordInfo.Rank, &wordInfo.DateMarked)
		if err != nil && err != sql.ErrNoRows {
			w.WriteHeader(http.StatusInternalServerError)
			gw.Write([]byte(`{ "message": "` + "failure to get word info: " + err.Error() + `"}`))
			return
		}

		wordInfoMap[word.BaseForm] = wordInfo
	}

	json.NewEncoder(gw).Encode(bson.M{"words": words, "wordInfoMap": wordInfoMap})
}

func getStoryWords(storyIds []int64, sqldb *sql.DB) (map[string]bool, error) {
	baseForms := make(map[string]bool)

	for _, id := range storyIds {
		rows, err := sqldb.Query(`SELECT lines FROM stories WHERE id = $1`, id)
		if err != nil {
			return nil, fmt.Errorf("failure to get story words: " + err.Error())
		}
		defer rows.Close()

		for rows.Next() {
			var linesJSON string
			var lines []Line
			err = rows.Scan(&linesJSON)
			if err != nil {
				return nil, fmt.Errorf("failure to scan story line: " + err.Error())
			}
			err := json.Unmarshal([]byte(linesJSON), &lines)
			if err != nil {
				return nil, fmt.Errorf("failure to unmarshall story lines: " + err.Error())
			}

			for _, line := range lines {
				for _, kanji := range line.Kanji {
					baseForms[kanji.Character] = true
				}
				for _, word := range line.Words {
					baseForms[word.BaseForm] = true
				}
			}
		}
	}

	return baseForms, nil
}

func UpdateWord(w http.ResponseWriter, r *http.Request) {
	dbPath, redirect, err := GetUserDb(w, r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}
	if redirect {
		return
	}

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

	_, err = sqldb.Exec(`UPDATE words SET rank = $1, date_marked = $2 WHERE base_form = $3;`,
		word.Rank, word.DateMarked, word.BaseForm)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + "failure to update word: " + err.Error() + `"}`))
		return
	}

	json.NewEncoder(w).Encode(word)
}

// [END indexHandler]
// [END gae_go111_app]
