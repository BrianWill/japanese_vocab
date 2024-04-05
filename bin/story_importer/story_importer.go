package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

const DB_PATH = "../users/test.db"

type Story struct {
	Title         string `json:"title,omitempty"`
	Date          int64  `json:"date,omitempty"`
	Link          string `json:"link,omitempty"`
	EpisodeNumber string `json:"episodeNumber,omitempty"`
	Audio         string `json:"audio,omitempty"`
	Video         string `json:"video,omitempty"`
	Content       string `json:"content,omitempty"`
	ContentFormat string `json:"contentFormat,omitempty"`
}

// where should user put the video and audio files? designated directories? let user specify the dest dir?

func main() {
	jsonPath := os.Args[1]
	basePath := os.Args[2] // the audio and video paths will be relative from this path

	jsonBytes, err := os.ReadFile(jsonPath)
	if err != nil {
		panic(err)
	}

	// parse the json file
	stories := make([]Story, 0)
	err = json.Unmarshal(jsonBytes, &stories)
	if err != nil {
		panic(err)
	}

	// open the db
	sqldb, err := sql.Open("sqlite3", DB_PATH)
	if err != nil {
		panic(err)
	}
	defer sqldb.Close()

	// if title is new, add the story to the catalog_stories database

	for _, s := range stories {
		fmt.Println(s)

		// check if story with same title already exists, and continue if so

		// (the story will be parsed into lines and words only when its added from the catalog to the main story table,
		// so in this importer, we just check that the data is valid)

		// check that the content can be parsed as the specified format

		// make audio and video paths relative from the base path (but check if they are urls first instead of paths)
		if isUrl(s.Video) {
			s.Video = basePath + "/" + s.Video
		}

		if isUrl(s.Audio) {
			s.Audio = basePath + "/" + s.Audio
		}
	}
}

func isUrl(s string) bool {
	return false
}
