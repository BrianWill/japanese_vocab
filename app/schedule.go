package main

import (
	"encoding/json"
	"net/http"
	"sort"

	"database/sql"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"go.mongodb.org/mongo-driver/bson"
)

func GetSchedule2(w http.ResponseWriter, r *http.Request) {
	dbPath := MAIN_USER_DB_PATH

	w.Header().Set("Content-Type", "application/json")

	sqldb, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}
	defer sqldb.Close()

	rows, err := sqldb.Query(`SELECT id, title, source, repetitions, reps_logged, reps_todo FROM stories WHERE reps_todo is not null;`)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + "failure to read schedule entry: " + err.Error() + `"}`))
		return
	}
	defer rows.Close()

	stories := make([]Story, 0)

	for rows.Next() {
		var story Story
		if err := rows.Scan(&story.ID, &story.Title, &story.Source, &story.Repetitions, &story.RepsLogged, &story.RepsTodo); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{ "message": "` + "failure to read schedule entry: " + err.Error() + `"}`))
			return
		}
		stories = append(stories, story)
	}

	json.NewEncoder(w).Encode(bson.M{"schedule": stories})
}

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
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + "failure to get reps_todo: " + err.Error() + `"}`))
		return
	}

	var repsStr string
	err = row.Scan(&repsStr)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + "failure to parse reps_todo: " + err.Error() + `"}`))
		return
	}

	reps := make([]int64, 0)

	if repsStr != "" {
		err = json.Unmarshal([]byte(repsStr), &reps)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{ "message": "` + err.Error() + `"}`))
			return
		}
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

func SetRepType(w http.ResponseWriter, r *http.Request) {
	dbPath := MAIN_USER_DB_PATH

	w.Header().Set("Content-Type", "application/json")

	sqldb, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}
	defer sqldb.Close()

	var body SetRepTypeRequest
	err = json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}

	row := sqldb.QueryRow(`SELECT reps_todo FROM stories WHERE id = $1;`, body.StoryID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + "failure to get reps_todo: " + err.Error() + `"}`))
		return
	}

	var repsStr string
	err = row.Scan(&repsStr)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + "failure to parse reps_todo: " + err.Error() + `"}`))
		return
	}

	if repsStr == "" { // no reps
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{ "message": "` + "cannot change rep type: index is out of bounds " + `"}`))
	}

	reps := make([]int64, 0)
	err = json.Unmarshal([]byte(repsStr), &reps)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}

	if body.RepIndex >= int64(len(reps)) || body.RepIndex < 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{ "message": "` + "cannot change rep type: index is out of bounds " + `"}`))
		return
	}

	reps[body.RepIndex] = body.Type

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

func GetLog(w http.ResponseWriter, r *http.Request) {
	dbPath := MAIN_USER_DB_PATH

	w.Header().Set("Content-Type", "application/json")

	sqldb, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}
	defer sqldb.Close()

	// make sure the story actually exists
	rows, err := sqldb.Query(`SELECT id, story, date, type FROM log_entries;`)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + "failure to get story: " + err.Error() + `"}`))
		return
	}
	defer rows.Close()

	entries := make([]ScheduleLogEntry, 0)

	for rows.Next() {
		var entry ScheduleLogEntry
		if err := rows.Scan(&entry.ID, &entry.Story, &entry.Date, &entry.Type); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{ "message": "` + "failure to read story list: " + err.Error() + `"}`))
			return
		}
		entries = append(entries, entry)
	}

	json.NewEncoder(w).Encode(entries)
}

func ScheduleAdjust(w http.ResponseWriter, r *http.Request) {
	dbPath := MAIN_USER_DB_PATH

	var body ScheduleStoryRequest
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

	w.Header().Set("Content-Type", "application/json")

	var entry ScheduleLogEntry
	row := sqldb.QueryRow(`SELECT day_offset, story FROM schedule_entries WHERE id = $1`, body.ID)
	err = row.Scan(&entry.DayOffset, &entry.Story)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + "failure to get schedule entry: " + err.Error() + `"}`))
		return
	}

	// get all other entries for the same story
	rows, err := sqldb.Query(`SELECT id, type, day_offset FROM schedule_entries WHERE story = $1`, entry.Story)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + "failure to get schedule entries: " + err.Error() + `"}`))
		return
	}
	defer rows.Close()

	scheduleEntries := make([]ScheduleLogEntry, 0)

	for rows.Next() {
		var entry ScheduleLogEntry
		if err := rows.Scan(&entry.ID, &entry.Type, &entry.DayOffset); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{ "message": "` + "failure to read schedule entry: " + err.Error() + `"}`))
			return
		}
		scheduleEntries = append(scheduleEntries, entry)
	}

	sort.Slice(scheduleEntries, func(i, j int) bool {
		return scheduleEntries[i].DayOffset < scheduleEntries[j].DayOffset
	})

	idx := 0
	for i := range scheduleEntries {
		if scheduleEntries[i].ID == body.ID {
			idx = i
		}
	}

	minOffset := int64(0)
	if idx > 0 {
		minOffset = scheduleEntries[idx-1].DayOffset + 1
	}

	if scheduleEntries[idx].DayOffset+body.OffsetAdjustment < minOffset {
		body.OffsetAdjustment = minOffset - scheduleEntries[idx].DayOffset
	}

	for _, entry := range scheduleEntries[idx:] {
		newOffset := entry.DayOffset + body.OffsetAdjustment
		_, err = sqldb.Exec(`UPDATE schedule_entries SET day_offset = $1 WHERE id = $2;`, newOffset, entry.ID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{ "message": "` + "failure to add schedule entry: " + err.Error() + `"}`))
			return
		}
	}

	json.NewEncoder(w).Encode(bson.M{"status": "success"})
}

// add a single rep after a specified rep
func ScheduleAdd(w http.ResponseWriter, r *http.Request) {
	dbPath := MAIN_USER_DB_PATH

	var body ScheduleStoryRequest
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

	w.Header().Set("Content-Type", "application/json")

	var entry ScheduleLogEntry
	row := sqldb.QueryRow(`SELECT day_offset, type, story FROM schedule_entries WHERE id = $1`, body.ID)
	err = row.Scan(&entry.DayOffset, &entry.Type, &entry.Story)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + "failure to get schedule entry: " + err.Error() + `"}`))
		return
	}

	// get all other entries for the same story
	rows, err := sqldb.Query(`SELECT id, type, day_offset FROM schedule_entries WHERE story = $1`, entry.Story)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + "failure to get schedule entries: " + err.Error() + `"}`))
		return
	}
	defer rows.Close()

	scheduleEntries := make([]ScheduleLogEntry, 0)

	for rows.Next() {
		var entry ScheduleLogEntry
		if err := rows.Scan(&entry.ID, &entry.Type, &entry.DayOffset); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{ "message": "` + "failure to read schedule entry: " + err.Error() + `"}`))
			return
		}
		scheduleEntries = append(scheduleEntries, entry)
	}

	sort.Slice(scheduleEntries, func(i, j int) bool {
		return scheduleEntries[i].DayOffset < scheduleEntries[j].DayOffset
	})

	// if target day is occupied, do nothing
	targetDay := entry.DayOffset + 1
	for i := range scheduleEntries {
		if scheduleEntries[i].DayOffset == targetDay {
			json.NewEncoder(w).Encode(bson.M{"status": "target day is occupied"})
			return
		}
	}

	// create the new rep
	_, err = sqldb.Exec(`INSERT INTO schedule_entries (story, day_offset, type) VALUES($1, $2, $3);`,
		entry.Story, targetDay, entry.Type)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + "failure to add schedule entry: " + err.Error() + `"}`))
		return
	}

	json.NewEncoder(w).Encode(bson.M{"status": "success"})
}

func GetSchedule(w http.ResponseWriter, r *http.Request) {
	dbPath := MAIN_USER_DB_PATH

	w.Header().Set("Content-Type", "application/json")

	sqldb, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}
	defer sqldb.Close()

	rows, err := sqldb.Query(`SELECT e.id, story, day_offset, type, 
		title, source, repetitions 
		FROM schedule_entries as e INNER JOIN stories as s 
		ON e.story = s.id;`)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + "failure to read schedule entry: " + err.Error() + `"}`))
		return
	}
	defer rows.Close()

	scheduleEntries := make([]ScheduleLogEntry, 0)

	for rows.Next() {
		var entry ScheduleLogEntry
		if err := rows.Scan(&entry.ID, &entry.Story, &entry.DayOffset, &entry.Type,
			&entry.Title, &entry.Source, &entry.Repetitions); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{ "message": "` + "failure to read schedule entry: " + err.Error() + `"}`))
			return
		}
		scheduleEntries = append(scheduleEntries, entry)
	}

	unixtime := time.Now().Unix() - 60*60*48 // 48 hours ago

	// make sure the story actually exists
	rows, err = sqldb.Query(`SELECT e.id, story, e.date, type, 
		title, source, repetitions
		FROM log_entries as e INNER JOIN stories as s 
		ON e.story = s.id AND e.date > $1;`, unixtime)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + "failure to read schedule entry: " + err.Error() + `"}`))
		return
	}
	defer rows.Close()

	logEntries := make([]ScheduleLogEntry, 0)

	for rows.Next() {
		var entry ScheduleLogEntry
		if err := rows.Scan(&entry.ID, &entry.Story, &entry.Date, &entry.Type,
			&entry.Title, &entry.Source, &entry.Repetitions); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{ "message": "` + "failure to read schedule entry: " + err.Error() + `"}`))
			return
		}
		logEntries = append(logEntries, entry)
	}

	json.NewEncoder(w).Encode(bson.M{"schedule": scheduleEntries, "log": logEntries})
}

func UnscheduleStory(w http.ResponseWriter, r *http.Request) {
	dbPath := MAIN_USER_DB_PATH

	var body ScheduleStoryRequest
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

	if body.ID > 0 {
		_, err = sqldb.Exec(`DELETE FROM schedule_entries WHERE id = $1;`, body.ID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{ "message": "` + "failure to add schedule entry: " + err.Error() + `"}`))
			return
		}
	} else {
		_, err = sqldb.Exec(`DELETE FROM schedule_entries WHERE story = $1;`, body.Story)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{ "message": "` + "failure to add schedule entry: " + err.Error() + `"}`))
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(bson.M{"status": "success"})
}

func ScheduleStory(w http.ResponseWriter, r *http.Request) {
	dbPath := MAIN_USER_DB_PATH

	var scheduleRequest ScheduleStoryRequest
	err := json.NewDecoder(r.Body).Decode(&scheduleRequest)
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

	for i := 0; i < NUM_SCHEDULED_REPETITIONS; i++ {
		day := i * 3

		// add rep
		_, err := sqldb.Exec(`INSERT INTO schedule_entries (story, day_offset, type) VALUES($1, $2, $3);`,
			scheduleRequest.Story, day, LISTENING)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{ "message": "` + "failure to add schedule entry: " + err.Error() + `"}`))
			return
		}

		// add drill except after last rep
		if i < NUM_SCHEDULED_REPETITIONS-1 {
			_, err = sqldb.Exec(`INSERT INTO schedule_entries (story, day_offset, type) VALUES($1, $2, $3);`,
				scheduleRequest.Story, day+1, DRILLING)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{ "message": "` + "failure to add schedule entry: " + err.Error() + `"}`))
				return
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(bson.M{"status": "success"})
}

func LogStory(w http.ResponseWriter, r *http.Request) {
	dbPath := MAIN_USER_DB_PATH

	var body ScheduleStoryRequest
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

	unixtime := time.Now().Unix()

	var entry ScheduleLogEntry
	row := sqldb.QueryRow(`SELECT type, story FROM schedule_entries WHERE id = $1`, body.ID)
	err = row.Scan(&entry.Type, &entry.Story)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + "failure to get schedule entry: " + err.Error() + `"}`))
		return
	}

	_, err = sqldb.Exec(`INSERT INTO log_entries (story, date, type) VALUES($1, $2, $3);`,
		entry.Story, unixtime, entry.Type)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + "failure to add schedule entry: " + err.Error() + `"}`))
		return
	}

	_, err = sqldb.Exec(`DELETE FROM schedule_entries WHERE id = $1;`, body.ID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + "failure to add schedule entry: " + err.Error() + `"}`))
		return
	}

	// todo should type of drill be included in the request?
	if len(body.Words) == 0 {
		_, err = sqldb.Exec(`UPDATE stories SET repetitions = repetitions + 1 WHERE id = $1;`, entry.Story)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{ "message": "` + "failure to update story reps: " + err.Error() + `"}`))
			return
		}
	}

	err = incrementWordRepetitions(body.Words, sqldb)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + "failure to update word repetition counts: " + err.Error() + `"}`))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(bson.M{"status": "success"})
}

// toggle among rep types (currently just two: listening and drill)
func ScheduleChangeType(w http.ResponseWriter, r *http.Request) {
	dbPath := MAIN_USER_DB_PATH

	w.Header().Set("Content-Type", "application/json")

	var body ScheduleStoryRequest
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

	if body.RepType != DRILLING && body.RepType != LISTENING && body.RepType != READING {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "invalid rep type"}`))
		return
	}

	_, err = sqldb.Exec(`UPDATE schedule_entries SET type = $1 WHERE id = $2;`, body.RepType, body.ID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + "failure to add schedule entry: " + err.Error() + `"}`))
		return
	}

	json.NewEncoder(w).Encode(bson.M{"status": "success"})
}

func incrementWordRepetitions(wordIds []int64, sqldb *sql.DB) error {
	for _, wordId := range wordIds {
		_, err := sqldb.Exec(`UPDATE words SET repetitions = repetitions + 1 WHERE id = $1;`, wordId)
		if err != nil {
			return err
		}
	}
	return nil
}
