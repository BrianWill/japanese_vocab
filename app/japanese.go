package main

import (
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"sort"
	"strings"
	"unicode/utf8"

	"crypto/md5"
	"encoding/hex"
	"log"
	"net/http"
	"os"

	"database/sql"
	"time"

	"github.com/ikawaha/kagome-dict/ipa"
	"github.com/ikawaha/kagome/v2/tokenizer"
	"golang.org/x/crypto/bcrypt"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	_ "github.com/mattn/go-sqlite3"
	"go.mongodb.org/mongo-driver/bson"
)

var allKanji KanjiDict
var allEntries JMDict
var allEntriesByReading map[string][]*JMDictEntry
var allEntriesByKanjiSpellings map[string][]*JMDictEntry

var definitionsCache map[string][]JMDictEntry // base form to []JMDictEntry
//var definitionsJSONCache map[string]string    // base form to JSON string of []JMDictEntry

var reHasKanji *regexp.Regexp

var sessionStore *sessions.CookieStore

const SESSION_KEY = "supersecret" // todo insecure; use env variable instead

var tok *tokenizer.Tokenizer

const SQL_USERS_FILE = "../users.db"
const SALT = "QWOpVRp6SObKeO6bBth5"

const DRILL_COOLDOWN_RANK_4 = 60 * 60 * 24 * 1000 // 1000 days in seconds
const DRILL_COOLDOWN_RANK_3 = 60 * 60 * 24 * 30   // 30 days in seconds
const DRILL_COOLDOWN_RANK_2 = 60 * 60 * 24 * 4    // 4 days in seconds
const DRILL_COOLDOWN_RANK_1 = 60 * 60 * 5         // 5 hours in second
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

func main() {
	var err error
	tok, err = tokenizer.New(ipa.Dict(), tokenizer.OmitBosEos())
	if err != nil {
		panic(err)
	}

	//cookieStore = sessions.NewCookieStore([]byte(os.Getenv("SESSION_KEY")))
	sessionStore = sessions.NewCookieStore([]byte(SESSION_KEY)) // todo insecure

	makeMainDB()
	initialize()

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

	for _, s := range os.Args {
		if s == "-dev" {
			router.Use(devMiddleware)
		}
	}

	router.HandleFunc("/add_log_event/{id}", AddLogEvent).Methods("GET")
	router.HandleFunc("/remove_log_event/{id}", RemoveLogEvent).Methods("GET")
	router.HandleFunc("/log_events/{since}", GetLogEvents).Methods("GET")
	router.HandleFunc("/loginauth", PostLoginAuth).Methods("POST")
	router.HandleFunc("/logout", PostLogout).Methods("POST")
	router.HandleFunc("/register", PostRegisterUser).Methods("POST")
	router.HandleFunc("/word_search", PostWordSearch).Methods("POST")
	router.HandleFunc("/word_type_search", PostWordTypeSearch).Methods("POST")
	router.HandleFunc("/update_story_status", UpdateStoryStatus).Methods("POST")
	router.HandleFunc("/update_story_counts", UpdateStoryCounts).Methods("POST")
	router.HandleFunc("/create_story", CreateStory).Methods("POST")
	router.HandleFunc("/retokenize_story", RetokenizeStory).Methods("POST")
	router.HandleFunc("/story/{id}", GetStory).Methods("GET")
	router.HandleFunc("/story_consolidate_line", ConsolidateLine).Methods("POST")
	router.HandleFunc("/story_split_line", SplitLine).Methods("POST")
	router.HandleFunc("/story_set_timestamp", SetTimestamp).Methods("POST")
	router.HandleFunc("/stories_list", GetStoriesList).Methods("GET")
	router.HandleFunc("/kanji", Kanji).Methods("POST")
	router.HandleFunc("/words", WordDrill).Methods("POST")
	router.HandleFunc("/update_word", UpdateWord).Methods("POST")
	router.HandleFunc("/schedule_story/{storyId}", EnqueueStory).Methods("GET")
	router.HandleFunc("/get_story_todo", GetStoriesTodo).Methods("GET")
	router.HandleFunc("/refresh_todo", RefreshQueue).Methods("GET")
	router.HandleFunc("/log_scheduled_story/{id}/{storyId}", MarkQueuedStory).Methods("GET")
	router.HandleFunc("/remove_scheduled_story/{id}", RemoveQueuedStory).Methods("GET")
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

// everything requiring init for production and testing
func initialize() {
	reHasKanji = regexp.MustCompile(`[\x{4E00}-\x{9FAF}]`)
	definitionsCache = make(map[string][]JMDictEntry)
	//definitionsJSONCache = make(map[string]string)
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

func auth(handlerFunc http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, err := sessionStore.Get(r, "session")
		if err != nil {
			http.Redirect(w, r, "/login.html", http.StatusFound)
			return
		}

		if session.IsNew {
			http.Redirect(w, r, "/login.html", http.StatusFound)
		}

		handlerFunc.ServeHTTP(w, r)
	}
}

func makeMainDB() {
	sqldb, err := sql.Open("sqlite3", SQL_USERS_FILE)
	if err != nil {
		log.Fatal(err)
	}
	defer sqldb.Close()

	statement, err := sqldb.Prepare(`CREATE TABLE IF NOT EXISTS users 
	(id INTEGER PRIMARY KEY,
		email TEXT NOT NULL, 
		passwordHash TEXT NOT NULL)`)
	if err != nil {
		log.Fatal(err)
	}
	if _, err := statement.Exec(); err != nil {
		log.Fatal(err)
	}

}

func makeUserDB(userhash string) {
	sqldb, err := sql.Open("sqlite3", "../users/"+userhash+".db")
	if err != nil {
		log.Fatal(err)
	}
	defer sqldb.Close()

	statement, err := sqldb.Prepare(`CREATE TABLE IF NOT EXISTS words 
		(id INTEGER PRIMARY KEY,
			base_form TEXT NOT NULL UNIQUE,
			drill_count INTEGER NOT NULL,
			category INTEGER NOT NULL,
			date_marked INTEGER NOT NULL,
			date_added INTEGER NOT NULL,
			rank INTEGER NOT NULL)`)
	if err != nil {
		log.Fatal(err)
	}
	if _, err := statement.Exec(); err != nil {
		log.Fatal(err)
	}

	statement, err = sqldb.Prepare(`CREATE TABLE IF NOT EXISTS stories 
		(id INTEGER PRIMARY KEY, 
			lines TEXT,         
			title	TEXT UNIQUE,
			link	TEXT UNIQUE,
			countdown INTEGER,
			read_count INTEGER,
			date_last_read INTEGER,
			status INTEGER NOT NULL,
			audio	TEXT,
			date_added INTEGER NOT NULL)`)
	if err != nil {
		log.Fatal(err)
	}
	if _, err := statement.Exec(); err != nil {
		log.Fatal(err)
	}

	statement, err = sqldb.Prepare(`CREATE TABLE IF NOT EXISTS log_events 
	(id INTEGER PRIMARY KEY, 
		story INTEGER NOT NULL,
		date INTEGER NOT NULL,
		FOREIGN KEY (story) REFERENCES stories (id))`)
	if err != nil {
		log.Fatal(err)
	}
	if _, err := statement.Exec(); err != nil {
		log.Fatal(err)
	}

	statement, err = sqldb.Prepare(`CREATE TABLE IF NOT EXISTS queued_stories 
	(id INTEGER PRIMARY KEY, 
		story INTEGER NOT NULL,
		date INTEGER NOT NULL,
		FOREIGN KEY (story) REFERENCES stories (id))`)
	if err != nil {
		log.Fatal(err)
	}
	if _, err := statement.Exec(); err != nil {
		log.Fatal(err)
	}
}

// [END main_func]

// [START indexHandler]

func Kanji(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("Content-Type", "application/json")

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

func GetMain(response http.ResponseWriter, request *http.Request) {

	session, err := sessionStore.Get(request, "session")
	if err != nil {
		http.Error(response, err.Error(), http.StatusInternalServerError)
		return
	}

	if session.IsNew {
		http.Redirect(response, request, "/login.html", http.StatusSeeOther)
	}

	//dbPath := session.Values["user_db_path"]

	http.ServeFile(response, request, "../static/index.html")
}

func PostLoginAuth(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + "could not parse login form: " + err.Error() + `"}`))
		return
	}
	email := r.FormValue("email")

	sqldb, err := sql.Open("sqlite3", SQL_USERS_FILE)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}
	defer sqldb.Close()

	row := sqldb.QueryRow(`SELECT passwordHash FROM users WHERE email = ?;`, email)
	var passwordHash string
	err = row.Scan(&passwordHash)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + "failure to access password for user; user may not exist" + `"}`))
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(r.FormValue("password")+SALT))
	if err != nil {
		http.Redirect(w, r, "/login.html", http.StatusSeeOther)
		return
	}

	session, err := sessionStore.Get(r, "session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	session.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7, // 7 days
		HttpOnly: true,
	}
	session.Values["email"] = email
	hash := md5.Sum([]byte(email))
	userDbPath := "../users/" + hex.EncodeToString(hash[:]) + ".db"
	session.Values["user_db_path"] = userDbPath

	err = session.Save(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// vacuuming the db will compact it to free up wasted space
	err = VacuumDb(userDbPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func PostLogout(response http.ResponseWriter, request *http.Request) {
	session, _ := sessionStore.Get(request, "session")

	delete(session.Values, "userId")
	session.Save(request, response)

	http.Redirect(response, request, "/login.html", http.StatusSeeOther)
}

func PostRegisterUser(response http.ResponseWriter, request *http.Request) {
	err := request.ParseForm()
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + "could not parse login form: " + err.Error() + `"}`))
		return
	}
	email := request.FormValue("email")

	sqldb, err := sql.Open("sqlite3", SQL_USERS_FILE)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}
	defer sqldb.Close()

	row := sqldb.QueryRow(`SELECT passwordHash FROM users WHERE email = ?;`, email)
	var hash string
	err = row.Scan(&hash)
	if err == nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + "user with that email already exists" + `"}`))
		return
	}

	password := request.FormValue("password")
	if password != request.FormValue("password2") {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + "the typed passwords do not match: " + `"}`))
		return
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password+SALT), 8) // 8 is arbitrarily chosen cost
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + "failure to hash password: " + err.Error() + `"}`))
		return
	}

	_, err = sqldb.Exec(`INSERT INTO users (email, passwordHash) VALUES($1, $2);`, email, string(passwordHash))
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + "failure to insert user: " + err.Error() + `"}`))
		return
	}

	// create user DB
	bytes := md5.Sum([]byte(email))
	makeUserDB(hex.EncodeToString(bytes[:]))

	// var cookie = http.Cookie{Name: "user", Value: "test", Expires: time.Now().Add(365 * 24 * time.Hour)}
	// http.SetCookie(response, &cookie)

	http.Redirect(response, request, "/login.html", http.StatusSeeOther)
}

func PostWordSearch(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("Content-Type", "application/json")

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
	response.Header().Set("Content-Type", "application/json")

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
