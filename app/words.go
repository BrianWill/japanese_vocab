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

const DRILL_ALL_TOP_RANK = -1

func WordDrillEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Add("content-type", "application/json")

	var drillRequest DrillRequest
	json.NewDecoder(request.Body).Decode(&drillRequest)
	drillRequest.WrongWithin *= 60

	sqldb, err := sql.Open("sqlite3", SQL_FILE)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}
	defer sqldb.Close()

	rows, err := sqldb.Query(`SELECT id, base_form, countdown, drill_count, read_count, 
			date_last_read, date_last_drill, definitions, drill_type, date_last_wrong, date_added FROM words WHERE user = $1;`, USER_ID)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + "failure to get word: " + err.Error() + `"}`))
		return
	}
	defer rows.Close()

	words := make([]DrillWord, 0)
	for rows.Next() {
		var word DrillWord
		err = rows.Scan(&word.ID, &word.BaseForm, &word.Countdown,
			&word.DrillCount, &word.ReadCount,
			&word.DateLastRead, &word.DateLastDrill,
			&word.Definitions, &word.DrillType, &word.DateLastWrong, &word.DateAdded)
		if err != nil {
			response.WriteHeader(http.StatusInternalServerError)
			response.Write([]byte(`{ "message": "` + "failure to scan word: " + err.Error() + `"}`))
			return
		}
		words = append(words, word)
	}

	wordAllCount := len(words)

	storyWords, allStories, err := getStoryWords(drillRequest.StoryIds, response, sqldb)
	if err != nil {
		return
	}

	wordOffCooldownCount := 0
	temp := make([]DrillWord, 0)
	t := time.Now().Unix()
	for _, w := range words {
		isCountdownZero := w.Countdown <= 0
		isInStory := allStories || storyWords[w.ID]
		isOffCooldown := ((t-w.DateLastDrill) > DRILL_COOLDOWN && (t-w.DateLastWrong) > DRILL_COOLDOWN)
		isDrillType := isDrillType(w.DrillType, drillRequest.Type)

		if isOffCooldown && !isCountdownZero {
			wordOffCooldownCount++
		}

		if !isInStory || isCountdownZero || !isDrillType {
			continue
		}

		if drillRequest.WrongWithin > 0 {
			if (t - w.DateLastWrong) < drillRequest.WrongWithin {
				temp = append(temp, w)
			}
		} else if drillRequest.IgnoreCooldown || isOffCooldown {
			temp = append(temp, w)
		}
	}
	words = temp

	wordMatchCount := len(words)
	count := drillRequest.Count
	if count > 0 && count < len(words) {
		words = words[:count]
	}

	json.NewEncoder(response).Encode(bson.M{
		"wordOffCooldownCount": wordOffCooldownCount,
		"wordAllCount":         wordAllCount,
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
		rows, err := sqldb.Query(`SELECT id FROM stories WHERE user = $1 AND rank >= $2;`, USER_ID, -rankThreshold)
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

	for storyId, _ := range storyIdMap {
		rows, err := sqldb.Query(`SELECT words FROM stories WHERE user = $1 AND id = $2;`, USER_ID, storyId)
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

func UpdateWordEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Add("content-type", "application/json")
	var word DrillWord
	json.NewDecoder(request.Body).Decode(&word)

	sqldb, err := sql.Open("sqlite3", SQL_FILE)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}
	defer sqldb.Close()

	rows, err := sqldb.Query(`SELECT id FROM words WHERE base_form = $1 AND user = $2;`, word.BaseForm, USER_ID)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + "failure to get word: " + err.Error() + `"}`))
		return
	}
	exists := rows.Next()
	rows.Close()

	if !exists {
		return
	}

	_, err = sqldb.Exec(`UPDATE words SET countdown = $1, drill_count = $2, date_last_drill = $3, date_last_wrong = $4  WHERE base_form = $5 AND user = $6;`,
		word.Countdown, word.DrillCount, word.DateLastDrill, word.DateLastWrong, word.BaseForm, USER_ID)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + "failure to update drill word: " + err.Error() + `"}`))
		return
	}

	json.NewEncoder(response).Encode(word)
}

// [END indexHandler]
// [END gae_go111_app]
