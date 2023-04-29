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
	"math"
	"regexp"
	"sort"

	"strings"
	"unicode/utf8"

	//"strconv"

	"log"
	"net/http"
	"os"

	"context"
	"time"

	"github.com/ikawaha/kagome-dict/ipa"
	"github.com/ikawaha/kagome/v2/tokenizer"

	"go.mongodb.org/mongo-driver/bson"
	// "go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	"database/sql"

	_ "github.com/mattn/go-sqlite3"

	//"github.com/hedhyw/rex/pkg/rex"  // regex builder

	"github.com/gorilla/mux"

	// Note: If connecting using the App Engine Flex Go runtime, use
	// "github.com/jackc/pgx/stdlib" instead, since v4 requires
	// Go modules which are not supported by App Engine Flex.
	_ "github.com/jackc/pgx/v4/stdlib"
)

// [END import]
// [START main_func]

var client *mongo.Client
var db *mongo.Database
var storiesCollection *mongo.Collection
var jmdictCollection *mongo.Collection
var kanjiCollection *mongo.Collection

var tok *tokenizer.Tokenizer

const SQL_FILE = "../testsql.db"
const USER_ID = 0 // TODO for now we hardcode for just one user
const INITIAL_COUNTDOWN = 7
const DRILL_COOLDOWN = 60 * 60 * 3 // in seconds
const DRILL_TYPE_KATAKANA = 1
const DRILL_TYPE_ICHIDAN = 2
const DRILL_TYPE_GODAN_SU = 8
const DRILL_TYPE_GODAN_RU = 16
const DRILL_TYPE_GODAN_U = 32
const DRILL_TYPE_GODAN_TSU = 64
const DRILL_TYPE_GODAN_KU = 128
const DRILL_TYPE_GODAN_GU = 256
const DRILL_TYPE_GODAN_MU = 512
const DRILL_TYPE_GODAN_BU = 1024
const DRILL_TYPE_GODAN_NU = 2048
const DRILL_TYPE_GODAN = DRILL_TYPE_GODAN_SU | DRILL_TYPE_GODAN_RU | DRILL_TYPE_GODAN_U | DRILL_TYPE_GODAN_TSU |
	DRILL_TYPE_GODAN_KU | DRILL_TYPE_GODAN_GU | DRILL_TYPE_GODAN_MU | DRILL_TYPE_GODAN_BU | DRILL_TYPE_GODAN_NU

func main() {
	var err error
	tok, err = tokenizer.New(ipa.Dict(), tokenizer.OmitBosEos())
	if err != nil {
		panic(err)
	}

	makeSqlDB()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err = mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))

	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()

	db = client.Database("JapaneseEnglish")
	storiesCollection = db.Collection("stories")
	jmdictCollection = db.Collection("jmdict")
	kanjiCollection = db.Collection("kanjidict")

	ctx, cancel = context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	err = client.Ping(ctx, readpref.Primary())

	//fmt.Print(client, err)

	// [START setting_port]
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}

	router := mux.NewRouter()

	router.HandleFunc("/read/{id}", ReadEndpoint).Methods("GET")
	router.HandleFunc("/word_search", PostWordSearch).Methods("POST")
	router.HandleFunc("/word_type_search", PostWordTypeSearch).Methods("POST")
	router.HandleFunc("/mark/{action}/{id}", MarkStoryEndpoint).Methods("GET")
	router.HandleFunc("/story", CreateStoryEndpoint).Methods("POST")
	router.HandleFunc("/story/{id}", GetStoryEndpoint).Methods("GET")
	router.HandleFunc("/story_retokenize/{id}", RetokenizeStoryEndpoint).Methods("GET")
	router.HandleFunc("/stories_list", GetStoriesListEndpoint).Methods("GET")
	router.HandleFunc("/kanji", KanjiEndpoint).Methods("POST")
	//router.HandleFunc("/add_word", AddWordEndpoint).Methods("POST")
	router.HandleFunc("/add_words", AddWordsEndpoint).Methods("POST")
	router.HandleFunc("/identify_verbs", IdentifyDrillTypeEndpoint).Methods("GET")
	router.HandleFunc("/drill", DrillEndpoint).Methods("POST")
	router.HandleFunc("/update_word", UpdateWordEndpoint).Methods("POST")
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("../static")))

	log.Printf("Listening on port %s", port)
	if err := http.ListenAndServe(":"+port, router); err != nil {
		log.Fatal(err)
	}
	// [END setting_port]
}

func makeSqlDB() {
	sqldb, err := sql.Open("sqlite3", SQL_FILE)
	if err != nil {
		log.Fatal(err)
	}
	defer sqldb.Close()

	statement, err := sqldb.Prepare("CREATE TABLE IF NOT EXISTS users (id INTEGER PRIMARY KEY, name TEXT)")
	if err != nil {
		log.Fatal(err)
	}
	if _, err := statement.Exec(); err != nil {
		log.Fatal(err)
	}

	// statement, err = sqldb.Prepare(`CREATE TABLE IF NOT EXISTS known_words
	// 	(id INTEGER PRIMARY KEY, user INTEGER NOT NULL, word INTEGER NOT NULL, FOREIGN KEY(user) REFERENCES users(id))`)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// if _, err := statement.Exec(); err != nil {
	// 	log.Fatal(err)
	// }

	statement, err = sqldb.Prepare(`CREATE TABLE IF NOT EXISTS words 
		(id INTEGER PRIMARY KEY, user INTEGER NOT NULL, 
			base_form TEXT NOT NULL, 
			countdown INTEGER NOT NULL,
			drill_count INTEGER NOT NULL,
			read_count INTEGER NOT NULL,
			date_last_read INTEGER NOT NULL,
			date_last_drill INTEGER NOT NULL,
			definitions TEXT NOT NULL,
			FOREIGN KEY(user) REFERENCES users(id))`)
	if err != nil {
		log.Fatal(err)
	}
	if _, err := statement.Exec(); err != nil {
		log.Fatal(err)
	}

	statement, err = sqldb.Prepare("CREATE TABLE IF NOT EXISTS stories (id INTEGER PRIMARY KEY, user INTEGER NOT NULL, story TEXT NOT NULL, state TEXT NOT NULL, FOREIGN KEY(user) REFERENCES users(id))")
	if err != nil {
		log.Fatal(err)
	}
	if _, err := statement.Exec(); err != nil {
		log.Fatal(err)
	}

	// statement, err = sqldb.Prepare("INSERT INTO users (name) VALUES (?)")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// statement.Exec("Mario Mario")
	// rows, _ := sqldb.Query("SELECT id, name FROM users")
	// var id int
	// var name string
	// for rows.Next() {
	// 	rows.Scan(&id, &name)
	// 	fmt.Println(strconv.Itoa(id) + ": " + name)
	// }
}

// [END main_func]

// [START indexHandler]

func KanjiEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Add("content-type", "application/json")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var str string
	json.NewDecoder(request.Body).Decode(&str)

	var re = regexp.MustCompile(`[\x{4E00}-\x{9FAF}]`)
	kanji := re.FindAllString(str, -1)

	if len(kanji) == 0 {
		json.NewEncoder(response).Encode(bson.M{
			"kanji": bson.A{}})
		return
	}

	arr := bson.A{}
	for _, k := range kanji {
		arr = append(arr, bson.D{{"literal", k}})
	}
	kanjiQuery := bson.D{{Key: "$or", Value: arr}}

	cursor, err := kanjiCollection.Find(ctx, kanjiQuery)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}
	defer cursor.Close(ctx)

	kanjiCharacters := make([]KanjiCharacter, 0)

	for cursor.Next(ctx) {
		var ch KanjiCharacter
		cursor.Decode(&ch)
		kanjiCharacters = append(kanjiCharacters, ch)
	}

	json.NewEncoder(response).Encode(bson.M{
		"kanji": kanjiCharacters})
}

// func AddWordEndpoint(response http.ResponseWriter, request *http.Request) {
// 	response.Header().Add("content-type", "application/json")
// 	var token JpToken
// 	json.NewDecoder(request.Body).Decode(&token)

// 	sqldb, err := sql.Open("sqlite3", SQL_FILE)
// 	if err != nil {
// 		response.WriteHeader(http.StatusInternalServerError)
// 		response.Write([]byte(`{ "message": "` + err.Error() + `"}`))
// 		return
// 	}
// 	defer sqldb.Close()

// 	if token.BaseForm == "" {
// 		token.BaseForm = token.Surface
// 	}

// 	rows, err := sqldb.Query(`SELECT id FROM words WHERE base_form = $1 AND user = $2;`, token.BaseForm, USER_ID)
// 	if err != nil {
// 		response.WriteHeader(http.StatusInternalServerError)
// 		response.Write([]byte(`{ "message": "` + "failure to get word: " + err.Error() + `"}`))
// 		return
// 	}
// 	exists := rows.Next()
// 	rows.Close()

// 	unixtime := time.Now().Unix()

// 	if !exists {
// 		fmt.Printf("\nadding word: %s %d\n", token.BaseForm, len(token.Definitions))

// 		defs := make([]JMDictEntry, len(token.Definitions))

// 		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
// 		defer cancel()

// 		for i, def := range token.Definitions {
// 			var entry JMDictEntry
// 			err := jmdictCollection.FindOne(ctx, bson.M{"_id": def}).Decode(&entry)
// 			if err != nil {
// 				response.WriteHeader(http.StatusInternalServerError)
// 				response.Write([]byte(`{ "message": "` + err.Error() + `"}`))
// 				return
// 			}
// 			defs[i] = entry
// 		}

// 		defsJson, err := json.Marshal(defs)
// 		if err != nil {
// 			response.WriteHeader(http.StatusInternalServerError)
// 			response.Write([]byte(`{ "message": "` + "failure to encode json: " + err.Error() + `"}`))
// 			return
// 		}

// 		_, err = sqldb.Exec(`INSERT INTO words (base_form, user, countdown, drill_count,
// 				read_count, date_last_read, date_last_drill, definitions) VALUES($1, $2, $3, $4, $5, $6, $7, $8);`,
// 			token.BaseForm, USER_ID, INITIAL_COUNTDOWN, 0, 0, unixtime, unixtime, defsJson)
// 		if err != nil {
// 			response.WriteHeader(http.StatusInternalServerError)
// 			response.Write([]byte(`{ "message": "` + "failure to insert word: " + err.Error() + `"}`))
// 			return
// 		}
// 	}

// 	json.NewEncoder(response).Encode(token)
// }

func PostWordSearch(response http.ResponseWriter, request *http.Request) {
	response.Header().Add("content-type", "application/json")

	var wordSearch WordSearch
	json.NewDecoder(request.Body).Decode(&wordSearch)

	fmt.Printf("\nword search: %v\n", wordSearch.Word)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var field string

	var re = regexp.MustCompile(`[\x{4E00}-\x{9FAF}]`)
	kanji := re.FindAllString(wordSearch.Word, -1)
	hasKanji := len(re.FindStringIndex(wordSearch.Word)) > 0
	if hasKanji { // if has kanji
		field = "kanji_spellings.kanji_spelling"
	} else {
		field = "readings.reading"
	}

	// only matches at start of string
	startOnlyQuery := bson.D{{field, bson.D{{"$regex", "^" + wordSearch.Word}}}}

	// only matches NOT at start of string
	notStartQuery := bson.D{
		{"$and",
			bson.A{
				bson.D{{field, bson.D{{"$not", bson.D{{"$regex", "^" + wordSearch.Word}}}}}},
				bson.D{{field, bson.D{{"$regex", wordSearch.Word}}}},
			},
		},
	}

	arr := bson.A{}
	for _, k := range kanji {
		arr = append(arr, bson.D{{"literal", k}})
	}
	kanjiQuery := bson.D{{Key: "$or", Value: arr}}

	cursor, err := jmdictCollection.Find(ctx, startOnlyQuery)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}
	defer cursor.Close(ctx)
	entriesStart := make([]JMDictEntry, 0)
	for cursor.Next(ctx) {
		var entry JMDictEntry
		cursor.Decode(&entry)
		entriesStart = append(entriesStart, entry)
	}

	cursor, err = jmdictCollection.Find(ctx, notStartQuery)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}
	defer cursor.Close(ctx)
	entriesMid := make([]JMDictEntry, 0)
	for cursor.Next(ctx) {
		var entry JMDictEntry
		cursor.Decode(&entry)
		entriesMid = append(entriesMid, entry)
	}

	kanjiCharacters := make([]KanjiCharacter, 0)
	if hasKanji {
		cursor, err = kanjiCollection.Find(ctx, kanjiQuery)
		if err != nil {
			response.WriteHeader(http.StatusInternalServerError)
			response.Write([]byte(`{ "message": "` + err.Error() + `"}`))
			return
		}
		defer cursor.Close(ctx)

		for cursor.Next(ctx) {
			var ch KanjiCharacter
			cursor.Decode(&ch)
			kanjiCharacters = append(kanjiCharacters, ch)
		}
	}

	sortResults(entriesStart, hasKanji, wordSearch.Word)
	sortResults(entriesMid, hasKanji, wordSearch.Word)

	nEntriesStart := len(entriesStart)
	if len(entriesStart) > 50 {
		entriesStart = entriesStart[:50]
	}

	nEntriesMid := len(entriesMid)
	if len(entriesMid) > 50 {
		entriesMid = entriesMid[:50]
	}

	json.NewEncoder(response).Encode(bson.M{
		"entries_start": entriesStart,
		"count_start":   nEntriesStart,
		"entries_mid":   entriesMid,
		"count_mid":     nEntriesMid,
		"kanji":         kanjiCharacters})
}

func PostWordTypeSearch(response http.ResponseWriter, request *http.Request) {
	response.Header().Add("content-type", "application/json")

	var wordSearch WordSearch
	json.NewDecoder(request.Body).Decode(&wordSearch)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// arr := bson.A{}
	// for _, k := range kanji {
	// 	arr = append(arr, bson.D{{"literal", k}})
	// }
	// kanjiQuery := bson.D{{Key: "$or", Value: arr}}

	cursor, err := jmdictCollection.Find(ctx, bson.D{{Key: "senses.parts_of_speech", Value: wordSearch.Word}})
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}
	defer cursor.Close(ctx)

	entries := make([]JMDictEntry, 0)
	for cursor.Next(ctx) {
		var entry JMDictEntry
		cursor.Decode(&entry)
		entries = append(entries, entry)
	}

	json.NewEncoder(response).Encode(bson.M{"entries": entries})
}

func sortResults(entries []JMDictEntry, hasKanji bool, word string) {
	// compute shortest readings and kanji spellings
	// TODO this could be stored in the DB
	for i := range entries {
		entries[i].ShortestKanjiSpelling = math.MaxInt32
		entries[i].ShortestReading = math.MaxInt32
		for _, ele := range entries[i].K_ele {
			if strings.Contains(ele.Keb, word) {
				count := utf8.RuneCountInString(ele.Keb)
				if count < entries[i].ShortestKanjiSpelling {
					entries[i].ShortestKanjiSpelling = count
				}
			}
		}
		for _, ele := range entries[i].R_ele {
			if strings.Contains(ele.Reb, word) {
				count := utf8.RuneCountInString(ele.Reb)
				if count < entries[i].ShortestReading {
					entries[i].ShortestReading = count
				}
			}
		}
	}

	if hasKanji {
		sort.Slice(entries, func(i, j int) bool {
			return entries[i].ShortestKanjiSpelling < entries[j].ShortestKanjiSpelling
		})
	} else {
		sort.Slice(entries, func(i, j int) bool {
			return entries[i].ShortestReading < entries[j].ShortestReading
		})
	}
}

// [END indexHandler]
// [END gae_go111_app]
