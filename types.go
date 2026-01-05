package main

const (
	SCREEN_MAIN = iota
	SCREEN_DRILL
)

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

type IOErrorMsg struct {
	err error
}

type MenuState int

const (
	MAIN_MENU MenuState = iota
	DRILL_MENU
)

type MainModel struct {
	width      int
	height     int
	menuState  MenuState
	drillModel DrillModel
}

type DrillState struct {
	Vocab      []*Vocab
	CurrentIdx int
	NumCorrect int
	Done       bool
}

type DrillModel struct {
	DB                 VocabDB
	Vocab              []*Vocab
	CurrentIdx         int
	NumCorrect         int
	NumToRepeat        int
	AnsweredMarkersStr string
	PriorVocab         *Vocab
	Done               bool
}

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
		Weight    float32 // used in random selection
		IsWrong   bool    // has been answered wrong at least once this round
		IsCorrect bool    // has been answered correctly this round (but not necessarily the first time)
	}
}

type KanjiInfo struct {
	Kanji         string `json:"kanji"`
	Pronunciation string `json:"pronunciation"`
	Meaning       string `json:"meaning"`
}
