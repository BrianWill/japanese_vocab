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
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

func WordDrill(response http.ResponseWriter, request *http.Request) {
	dbPath, redirect, err := GetUserDb(response, request)
	if redirect {
		return
	}
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `"}`))
	}

	response.Header().Set("Content-Type", "application/json")

	var drillRequest DrillRequest
	json.NewDecoder(request.Body).Decode(&drillRequest)

	sqldb, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}
	defer sqldb.Close()

	rows, err := sqldb.Query(`SELECT id, base_form, rank, drill_count, 
			date_last_read, date_last_drill, definitions, drill_type, date_last_wrong, date_added FROM words;`)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + "failure to get word: " + err.Error() + `"}`))
		return
	}
	defer rows.Close()

	words := make([]DrillWord, 0)
	for rows.Next() {
		var word DrillWord
		err = rows.Scan(&word.ID, &word.BaseForm,
			&word.Rank, &word.DrillCount,
			&word.DateLastRead, &word.DateLastDrill,
			&word.Definitions, &word.DrillType,
			&word.DateLastWrong, &word.DateAdded)
		if err != nil {
			response.WriteHeader(http.StatusInternalServerError)
			response.Write([]byte(`{ "message": "` + "failure to scan word: " + err.Error() + `"}`))
			return
		}
		words = append(words, word)
	}

	countAllWords := len(words)

	storyWords, allStories, err := getStoryWords(drillRequest.StoryIds, response, sqldb)
	if err != nil {
		return
	}

	countsByRank := [4]int{}
	cooldownCountsByRank := [4]int{}

	countWordsInStory := 0

	includeOnCooldown := false
	includeOffCooldown := false
	switch drillRequest.Filter {
	case DRILL_FILTER_ON_COOLDOWN:
		includeOnCooldown = true
	case DRILL_FILTER_OFF_COOLDOWN:
		includeOffCooldown = true
	case DRILL_FILTER_ALL:
		includeOnCooldown = true
		includeOffCooldown = true
	}

	cooldowns := [5]int64{0, DRILL_COOLDOWN_RANK_1, DRILL_COOLDOWN_RANK_2, DRILL_COOLDOWN_RANK_3, DRILL_COOLDOWN_RANK_4}

	temp := make([]DrillWord, 0)
	t := time.Now().Unix()
	for _, w := range words {
		isInStory := allStories || storyWords[w.ID]
		var cooldown = cooldowns[w.Rank]
		isOnCooldown := ((t-w.DateLastDrill) < cooldown || (t-w.DateLastWrong) < cooldown)
		isDrillType := isDrillType(w.DrillType, drillRequest.Type)

		if !isInStory {
			continue
		}

		countWordsInStory++

		if isOnCooldown {
			cooldownCountsByRank[w.Rank-1]++
		} else {
			countsByRank[w.Rank-1]++
		}

		if !isDrillType {
			continue
		}

		if w.Rank < drillRequest.MinRank || w.Rank > drillRequest.MaxRank {
			continue
		}

		if isOnCooldown && !includeOnCooldown {
			continue
		}

		if !isOnCooldown && !includeOffCooldown {
			continue
		}

		temp = append(temp, w)
	}
	words = temp

	wordMatchCount := len(words)
	if !includeOnCooldown {
		count := drillRequest.Count
		if count > 0 && count < len(words) {
			words = words[:count]
		}
	}

	json.NewEncoder(response).Encode(bson.M{
		"countAllWords":        countAllWords,
		"countWordsInStory":    countWordsInStory,
		"countsByRank":         countsByRank,
		"cooldownCountsByRank": cooldownCountsByRank,
		"words":                words,
		"wordMatchCount":       wordMatchCount})
}

func getStoryWords(storyIds []int64, response http.ResponseWriter, sqldb *sql.DB) (map[int64]bool, bool, error) {
	storyIdMap := make(map[int64]bool)
	storyWords := make(map[int64]bool)

	var rankThreshold int64 = math.MinInt64
	for _, id := range storyIds {
		if id == 0 {
			// return true if all stories included
			return nil, true, nil
		}
		if id < 0 && id > rankThreshold {
			rankThreshold = id
		}
		if id > 0 {
			storyIdMap[id] = true
			fmt.Println("story id", id)
		}
	}

	if rankThreshold != math.MinInt64 {
		rows, err := sqldb.Query(`SELECT id FROM stories WHERE status >= $1;`, -rankThreshold)
		if err != nil {
			response.WriteHeader(http.StatusInternalServerError)
			response.Write([]byte(`{ "message": "` + "failure to get story words: " + err.Error() + `"}`))
			return nil, false, err
		}
		defer rows.Close()

		for rows.Next() {
			var id int64
			err = rows.Scan(&id)
			if err != nil {
				response.WriteHeader(http.StatusInternalServerError)
				response.Write([]byte(`{ "message": "` + "failure to scan story id: " + err.Error() + `"}`))
				return nil, false, err
			}
			storyIdMap[id] = true
		}
	}

	for storyId := range storyIdMap {
		rows, err := sqldb.Query(`SELECT words FROM stories WHERE id = $1;`, storyId)
		if err != nil {
			response.WriteHeader(http.StatusInternalServerError)
			response.Write([]byte(`{ "message": "` + "failure to get story words: " + err.Error() + `"}`))
			return nil, false, err
		}
		defer rows.Close()

		for rows.Next() {
			var wordStr string
			err = rows.Scan(&wordStr)
			if err != nil {
				response.WriteHeader(http.StatusInternalServerError)
				response.Write([]byte(`{ "message": "` + "failure to scan word: " + err.Error() + `"}`))
				return nil, false, err
			}
			var words []int64
			err = json.Unmarshal([]byte(wordStr), &words)
			if err != nil {
				response.WriteHeader(http.StatusInternalServerError)
				response.Write([]byte(`{ "message": "` + "failure to unmarhsall story words: " + err.Error() + `"}`))
				return nil, false, err
			}

			for _, word := range words {
				storyWords[word] = true
			}
		}
	}
	return storyWords, false, nil
}

func isDrillType(drillType int, requestedType string) bool {
	switch requestedType {
	case "all":
		return true
	case "kanji":
		return (drillType & DRILL_TYPE_KANJI) > 0
	case "ichidan":
		return (drillType & DRILL_TYPE_ICHIDAN) > 0
	case "godan":
		return (drillType & DRILL_TYPE_GODAN) > 0
	case "katakana":
		return (drillType & DRILL_TYPE_KATAKANA) > 0
	}
	return false
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

	_, err = sqldb.Exec(`UPDATE words SET rank = $1, date_last_drill = $2 WHERE base_form = $3;`,
		word.Rank, word.DateLastDrill, word.BaseForm)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + "failure to update word: " + err.Error() + `"}`))
		return
	}

	json.NewEncoder(w).Encode(word)
}

// [END indexHandler]
// [END gae_go111_app]
