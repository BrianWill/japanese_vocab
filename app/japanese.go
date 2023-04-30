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
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"sort"
	"strings"
	"unicode/utf8"

	"log"
	"net/http"
	"os"

	"context"
	"database/sql"
	"time"

	"github.com/ikawaha/kagome-dict/ipa"
	"github.com/ikawaha/kagome/v2/tokenizer"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	// Note: If connecting using the App Engine Flex Go runtime, use
	// "github.com/jackc/pgx/stdlib" instead, since v4 requires
	// Go modules which are not supported by App Engine Flex.
	_ "github.com/jackc/pgx/v4/stdlib"
)

// [END import]
// [START main_func]

var client *mongo.Client
var db *mongo.Database
var jmdictCollection *mongo.Collection
var kanjiCollection *mongo.Collection
var allKanji KanjiDict
var allEntries JMDict

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
	jmdictCollection = db.Collection("jmdict")
	kanjiCollection = db.Collection("kanjidict")

	ctx, cancel = context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	err = client.Ping(ctx, readpref.Primary())

	start := time.Now()
	bytes, err := unzipSource("../kanji.zip")
	if err != nil {
		panic(err)
	}
	err = bson.Unmarshal(bytes, &allKanji)
	if err != nil {
		panic(err)
	}
	duration := time.Since(start)
	fmt.Println("time to load kanji: ", duration)

	start = time.Now()
	bytes, err = unzipSource("../entries.zip")
	if err != nil {
		panic(err)
	}
	duration = time.Since(start)
	fmt.Println("time to unzip entries: ", duration)

	start = time.Now()
	err = bson.Unmarshal(bytes, &allEntries)
	if err != nil {
		panic(err)
	}
	duration = time.Since(start)
	fmt.Println("time to load entries: ", duration)

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
	router.HandleFunc("/create_story", CreateStoryEndpoint).Methods("POST")
	//router.HandleFunc("/load_stories", LoadStoriesEndpoint).Methods("GET")
	router.HandleFunc("/story/{id}", GetStoryEndpoint).Methods("GET")
	router.HandleFunc("/stories_list", GetStoriesListEndpoint).Methods("GET")
	router.HandleFunc("/kanji", KanjiEndpoint).Methods("POST")
	router.HandleFunc("/dump_kanji", DumpKanjiEndpoint).Methods("GET")
	router.HandleFunc("/dump_entries", DumpEntriesEndpoint).Methods("GET")
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

	statement, err = sqldb.Prepare(`CREATE TABLE IF NOT EXISTS stories 
		(id INTEGER PRIMARY KEY, user INTEGER NOT NULL,
			state   TEXT NOT NULL,
			words	TEXT NOT NULL,
			content	TEXT,
			title	TEXT,
			link	TEXT,
			tokens	TEXT,
			FOREIGN KEY(user) REFERENCES users(id))`)
	if err != nil {
		log.Fatal(err)
	}
	if _, err := statement.Exec(); err != nil {
		log.Fatal(err)
	}
}

// [END main_func]

// [START indexHandler]

func KanjiEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Add("content-type", "application/json")

	// ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	// defer cancel()

	var str string
	json.NewDecoder(request.Body).Decode(&str)

	var re = regexp.MustCompile(`[\x{4E00}-\x{9FAF}]`)
	kanji := re.FindAllString(str, -1)

	json.NewEncoder(response).Encode(bson.M{"kanji": getKanji(kanji)})
}

func getKanji(characters []string) []KanjiCharacter {
	kanjiSet := make(map[string]KanjiCharacter)
	for _, ch := range characters {
		for _, k := range allKanji.Characters {
			if k.Literal == ch {
				kanjiSet[ch] = k
				break
			}
		}
	}

	kanji := make([]KanjiCharacter, 0)
	for _, k := range kanjiSet {
		kanji = append(kanji, k)
	}
	return kanji
}

func DumpEntriesEndpoint(response http.ResponseWriter, request *http.Request) {
	dumpEntries := func() ([]JMDictEntry, error) {
		ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
		defer cancel()

		cursor, err := jmdictCollection.Find(ctx, bson.D{{}})
		if err != nil {
			return nil, err
		}
		defer cursor.Close(ctx)

		entries := make([]JMDictEntry, 0)

		for cursor.Next(ctx) {
			var entry JMDictEntry
			cursor.Decode(&entry)
			entries = append(entries, entry)
		}

		dict := JMDict{Entries: entries}

		bytes, err := bson.Marshal(dict)
		if err != nil {
			return nil, err
		}

		err = os.WriteFile("../entries.bson", bytes, 0644)
		if err != nil {
			return nil, err
		}
		return entries, nil
	}

	kanji, err := dumpEntries()
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}

	json.NewEncoder(response).Encode(bson.M{
		"kanji": len(kanji)})
}

func DumpKanjiEndpoint(response http.ResponseWriter, request *http.Request) {
	dumpKanji := func() ([]KanjiCharacter, error) {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		cursor, err := kanjiCollection.Find(ctx, bson.D{{}})
		if err != nil {
			return nil, err
		}
		defer cursor.Close(ctx)

		kanjiCharacters := make([]KanjiCharacter, 0)

		for cursor.Next(ctx) {
			var ch KanjiCharacter
			cursor.Decode(&ch)
			//fmt.Printf("kanji: %s", ch.Literal)
			kanjiCharacters = append(kanjiCharacters, ch)
		}

		dict := KanjiDict{Characters: kanjiCharacters}

		bytes, err := bson.Marshal(dict)
		if err != nil {
			return nil, err
		}

		err = os.WriteFile("../kanji.bson", bytes, 0644)
		if err != nil {
			return nil, err
		}
		return kanjiCharacters, nil
	}

	kanji, err := dumpKanji()
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}

	json.NewEncoder(response).Encode(bson.M{
		"kanji": len(kanji)})
}

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

	kanjiCharacters := getKanji(kanji)

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

	entries := make([]JMDictEntry, 0)

	start := time.Now()
	for _, entry := range allEntries.Entries {
		for _, sense := range entry.Senses {
			include := false
			for _, pos := range sense.Pos {
				if pos == wordSearch.Word {
					include = true
					break
				}
			}
			if include {
				entries = append(entries, entry)
				break
			}
		}
	}
	duration := time.Since(start)
	fmt.Println("time to search entries: ", duration)

	json.NewEncoder(response).Encode(bson.M{"entries": entries})
}

func sortResults(entries []JMDictEntry, hasKanji bool, word string) {
	// compute shortest readings and kanji spellings
	// TODO this could be stored in the DB
	for i := range entries {
		entries[i].ShortestKanjiSpelling = math.MaxInt32
		entries[i].ShortestReading = math.MaxInt32
		for _, ele := range entries[i].KanjiSpellings {
			if strings.Contains(ele.KanjiSpelling, word) {
				count := utf8.RuneCountInString(ele.KanjiSpelling)
				if count < entries[i].ShortestKanjiSpelling {
					entries[i].ShortestKanjiSpelling = count
				}
			}
		}
		for _, ele := range entries[i].Readings {
			if strings.Contains(ele.Reading, word) {
				count := utf8.RuneCountInString(ele.Reading)
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
