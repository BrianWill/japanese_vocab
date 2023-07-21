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
	"os/exec"
	"regexp"
	"sort"
	"strings"
	"unicode/utf8"

	"log"
	"net/http"
	"os"

	"database/sql"
	"time"

	"github.com/ikawaha/kagome-dict/ipa"
	"github.com/ikawaha/kagome/v2/tokenizer"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
	"go.mongodb.org/mongo-driver/bson"
	// Note: If connecting using the App Engine Flex Go runtime, use
	// "github.com/jackc/pgx/stdlib" instead, since v4 requires
	// Go modules which are not supported by App Engine Flex.
	//_ "github.com/jackc/pgx/v4/stdlib"
)

// [END import]
// [START main_func]

var allKanji KanjiDict
var allEntries JMDict
var allEntriesByReading map[string][]*JMDictEntry
var allEntriesByKanjiSpellings map[string][]*JMDictEntry

var tok *tokenizer.Tokenizer

const SQL_FILE = "../testsql.db"
const USER_ID = 0 // TODO for now we hardcode for just one user

const DRILL_COOLDOWN_RANK_4 = 60 * 60 * 3       // 3 hours in seconds
const DRILL_COOLDOWN_RANK_3 = 60 * 60 * 24 * 2  // 2 days in seconds
const DRILL_COOLDOWN_RANK_2 = 60 * 60 * 24 * 7  // 7 days in seconds
const DRILL_COOLDOWN_RANK_1 = 60 * 60 * 24 * 30 // 30 days weeks in seconds
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
const DRILL_TYPE_KANJI = 4096
const DRILL_TYPE_GODAN = DRILL_TYPE_GODAN_SU | DRILL_TYPE_GODAN_RU | DRILL_TYPE_GODAN_U | DRILL_TYPE_GODAN_TSU |
	DRILL_TYPE_GODAN_KU | DRILL_TYPE_GODAN_GU | DRILL_TYPE_GODAN_MU | DRILL_TYPE_GODAN_BU | DRILL_TYPE_GODAN_NU

func main() {
	var err error
	tok, err = tokenizer.New(ipa.Dict(), tokenizer.OmitBosEos())
	if err != nil {
		panic(err)
	}

	makeSqlDB()

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

	start = time.Now()
	buildEntryMaps()
	if err != nil {
		panic(err)
	}
	duration = time.Since(start)
	fmt.Println("time to build entry maps: ", duration)

	// [START setting_port]
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}

	router := mux.NewRouter()

	router.HandleFunc("/add_log_event/{id}", AddLogEvent).Methods("GET")
	router.HandleFunc("/remove_log_event/{id}", RemoveLogEvent).Methods("GET")
	router.HandleFunc("/log_events/{since}", GetLogEvents).Methods("GET")
	router.HandleFunc("/read/{id}", ReadEndpoint).Methods("GET")
	router.HandleFunc("/word_search", PostWordSearch).Methods("POST")
	router.HandleFunc("/word_type_search", PostWordTypeSearch).Methods("POST")
	router.HandleFunc("/update_story", UpdateStoryEndpoint).Methods("POST")
	router.HandleFunc("/create_story", CreateStoryEndpoint).Methods("POST")
	router.HandleFunc("/retokenize_story", RetokenizeStoryEndpoint).Methods("POST")
	router.HandleFunc("/load_stories", LoadStoriesFromDumpEndpoint).Methods("GET")
	router.HandleFunc("/story/{id}", GetStoryEndpoint).Methods("GET")
	router.HandleFunc("/stories_list", GetStoriesListEndpoint).Methods("GET")
	router.HandleFunc("/kanji", KanjiEndpoint).Methods("POST")
	router.HandleFunc("/words", WordDrillEndpoint).Methods("POST")
	router.HandleFunc("/update_word", UpdateWordEndpoint).Methods("POST")
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("../static")))

	log.Printf("Running on http://localhost:%s", port)
	if err := http.ListenAndServe(":"+port, router); err != nil {
		log.Fatal(err)
	}

	exec.Command("open", "http://localhost:8080/").Run()
	// [END setting_port]
}

func buildEntryMaps() {
	allEntriesByKanjiSpellings = make(map[string][]*JMDictEntry)
	allEntriesByReading = make(map[string][]*JMDictEntry)
	for i, entry := range allEntries.Entries {
		for _, k_ele := range entry.KanjiSpellings {
			if k_ele.KanjiSpelling != "" {
				if entries, ok := allEntriesByKanjiSpellings[k_ele.KanjiSpelling]; ok {
					allEntriesByKanjiSpellings[k_ele.KanjiSpelling] = append(entries, &allEntries.Entries[i])
				} else {
					allEntriesByKanjiSpellings[k_ele.KanjiSpelling] = []*JMDictEntry{&allEntries.Entries[i]}
				}
			}
		}

		for _, r_ele := range entry.Readings {
			if r_ele.Reading != "" {
				if entries, ok := allEntriesByReading[r_ele.Reading]; ok {
					allEntriesByReading[r_ele.Reading] = append(entries, &allEntries.Entries[i])
				} else {
					allEntriesByReading[r_ele.Reading] = []*JMDictEntry{&allEntries.Entries[i]}
				}
			}
		}
	}
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
			drill_count INTEGER NOT NULL,
			date_last_read INTEGER NOT NULL,
			date_last_drill INTEGER NOT NULL,
			date_last_wrong INTEGER NOT NULL,
			date_added INTEGER NOT NULL,
			rank INTEGER NOT NULL,
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
			words	TEXT NOT NULL,
			content	TEXT,
			title	TEXT,
			link	TEXT,
			tokens	TEXT,
			status INTEGER NOT NULL,
			date_last_read INTEGER NOT NULL,
			date_added INTEGER NOT NULL,
			FOREIGN KEY(user) REFERENCES users(id))`)
	if err != nil {
		log.Fatal(err)
	}
	if _, err := statement.Exec(); err != nil {
		log.Fatal(err)
	}

	statement, err = sqldb.Prepare(`CREATE TABLE IF NOT EXISTS log_events 
	(id INTEGER PRIMARY KEY, 
		user INTEGER NOT NULL,
		story INTEGER NOT NULL,
		date INTEGER NOT NULL,
		FOREIGN KEY (story) REFERENCES stories (id),
		FOREIGN KEY (user) REFERENCES users (id))`)
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

func PostWordSearch(response http.ResponseWriter, request *http.Request) {
	response.Header().Add("content-type", "application/json")

	var wordSearch WordSearch
	json.NewDecoder(request.Body).Decode(&wordSearch)

	fmt.Printf("\nword search: %v\n", wordSearch.Word)

	var re = regexp.MustCompile(`[\x{4E00}-\x{9FAF}]`)
	kanji := re.FindAllString(wordSearch.Word, -1)
	hasKanji := len(re.FindStringIndex(wordSearch.Word)) > 0

	reStart := regexp.MustCompile(`^` + wordSearch.Word)
	entriesStart := make([]JMDictEntry, 0)
	entriesMid := make([]JMDictEntry, 0)

	if hasKanji { // if has kanji
		for _, entry := range allEntries.Entries {
			start := false
			interior := false
			for _, kanjiSpelling := range entry.KanjiSpellings {
				if reStart.MatchString(kanjiSpelling.KanjiSpelling) { // todo regex "^word"
					start = true
				} else if strings.Contains(kanjiSpelling.KanjiSpelling, wordSearch.Word) {
					interior = true
				}
			}
			if start {
				entriesStart = append(entriesStart, entry)
			} else if interior {
				entriesMid = append(entriesMid, entry)
			}
		}
	} else {
		for _, entry := range allEntries.Entries {
			start := false
			interior := false
			for _, reading := range entry.Readings {
				if reStart.MatchString(reading.Reading) { // todo regex "^word"
					start = true
				} else if strings.Contains(reading.Reading, wordSearch.Word) {
					interior = true
				}
			}
			if start {
				entriesStart = append(entriesStart, entry)
			} else if interior {
				entriesMid = append(entriesMid, entry)
			}
		}
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
