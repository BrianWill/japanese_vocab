package main

import ()

type Story struct {
	ID                int64   `json:"id,omitempty"`
	Title             string  `json:"title,omitempty"`
	Source            string  `json:"source,omitempty"`
	Date              string  `json:"date,omitempty"`
	Link              string  `json:"link,omitempty"`
	Video             string  `json:"video,omitempty"`
	WordCount         int     `json:"word_count"`
	ArchivedWordCount int     `json:"archived_word_count"`
	SubtitlesENJson   string  `json:"subtitles_en,omitempty"`
	SubtitlesJAJson   string  `json:"subtitles_ja,omitempty"`
	SubtitlesENOffset float64 `json:"subtitles_en_offset"`
	SubtitlesJAOffset float64 `json:"subtitles_ja_offset"`
	SubtitlesEN       []Subtitle
	SubtitlesJA       []Subtitle
	Log               []LogItem `json:"log"`
}

type LoggedRep struct {
	Date int64 `json:"date,omitempty"`
}

// paths are relative from source dir
type StoryFilePaths struct {
	Name              string
	Video             string
	JapaneseSubtitles string
	EnglishSubtitles  string
}

type LogStoryRequest struct {
	StoryID int64 `json:"story_id"`
	Date    int64 `json:"date"`
}

type ImportSourceRequest struct {
	Source string `json:"source"`
}

type IncWordsRequest struct {
	Words []int64 `json:"words,omitempty"` // the words whose repetitions needs to be incremented
}

type DrillRequest struct {
	StoryId int64 `json:"story_id"`
}

type DrillWord struct {
	ID          int64  `json:"id,omitempty"`
	BaseForm    string `json:"base_form"`
	Archived    int64  `json:"archived"`
	Category    int    `json:"category"`
	Definitions string `json:"definitions,omitempty"`
	Repetitions int    `json:"repetitions"`
	DateLastRep int64  `json:"date_last_rep"`
}

type WordUpdate struct {
	BaseForm string `json:"base_form"`
	Archived int64  `json:"archived"`
	// Audio      string  `json:"audio"`
	// AudioStart float32 `json:"audio_start"`
	// AudioEnd   float32 `json:"audio_end"`
}

type OpenTranscriptRequest struct {
	StoryName  string `json:"story_name"`
	SourceName string `json:"source_name"`
	Lang       string `json:"lang"`
}

type ImportStoryRequest struct {
	StoryTitle string `json:"story_title"`
	Source     string `json:"source"`
}

type Subtitle struct {
	StartTime float64 `json:"start_time"`
	EndTime   float64 `json:"end_time"`
	Text      string  `json:"text"`
	Words     []Word  `json:"words"`
}

type Word struct {
	/* the text that gets displayed (including trailing whitespace and punctuation)
	The first word of a line will also potentially include leading puntuation. */
	Display string `json:"display,omitempty"`

	BaseForm string `json:"base_form"`
	POS      string `json:"POS"`
}

type LogItem struct {
	Date int64 `json:"date"`
}
