package main

import (
	"encoding/json"
	"net/http"

	"database/sql"

	_ "github.com/mattn/go-sqlite3"
	"go.mongodb.org/mongo-driver/bson"
)

func AddReps(w http.ResponseWriter, r *http.Request) {
	dbPath := MAIN_USER_DB_PATH

	w.Header().Set("Content-Type", "application/json")

	sqldb, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}
	defer sqldb.Close()

	var body AddRepsRequest
	err = json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}

	row := sqldb.QueryRow(`SELECT reps_todo FROM stories WHERE id = $1;`, body.StoryID)

	var repsStr string
	err = row.Scan(&repsStr)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + "failure to parse reps_todo: " + err.Error() + `"}`))
		return
	}

	reps := make([]int64, 0)

	err = json.Unmarshal([]byte(repsStr), &reps)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}

	reps = append(reps, body.Reps...)

	repsJSON, err := json.Marshal(reps)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + "failure to marshall new reps: " + err.Error() + `"}`))
		return
	}

	_, err = sqldb.Exec(`UPDATE stories SET reps_todo = $1 WHERE id = $2;`, string(repsJSON), body.StoryID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + "failure to set reps_todo: " + err.Error() + `"}`))
		return
	}

	json.NewEncoder(w).Encode(bson.M{"status": "success"})
}

func UpdateReps(w http.ResponseWriter, r *http.Request) {
	dbPath := MAIN_USER_DB_PATH

	var body UpdateRepsRequest
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}

	sqldb, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}
	defer sqldb.Close()

	repsLoggedJSON, err := json.Marshal(body.RepsLogged)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + err.Error() + `"}`))
	}

	repsTodoJSON, err := json.Marshal(body.RepsTodo)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + err.Error() + `"}`))
	}

	_, err = sqldb.Exec(`UPDATE stories SET reps_logged = $1, reps_todo = $2 WHERE id = $3;`,
		string(repsLoggedJSON), string(repsTodoJSON), body.StoryID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + ` ` + err.Error() + `"}`))
	}

	json.NewEncoder(w).Encode(bson.M{"status": "success"})
}

func IncWords(w http.ResponseWriter, r *http.Request) {
	dbPath := MAIN_USER_DB_PATH

	var body IncWordsRequest
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}

	sqldb, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}
	defer sqldb.Close()

	for _, wordId := range body.Words {
		_, err := sqldb.Exec(`UPDATE words SET repetitions = repetitions + 1 WHERE id = $1;`, wordId)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{ "message": "` + "failure to update word repetition counts: " + err.Error() + `"}`))
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(bson.M{"status": "success"})
}
