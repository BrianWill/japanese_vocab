package main

import (
	"log"
	"time"
	// vlc "japanese_vocab_cmdline/vlc_control"
	_ "modernc.org/sqlite"
)

var startTime time.Time = time.Now()

func main() {
	vocabList, err := LoadVocabCSVFile("add_words.csv")
	if err != nil {
		log.Fatal(err)
	}

	// // Optionally save back
	// err = SaveVocabCSVFile("vocab_out.csv", vocabList)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	db, err := InitVocabDB()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// upsert all words into the db
	err = db.UpsertAll(vocabList)
	if err != nil {
		log.Fatal(err)
	}

	tuiLoop(db)
}
