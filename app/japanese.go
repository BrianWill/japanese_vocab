// Copyright 2019 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// [START gae_go111_app]

package main

// [START import]
import (
	"database/sql"
	"encoding/json"
	"fmt"
	//"strconv"

	//"strings"

	"log"
	"net/http"
	"os"

	"context"
	"time"

	"github.com/ikawaha/kagome-dict/ipa"
	"github.com/ikawaha/kagome/v2/tokenizer"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	// github.com/hedhyw/rex   // regex builder

	"github.com/gorilla/mux"

	// Note: If connecting using the App Engine Flex Go runtime, use
	// "github.com/jackc/pgx/stdlib" instead, since v4 requires
	// Go modules which are not supported by App Engine Flex.
	_ "github.com/jackc/pgx/v4/stdlib"
)

// [END import]
// [START main_func]

var client *mongo.Client
var db *mongo.Database
var storiesCollection *mongo.Collection

var tok *tokenizer.Tokenizer

type Person struct {
	ID        primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	FirstName string             `json:"firstname,omitempty" bson:"firstname,omitempty"`
	Lastname  string             `json:"lastname,omitempty" bson:"lastname,omitempty"`
}

type Story struct {
	ID      primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Content string             `json:"content,omitempty" bson:"content,omitempty"`
	Title   string             `json:"title,omitempty" bson:"title,omitempty"`
	Tokens  []JpToken          `json:"tokens,omitempty" bson:"tokens,omitempty"`
}

type JpToken struct {
	Surface          string `json:"surface,omitempty" bson:"surface,omitempty"`
	POS              string `json:"pos,omitempty" bson:"pos"`
	POS_1            string `json:"pos1,omitempty" bson:"pos1"`
	POS_2            string `json:"pos2,omitempty" bson:"pos2"`
	POS_3            string `json:"pos3,omitempty" bson:"pos3"`
	InflectionalType string `json:"inflectionalType,omitempty" bson:"inflectionalType"`
	InflectionalForm string `json:"inflectionalForm,omitempty" bson:"inflectionalForm"`
	BaseForm         string `json:"baseForm,omitempty" bson:"baseForm"`
	Reading          string `json:"reading,omitempty" bson:"reading"`
	Pronunciation    string `json:"pronunciation,omitempty" bson:"pronunciation"`
}

func main() {
	var err error
	tok, err = tokenizer.New(ipa.Dict(), tokenizer.OmitBosEos())
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err = mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))

	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()

	db = client.Database("JapaneseEnglish")
	storiesCollection = db.Collection("stories")

	ctx, cancel = context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	err = client.Ping(ctx, readpref.Primary())

	//fmt.Print(client, err)

	// [START setting_port]
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}

	router := mux.NewRouter()

	router.HandleFunc("/story", CreateStoryEndpoint).Methods("POST")
	router.HandleFunc("/story/{id}", GetStoryEndpoint).Methods("GET")
	router.HandleFunc("/stories_list", GetStoriesListEndpoint).Methods("GET")
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("../static")))

	log.Printf("Listening on port %s", port)
	if err := http.ListenAndServe(":"+port, router); err != nil {
		log.Fatal(err)
	}
	// [END setting_port]
}

// [END main_func]

// [START indexHandler]

func CreateStoryEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Add("content-type", "application/json")
	var story Story
	json.NewDecoder(request.Body).Decode(&story)
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	result, err := storiesCollection.InsertOne(ctx, story)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}
	json.NewEncoder(response).Encode(result)
}

func GetStoriesListEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Add("content-type", "application/json")
	var stories []Story
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
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
	json.NewEncoder(response).Encode(stories)
}

func GetStoryEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Add("content-type", "application/json")
	params := mux.Vars(request)
	id, _ := primitive.ObjectIDFromHex(params["id"])
	var story Story
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err := storiesCollection.FindOne(ctx, Story{ID: id}).Decode(&story)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}

	tokens := tok.Analyze(story.Content, tokenizer.Search)
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

		// if len(features) < 9 {

		// } else {
		// 	fmt.Println("features: ", strings.Join(features, ","))
		// }

	}

	json.NewEncoder(response).Encode(story)
}

// connectUnixSocket initializes a Unix socket connection pool for
// a Cloud SQL instance of Postgres.
func connectUnixSocket() (*sql.DB, error) {
	mustGetenv := func(k string) string {
		v := os.Getenv(k)
		if v == "" {
			log.Fatalf("Warning: %s environment variable not set.\n", k)
		}
		return v
	}
	// Note: Saving credentials in environment variables is convenient, but not
	// secure - consider a more secure solution such as
	// Cloud Secret Manager (https://cloud.google.com/secret-manager) to help
	// keep secrets safe.
	var (
		dbUser         = mustGetenv("DB_USER")              // e.g. 'my-db-user'
		dbPwd          = mustGetenv("DB_PASS")              // e.g. 'my-db-password'
		unixSocketPath = mustGetenv("INSTANCE_UNIX_SOCKET") // e.g. '/cloudsql/project:region:instance'
		dbName         = mustGetenv("DB_NAME")              // e.g. 'my-database'
	)

	dbURI := fmt.Sprintf("user=%s password=%s database=%s host=%s",
		dbUser, dbPwd, dbName, unixSocketPath)

	// dbPool is the pool of database connections.
	dbPool, err := sql.Open("pgx", dbURI)
	if err != nil {
		return nil, fmt.Errorf("sql.Open: %v", err)
	}

	// ...

	return dbPool, nil
}

// [END indexHandler]
// [END gae_go111_app]
