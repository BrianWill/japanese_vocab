package main

import (
	"compress/gzip"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"time"

	//"unicode/utf8"

	"github.com/gorilla/mux"
	"github.com/ikawaha/kagome/v2/tokenizer"
	_ "github.com/mattn/go-sqlite3"
	"go.mongodb.org/mongo-driver/bson"
)

func tokenize(content string) ([]*JpToken, []string, error) {
	analyzerTokens := tok.Analyze(content, tokenizer.Normal)
	tokens := make([]*JpToken, len(analyzerTokens))

	for i, t := range analyzerTokens {
		features := t.Features()
		if len(features) < 9 {

			tokens[i] = &JpToken{
				Surface: t.Surface,
				POS:     features[0],
				POS_1:   features[1],
			}
		} else {
			tokens[i] = &JpToken{
				Surface:          t.Surface,
				POS:              features[0],
				POS_1:            features[1],
				POS_2:            features[2],
				POS_3:            features[3],
				InflectionalType: features[4],
				InflectionalForm: features[5],
				BaseForm:         features[6],
				Reading:          features[7],
				Pronunciation:    features[8],
			}
		}
		if tokens[i].BaseForm == "" {
			tokens[i].BaseForm = tokens[i].Surface
		}
	}

	kanji, err := extractKanji(tokens)
	if err != nil {
		return nil, nil, fmt.Errorf(`failure to extract kanji` + err.Error())
	}

	return tokens, kanji, nil
}

func getTokenPOS(token *JpToken, priorToken *JpToken) string {
	if token.Surface == "。" {
		return ""
	} else if token.Surface == "\n\n" {
		return ""
	} else if token.Surface == "\n" {
		return ""
	} else if token.Surface == " " {
		return ""
	}

	if (token.POS == "動詞" && token.POS_1 == "接尾") ||
		(token.POS == "助動詞") ||
		(token.Surface == "で" && token.POS == "助詞" && token.POS_1 == "接続助詞") ||
		(token.Surface == "て" && token.POS == "助詞" && token.POS_1 == "接続助詞") ||
		(token.Surface == "じゃ" && token.POS == "助詞" && token.POS_1 == "副助詞") ||
		(token.Surface == "し" && token.POS == "動詞" && token.POS_1 == "自立") { // auxilliary verb
		return "verb_auxiliary"
	} else if token.POS == "動詞" && token.POS_1 == "非自立" { // auxilliary verb
		return "verb_auxiliary"
	} else if (token.POS == "助詞" && token.POS_1 == "格助詞") || // case particle
		(token.POS == "助詞" && token.POS_1 == "接続助詞") || // conjunction particle
		(token.POS == "助詞" && token.POS_1 == "係助詞") || // binding particle (も　は)
		(token.POS == "助詞" && token.POS_1 == "副助詞") { // auxiliary particle
		return "particle"
	} else if token.POS == "副詞" {
		return "adverb"
	} else if token.POS == "接続詞" && token.POS_1 == "*" { // conjunction
		return "conjunction"
	} else if (token.POS == "助詞" && token.POS_1 == "連体化") || // connecting particle　(の)
		(token.POS == "助詞" && token.POS_1 == "並立助詞") { // connecting particle (や)
		return "connecting_particle"
	} else if token.POS == "形容詞" { // i-adj
		return "i_adjective pad_left"
	} else if token.POS == "名詞" && token.POS_1 == "代名詞" { // pronoun
		return "pronoun pad_left"
	} else if token.POS == "連体詞" { // adnominal adjective
		return "admoninal_adjective pad_left"
	} else if token.POS == "動詞" { //　verb
		return "verb pad_left"
	} else if token.POS == "名詞" && token.POS_1 == "接尾" { // noun suffix
		return "noun"
	} else if (priorToken.POS == "助詞" && (priorToken.POS_1 == "連体化" || priorToken.POS_1 == "並立助詞")) || // preceded by connective particle
		(priorToken.POS == "接頭詞" && priorToken.POS_1 == "名詞接続") { // preceded by prefix
		return "noun"
	} else if token.POS == "名詞" { // noun
		return "noun"
	} else if token.POS == "記号" { // symbol
		return ""
	} else if token.POS == "号" { // counter
		return "counter"
	} else {
		return ""
	}
}

func extractKanji(tokens []*JpToken) ([]string, error) {
	kanjiMap := make(map[string]bool)

	for _, t := range tokens {
		var re = regexp.MustCompile(`[\x{4E00}-\x{9FAF}]`)
		kanji := re.FindAllString(t.Surface+t.BaseForm, -1)

		for _, s := range kanji {
			kanjiMap[s] = true
		}
	}

	kanji := make([]string, len(kanjiMap))

	i := 0
	for k := range kanjiMap {
		kanji[i] = k
		i++
	}

	return kanji, nil
}

func addWords(tokens []*JpToken, kanjiSet []string, sqldb *sql.DB) (wordIds []int64, newWordCount int, err error) {
	var reHasKanji = regexp.MustCompile(`[\x{4E00}-\x{9FAF}]`)
	var reHasKatakana = regexp.MustCompile(`[ア-ン]`)
	var reHasKana = regexp.MustCompile(`[ア-ンァ-ヴぁ-ゔ]`)

	newWordCount = 0
	unixtime := time.Now().Unix()

	keptTokens := make([]*JpToken, 0)

	// filter out tokens that have no part of speech
	for i, token := range tokens {
		priorToken := &JpToken{}
		if i > 0 {
			priorToken = tokens[i-1]
		}

		pos := getTokenPOS(token, priorToken)
		if pos == "" {
			continue
		}

		keptTokens = append(keptTokens, token)
	}

	words := make(map[string]bool)
	for _, t := range keptTokens {
		words[t.BaseForm] = true
	}

	for _, k := range kanjiSet {
		words[k] = true
	}

	wordIdsMap := make(map[int64]bool)

	for baseForm := range words {
		hasKatakana := len(reHasKatakana.FindStringIndex(baseForm)) > 0
		hasKana := len(reHasKana.FindStringIndex(baseForm)) > 0
		hasKanji := len(reHasKanji.FindStringIndex(baseForm)) > 0
		isKanji := len([]rune(baseForm)) == 1 && reHasKanji.FindStringIndex(baseForm) != nil

		if !hasKana && !hasKanji { // not a valid word
			continue
		}

		category := 0

		// has katakana
		if hasKatakana {
			category |= DRILL_CATEGORY_KATAKANA
		}

		entries := getDefinitions(baseForm)
		for _, entry := range entries {
			for _, sense := range entry.Senses {
				category |= getVerbCategory(sense)
			}
		}

		entriesJSON, err := json.Marshal(entries)
		if err != nil {
			return nil, 0, err
		}

		kanjiDef := KanjiCharacter{}

		if isKanji {
			for _, ch := range allKanji.Characters {
				if ch.Literal == baseForm {
					kanjiDef = ch
					break
				}
			}

			category |= DRILL_CATEGORY_KANJI
		}

		kanjiDefJSON, err := json.Marshal(kanjiDef)
		if err != nil {
			return nil, 0, err
		}

		var id int64
		err = sqldb.QueryRow(`SELECT id FROM words WHERE base_form = $1`, baseForm).Scan(&id)
		// if error other than no rows
		if err != nil && err != sql.ErrNoRows {
			return nil, 0, err
		}
		// if no error and
		if err == nil {
			wordIdsMap[id] = true
			continue
		}

		insertResult, err := sqldb.Exec(`INSERT INTO words (base_form, 
			date_added, category, repetitions, definitions, kanji, date_last_rep, archived) 
			VALUES($1, $2, $3, $4, $5, $6, $7, $8);`,
			baseForm, unixtime, category, 0, entriesJSON, kanjiDefJSON, 1, 0)
		if err != nil {
			return nil, 0, fmt.Errorf("failure to insert word: " + err.Error())
		}

		id, err = insertResult.LastInsertId()
		if err != nil {
			return nil, 0, fmt.Errorf("failure to get id of inserted word: " + err.Error())
		}

		fmt.Println("inserted word: ", baseForm, id)

		newWordCount++
		wordIdsMap[id] = true
	}

	wordIds = make([]int64, 0, len(wordIdsMap))
	for id, _ := range wordIdsMap {
		wordIds = append(wordIds, id)
	}

	return wordIds, newWordCount, nil
}

func getDefinitions(baseForm string) []JMDictEntry {
	if entries, ok := definitionsCache[baseForm]; ok {
		return entries
	}

	entries := make([]JMDictEntry, 0)

	hasKanji := len(reHasKanji.FindStringIndex(baseForm)) > 0
	if hasKanji {
		for _, e := range allEntriesByKanjiSpellings[baseForm] {
			entries = append(entries, *e)
		}
	} else {
		for _, e := range allEntriesByReading[baseForm] {
			entries = append(entries, *e)
		}
	}

	definitionsCache[baseForm] = entries
	return entries
}

type BaseFormCategoryPair struct {
	BaseForm string
	Category int
}

func GetStories(response http.ResponseWriter, request *http.Request) {
	dbPath := MAIN_USER_DB_PATH

	response.Header().Set("Content-Type", "application/json")

	sqldb, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}
	defer sqldb.Close()

	ips, err := GetOutboundIP()
	if err != nil {
		fmt.Println(err)
	}

	for _, ip := range ips {
		fmt.Println("ip: ", ip)
	}

	stmt := `SELECT s.id, s.title, s.source, s.link, s.video, s.date, s.log, ifnull(q1.word_count, 0), ifnull(q2.archived_word_count, 0)
				FROM stories s
				LEFT JOIN (SELECT story_id, count(*) AS word_count
							FROM stories_x_words
							GROUP BY story_id) q1
					ON s.id = q1.story_id
				LEFT JOIN (SELECT story_id, count(*) AS archived_word_count
							FROM stories_x_words
							INNER JOIN words ON words.id = stories_x_words.word_id
							WHERE words.archived = 1
							GROUP BY story_id) q2
					ON s.id = q2.story_id`

	rows, err := sqldb.Query(stmt)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + "failure to get story: " + err.Error() + `"}`))
		return
	}
	defer rows.Close()

	var stories []Story
	var log string
	for rows.Next() {
		var story Story
		if err := rows.Scan(&story.ID, &story.Title, &story.Source, &story.Link,
			&story.Video, &story.Date, &log, &story.WordCount, &story.ArchivedWordCount); err != nil {
			response.WriteHeader(http.StatusInternalServerError)
			response.Write([]byte(`{ "message": "` + "failure to read story list: " + err.Error() + `"}`))
			return
		}
		err = json.Unmarshal([]byte(log), &story.Log)
		if err != nil {
			response.WriteHeader(http.StatusInternalServerError)
			response.Write([]byte(`{ "message": "` + "failure to get story log: " + err.Error() + `"}`))
		}
		stories = append(stories, story)
	}

	wordStats, err := getWordStats(sqldb)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + "failure to get word stats: " + err.Error() + `"}`))
	}

	json.NewEncoder(response).Encode(bson.M{"stories": stories, "word_stats": wordStats})
}

func getWordStats(sqldb *sql.DB) (bson.M, error) {
	stmt := `SELECT archived, count(*) as word_count
				FROM words
				GROUP BY archived`

	rows, err := sqldb.Query(stmt)
	if err != nil {
		return bson.M{}, nil
	}
	defer rows.Close()

	var wordsTotal int64
	var wordsArchived int64

	var archived int64
	var wordCount int64
	for rows.Next() {
		if err := rows.Scan(&archived, &wordCount); err != nil {
			return bson.M{}, nil
		}
		if archived == 1 {
			wordsArchived = wordCount
		} else if archived == 0 {
			wordsTotal = wordCount
		}
	}

	return bson.M{"total": wordsTotal, "archived": wordsArchived}, nil
}

func GetStory(w http.ResponseWriter, r *http.Request) {
	dbPath := MAIN_USER_DB_PATH

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("content-encoding", "gzip")

	gw := gzip.NewWriter(w)
	defer gw.Close()

	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		gw.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}

	sqldb, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		gw.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}
	defer sqldb.Close()

	row := sqldb.QueryRow(`SELECT title, source, link, date, video, 
		log, subtitles_en, subtitles_ja, subtitles_ja_offset, subtitles_en_offset
		FROM stories WHERE id = $1;`, id)

	var log string
	story := Story{ID: int64(id)}
	if err := row.Scan(&story.Title, &story.Source, &story.Link, &story.Date,
		&story.Video, &log,
		&story.SubtitlesENJson, &story.SubtitlesJAJson, &story.SubtitlesJAOffset, &story.SubtitlesENOffset); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		gw.Write([]byte(`{ "message": failure to scan story row:"` + err.Error() + `"}`))
		return
	}

	err = json.Unmarshal([]byte(log), &story.Log)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		gw.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}

	err = json.NewEncoder(gw).Encode(story)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		gw.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}
}

func UpdateSubtitles(w http.ResponseWriter, r *http.Request) {
	dbPath := MAIN_USER_DB_PATH

	w.Header().Set("Content-Type", "application/json")

	var story Story
	err := json.NewDecoder(r.Body).Decode(&story)
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

	// make sure the story actually exists
	rows, err := sqldb.Query(`SELECT id FROM stories WHERE id = $1;`, story.ID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + "failure to get story: " + err.Error() + `"}`))
		return
	}

	if !rows.Next() {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + "story with ID does not exist: " + strconv.FormatInt(story.ID, 10) + `"}`))
		rows.Close()
		return
	}
	rows.Close()

	_, err = sqldb.Exec(`UPDATE stories SET 
			subtitles_en = CASE WHEN $1 = '' THEN subtitles_en ELSE $1 END,
			subtitles_ja = CASE WHEN $2 = '' THEN subtitles_ja ELSE $2 END,
			subtitles_en_offset = $3,
			subtitles_ja_offset = $4
			WHERE id = $5;`,
		story.SubtitlesENJson, story.SubtitlesJAJson, story.SubtitlesENOffset, story.SubtitlesJAOffset, story.ID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + "failure to update subtitles: " + err.Error() + `"}`))
		return
	}

	err = updateStorySubtitleFiles(story)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + "failure updating story subtitle files: " + err.Error() + `"}`))
		rows.Close()
	}

	json.NewEncoder(w).Encode(bson.M{"status": "success"})
}

func LogStory(w http.ResponseWriter, r *http.Request) {
	dbPath := MAIN_USER_DB_PATH

	w.Header().Set("Content-Type", "application/json")

	var body LogStoryRequest
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

	var logStr string
	err = sqldb.QueryRow(`SELECT log FROM stories WHERE id = $1;`, body.StoryID).Scan(&logStr)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + "failure to get story: " + err.Error() + `"}`))
		return
	}

	var log []LogItem
	err = json.Unmarshal([]byte(logStr), &log)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}

	log = append(log, LogItem{Date: body.Date})

	logJson, err := json.Marshal(log)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + "failure to marshall story log: " + err.Error() + `"}`))
		return
	}

	_, err = sqldb.Exec(`UPDATE stories SET log = $1 WHERE id = $2;`,
		logJson, body.StoryID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + "failure to update story: " + err.Error() + `"}`))
		return
	}

	json.NewEncoder(w).Encode(bson.M{"status": "success"})
}

func GetIP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	ips, err := GetOutboundIP()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + "failure to get IPs: " + err.Error() + `"}`))
		return
	}

	json.NewEncoder(w).Encode(ips)
}

func OpenTranscript(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var body OpenTranscriptRequest
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}

	wd, err := os.Getwd()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + "failure to get working directory: " + err.Error() + `"}`))
	}

	path := wd + "\\..\\sources\\" + body.SourceName + "\\" + body.StoryName + "." + body.Lang + ".vtt"
	fmt.Println("open transcript path: ", path)

	if runtime.GOOS == "windows" {
		cmd := exec.Command("cmd", "/C", "start", "", path)
		err = cmd.Run()
	} else if runtime.GOOS == "darwin" { // mac
		cmd := exec.Command("open", path)
		err = cmd.Run()
	} else { // linux
		cmd := exec.Command("xdg-open", path) // todo untested
		err = cmd.Run()
	}

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + "failure to open transcript in default program: " + err.Error() + `"}`))
		return
	}

	json.NewEncoder(w).Encode(bson.M{"status": "success"})
}
