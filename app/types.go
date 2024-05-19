package main

import (
	"encoding/xml"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CatalogStory struct {
	ID                  int64   `json:"id,omitempty"`
	Title               string  `json:"title,omitempty"`
	Source              string  `json:"source,omitempty"`
	Status              string  `json:"status,omitempty"`
	Date                string  `json:"date"`
	DateMarked          int     `json:"date_marked,omitempty"`
	EpisodeNumber       int     `json:"episode_number,omitempty"`
	Level               string  `json:"level,omitempty"`
	Content             string  `json:"content,omitempty"`
	ContentFormat       string  `json:"content_format,omitempty"`
	Link                string  `json:"link,omitempty"`
	Audio               string  `json:"audio,omitempty"`
	Video               string  `json:"video,omitempty"`
	TranscriptEN        string  `json:"transcript_en"`
	TranscriptJA        string  `json:"transcript_ja"`
	LifetimeRepetitions int64   `json:"lifetime_repetitions"`
	Words               []int64 `json:"words,omitempty"`
}

type StoryImport struct {
	Title         string `json:"title,omitempty"`
	Source        string `json:"source,omitempty"`
	Date          string `json:"date,omitempty"`
	EpisodeNumber string `json:"episode_number,omitempty"`
	Level         string `json:"level,omitempty"`
	Content       string `json:"content,omitempty"`
	ContentFormat string `json:"content_format,omitempty"`
	Link          string `json:"link,omitempty"`
	Audio         string `json:"audio,omitempty"`
	Video         string `json:"video,omitempty"`
	TranscriptEN  string `json:"transcript_en,omitempty"`
	TranscriptJA  string `json:"transcript_ja,omitempty"`
}

type ScheduleEntry struct {
	ID                  int64  `json:"id,omitempty"`
	Story               int64  `json:"story"`
	DayOffset           int64  `json:"day_offset"`
	Type                int64  `json:"type"`
	Title               string `json:"title,omitempty"`
	Source              string `json:"source,omitempty"`
	LifetimeRepetitions int64  `json:"lifetime_repetitions,omitempty"`
	Level               string `json:"level,omitempty"`
}

type ScheduleStoryRequest struct {
	ID    int64 `json:"id,omitempty"`    // used for removing a specific repetition
	Story int64 `json:"story,omitempty"` // used for adding/removing all reps of a story
}

type LogEntry struct {
	ID    int64 `json:"id,omitempty"`
	Story int64 `json:"story"`
	Date  int64 `json:"date"`
	Type  int64 `json:"type"`
}

type StoryImportJSON struct {
	Source        string        `json:"source,omitempty"`
	ContentFormat string        `json:"content_format,omitempty"`
	Stories       []StoryImport `json:"stories,omitempty"`
}

type WordInfo struct {
	Definitions []JMDictEntry `json:"definitions,omitempty"`
	DateMarked  int64         `json:"date_marked"`
	Audio       string        `json:"audio"`
	AudioStart  float32       `json:"audio_start"`
	AudioEnd    float32       `json:"audio_end"`
}

type StoryWordStatusUpdateRequest struct {
	ID        int64  `json:"id,omitempty"`
	OldStatus string `json:"old_status"`
	NewStatus string `json:"new_status"`
}

type LineWord struct {
	ID       int64  `json:"id,omitempty"`
	BaseForm string `json:"baseform,omitempty"`
	Surface  string `json:"surface"`
	POS      string `json:"pos,omitempty"` // highlight color
	Category int    `json:"Category,omitempty"`
}

type LineKanji struct {
	ID        int64  `json:"id,omitempty"`
	Character string `json:"character,omitempty"`
}

type LogEvent struct {
	ID      int64  `json:"id,omitempty"`
	StoryID int64  `json:"story_id,omitempty"`
	Date    int64  `json:"date,omitempty"`
	Title   string `json:"title,omitempty"`
}

type SplitLineRequest struct {
	StoryID     int64   `json:"story_id,omitempty"`
	LineToSplit int     `json:"line_to_split"`
	WordIdx     int     `json:"word_idx"`
	Timestamp   float64 `json:"timestamp"`
}

type ConsolidateLineRequest struct {
	StoryID      int64 `json:"story_id,omitempty"`
	LineToRemove int   `json:"line_to_remove,omitempty"`
}

type SetTimestampRequest struct {
	StoryID   int64   `json:"story_id,omitempty"`
	LineIdx   int     `json:"line_idx"`
	Timestamp float64 `json:"timestamp"`
}

type SetLineMarkRequest struct {
	StoryID int64 `json:"story_id,omitempty"`
	LineIdx int   `json:"line_idx"`
	Marked  bool  `json:"marked"`
}

type DrillRequest struct {
	StoryId int64 `json:"story_id"`
}

type EnqueueRequest struct {
	StoryId int64 `json:"story_id,omitempty"`
	Count   int   `json:"count,omitempty"`
}

type EnqueuedStory struct {
	Date    int    `json:"date"`
	ID      int64  `json:"id,omitempty"`
	StoryID int64  `json:"story_id,omitempty"`
	Title   string `json:"title,omitempty"`
	Link    string `json:"link,omitempty"`
}

type DrillWord struct {
	ID                  int64   `json:"id,omitempty"`
	BaseForm            string  `json:"base_form"`
	Status              string  `json:"status"`
	DateMarked          int64   `json:"date_marked"`
	Category            int     `json:"category"`
	Definitions         string  `json:"definitions,omitempty"`
	Audio               string  `json:"audio,omitempty"`
	AudioStart          float32 `json:"audio_start,omitempty"`
	AudioEnd            float32 `json:"audio_end,omitempty"`
	LifetimeRepetitions int     `json:"lifetime_repetitions"`
}

type WordUpdate struct {
	BaseForm   string  `json:"base_form"`
	Status     string  `json:"status"`
	DateMarked int64   `json:"date_marked"`
	Audio      string  `json:"audio"`
	AudioStart float32 `json:"audio_start"`
	AudioEnd   float32 `json:"audio_end"`
}

/* TOKENIZATION */

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
	//Entries          []JMDictEntry `json:"entries,omitempty" bson:"entries,omitempty"`
	// actually, the related words (component words and homynms) should be stored in monogo with the definition
	// also, should distinguish between words the user has encountered vs those related words which they haven't
	// ComponentWords []primitive.ObjectID `json:"componentWords" bson:"componentWords,omitempty"`
	// PitchHomonyms  []primitive.ObjectID `json:"pitchHomonyms" bson:"pitchHomonyms,omitempty"`
	// MoraHomonyms   []primitive.ObjectID `json:"moraHomonyms" bson:"moraHomonyms,omitempty"`
}

type Sentence struct {
	Words          []Word
	EndPunctuation string
}

type Word struct {
	Text       string             `json:"text,omitempty" bson:"text,omitempty"`
	Definition primitive.ObjectID `json:"definition,omitempty" bson:"definition,omitempty"`
	Form       string             `json:"form,omitempty" bson:"form,omitempty"` // e.g. 'ました' for a verb
}

type WordSearch struct {
	Word string `json:"word,omitempty" bson:"word,omitempty"`
}

// JMDict xml format
type JMDict struct {
	XMLName xml.Name      `xml:"JMDict"`
	Entries []JMDictEntry `xml:"entry" json:"entries"`
}

type Definition struct {
	Entries []JMDictEntry `bson:"entries, omitempty" json:"entries,omitempty"`
}

type JMDictEntry struct {
	//XMLName               *xml.Name          `xml:"entry" bson:"xmlname,omitempty" json:"xmlname,omitempty"`
	//ID                    primitive.ObjectID `bson:"_id, omitempty"`
	//Ent_seq               string             `xml:"ent_seq" bson:"sequence_number,omitempty" json:"sequence_number,omitempty"`
	Senses                []JMDictSense `xml:"sense" bson:"senses,omitempty" json:"senses,omitempty"`
	Readings              []JMDictR_ele `xml:"r_ele" bson:"readings,omitempty" json:"readings,omitempty"`
	KanjiSpellings        []JMDictK_ele `xml:"k_ele" bson:"kanji_spellings,omitempty" json:"kanji_spellings,omitempty"`
	ShortestKanjiSpelling int
	ShortestReading       int
}

type JMDictSense struct {
	//Stagk   []string        `xml:"stagk" bson:"restricted_to_kanji_spellings,omitempty" json:"restricted_to_kanji_spellings,omitempty"` //  indicate that the sense is restricted to the lexeme represented by the keb
	//Stagr   []string        `xml:"stagr" bson:"restricted_to_readings,omitempty" json:"restricted_to_readings,omitempty"` //  indicate that the sense is restricted to the lexeme represented by the reb
	Pos []string `xml:"pos" bson:"parts_of_speech,omitempty" json:"parts_of_speech,omitempty"` // part of speech
	//Ant     []string        `xml:"ant" bson:"antonyms,omitempty" json:"antonyms,omitempty"`               // ref to another entry which is an antonym of the current entry/sense
	Gloss []JMDictGloss `xml:"gloss" bson:"glosses,omitempty" json:"glosses,omitempty"`
	//Misc  []string      `xml:"misc" bson:"misc,omitempty" json:"misc,omitempty"`
	//Dial  []string      `xml:"dial" bson:"dialects,omitempty" json:"dialects,omitempty"` // associated with regional dialects in Japanese, the entity code for that dialect, e.g. ksb for Kansaiben.
	//Example []JMDictExample `xml:"example" bson:"examples,omitempty" json:"examples,omitempty"`
	//Xref    []string        `xml:"xref" bson:"related_words,omitempty" json:"related_words,omitempty"`
	//Lsource []JMDictLsource `xml:"lsource" bson:"source_languages,omitempty" json:"source_languages,omitempty"` // source language(s) of a loan-word/gairaigo
	//Field []string `xml:"field" bson:"applications,omitempty" json:"applications,omitempty"` // Information about the field of application of the entry/sense.
	//S_inf []string `xml:"s_inf" bson:"information,omitempty" json:"information,omitempty"`
}

type JMDictExample struct {
	Ex_srce *JMDictEx_srce  `xml:"ex_srce" bson:"source,omitempty" json:"source,omitempty"`
	Ex_text string          `xml:"ex_text" bson:"text,omitempty" json:"text,omitempty"`
	Ex_sent []JMDictEx_sent `xml:"ex_sent" bson:"sentence,omitempty" json:"sentence,omitempty"`
}

// reading element
type JMDictR_ele struct {
	Reading string `xml:"reb" bson:"reading,omitempty" json:"reading,omitempty"`
	//Re_nokanji string `xml:"re_nokanji" bson:"no_kanji,omitempty" json:"no_kanji,omitempty"`
	/* indicates that the reb, while associated with the keb,
	cannot be regarded as a true reading of the kanji. It is
	typically used for words such as foreign place names,
	gairaigo which can be in kanji or katakana, etc. */
	//Re_restr []string `xml:"re_restr" bson:"restrictions,omitempty" json:"restrictions,omitempty"` // reading only applies to a subset of the keb elements in the entry
	Re_inf []string `xml:"re_inf" bson:"information,omitempty" json:"information,omitempty"` // denotes orthography, e.g. okurigana irregularity
	//Re_pri   []string `xml:"re_pri" bson:"priority,omitempty" json:"priority,omitempty"`           // relative priority (see schema)
	Pitch string `bson:"pitch,omitempty" json:"pitch,omitempty"`
}

// kanji element
type JMDictK_ele struct {
	KanjiSpelling string `xml:"keb" bson:"kanji_spelling,omitempty" json:"kanji_spelling,omitempty"`
	//Ke_inf        []string `xml:"ke_inf" bson:"information,omitempty" json:"information,omitempty"` // denotes orthography, e.g. okurigana irregularity
	//Ke_pri        []string `xml:"ke_pri" bson:"priority,omitempty" json:"priority,omitempty"`       // relative priority (see schema)
}

type JMDictEx_srce struct {
	Exsrc_type string `xml:"exsrc_type,attr,omitempty" bson:"source_type,omitempty" json:"source_type,omitempty"`
	Value      string `xml:",chardata" bson:"value,omitempty" json:"value,omitempty"`
}

type JMDictEx_sent struct {
	Lang  string `xml:"xml:lang,attr,omitempty" bson:"language,omitempty" json:"language,omitempty"`
	Value string `xml:",chardata" bson:"value,omitempty" json:"value,omitempty"`
}

type JMDictLsource struct {
	Lang  string `xml:"xml:lang,attr,omitempty" bson:"language,omitempty" json:"language,omitempty"`
	Value string `xml:",chardata" bson:"value,omitempty" json:"value,omitempty"`
}

type JMDictGloss struct {
	Lang   string `xml:"xml:lang,attr,omitempty" bson:"language,omitempty" json:"language,omitempty"`
	G_type string `xml:"g_type,attr,omitempty" bson:"type,omitempty" json:"type,omitempty"`     // gloss is of a particular type, e.g. "lit" (literal), "fig" (figurative), "expl" (explanation).
	G_gend string `xml:"g_gend,attr,omitempty" bson:"gender,omitempty" json:"gender,omitempty"` //  gender of the gloss (typically a noun in the target language)
	Value  string `xml:",chardata" bson:"value,omitempty" json:"value,omitempty"`
}

// Kanji dicttionary

type KanjiDict struct {
	XMLName    xml.Name         `xml:"kanjidic2"`
	Characters []KanjiCharacter `xml:"character" json:"characters,omitempty"`
}

type KanjiCharacter struct {
	XMLName *xml.Name `xml:"character" json:"xmlname,omitempty"`
	Literal string    `xml:"literal" json:"literal,omitempty"`
	// Codepoint      []KanjiCodePoint      `xml:"cp_value"`
	Radical        *KanjiRadical        `xml:"radical,omitempty" json:"radical,omitempty"`
	Misc           *KanjiMisc           `xml:"misc,omitempty" json:"misc,omitempty"`
	ReadingMeaning *KanjiReadingMeaning `xml:"reading_meaning,omitempty" json:"readingmeaning,omitempty"`
}

// type KanjiCodePoint struct {
// 	Type  string `xml:"cp_type,attr,omitempty"`
// 	Value string `xml:",chardata"`
// }

type KanjiRadical struct {
	Values []KanjiRadicalValue `xml:"rad_value" json:"values,omitempty"`
}

type KanjiRadicalValue struct {
	Type  string `xml:"rad_type,attr,omitempty" json:"type,omitempty"`
	Value string `xml:",chardata" json:"value,omitempty"`
}

type KanjiMisc struct {
	Frequency   *int              `xml:"freq,omitempty" json:"frequency,omitempty"`
	StrokeCount *int              `xml:"stroke_count,omitempty" json:"stroke_count,omitempty"`
	Grade       *int              `xml:"grade,omitempty" json:"grade,omitempty"`
	JLPT        *int              `xml:"jlpt,omitempty" json:"jlpt,omitempty"`
	Variant     *KanjiMiscVariant `xml:"variant,omitempty" json:"variant,omitempty"`
}

type KanjiFrequency struct {
	Value string `xml:",chardata" json:"value,omitempty"`
}

type KanjiMiscVariant struct {
	Type  string `xml:"var_type,attr,omitempty" json:"type,omitempty"`
	Value string `xml:",chardata" json:"value,omitempty"`
}

type KanjiReadingMeaning struct {
	Group  []KanjiRMGroup `xml:"rmgroup,omitempty" json:"group,omitempty"`
	Nanori []string       `xml:"nanori.omitempty" json:"nanori,omitempty"`
}

type KanjiRMGroup struct {
	Reading []KanjiReading `xml:"reading,omitempty" json:"reading,omitempty"`
	Meaning []KanjiMeaning `xml:"meaning,omitempty" json:"meaning,omitempty"`
}

type KanjiReading struct {
	Value string `xml:",chardata" json:"value,omitempty"`
	Type  string `xml:"r_type,attr,omitempty" json:"type,omitempty"`
}

type KanjiMeaning struct {
	Value    string `xml:",chardata" json:"value,omitempty"`
	Language string `xml:"m_lang,attr,omitempty" json:"language,omitempty"`
}
