package main

import (
	// "context"
	// "compress/gzip"
	"database/sql"
	"encoding/json"
	// "errors"
	"fmt"
	"net/http"
	// "regexp"
	// "strconv"
	// "strings"
	// "time"
	// "unicode/utf8"

	//"github.com/gorilla/mux"
	//"github.com/ikawaha/kagome/v2/tokenizer"
	_ "github.com/mattn/go-sqlite3"
	//"go.mongodb.org/mongo-driver/bson"
	//"go.mongodb.org/mongo-driver/bson/primitive"
)

func EnqueueStory(w http.ResponseWriter, r *http.Request) {
	dbPath, redirect, err := GetUserDb(w, r)
	if redirect || err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + err.Error() + `"}`))
	}

	w.Header().Set("Content-Type", "application/json")

	var enqueueRequest EnqueueRequest
	json.NewDecoder(r.Body).Decode(&enqueueRequest)

	sqldb, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + err.Error() + `"}`))
	}
	defer sqldb.Close()

	daysFromNow := 0

	for i := 0; i < enqueueRequest.Count; i++ {
		_, err = sqldb.Exec(`INSERT INTO queued_stories (story, days_from_now) VALUES($1, $2);`, enqueueRequest.StoryId, daysFromNow)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{ "message": "` + "failure to enqueue story: " + err.Error() + `"}`))
			return
		}
		daysFromNow += 2
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
	return
	dbPath, redirect, err := GetUserDb(w, r)
	if redirect || err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + err.Error() + `"}`))
	}

	w.Header().Set("Content-Type", "application/json")

	var story Story
	json.NewDecoder(r.Body).Decode(&story)

	sqldb, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + err.Error() + `"}`))
	}
	defer sqldb.Close()

	_, newWordCount, err := addStory(story, sqldb, false)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}
	fmt.Println("total new words added:", newWordCount)
	json.NewEncoder(w).Encode("Success adding story")
}

func RemoveQueuedStory(w http.ResponseWriter, r *http.Request) {
	return
	dbPath, redirect, err := GetUserDb(w, r)
	if redirect || err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + err.Error() + `"}`))
	}

	w.Header().Set("Content-Type", "application/json")

	var story Story
	json.NewDecoder(r.Body).Decode(&story)

	sqldb, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + err.Error() + `"}`))
	}
	defer sqldb.Close()

	_, newWordCount, err := addStory(story, sqldb, false)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}
	fmt.Println("total new words added:", newWordCount)
	json.NewEncoder(w).Encode("Success adding story")
}

func BalanceQueue(w http.ResponseWriter, r *http.Request) {
	return
	dbPath, redirect, err := GetUserDb(w, r)
	if redirect || err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + err.Error() + `"}`))
	}

	w.Header().Set("Content-Type", "application/json")

	var story Story
	json.NewDecoder(r.Body).Decode(&story)

	sqldb, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + err.Error() + `"}`))
	}
	defer sqldb.Close()

	_, newWordCount, err := addStory(story, sqldb, false)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}
	fmt.Println("total new words added:", newWordCount)
	json.NewEncoder(w).Encode("Success adding story")
}
