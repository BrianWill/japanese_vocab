package main

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/table"
)

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

const DRILL_CATEGORY_KATAKANA = 1
const DRILL_CATEGORY_ICHIDAN = 2
const DRILL_CATEGORY_GODAN_SU = 8
const DRILL_CATEGORY_GODAN_RU = 16
const DRILL_CATEGORY_GODAN_U = 32
const DRILL_CATEGORY_GODAN_TSU = 64
const DRILL_CATEGORY_GODAN_KU = 128
const DRILL_CATEGORY_GODAN_GU = 256
const DRILL_CATEGORY_GODAN_MU = 512
const DRILL_CATEGORY_GODAN_BU = 1024
const DRILL_CATEGORY_GODAN_NU = 2048
const DRILL_CATEGORY_KANJI = 4096
const DRILL_CATEGORY_GODAN = DRILL_CATEGORY_GODAN_SU | DRILL_CATEGORY_GODAN_RU | DRILL_CATEGORY_GODAN_U | DRILL_CATEGORY_GODAN_TSU |
	DRILL_CATEGORY_GODAN_KU | DRILL_CATEGORY_GODAN_GU | DRILL_CATEGORY_GODAN_MU | DRILL_CATEGORY_GODAN_BU | DRILL_CATEGORY_GODAN_NU

type IOErrorMsg struct {
	err error
}

type ReturnToMainMsg struct{}

type ExtractedVocabMsg struct {
	Text string
	// Words []*Vocab
	Words []*JpToken
}

type MenuState int

const (
	MAIN_MENU MenuState = iota
	DRILL_MENU
	EXTRACT_MENU
)

type MainModel struct {
	width        int
	height       int
	menuState    MenuState
	drillModel   DrillModel
	extractModel ExtractModel
}

type ExtractModel struct {
	IsLoaded bool
	Text     string
	WordList list.Model
	Vocab    []*Vocab
	Words    []*JpToken
}

type ExtractedWordItem struct {
	Base string
	Kana string
}

func (i ExtractedWordItem) FilterValue() string { return i.Base }
func (i ExtractedWordItem) Title() string       { return i.Base }
func (i ExtractedWordItem) Description() string { return i.Kana }

type ExtractKeysMap struct {
	addDrills    key.Binding
	removeDrills key.Binding
	discardWord  key.Binding
	archiveWord  key.Binding
}

func newExtractKeysMap() *ExtractKeysMap {
	return &ExtractKeysMap{
		addDrills: key.NewBinding(
			key.WithKeys("e"),
			key.WithHelp("e", "add drills"),
		),
		removeDrills: key.NewBinding(
			key.WithKeys("q"),
			key.WithHelp("q", "remove drills"),
		),
		discardWord: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "not a word"),
		),
		archiveWord: key.NewBinding(
			key.WithKeys("a"),
			key.WithHelp("a", "archive word"),
		),
	}
}

type DrillState struct {
	Vocab      []*Vocab
	CurrentIdx int
	NumCorrect int
	Done       bool
}

type DrillModel struct {
	DB          VocabDB
	Vocab       []*Vocab
	CurrentIdx  int
	NumCorrect  int
	NumToRepeat int
	Done        bool
	VocabTable  table.Model
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
		IsShown   bool
	}
}

type KanjiInfo struct {
	Kanji         string `json:"kanji"`
	Pronunciation string `json:"pronunciation"`
	Meaning       string `json:"meaning"`
}

type JpToken struct {
	Surface          string `json:"surface,omitempty" bson:"surface,omitempty"`
	WordId           int64  `json:"wordId,omitempty" bson:"wordId,omitempty"`
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
