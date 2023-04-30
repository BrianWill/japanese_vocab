package main

// [START import]
import (
	// "context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/ikawaha/kagome/v2/tokenizer"
	_ "github.com/mattn/go-sqlite3"
	"go.mongodb.org/mongo-driver/bson"
	//"go.mongodb.org/mongo-driver/bson/primitive"
)

func CreateStoryEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Add("content-type", "application/json")
	var story Story
	json.NewDecoder(request.Body).Decode(&story)

	tokens := tok.Analyze(story.Content, tokenizer.Normal)
	story.Tokens = make([]JpToken, len(tokens))

	for i, r := range tokens {
		features := r.Features()
		if len(features) < 9 {

			story.Tokens[i] = JpToken{
				Surface: r.Surface,
				POS:     features[0],
				POS_1:   features[1],
			}

			//fmt.Println(strconv.Itoa(len(features)), features[0], r.Surface, "features: ", strings.Join(features, ","))
		} else {
			story.Tokens[i] = JpToken{
				Surface:          r.Surface,
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
	}

	fmt.Println("prefiltered tokens: ", len(story.Tokens))
	story.Tokens = filterPartsOfSpeech(story.Tokens)
	fmt.Println("filtered tokens: ", len(story.Tokens))

	err := getDefinitions(story.Tokens, response)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": failure to get definitions"` + err.Error() + `"}`))
		return
	}

	sqldb, err := sql.Open("sqlite3", SQL_FILE)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}
	defer sqldb.Close()

	wordIds, err := addDrillWords(story.Tokens, response)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + "failure to add words: " + err.Error() + `"}`))
		return
	}

	wordsJson, err := json.Marshal(wordIds)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + "failure to marshall wordIds: " + err.Error() + `"}`))
		return
	}

	tokensJson, err := json.Marshal(story.Tokens)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + "failure to marshall tokens: " + err.Error() + `"}`))
		return
	}

	_, err = sqldb.Exec(`INSERT INTO stories (user, state, words, content, title, link, tokens) VALUES($1, $2, $3, $4, $5, $6, $7);`,
		USER_ID, "unread", wordsJson, story.Content, story.Title, story.Link, tokensJson)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + "failure to insert story state: " + err.Error() + `"}`))
		return
	}
	json.NewEncoder(response).Encode("Success adding story")
}

func filterPartsOfSpeech(tokens []JpToken) []JpToken {
	filteredTokens := make([]JpToken, 0)
	var prior JpToken
	for _, t := range tokens {
		if t.Surface == "。" {
			continue
		} else if t.Surface == "、" {
			continue
		} else if t.Surface == " " {
			continue
		} else if t.POS == "動詞" && t.POS_1 == "非自立" { // auxilliary verb
			filteredTokens = append(filteredTokens, t)
		} else if t.POS == "副詞" { // adverb
			filteredTokens = append(filteredTokens, t)
		} else if t.POS == "接続詞" && t.POS_1 == "*" { // conjunction
			filteredTokens = append(filteredTokens, t)
		} else if t.POS == "形容詞" { // i-adj
			filteredTokens = append(filteredTokens, t)
		} else if t.POS == "名詞" && t.POS_1 == "代名詞" { // pronoun
			filteredTokens = append(filteredTokens, t)
		} else if t.POS == "連体詞" { // adnominal adjective
			filteredTokens = append(filteredTokens, t)
		} else if t.POS == "動詞" { //　verb
			filteredTokens = append(filteredTokens, t)
		} else if t.POS == "名詞" && t.POS_1 == "接尾" { // noun suffix
			filteredTokens = append(filteredTokens, t)
		} else if (prior.POS == "助詞" && (prior.POS_1 == "連体化" || prior.POS_1 == "並立助詞")) || // preceded by connective particle
			(prior.POS == "接頭詞" && prior.POS_1 == "名詞接続") { // preceded by prefix
			filteredTokens = append(filteredTokens, t)
		} else if t.POS == "名詞" { // noun
			filteredTokens = append(filteredTokens, t)
		} else if t.POS == "号" { // counter
			filteredTokens = append(filteredTokens, t)
		}
		prior = t
	}
	return filteredTokens
}

func addDrillWords(tokens []JpToken, response http.ResponseWriter) ([]int64, error) {
	sqldb, err := sql.Open("sqlite3", SQL_FILE)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return nil, err
	}
	defer sqldb.Close()

	var reHasKanji = regexp.MustCompile(`[\x{4E00}-\x{9FAF}]`)
	var reHasKana = regexp.MustCompile(`[あ-んア-ン]`)
	var reHasKatakana = regexp.MustCompile(`[ア-ン]`)

	// deduplicate
	tokenSet := make(map[string]JpToken)
	for _, token := range tokens {
		tokenSet[token.BaseForm] = token
	}

	tokens = nil
	for k := range tokenSet {
		tokens = append(tokens, tokenSet[k])
	}

	wordIds := make([]int64, 0)
	for _, token := range tokens {
		hasKanji := len(reHasKanji.FindStringIndex(token.BaseForm)) > 0
		hasKana := len(reHasKana.FindStringIndex(token.BaseForm)) > 0
		if !hasKanji && !hasKana {
			continue
		}

		rows, err := sqldb.Query(`SELECT id FROM words WHERE base_form = $1 AND user = $2;`, token.BaseForm, USER_ID)
		if err != nil {
			response.WriteHeader(http.StatusInternalServerError)
			response.Write([]byte(`{ "message": "` + "error while looking up word: " + err.Error() + `"}`))
			return nil, err
		}
		exists := rows.Next()

		unixtime := time.Now().Unix()

		var id int64
		if exists {
			rows.Scan(&id)
			wordIds = append(wordIds, id)
			fmt.Printf("getting word: %s %d \t %d\n", token.BaseForm, len(token.Entries), id)
		} else {
			drillType := 0
			hasKatakana := len(reHasKatakana.FindStringIndex(token.BaseForm)) > 0
			if hasKatakana {
				drillType |= DRILL_TYPE_KATAKANA
			}

			for _, entry := range token.Entries {
				for _, sense := range entry.Senses {
					drillType |= getVerbDrillType(sense)
				}
			}

			entriesJson, err := json.Marshal(token.Entries)
			if err != nil {
				response.WriteHeader(http.StatusInternalServerError)
				response.Write([]byte(`{ "message": "` + "failure to json encode entry: " + err.Error() + `"}`))
				rows.Close()
				return nil, err
			}

			fmt.Printf("\nadding word: %s %d \t %d\n", token.BaseForm, len(token.Entries), id)

			insertResult, err := sqldb.Exec(`INSERT INTO words (base_form, user, countdown, drill_count, 
					read_count, date_last_read, date_last_drill, date_added, date_last_wrong, definitions, drill_type) VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11);`,
				token.BaseForm, USER_ID, INITIAL_COUNTDOWN, 0, 0, unixtime, 0, unixtime, 0, entriesJson, drillType)
			if err != nil {
				response.WriteHeader(http.StatusInternalServerError)
				response.Write([]byte(`{ "message": "` + "failure to insert word: " + err.Error() + `"}`))
				rows.Close()
				return nil, err
			}

			id, err := insertResult.LastInsertId()
			if err != nil {
				response.WriteHeader(http.StatusInternalServerError)
				response.Write([]byte(`{ "message": "` + "failure to get id of inserted word: " + err.Error() + `"}`))
				rows.Close()
				return nil, err
			}

			wordIds = append(wordIds, id)

		}
		rows.Close()
	}

	return wordIds, nil
}

func getDefinitions(tokens []JpToken, response http.ResponseWriter) error {
	// ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	// defer cancel()

	start := time.Now()

	reHasKanji := regexp.MustCompile(`[\x{4E00}-\x{9FAF}]`)

	for i, token := range tokens {
		surface := token.Surface
		hasKanji := len(reHasKanji.FindStringIndex(surface)) > 0

		entries := make([]JMDictEntry, 0)

		if hasKanji {
			//wordQuery = bson.D{{"kanji_spellings.kanji_spelling", searchTerm}}
			for _, entry := range allEntries.Entries {
				for _, k_ele := range entry.KanjiSpellings {
					if k_ele.KanjiSpelling == surface {
						entries = append(entries, entry)
						break
					}
				}
			}
		} else {
			//wordQuery = bson.D{{"readings.reading", searchTerm}}
			for _, entry := range allEntries.Entries {
				for _, r_ele := range entry.Readings {
					if r_ele.Reading == surface {
						entries = append(entries, entry)
						break
					}
				}
			}

		}

		// cursor, err := jmdictCollection.Find(ctx, wordQuery)
		// if err != nil {
		// 	response.WriteHeader(http.StatusInternalServerError)
		// 	response.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		// 	return err
		// }
		// defer cursor.Close(ctx)

		// for cursor.Next(ctx) {
		// 	var entry JMDictEntry
		// 	cursor.Decode(&entry)
		// 	entries = append(entries, entry)
		// }

		//fmt.Printf("\"%v\" \t\t\t matches: %v \n ", surface, len(entries))

		// past certain point, too many matching words isn't useful (will require manual assignment of definition to the token)
		if len(entries) > 8 {
			entries = entries[:8]
		}

		tokens[i].Entries = entries
	}

	duration := time.Since(start)

	fmt.Printf("time to get definitions of %d tokens: %s \n ", len(tokens), duration)

	return nil
}

func getVerbDrillType(sense JMDictSense) int {
	drillType := 0
	for _, pos := range sense.Pos {
		switch pos {
		case "verb-ichidan":
			drillType |= DRILL_TYPE_ICHIDAN
		case "verb-godan-su":
			drillType |= DRILL_TYPE_GODAN_SU
		case "verb-godan-ku":
			drillType |= DRILL_TYPE_GODAN_KU
		case "verb-godan-gu":
			drillType |= DRILL_TYPE_GODAN_GU
		case "verb-godan-ru":
			drillType |= DRILL_TYPE_GODAN_RU
		case "verb-godan-u":
			drillType |= DRILL_TYPE_GODAN_U
		case "verb-godan-tsu":
			drillType |= DRILL_TYPE_GODAN_TSU
		case "verb-godan-mu":
			drillType |= DRILL_TYPE_GODAN_MU
		case "verb-godan-nu":
			drillType |= DRILL_TYPE_GODAN_NU
		case "verb-godan-bu":
			drillType |= DRILL_TYPE_GODAN_BU
		}
	}
	return drillType
}

func GetStoriesListEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Add("content-type", "application/json")

	sqldb, err := sql.Open("sqlite3", SQL_FILE)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}
	defer sqldb.Close()

	rows, err := sqldb.Query(`SELECT id, state, title, link FROM stories WHERE user = $1;`, USER_ID)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + "failure to get story: " + err.Error() + `"}`))
		return
	}
	defer rows.Close()

	var stories []StorySql
	for rows.Next() {
		var story StorySql
		if err := rows.Scan(&story.ID, &story.State, &story.Title, &story.Link); err != nil {
			response.WriteHeader(http.StatusInternalServerError)
			response.Write([]byte(`{ "message": "` + "failure to read story states: " + err.Error() + `"}`))
			return
		}
		stories = append(stories, story)
	}

	json.NewEncoder(response).Encode(stories)
}

func ReadEndpoint(response http.ResponseWriter, request *http.Request) {
	fmt.Println(request.URL.Path)
	http.ServeFile(response, request, "../static/index.html")
}

func GetStoryEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Add("content-type", "application/json")
	params := mux.Vars(request)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}
	fmt.Println("GET STORY id: ", id)

	sqldb, err := sql.Open("sqlite3", SQL_FILE)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}
	defer sqldb.Close()

	rows, err := sqldb.Query(`SELECT state, title, link, tokens, content FROM stories WHERE user = $1 AND id = $2;`, USER_ID, id)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + "failure to get story: " + err.Error() + `"}`))
		return
	}
	defer rows.Close()

	var story StorySql
	for rows.Next() {

		if err := rows.Scan(&story.State, &story.Title, &story.Link, &story.Tokens, &story.Content); err != nil {
			response.WriteHeader(http.StatusInternalServerError)
			response.Write([]byte(`{ "message": "` + "failure to read story states: " + err.Error() + `"}`))
			return
		}
	}

	json.NewEncoder(response).Encode(story)
}

func MarkStoryEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Add("content-type", "application/json")
	params := mux.Vars(request)
	storyId, err := strconv.Atoi(params["id"])
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}
	fmt.Println("MARK STORY id: ", storyId)

	sqldb, err := sql.Open("sqlite3", SQL_FILE)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}
	defer sqldb.Close()

	// make sure the story actually exists
	rows, err := sqldb.Query(`SELECT state FROM stories WHERE user = $1 AND id = $2;`, USER_ID, storyId)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + "failure to get story: " + err.Error() + `"}`))
		return
	}

	if !rows.Next() {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + "story with ID does not exist: " + err.Error() + `"}`))
		rows.Close()
		return
	}
	rows.Close()

	action := params["action"]
	if action != "inactive" && action != "unread" && action != "active" {
		response.WriteHeader(400)
		return
	}

	_, err = sqldb.Exec(`UPDATE stories SET state = $1 WHERE id = $2 AND user = $3;`, action, storyId, USER_ID)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + "failure to update story state: " + err.Error() + `"}`))
		return
	}

	json.NewEncoder(response).Encode(bson.M{"status": "success"})
}

// [END indexHandler]
// [END gae_go111_app]
