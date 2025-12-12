package main

type Vocab struct {
	ID            int
	Word          string      `json:"word"`
	Kana          string      `json:"kana"` // for pronunciation (not necessarily a common spelling of the word)
	PartOfSpeech  string      `json:"part_of_speech"`
	Definition    string      `json:"definition"`
	KanjiMeanings []KanjiInfo `json:"kanji_meanings"`
	DrillCount    int         `json:"drill_count"`
	Archived      bool        `json:"archived"`
	DrillInfo     struct {    // used in drills
		Weight  float32 // used in random selection
		Status  int     // see DRILL_STATUS_*
		Wrong   bool    // has been answered wrong at least once this round
		Correct bool    // has been answered correctly this round (but not necessarily the first time)
	}
}

const (
// DRILL_STATUS_UNANSWERED           = iota
// DRILL_STATUS_CORRECT_WILL_DISCARD // answered correctly first time and so will be discarded in next round
// DRILL_STATUS_CORRECT              // answered correctly this round but only after wrong answer (so will reappear next round)
// DRILL_STATUS_WRONG_1              // the most recently wrong
// DRILL_STATUS_WRONG_2              // the second most recently wrong
// DRILL_STATUS_DISCARD              // was answered correct on first try in prior round
)

type KanjiInfo struct {
	Kanji         string `json:"kanji"`
	Pronunciation string `json:"pronunciation"`
	Meaning       string `json:"meaning"`
}

const (
	ColorReset         = "\033[0m"
	Red                = "\033[31m"
	Green              = "\033[32m"
	Yellow             = "\033[33m"
	Blue               = "\033[34m"
	Magenta            = "\033[35m"
	Cyan               = "\033[36m"
	White              = "\033[37m"
	Grey               = "\x1b[90m"
	WordsToAddPath     = "words_add.csv"
	WordsToArchivePath = "words_archive.csv" // words that we're done drilling for good
)
