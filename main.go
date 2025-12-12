package main

import (
	// "bufio"
	"encoding/csv"
	"errors"
	"fmt"
	"golang.org/x/term"
	"io"

	"github.com/inancgumus/screen"  // fixes windows terminal clear
	"github.com/mattn/go-colorable" // fixes windows terminal ANSI color output
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	// vlc "japanese_vocab_cmdline/vlc_control"

	_ "modernc.org/sqlite"
)

var startTime time.Time

func main() {
	// hide terminal cursor
	out := colorable.NewColorableStdout()
	fmt.Fprintln(out, "\033[?25l")       // hide cursor
	defer fmt.Fprintln(out, "\033[?25h") // show cursor

	startTime = time.Now()

	vocabList, err := LoadVocabCSVFile("add_words.csv")
	if err != nil {
		log.Fatal(err)
	}

	// for _, v := range vocabList {
	// 	fmt.Printf("%+v\n", v)
	// }

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

	mainMenu(db)
	//tcellDemo()

	return
}

func mainMenu(db VocabDB) error {

loop:
	for {
		screen.Clear()
		screen.MoveTopLeft()

		out := colorable.NewColorableStdout()
		fmt.Fprintln(out, Magenta, "  JAPANESE VOCAB TRAINER\n", ColorReset)
		fmt.Fprintln(out, Cyan, "  d.  Drill random words", ColorReset)
		fmt.Fprintln(out, Cyan, "  l.  List all words", ColorReset)
		fmt.Fprintln(out, Cyan, "  q.  Exit", ColorReset)

		// Get current terminal state so we can restore it
		oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
		if err != nil {
			panic(err)
		}
		defer term.Restore(int(os.Stdin.Fd()), oldState)

		var b [1]byte
		_, err = os.Stdin.Read(b[:])
		if err != nil {
			return err
		}
		ch := strings.ToLower(string(b[:]))

		switch ch {
		case "d":
			drillLoop(db)
		case "l":
			// todo list all vocab (use paginator?)
		case "q":
			break loop
		}
	}

	return nil
}

// return false when user gets through all items, return true when user wants to quit early
func drillRound(vocab []*Vocab) (bool, error) {

	out := colorable.NewColorableStdout()

	rand.Shuffle(len(vocab), func(i, j int) {
		vocab[i], vocab[j] = vocab[j], vocab[i]
	})

	numCorrect := 0
	currentIdx := 0

	for {
		screen.Clear()
		screen.MoveTopLeft()

		elapsedTime := time.Since(startTime)
		minutes := int(elapsedTime / time.Minute)
		seconds := int((elapsedTime % time.Minute) / time.Second)

		fmt.Printf(" DRILL WORDS (elpased time %d min %02d sec):\n\n", minutes, seconds)

		currentVocab := vocab[currentIdx]

		// print unanswered and wrong
		for _, v := range vocab {
			def := "\t\t" + v.Kana + ": " + v.Definition

			if v.DrillInfo.Correct {
				if v.DrillInfo.Wrong {
					// correct but not on first try
					fmt.Fprintln(out, Blue, "  ◯  ", v.Word, def)
				} else {
					// correct on first try
					fmt.Fprintln(out, Green, "  ◯  ", v.Word, def)
				}
			} else {
				if v.DrillInfo.Wrong {
					if v == currentVocab {
						if len(vocab)-numCorrect == 1 { // case of one remaining
							fmt.Fprintln(out, Yellow, "  🗙  ", v.Word, def)
						} else {
							fmt.Fprintln(out, Yellow, "  🗙  ", v.Word)
						}
					} else {
						fmt.Fprintln(out, Red, "  🗙  ", v.Word, def)
					}
				} else {
					// not yet answered
					if v == currentVocab {
						fmt.Fprintln(out, Yellow, "  -  ", v.Word)
					} else {
						fmt.Fprintln(out, Grey, "  -  ", v.Word)
					}
				}
			}
		}

		fmt.Fprintln(out)
		fmt.Fprintln(out, Cyan, "  a.  Mark yellow word as wrong", ColorReset)
		fmt.Fprintln(out, Cyan, "  d.  Mark yellow word as correct", ColorReset)
		fmt.Fprintln(out, Cyan, "  q.  Back to main menu", ColorReset)

		// Get current terminal state so we can restore it
		oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
		if err != nil {
			panic(err)
		}
		defer term.Restore(int(os.Stdin.Fd()), oldState)

		var b [1]byte
		_, err = os.Stdin.Read(b[:])
		if err != nil {
			return false, err
		}

		ch := strings.ToLower(string(b[:]))

		switch ch {
		case "a":
			currentVocab.DrillInfo.Wrong = true
		case "d":
			currentVocab.DrillInfo.Correct = true
			numCorrect++
		case "q":
			return true, nil
		}

		// pick the first item that is not correct and isn't at currentIdx
		// (for case of one remaining, stays at currentIdx)
		for i, v := range vocab {
			if i != currentIdx && !v.DrillInfo.Correct {
				currentIdx = i
				break
			}
		}

		if numCorrect == len(vocab) {
			return false, nil
		}
	}
}

func drillLoop(db VocabDB) error {
	allVocab, err := db.ListAll(true)
	if err != nil {
		return err
	}

	vocab, err := PickRandomN(allVocab, 12)
	if err != nil {
		log.Fatal(err)
	}

	for len(vocab) > 0 {
		quit, err := drillRound(vocab)
		if err != nil {
			return err
		}

		if quit {
			break
		}

		// filter out the non-wrong vocab and reset drill info
		// (increment and update drill counts of the non-wrong vocab)
		tmp := make([]*Vocab, 0)
		for _, v := range vocab {
			if v.DrillInfo.Wrong {
				tmp = append(tmp, v)
			} else {
				v.DrillCount++
				db.Update(v)
			}
			v.DrillInfo = Vocab{}.DrillInfo
		}
		vocab = tmp
	}

	return nil
}

func ParseVocabCSVLine(line string) (Vocab, error) {
	r := csv.NewReader(strings.NewReader(line))
	r.FieldsPerRecord = -1

	fields, err := r.Read()
	if err != nil {
		return Vocab{}, err
	}
	if len(fields) < 5 {
		return Vocab{}, errors.New("invalid line: expected 5 fields")
	}

	kanjiInfos := parseKanjiMeanings(fields[4])

	return Vocab{
		Word:          fields[0],
		Kana:          fields[1],
		PartOfSpeech:  fields[2],
		Definition:    fields[3],
		KanjiMeanings: kanjiInfos,
	}, nil
}

func parseKanjiMeanings(s string) []KanjiInfo {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}

	parts := strings.Split(s, "/")

	var result []KanjiInfo
	for _, part := range parts {
		p := strings.TrimSpace(part)
		if p == "" {
			continue
		}

		// Expect something like: "躊: meaning here"
		kv := strings.SplitN(p, ":", 2)
		if len(kv) != 2 {
			continue // skip malformed segments
		}

		meaning := strings.TrimSpace(kv[1])
		kv = strings.SplitN(kv[0], "(", 2)
		kanji := strings.TrimSpace(kv[0])
		pronunciation := strings.TrimSpace(kv[1])
		pronunciation = strings.TrimSuffix(pronunciation, ")")

		if kanji != "" && meaning != "" {
			result = append(result, KanjiInfo{
				Kanji:         kanji,
				Pronunciation: pronunciation,
				Meaning:       meaning,
			})
		}
	}

	return result
}

func LoadVocabCSVFile(path string) ([]Vocab, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	reader := csv.NewReader(f)
	reader.FieldsPerRecord = -1

	var vocabList []Vocab
	for {
		fields, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if len(fields) < 5 {
			continue // skip invalid lines
		}

		v := Vocab{
			Word:          fields[0],
			Kana:          fields[1],
			PartOfSpeech:  fields[2],
			Definition:    fields[3],
			KanjiMeanings: parseKanjiMeanings(fields[4]),
		}
		vocabList = append(vocabList, v)
	}

	return vocabList, nil
}

func VocabToCSVLine(v Vocab) string {
	// Serialize kanji meanings
	var kmParts []string
	for _, k := range v.KanjiMeanings {
		kmParts = append(kmParts, k.Kanji+"("+k.Pronunciation+"): "+k.Meaning)
	}
	kmField := strings.Join(kmParts, " / ")

	// Escape fields with quotes if necessary
	escape := func(s string) string {
		if strings.ContainsAny(s, `",`) {
			s = strings.ReplaceAll(s, `"`, `""`)
			return `"` + s + `"`
		}
		return s
	}

	return strings.Join([]string{
		escape(v.Word),
		escape(v.Kana),
		escape(v.PartOfSpeech),
		escape(v.Definition),
		escape(kmField),
	}, ",")
}

func SaveVocabCSVFile(path string, vocabList []Vocab) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	writer := csv.NewWriter(f)
	defer writer.Flush()

	// Write header
	writer.Write([]string{"word", "romaji", "part_of_speech", "definition", "kanji_meanings"})

	for _, v := range vocabList {
		line := VocabToCSVLine(v)
		// csv.Writer handles proper escaping again
		if err := writer.Write(strings.Split(line, ",")); err != nil {
			return err
		}
	}
	return nil
}
