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
	//"database/sql"
	"encoding/json"
	"fmt"
	"regexp"
	// "sort"

	// "strings"
	// "unicode/utf8"

	//"strconv"

	// "log"
	"net/http"
	// "os"

	"context"
	"time"

	// "github.com/ikawaha/kagome-dict/ipa"
	// "github.com/ikawaha/kagome/v2/tokenizer"

	"go.mongodb.org/mongo-driver/bson"
	// "go.mongodb.org/mongo-driver/bson/primitive"
	// "go.mongodb.org/mongo-driver/mongo"
	// "go.mongodb.org/mongo-driver/mongo/options"
	// "go.mongodb.org/mongo-driver/mongo/readpref"

	"database/sql"
	//_ "github.com/mattn/go-sqlite3"
	//"github.com/hedhyw/rex/pkg/rex"  // regex builder
	// "github.com/gorilla/mux"
	// Note: If connecting using the App Engine Flex Go runtime, use
	// "github.com/jackc/pgx/stdlib" instead, since v4 requires
	// Go modules which are not supported by App Engine Flex.
	//_ "github.com/jackc/pgx/v4/stdlib"
)

func DrillEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Add("content-type", "application/json")

	var drillRequest DrillRequest
	json.NewDecoder(request.Body).Decode(&drillRequest)
	drillRequest.Recency *= 60
	drillRequest.Wrong *= 60

	fmt.Println("ignore cooldown:", drillRequest.IgnoreCooldown)

	sqldb, err := sql.Open("sqlite3", SQL_FILE)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}
	defer sqldb.Close()

	rows, err := sqldb.Query(`SELECT base_form, countdown, drill_count, read_count, 
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
		err = rows.Scan(&word.BaseForm, &word.Countdown,
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

	total := len(words)

	activeCount := 0
	temp := make([]DrillWord, 0)
	t := time.Now().Unix()
	for _, w := range words {
		if !drillRequest.IgnoreCooldown && (t-w.DateLastDrill) < DRILL_COOLDOWN {
			continue
		}
		if drillRequest.Recency > 0 && (t-w.DateAdded) > drillRequest.Recency {
			continue
		}
		if drillRequest.Wrong > 0 && (t-w.DateLastWrong) > drillRequest.Wrong {
			continue
		}
		if w.Countdown <= 0 {
			continue
		}
		if !isDrillType(w.DrillType, drillRequest.Type) {
			continue
		}
		temp = append(temp, w)
		activeCount++
	}
	words = temp

	count := drillRequest.Count
	if count > 0 && count < len(words) {
		words = words[:count]
	}

	json.NewEncoder(response).Encode(bson.M{
		"wordCount":       len(words),
		"wordCountActive": activeCount,
		"wordCountTotal":  total,
		"words":           words})
}

func isDrillType(drillType int, requestedType string) bool {
	switch requestedType {
	case "all":
		return true
	case "ichidan":
		return (drillType & DRILL_TYPE_ICHIDAN) > 0
	case "godan":
		return (drillType & DRILL_TYPE_GODAN) > 0
	case "katakana":
		return (drillType & DRILL_TYPE_KATAKANA) > 0
	}
	return false
}

func AddWordsEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Add("content-type", "application/json")
	var tokens []JpToken
	json.NewDecoder(request.Body).Decode(&tokens)

	sqldb, err := sql.Open("sqlite3", SQL_FILE)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}
	defer sqldb.Close()

	var reHasKanji = regexp.MustCompile(`[\x{4E00}-\x{9FAF}]`)
	var reHasKana = regexp.MustCompile(`[あ-んア-ン]`)
	var reHasKatakana = regexp.MustCompile(`[ア-ン]`)

	wordIds := make([]int64, 0)

	for _, token := range tokens {
		hasKanji := len(reHasKanji.FindStringIndex(token.BaseForm)) > 0
		hasKana := len(reHasKana.FindStringIndex(token.BaseForm)) > 0
		if !hasKanji && !hasKana {
			continue
		}

		rows, err := sqldb.Query(`SELECT id FROM words WHERE base_form = $1 AND user = $2;`, token.BaseForm, USER_ID)
		if err != nil {
			response.WriteHeader(http.StatusInternalServerError)
			response.Write([]byte(`{ "message": "` + "failure to get word: " + err.Error() + `"}`))
			return
		}
		exists := rows.Next()
		rows.Close()

		unixtime := time.Now().Unix()

		var id int64
		if exists {
			rows.Scan(&id)
			wordIds = append(wordIds, id)
		} else {
			fmt.Printf("\nadding word: %s %d\n", token.BaseForm, len(token.Definitions))

			defs := make([]JMDictEntry, len(token.Definitions))

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			drillType := 0
			hasKatakana := len(reHasKatakana.FindStringIndex(token.BaseForm)) > 0
			if hasKatakana {
				drillType |= DRILL_TYPE_KATAKANA
			}

			for i, def := range token.Definitions {
				var entry JMDictEntry
				err := jmdictCollection.FindOne(ctx, bson.M{"_id": def}).Decode(&entry)
				if err != nil {
					response.WriteHeader(http.StatusInternalServerError)
					response.Write([]byte(`{ "message": "` + err.Error() + `"}`))
					return
				}
				defs[i] = entry
				for _, sense := range entry.Sense {
					drillType |= getVerbDrillType(sense)
				}
			}

			defsJson, err := json.Marshal(defs)
			if err != nil {
				response.WriteHeader(http.StatusInternalServerError)
				response.Write([]byte(`{ "message": "` + "failure to encode json: " + err.Error() + `"}`))
				return
			}

			insertResult, err := sqldb.Exec(`INSERT INTO words (base_form, user, countdown, drill_count, 
					read_count, date_last_read, date_last_drill, date_added, date_last_wrong, definitions, drill_type) VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11);`,
				token.BaseForm, USER_ID, INITIAL_COUNTDOWN, 0, 0, unixtime, 0, unixtime, 0, defsJson, drillType)
			if err != nil {
				response.WriteHeader(http.StatusInternalServerError)
				response.Write([]byte(`{ "message": "` + "failure to insert word: " + err.Error() + `"}`))
				return
			}

			id, err := insertResult.LastInsertId()
			if err != nil {
				response.WriteHeader(http.StatusInternalServerError)
				response.Write([]byte(`{ "message": "` + "failure to get id of inserted word: " + err.Error() + `"}`))
				return
			}

			wordIds = append(wordIds, id)
		}
	}

	// // add words to story
	wordIdsJson, err := json.Marshal(wordIds)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + "failure to marshall word ids: " + err.Error() + `"}`))
		return
	}
	// _, err = sqldb.Exec(`UPDATE stories SET words = $1 WHERE base_form = $2 AND user = $3;`,
	// 	wordIdsJson, storyId, USER_ID)

	json.NewEncoder(response).Encode(tokens)
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

func IdentifyDrillTypeEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Add("content-type", "application/json")

	sqldb, err := sql.Open("sqlite3", SQL_FILE)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}
	defer sqldb.Close()

	// var reHasKanji = regexp.MustCompile(`[\x{4E00}-\x{9FAF}]`)
	// var reHasKana = regexp.MustCompile(`[あ-んア-ン]`)
	var reHasKatakana = regexp.MustCompile(`[ア-ン]`)

	rows, err := sqldb.Query(`SELECT base_form, definitions FROM words WHERE user = $1;`, USER_ID)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + "failure to get words: " + err.Error() + `"}`))
		return
	}

	words := make([]string, 0)
	defs := make([]string, 0)
	for rows.Next() {
		var str string
		var def string
		rows.Scan(&str, &def)
		words = append(words, str)
		defs = append(defs, def)
	}
	rows.Close()

	fmt.Println("identifying drill type number words: ", len(words))

	for i, word := range words {
		fmt.Println("identifying drill type: ", word, defs[i])

		drillType := 0
		hasKatakana := len(reHasKatakana.FindStringIndex(word)) > 0
		if hasKatakana {
			drillType |= DRILL_TYPE_KATAKANA
		}

		var entries []JMDictEntry
		err := json.Unmarshal([]byte(defs[i]), &entries)
		if err != nil {
			response.WriteHeader(http.StatusInternalServerError)
			response.Write([]byte(`{ "message": "` + "failure to unmarshal dict entry: " + err.Error() + `"}`))
			return
		}

		for _, entry := range entries {
			for _, sense := range entry.Sense {
				drillType |= getVerbDrillType(sense)
			}
		}

		_, err = sqldb.Exec(`UPDATE words SET drill_type = $1 WHERE base_form = $2 AND user = $3;`,
			drillType, word, USER_ID)
		if err != nil {
			response.WriteHeader(http.StatusInternalServerError)
			response.Write([]byte(`{ "message": "` + "failure to update drill word: " + err.Error() + `"}`))
			return
		}
	}

	json.NewEncoder(response).Encode("done")
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
