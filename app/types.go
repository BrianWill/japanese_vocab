package main

import ()

type Story struct {
	ID          int64     `json:"id,omitempty"`
	Title       string    `json:"title,omitempty"`
	Source      string    `json:"source,omitempty"`
	Date        string    `json:"date,omitempty"`
	Content     string    `json:"content,omitempty"`
	Link        string    `json:"link,omitempty"`
	Video       string    `json:"video,omitempty"`
	DateLastRep int64     `json:"date_last_rep"`
	HasRepsTodo int       `json:"has_reps_todo"`
	SubtitlesEN string    `json:"subtitles_en,omitempty"`
	SubtitlesJA string    `json:"subtitles_ja,omitempty"`
	Excerpts    []Excerpt `json:"excerpts"`
}

type Excerpt struct {
	StoryID    int64       `json:"story_id,omitempty"` // used in update requests
	Hash       int64       `json:"hash"`               // a random id for the excerpt
	StartTime  float64     `json:"start_time"`
	EndTime    float64     `json:"end_time"`
	RepsTodo   int64       `json:"reps_todo"` // number of reps todo
	RepsLogged []LoggedRep `json:"reps_logged"`
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

type UpdateExcerptsRequest struct {
	StoryID  int64     `json:"story_id"`
	Excerpts []Excerpt `json:"excerpts"`
}

type ImportSourceRequest struct {
	Source string `json:"source"`
}

type IncWordsRequest struct {
	Words []int64 `json:"words,omitempty"` // the words whose repetitions needs to be incremented
}

type DrillRequest struct {
	StoryId     int64 `json:"story_id"`
	ExcerptHash int64 `json:"excerpt_hash"`
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
	BaseForm   string  `json:"base_form"`
	Archived   int64   `json:"archived"`
	Audio      string  `json:"audio"`
	AudioStart float32 `json:"audio_start"`
	AudioEnd   float32 `json:"audio_end"`
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
}
