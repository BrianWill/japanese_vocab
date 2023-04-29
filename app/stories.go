package main

// [START import]
import (
	//"database/sql"
	"encoding/json"
	"fmt"
	// "math"
	"regexp"
	// "sort"

	// "strings"
	// "unicode/utf8"

	// //"strconv"

	// "log"
	"net/http"
	// "os"

	"context"
	"time"

	// "github.com/ikawaha/kagome-dict/ipa"
	"github.com/ikawaha/kagome/v2/tokenizer"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	// "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	// "go.mongodb.org/mongo-driver/mongo/readpref"

	"database/sql"

	_ "github.com/mattn/go-sqlite3"

	//"github.com/hedhyw/rex/pkg/rex"  // regex builder

	"github.com/gorilla/mux"
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

	getDefinitions(story.Tokens, response)

	sqldb, err := sql.Open("sqlite3", SQL_FILE)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}
	defer sqldb.Close()

	wordIds := addWords(story.Tokens)

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

func addWords(tokens []JpToken, response http.ResponseWriter) ([]int64, error) {
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
		rows.Close()

		unixtime := time.Now().Unix()

		var id int64
		if exists {
			rows.Scan(&id)
			wordIds = append(wordIds, id)
		} else {
			fmt.Printf("\nadding word: %s %d\n", token.BaseForm, len(token.Definitions))

			defs := make([]JMDictEntry, len(token.Definitions))

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			drillType := 0
			hasKatakana := len(reHasKatakana.FindStringIndex(token.BaseForm)) > 0
			if hasKatakana {
				drillType |= DRILL_TYPE_KATAKANA
			}

			for i, def := range token.Definitions {
				var entry JMDictEntry
				err := jmdictCollection.FindOne(ctx, bson.M{"_id": def}).Decode(&entry)
				if err != nil {
					response.WriteHeader(http.StatusInternalServerError)
					response.Write([]byte(`{ "message": "` + err.Error() + `"}`))
					return nil, err
				}
				defs[i] = entry
				for _, sense := range entry.Sense {
					drillType |= getVerbDrillType(sense)
				}
			}

			defsJson, err := json.Marshal(defs)
			if err != nil {
				response.WriteHeader(http.StatusInternalServerError)
				response.Write([]byte(`{ "message": "` + "failure to encode json: " + err.Error() + `"}`))
				return nil, err
			}

			insertResult, err := sqldb.Exec(`INSERT INTO words (base_form, user, countdown, drill_count, 
					read_count, date_last_read, date_last_drill, date_added, date_last_wrong, definitions, drill_type) VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11);`,
				token.BaseForm, USER_ID, INITIAL_COUNTDOWN, 0, 0, unixtime, 0, unixtime, 0, defsJson, drillType)
			if err != nil {
				response.WriteHeader(http.StatusInternalServerError)
				response.Write([]byte(`{ "message": "` + "failure to insert word: " + err.Error() + `"}`))
				return nil, err
			}

			id, err := insertResult.LastInsertId()
			if err != nil {
				response.WriteHeader(http.StatusInternalServerError)
				response.Write([]byte(`{ "message": "` + "failure to get id of inserted word: " + err.Error() + `"}`))
				return nil, err
			}

			wordIds = append(wordIds, id)
		}
	}

	return wordIds, nil
}

func getDefinitions(tokens []JpToken, response http.ResponseWriter) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	var re = regexp.MustCompile(`[\x{4E00}-\x{9FAF}]`)

	for _, token := range tokens {
		searchTerm := token.Surface

		var wordQuery primitive.D
		if len(re.FindStringIndex(searchTerm)) > 0 { // has kanji
			//kanji := re.FindAllString(searchTerm, -1)
			wordQuery = bson.D{{"kanji_spellings.kanji_spelling", searchTerm}}
		} else {
			wordQuery = bson.D{{"readings.reading", searchTerm}}
		}

		start := time.Now()

		cursor, err := jmdictCollection.Find(ctx, wordQuery)
		if err != nil {
			response.WriteHeader(http.StatusInternalServerError)
			response.Write([]byte(`{ "message": "` + err.Error() + `"}`))
			return err
		}
		defer cursor.Close(ctx)

		duration := time.Since(start)

		entries := make([]JMDictEntry, 0)
		for cursor.Next(ctx) {
			var entry JMDictEntry
			cursor.Decode(&entry)
			entries = append(entries, entry)
		}

		fmt.Printf("\"%v\" \t matches: %v \t %v \n ", searchTerm, len(entries), duration)

		// past certain point, too many matching words isn't useful (will require manual assignment of definition to the token)
		if len(entries) > 8 {
			entries = entries[:8]
		}

		token.Entries = entries
	}

	return nil
}

func GetStoriesListEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Add("content-type", "application/json")
	var stories []Story
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	opts := options.Find().SetProjection(bson.D{{"title", 1}, {"_id", 1}})
	cursor, err := storiesCollection.Find(ctx, bson.M{}, opts)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var story Story
		cursor.Decode(&story)
		stories = append(stories, story)
	}
	if err := cursor.Err(); err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}

	sqldb, err := sql.Open("sqlite3", SQL_FILE)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}
	defer sqldb.Close()

	rows, err := sqldb.Query(`SELECT story, state FROM stories WHERE user = $1;`, USER_ID)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + "failure to get story: " + err.Error() + `"}`))
		return
	}
	defer rows.Close()

	states := make(map[string]string)
	for rows.Next() {
		var state string
		var storyId string
		if err := rows.Scan(&storyId, &state); err != nil {
			response.WriteHeader(http.StatusInternalServerError)
			response.Write([]byte(`{ "message": "` + "failure to read story states: " + err.Error() + `"}`))
			return
		}
		states[storyId] = state
		fmt.Println("STATUS", storyId, state)
	}

	activeStories := make([]Story, 0)
	inactiveStories := make([]Story, 0)
	unreadStories := make([]Story, 0)
	for _, story := range stories {
		state, ok := states[story.ID.Hex()]
		if !ok || state == "unread" {
			unreadStories = append(unreadStories, story)
		} else if state == "inactive" {
			inactiveStories = append(inactiveStories, story)
		} else if state == "active" {
			activeStories = append(activeStories, story)
		}
	}

	json.NewEncoder(response).Encode(bson.M{
		"unreadStories":   unreadStories,
		"inactiveStories": inactiveStories,
		"activeStories":   activeStories})
}

func ReadEndpoint(response http.ResponseWriter, request *http.Request) {
	fmt.Println(request.URL.Path)
	http.ServeFile(response, request, "../static/index.html")
}

func GetStoryEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Add("content-type", "application/json")
	params := mux.Vars(request)
	id, _ := primitive.ObjectIDFromHex(params["id"])
	fmt.Println("story id: ", id)
	var story Story
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err := storiesCollection.FindOne(ctx, bson.M{"_id": id}).Decode(&story)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}

	tokenDefinitions := make([][]JMDictEntry, len(story.Tokens))

	for i, token := range story.Tokens {
		tokenDefinitions[i] = make([]JMDictEntry, len(token.Definitions))
		for j, def := range token.Definitions {
			var entry JMDictEntry
			err := jmdictCollection.FindOne(ctx, bson.M{"_id": def}).Decode(&entry)
			if err != nil {
				response.WriteHeader(http.StatusInternalServerError)
				response.Write([]byte(`{ "message": "` + err.Error() + `"}`))
				return
			}
			tokenDefinitions[i][j] = entry
		}
	}

	json.NewEncoder(response).Encode(bson.M{
		"story":       story,
		"definitions": tokenDefinitions,
	})
}

func MarkStoryEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Add("content-type", "application/json")
	params := mux.Vars(request)
	id, err := primitive.ObjectIDFromHex(params["id"])
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}

	// make sure the story actually exists
	var story Story
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	err = storiesCollection.FindOne(ctx, Story{ID: id}).Decode(&story)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}

	action := params["action"]

	if action != "inactive" && action != "unread" && action != "active" {
		response.WriteHeader(400)
		return
	}

	sqldb, err := sql.Open("sqlite3", SQL_FILE)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}
	defer sqldb.Close()

	storyID := story.ID.Hex()

	rows, err := sqldb.Query(`SELECT id FROM stories WHERE story = $1 AND user = $2;`, storyID, USER_ID)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + "failure to get story: " + err.Error() + `"}`))
		return
	}
	exists := rows.Next()
	rows.Close()

	fmt.Println("query ", exists, storyID, USER_ID)

	if exists {
		_, err = sqldb.Exec(`UPDATE stories SET state = $1 WHERE story = $2 AND user = $3;`, action, storyID, USER_ID)
		if err != nil {
			response.WriteHeader(http.StatusInternalServerError)
			response.Write([]byte(`{ "message": "` + "failure to update story state: " + err.Error() + `"}`))
			return
		}
	} else {
		_, err = sqldb.Exec(`INSERT INTO stories (story, state, user) VALUES($1, $2, $3);`, storyID, action, USER_ID)
		if err != nil {
			response.WriteHeader(http.StatusInternalServerError)
			response.Write([]byte(`{ "message": "` + "failure to insert story state: " + err.Error() + `"}`))
			return
		}
	}

	json.NewEncoder(response).Encode(bson.M{"status": "success"})
}

func RetokenizeStoryEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Add("content-type", "application/json")
	params := mux.Vars(request)
	id, err := primitive.ObjectIDFromHex(params["id"])
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}

	var story Story
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	err = storiesCollection.FindOne(ctx, Story{ID: id}).Decode(&story)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}

	tokens := tok.Analyze(story.Content, tokenizer.Normal)
	story.Tokens = make([]JpToken, len(tokens))

	var re = regexp.MustCompile(`[\x{4E00}-\x{9FAF}]`)

	for i, r := range tokens {
		features := r.Features()
		var searchTerm string
		if len(features) < 9 {
			searchTerm = r.Surface
			story.Tokens[i] = JpToken{
				Surface: r.Surface,
				POS:     features[0],
				POS_1:   features[1],
			}

			//fmt.Println(strconv.Itoa(len(features)), features[0], r.Surface, "features: ", strings.Join(features, ","))
		} else {
			searchTerm = features[6] // base form
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
		var wordQuery primitive.D
		if len(re.FindStringIndex(searchTerm)) > 0 { // has kanji
			//kanji := re.FindAllString(searchTerm, -1)
			wordQuery = bson.D{{"kanji_spellings.kanji_spelling", searchTerm}}
		} else {
			wordQuery = bson.D{{"readings.reading", searchTerm}}
		}

		start := time.Now()

		cursor, err := jmdictCollection.Find(ctx, wordQuery)
		if err != nil {
			response.WriteHeader(http.StatusInternalServerError)
			response.Write([]byte(`{ "message": "` + err.Error() + `"}`))
			return
		}
		defer cursor.Close(ctx)

		duration := time.Since(start)

		wordIDs := make([]primitive.ObjectID, 0)
		for cursor.Next(ctx) {
			var entry JMDictEntry
			cursor.Decode(&entry)
			wordIDs = append(wordIDs, entry.ID)
		}

		// todo past certain point, too many matching words isn't useful (will require manual assignment of definition to the token)

		fmt.Printf("\"%v\" \t matches: %v \t %v \n ", searchTerm, len(wordIDs), duration)
		if len(wordIDs) < 8 {
			story.Tokens[i].Definitions = wordIDs
		}
	}

	_, err = storiesCollection.UpdateByID(ctx, id, bson.M{"$set": story})
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}

	json.NewEncoder(response).Encode(story)
}

// [END indexHandler]
// [END gae_go111_app]
