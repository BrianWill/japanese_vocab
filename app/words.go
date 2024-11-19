package main

import (
	"compress/gzip"
	"database/sql"
	"encoding/json"
	"regexp"
	"time"

	// "math"
	"net/http"
	//"time"
	"go.mongodb.org/mongo-driver/bson"
)

func GetWords(w http.ResponseWriter, r *http.Request) {
	dbPath := MAIN_USER_DB_PATH

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("content-encoding", "gzip")

	gw := gzip.NewWriter(w)
	defer gw.Close()

	var body DrillRequest
	json.NewDecoder(r.Body).Decode(&body)

	sqldb, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		gw.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}
	defer sqldb.Close()

	var story_title string
	var story_source string
	var story_link string

	if body.StoryId != 0 {
		row := sqldb.QueryRow(`SELECT title, source, link FROM stories WHERE id = $1;`, body.StoryId)
		err = row.Scan(&story_title, &story_source, &story_link)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			gw.Write([]byte(`{ "message": "` + err.Error() + `"}`))
			return
		}
	}

	words, err := getWordsFromStory(sqldb, body.StoryId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		gw.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}

	json.NewEncoder(gw).Encode(bson.M{"words": words, "story_link": story_link, "story_title": story_title, "story_source": story_source})
}

func getStoryTokens(sqldb *sql.DB, storyId int64) ([]*JpToken, error) {
	var subtitlesJA string
	row := sqldb.QueryRow(`SELECT subtitles_ja FROM stories WHERE id = $1;`, storyId)
	err := row.Scan(&subtitlesJA)
	if err != nil {
		return nil, err
	}

	text, err := getSubtitlesString(subtitlesJA)
	if err != nil {
		return nil, err
	}

	tokens, _, err := tokenize(text)
	if err != nil {
		return nil, err
	}

	return tokens, nil
}

func getStoriesRecentlyLogged(sqldb *sql.DB) ([]Story, error) {
	rows, err := sqldb.Query(`SELECT id, title, source, link, video, 
			date, log FROM stories;`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	unixtime := time.Now().Unix()

	var stories []Story
	var log string
	for rows.Next() {
		var story Story
		if err := rows.Scan(&story.ID, &story.Title, &story.Source, &story.Link,
			&story.Video, &story.Date, &log); err != nil {
			return nil, err
		}
		err = json.Unmarshal([]byte(log), &story.Log)
		if err != nil {
			return nil, err
		}
		var dateLastLogged int64
		for _, logItem := range story.Log {
			if logItem.Date > int64(dateLastLogged) {
				dateLastLogged = logItem.Date
			}
		}

		if (unixtime - dateLastLogged) < STORY_RECENTLY_LOGGED_PERIOD {
			stories = append(stories, story)
		}
	}

	return stories, nil
}

func getWordsFromStory(sqldb *sql.DB, storyId int64) ([]DrillWord, error) {
	var tokens []*JpToken
	var err error
	if storyId == 0 {
		tokens = make([]*JpToken, 0)
		stories, err := getStoriesRecentlyLogged(sqldb)
		if err != nil {
			return nil, err
		}
		for _, story := range stories {
			tokens_, err := getStoryTokens(sqldb, story.ID)
			if err != nil {
				return nil, err
			}
			tokens = append(tokens, tokens_...)
		}
	} else {
		tokens, err = getStoryTokens(sqldb, storyId)
		if err != nil {
			return nil, err
		}
	}

	// get word info
	wordMap := make(map[string]DrillWord)

	for _, token := range tokens {
		word := DrillWord{}
		word.BaseForm = token.BaseForm
		row := sqldb.QueryRow(`SELECT id, archived, category,
				repetitions, definitions, date_last_rep FROM words WHERE base_form = $1;`, token.BaseForm)
		err := row.Scan(&word.ID, &word.Archived, &word.Category,
			&word.Repetitions, &word.Definitions, &word.DateLastRep)
		if err == sql.ErrNoRows {
			continue
		}
		if err != nil {
			return nil, err
		}
		wordMap[word.BaseForm] = word
	}

	words := make([]DrillWord, len(wordMap))
	i := 0
	for _, word := range wordMap {
		words[i] = word
		i++
	}

	return words, nil
}

func UpdateWordArchiveState(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var word WordUpdate
	json.NewDecoder(r.Body).Decode(&word)

	sqldb, err := sql.Open("sqlite3", MAIN_USER_DB_PATH)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}
	defer sqldb.Close()

	row := sqldb.QueryRow(`SELECT id FROM words WHERE base_form = $1;`, word.BaseForm)
	var id int64

	err = row.Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{ "message": "` + "cannot update word; word not found" + err.Error() + `"}`))
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + "error looking up word " + err.Error() + `"}`))
		return
	}

	_, err = sqldb.Exec(`UPDATE words SET archived = $1 WHERE base_form = $2;`,
		word.Archived, word.BaseForm)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + "failure to update word: " + err.Error() + `"}`))
		return
	}

	json.NewEncoder(w).Encode(word)
}

func GetKanji(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var str string
	json.NewDecoder(r.Body).Decode(&str)

	var re = regexp.MustCompile(`[\x{4E00}-\x{9FAF}]`)
	kanjiStrings := re.FindAllString(str, -1)

	sqldb, err := sql.Open("sqlite3", MAIN_USER_DB_PATH)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}
	defer sqldb.Close()

	// remove duplicate kanji but maintain original order
	kanjiSet := make([]string, 0)
outer:
	for _, ch := range kanjiStrings {
		for _, k := range kanjiSet {
			if k == ch {
				continue outer
			}
		}
		kanjiSet = append(kanjiSet, ch)
	}

	kanjiDefs := make([]string, 0)
	for _, ch := range kanjiSet {
		row := sqldb.QueryRow(`SELECT kanji FROM words WHERE base_form = $1;`, ch)
		var def string
		err = row.Scan(&def)
		if err != nil {
			if err == sql.ErrNoRows {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(`{ "message": "` + "kanji not found: " + err.Error() + `"}`))
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{ "message": "` + "error looking up kanji: " + err.Error() + `"}`))
			return
		}
		kanjiDefs = append(kanjiDefs, def)
	}

	json.NewEncoder(w).Encode(kanjiDefs)
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

	unixtime := time.Now().Unix()

	for _, wordId := range body.Words {
		_, err := sqldb.Exec(`UPDATE words SET repetitions = repetitions + 1, date_last_rep = $1 WHERE id = $2;`, unixtime, wordId)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{ "message": "` + "failure to update word repetition counts: " + err.Error() + `"}`))
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(bson.M{"status": "success"})
}
