package main

import (
	"fmt"
	"regexp"
	"sync"

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
)

var allKanji KanjiDict
var allEntries JMDict
var allEntriesByReading map[string][]*JMDictEntry
var allEntriesByKanjiSpellings map[string][]*JMDictEntry
var dictionaryLoaded bool

var importLock sync.Mutex // while any import is in process

var definitionsCache map[string][]JMDictEntry // base form to []JMDictEntry

var reHasKanji *regexp.Regexp

var tok *tokenizer.Tokenizer

const DRILL_CATEGORY_KATAKANA = 1
const DRILL_CATEGORY_ICHIDAN = 2
const DRILL_CATEGORY_GODAN_SU = 8
const DRILL_CATEGORY_GODAN_RU = 16
const DRILL_CATEGORY_GODAN_U = 32
const DRILL_CATEGORY_GODAN_TSU = 64
const DRILL_CATEGORY_GODAN_KU = 128
const DRILL_CATEGORY_GODAN_GU = 256
const DRILL_CATEGORY_GODAN_MU = 512
const DRILL_CATEGORY_GODAN_BU = 1024
const DRILL_CATEGORY_GODAN_NU = 2048
const DRILL_CATEGORY_KANJI = 4096
const DRILL_CATEGORY_GODAN = DRILL_CATEGORY_GODAN_SU | DRILL_CATEGORY_GODAN_RU | DRILL_CATEGORY_GODAN_U | DRILL_CATEGORY_GODAN_TSU |
	DRILL_CATEGORY_GODAN_KU | DRILL_CATEGORY_GODAN_GU | DRILL_CATEGORY_GODAN_MU | DRILL_CATEGORY_GODAN_BU | DRILL_CATEGORY_GODAN_NU

const STORY_TRACKING_PERIOD = 60 * 60 * 24 * 7 * 8 // 2 weeks

const MAIN_USER_DB_PATH = "./data/data.db"

func main() {
	var err error
	tok, err = tokenizer.New(ipa.Dict(), tokenizer.OmitBosEos())
	if err != nil {
		panic(err)
	}

	wd, err := os.Getwd()
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println("working dir:", wd)

	makeUserDB(MAIN_USER_DB_PATH)
	reHasKanji = regexp.MustCompile(`[\x{4E00}-\x{9FAF}]`)
	definitionsCache = make(map[string][]JMDictEntry)

	if len(os.Args) > 1 && os.Args[1] == "import" {
		loadDictionary()

		fmt.Println("db: ", MAIN_USER_DB_PATH)
		err := importSources(MAIN_USER_DB_PATH)
		if err != nil {
			log.Fatalln(err)
		}
		return
	}

	// [START setting_port]
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}

	router := mux.NewRouter()

	for _, s := range os.Args {
		if s == "dev" {
			router.Use(devMiddleware)
			fmt.Println("In dev mode")
		}
	}

	fmt.Println("db: ", MAIN_USER_DB_PATH)

	//router.Use(middleware)

	router.HandleFunc("/update_subtitles", UpdateSubtitles).Methods("POST")
	router.HandleFunc("/log_story", LogStory).Methods("POST")
	router.HandleFunc("/story/{id}", GetStory).Methods("GET")
	router.HandleFunc("/ip", GetIP).Methods("GET")
	router.HandleFunc("/stories", GetStories).Methods("GET")
	router.HandleFunc("/sources", GetSources).Methods("GET")
	router.HandleFunc("/import_source", ImportSource).Methods("POST")
	router.HandleFunc("/import_story", ImportStory).Methods("POST")
	router.HandleFunc("/remove_source", RemoveSource).Methods("POST")
	router.HandleFunc("/kanji", GetKanji).Methods("POST")
	router.HandleFunc("/words", GetWords).Methods("POST")
	router.HandleFunc("/update_word", UpdateWordArchiveState).Methods("POST")
	router.HandleFunc("/inc_words", IncWords).Methods("POST")
	router.HandleFunc("/open_transcript", OpenTranscript).Methods("POST")
	router.PathPrefix("/sources").Handler(http.StripPrefix("/sources/", http.FileServer(http.Dir("../sources"))))
	router.HandleFunc("/", GetMain).Methods("GET")
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("../static")))

	log.Printf("Running on port: %s", port)
	if err := http.ListenAndServe(":"+port, router); err != nil {
		log.Fatal(err)
	}

	//exec.Command("open", "http://localhost:8080/").Run()
	// [END setting_port]
}

func devMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
		h.ServeHTTP(w, r)
	})
}

// func middleware(h http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		w.Header().Set("Permissions-Policy", "fullscreen=self")
// 		h.ServeHTTP(w, r)
// 	})
// }

func loadDictionary() {
	// only load if dictionary has not already been loaded
	if dictionaryLoaded {
		return
	}

	start := time.Now()
	bytes, err := unzipSource("./data/kanji.zip")
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
	bytes, err = unzipSource("./data/entries.zip")
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

	duration = time.Since(start)
	fmt.Println("time to build entry maps: ", duration)

	dictionaryLoaded = true
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

func makeUserDB(path string) {
	sqldb, err := sql.Open("sqlite3", path)
	if err != nil {
		log.Fatal(err)
	}
	defer sqldb.Close()

	statement, err := sqldb.Prepare(`CREATE TABLE IF NOT EXISTS words 
		(id INTEGER PRIMARY KEY,
			base_form TEXT NOT NULL UNIQUE,
			archived INTEGER NOT NULL,
			repetitions INTEGER NOT NULL,
			category INTEGER NOT NULL,
			date_added INTEGER NOT NULL,
			date_last_rep INTEGER NOT NULL,
			definitions TEXT,
			kanji TEXT)`)
	if err != nil {
		log.Fatal(err)
	}
	if _, err := statement.Exec(); err != nil {
		log.Fatal(err)
	}

	statement, err = sqldb.Prepare(`CREATE TABLE IF NOT EXISTS stories
		(id INTEGER PRIMARY KEY,
			title TEXT NOT NULL,
			source TEXT NOT NULL,
			date TEXT,
			link TEXT,
			video TEXT,
			subtitles_en TEXT,
			subtitles_ja TEXT,
			subtitles_ja_offset REAL NOT NULL,
			subtitles_en_offset REAL NOT NULL,
			tracking_date INTEGER NOT NULL,
			log TEXT);`)
	if err != nil {
		log.Fatal(err)
	}
	if _, err := statement.Exec(); err != nil {
		log.Fatal(err)
	}

	statement, err = sqldb.Prepare(`CREATE TABLE IF NOT EXISTS stories_x_words 
		(id INTEGER PRIMARY KEY,
			story_id INTEGER,
			word_id INTEGER,
			FOREIGN KEY (story_id) REFERENCES stories(id)
			FOREIGN KEY (word_id) REFERENCES words(id));`)
	if err != nil {
		log.Fatal(err)
	}
	if _, err := statement.Exec(); err != nil {
		log.Fatal(err)
	}
}

// [END main_func]

// [START indexHandler]

func GetMain(response http.ResponseWriter, request *http.Request) {
	http.ServeFile(response, request, "../static/index.html")
}

func VacuumDb(userDbPath string) error {
	sqldb, err := sql.Open("sqlite3", userDbPath)
	if err != nil {
		return fmt.Errorf("failure to open user db: " + err.Error())
	}
	defer sqldb.Close()

	_, err = sqldb.Exec(`VACUUM;`)
	if err != nil {
		return fmt.Errorf("failure to vacuum user db: " + err.Error())
	}

	return nil
}
