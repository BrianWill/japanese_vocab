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
)

const DEFAULT_ENQUEUED_REPETITIONS = 5

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

	daysFromNow := 0

	for i := 0; i < DEFAULT_ENQUEUED_REPETITIONS; i++ {
		_, err = sqldb.Exec(`INSERT INTO queued_stories (story, days_from_now) VALUES($1, $2);`, storyId, daysFromNow)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{ "message": "` + "failure to enqueue story: " + err.Error() + `"}`))
			return
		}
		daysFromNow += 3
	}

	json.NewEncoder(w).Encode("Success enqueuing story")
}

func GetEnqueuedStories(w http.ResponseWriter, r *http.Request) {
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

	rows, err := sqldb.Query(`SELECT q.id, q.story, q.days_from_now, s.title, s.link
								FROM queued_stories as q
								INNER JOIN stories as s ON q.story = s.id
								ORDER BY q.days_from_now ASC;`)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + "failure to get enqueued story list: " + err.Error() + `"}`))
		return
	}
	defer rows.Close()

	queuedStories := make([]EnqueuedStory, 0)
	for rows.Next() {
		var qs EnqueuedStory
		err = rows.Scan(&qs.ID, &qs.StoryID, &qs.DaysFromNow, &qs.Title, &qs.Link)
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

func BalanceQueue(w http.ResponseWriter, r *http.Request) {
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
		StoryID        int64
		Count          int
		MinDaysFromNow int
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
		return occurences[i].Count > occurences[j].Count
	})

	// delete all entries in story queue
	_, err = sqldb.Exec(`DELETE FROM queued_stories`)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + "failure to delete all enqueued stories: " + err.Error() + `"}`))
		return
	}

	const MAX_STORIES_PER_DAY = 5

	numExhausted := 0
	daysFromNow := 0
	storiesPerDay := 0
	for numExhausted < len(occurences) {
		for i := 0; i < len(occurences); i++ {
			occ := &occurences[i]
			if occ.Count > 0 {
				if daysFromNow < occ.MinDaysFromNow {
					storiesPerDay = 0
					daysFromNow++
				}

				_, err = sqldb.Exec(`INSERT INTO queued_stories (story, days_from_now)
									VALUES ($1, $2);`, occ.StoryID, daysFromNow)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte(`{ "message": "` + "failure to insert queued stories: " + err.Error() + `"}`))
					return
				}

				occ.MinDaysFromNow = daysFromNow + 1

				storiesPerDay++
				if storiesPerDay > MAX_STORIES_PER_DAY {
					storiesPerDay = 0
					daysFromNow++
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
