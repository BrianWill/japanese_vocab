package main

import (
	"archive/zip"
	"database/sql"
	// "fmt"
	"io"
	"os"

	"go.mongodb.org/mongo-driver/bson"
)

func unzipSource(path string) ([]byte, error) {
	reader, err := zip.OpenReader(path)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	bytes, err := unzipFile(reader.File[0])
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

func unzipFile(f *zip.File) ([]byte, error) {
	zippedFile, err := f.Open()
	if err != nil {
		return nil, err
	}
	defer zippedFile.Close()

	bytes, err := io.ReadAll(zippedFile)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

func saveStoryDump() {
	sqldb, err := sql.Open("sqlite3", SQL_FILE)
	if err != nil {
		panic(err)
	}
	defer sqldb.Close()

	rows, err := sqldb.Query(`SELECT id, content, title, link FROM stories WHERE user = $1;`, 0)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	var stories []Story
	for rows.Next() {
		var story Story
		if err := rows.Scan(&story.ID, &story.Content, &story.Title, &story.Link); err != nil {
			panic(err)
		}
		stories = append(stories, story)
	}

	bytes, err := bson.Marshal(StoryList{Stories: stories})
	if err != nil {
		panic(err)
	}

	err = os.WriteFile("../stories_temp.bson", bytes, 0644)
	if err != nil {
		panic(err)
	}
}

func loadStoryDump() (StoryList, error) {
	file, err := os.Open("../stories_temp.bson")
	if err != nil {
		return StoryList{}, err
	}
	defer file.Close()

	bytes, err := io.ReadAll(file)
	if err != nil {
		return StoryList{}, err
	}
	var storyList StoryList
	err = bson.Unmarshal(bytes, &storyList)
	if err != nil {
		return StoryList{}, err
	}
	return storyList, nil
}
