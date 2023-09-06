package main

import (
	"database/sql"
	"encoding/json"
	"sort"
	"strconv"
	"time"

	"net/http"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
	"go.mongodb.org/mongo-driver/bson"
)

const DEFAULT_ENQUEUED_REPETITIONS = 5
const SECONDS_IN_DAY = 60 * 60 * 24

func EnqueueStory(w http.ResponseWriter, r *http.Request) {
	dbPath, redirect, err := GetUserDb(w, r)
	if redirect || err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + err.Error() + `"}`))
	}

	w.Header().Set("Content-Type", "application/json")

	sqldb, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + err.Error() + `"}`))
	}
	defer sqldb.Close()

	params := mux.Vars(r)
	var storyId int64
	id, err := strconv.Atoi(params["storyId"])
	storyId = int64(id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}

	unixtime := time.Now().Unix()

	for i := 0; i < DEFAULT_ENQUEUED_REPETITIONS; i++ {
		_, err = sqldb.Exec(`INSERT INTO queued_stories (story, date) VALUES($1, $2);`, storyId, unixtime)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{ "message": "` + "failure to enqueue story: " + err.Error() + `"}`))
			return
		}
		unixtime += SECONDS_IN_DAY
	}

	json.NewEncoder(w).Encode("Success enqueuing story")
}

func GetStoriesTodo(w http.ResponseWriter, r *http.Request) {
	dbPath, redirect, err := GetUserDb(w, r)
	if redirect || err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + err.Error() + `"}`))
	}

	w.Header().Set("Content-Type", "application/json")

	sqldb, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + err.Error() + `"}`))
	}
	defer sqldb.Close()

	rows, err := sqldb.Query(`SELECT q.id, q.story, q.date, s.title, s.link
								FROM queued_stories as q
								INNER JOIN stories as s ON q.story = s.id
								ORDER BY q.date ASC;`)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + "failure to get enqueued story list: " + err.Error() + `"}`))
		return
	}
	defer rows.Close()

	queuedStories := make([]EnqueuedStory, 0)
	for rows.Next() {
		var qs EnqueuedStory
		err = rows.Scan(&qs.ID, &qs.StoryID, &qs.Date, &qs.Title, &qs.Link)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{ "message": "` + "failure to get enqueued story list: " + err.Error() + `"}`))
			return
		}
		queuedStories = append(queuedStories, qs)
	}

	json.NewEncoder(w).Encode(queuedStories)
}

func MarkQueuedStory(w http.ResponseWriter, r *http.Request) {
	dbPath, redirect, err := GetUserDb(w, r)
	if redirect || err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + err.Error() + `"}`))
	}

	w.Header().Set("Content-Type", "application/json")

	params := mux.Vars(r)
	var logId int64
	id, err := strconv.Atoi(params["id"])
	logId = int64(id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}

	var storyId int64
	id, err = strconv.Atoi(params["storyId"])
	storyId = int64(id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}

	sqldb, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + err.Error() + `"}`))
	}
	defer sqldb.Close()

	unixtime := time.Now().Unix()
	cooldownWindowStart := unixtime - STORY_LOG_COOLDOWN

	row := sqldb.QueryRow(`SELECT date FROM log_events WHERE date > $1 AND story = $2`, cooldownWindowStart, storyId)

	var existingId int64
	err = row.Scan(&existingId)
	if err == nil {
		w.Write([]byte(`{ "message": "No entry added to the read log. This story was previously logged within the 8 hour cooldown window."}`))
		return
	} else if err != sql.ErrNoRows {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}

	_, err = sqldb.Exec(`INSERT INTO log_events (date, story) 
			VALUES($1, $2);`,
		unixtime, storyId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + "failure to insert log event: " + err.Error() + `"}`))
		return
	}

	_, err = sqldb.Exec(`DELETE FROM queued_stories WHERE id = $1`, logId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + "failure to delete enqueued story: " + err.Error() + `"}`))
		return
	}

	json.NewEncoder(w).Encode("Removed queued story entry")
}

func RemoveQueuedStory(w http.ResponseWriter, r *http.Request) {
	dbPath, redirect, err := GetUserDb(w, r)
	if redirect || err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + err.Error() + `"}`))
	}

	w.Header().Set("Content-Type", "application/json")

	params := mux.Vars(r)
	var logId int64
	id, err := strconv.Atoi(params["id"])
	logId = int64(id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}

	sqldb, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + err.Error() + `"}`))
	}
	defer sqldb.Close()

	_, err = sqldb.Exec(`DELETE FROM queued_stories WHERE id = $1`, logId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + "failure to delete enqueued story: " + err.Error() + `"}`))
		return
	}

	json.NewEncoder(w).Encode("Removed queued story entry")
}

func RefreshQueue(w http.ResponseWriter, r *http.Request) {
	dbPath, redirect, err := GetUserDb(w, r)
	if redirect || err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + err.Error() + `"}`))
	}

	w.Header().Set("Content-Type", "application/json")

	sqldb, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + err.Error() + `"}`))
	}
	defer sqldb.Close()

	rows, err := sqldb.Query(`SELECT story FROM queued_stories;`)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + "failure to get enqueued story list: " + err.Error() + `"}`))
		return
	}
	defer rows.Close()

	type StoryOccurences struct {
		StoryID int64
		Count   int
		MinDate int64
	}

	countById := make(map[int64]int)
	for rows.Next() {
		var storyId int64
		err = rows.Scan(&storyId)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{ "message": "` + "failure to get enqueued story list: " + err.Error() + `"}`))
			return
		}
		if val, ok := countById[storyId]; ok {
			countById[storyId] = val + 1
		} else {
			countById[storyId] = 1
		}
	}

	occurences := make([]StoryOccurences, len(countById))
	i := 0
	for id, count := range countById {
		occurences[i] = StoryOccurences{StoryID: id, Count: count}
		i++
	}
	sort.Slice(occurences, func(i, j int) bool {
		return occurences[i].Count < occurences[j].Count
	})

	// delete all entries in story queue
	_, err = sqldb.Exec(`DELETE FROM queued_stories`)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + "failure to delete all enqueued stories: " + err.Error() + `"}`))
		return
	}

	// distribute the stories across days, round robin (prioritizing stories with lower counts)
	const MAX_STORIES_PER_DAY = 5
	numExhausted := 0
	unixtime := time.Now().Unix()
	storiesPerDay := 0
	for numExhausted < len(occurences) {
		for i := 0; i < len(occurences); i++ {
			occ := &occurences[i]
			if occ.Count > 0 {
				if unixtime < occ.MinDate {
					storiesPerDay = 0
					unixtime = occ.MinDate
				}

				_, err = sqldb.Exec(`INSERT INTO queued_stories (story, date)
									VALUES ($1, $2);`, occ.StoryID, unixtime)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte(`{ "message": "` + "failure to insert queued stories: " + err.Error() + `"}`))
					return
				}

				occ.MinDate = unixtime + SECONDS_IN_DAY

				storiesPerDay++
				if storiesPerDay > MAX_STORIES_PER_DAY {
					storiesPerDay = 0
					unixtime += SECONDS_IN_DAY
				}

				occ.Count--
				if occ.Count == 0 {
					numExhausted++
				}
			}
		}
	}

	json.NewEncoder(w).Encode("rebalanced story queue")
}

func AddLogEvent(response http.ResponseWriter, request *http.Request) {
	dbPath, redirect, err := GetUserDb(response, request)
	if redirect || err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `"}`))
	}

	response.Header().Set("Content-Type", "application/json")

	params := mux.Vars(request)
	var storyId int64
	id, err := strconv.Atoi(params["id"])
	storyId = int64(id)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}

	sqldb, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}
	defer sqldb.Close()

	unixtime := time.Now().Unix()
	cooldownWindowStart := unixtime - STORY_LOG_COOLDOWN

	row := sqldb.QueryRow(`SELECT date FROM log_events WHERE date > $1 AND story = $2`, cooldownWindowStart, storyId)

	var existingId int64
	err = row.Scan(&existingId)
	if err == nil {
		response.Write([]byte(`{ "message": "No entry added to the read log. This story was previously logged within the 8 hour cooldown window."}`))
		return
	} else if err != sql.ErrNoRows {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}

	_, err = sqldb.Exec(`INSERT INTO log_events (date, story) 
			VALUES($1, $2);`,
		unixtime, storyId)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + "failure to insert log event: " + err.Error() + `"}`))
		return
	}

	response.Write([]byte(`{ "message": "Entry for story added to the read log."}`))
}

func RemoveLogEvent(response http.ResponseWriter, request *http.Request) {
	dbPath, redirect, err := GetUserDb(response, request)
	if redirect || err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `"}`))
	}

	response.Header().Set("Content-Type", "application/json")

	params := mux.Vars(request)
	var logId int64
	id, err := strconv.Atoi(params["id"])
	logId = int64(id)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}

	sqldb, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}
	defer sqldb.Close()

	_, err = sqldb.Exec(`DELETE FROM log_events WHERE id = $1;`, logId)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + "failure to delete log event: " + err.Error() + `"}`))
		return
	}

	json.NewEncoder(response).Encode(bson.M{"status": "success"})
}

func GetLogEvents(response http.ResponseWriter, request *http.Request) {
	dbPath, redirect, err := GetUserDb(response, request)
	if redirect {
		return
	}
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `"}`))
	}

	response.Header().Set("Content-Type", "application/json")

	sqldb, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}
	defer sqldb.Close()

	rows, err := sqldb.Query(`SELECT l.id, l.date, l.story, s.title 
								FROM log_events as l
								INNER JOIN stories as s ON l.story = s.id 
								ORDER BY date DESC;`)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + "failure to get story: " + err.Error() + `"}`))
		return
	}
	defer rows.Close()

	var logEvents = make([]LogEvent, 0)
	for rows.Next() {
		var le LogEvent
		if err := rows.Scan(&le.ID, &le.Date, &le.StoryID, &le.Title); err != nil {
			response.WriteHeader(http.StatusInternalServerError)
			response.Write([]byte(`{ "message": "` + "failure to read story: " + err.Error() + `"}`))
			return
		}
		logEvents = append(logEvents, le)
	}

	json.NewEncoder(response).Encode(bson.M{"logEvents": logEvents})
}
