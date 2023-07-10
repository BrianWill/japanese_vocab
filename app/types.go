package main

import (
	"encoding/xml"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Story struct {
	ID           int64  `json:"id,omitempty"`
	Words        string `json:"words,omitempty"`
	Content      string `json:"content,omitempty"`
	Title        string `json:"title,omitempty"`
	Link         string `json:"link,omitempty"`
	Tokens       string `json:"tokens,omitempty"`
	Status       int    `json:"status"`
	Countdown    int    `json:"countdown"`
	ReadCount    int    `json:"read_count"`
	DateLastRead int64  `json:"date_last_read"`
	DateAdded    int64  `json:"date_added"`
}

type StoryList struct {
	Stories []Story `json:"stories,omitempty"`
}

type DrillRequest struct {
	Count          int     `json:"count,omitempty"`
	WrongWithin    int64   `json:"wrong,omitempty"`
	Type           string  `json:"drill_type,omitempty"`
	StoryIds       []int64 `json:"storyIds,omitempty"`
	IgnoreCooldown bool    `json:"ignore_cooldown,omitempty"`
}

type DrillWord struct {
	ID            int64  `json:"id,omitempty"`
	BaseForm      string `json:"base_form"`
	Countdown     int    `json:"countdown"`
	CountdownMax  int    `json:"countdown_max"`
	DrillCount    int    `json:"drill_count"`
	ReadCount     int    `json:"read_count"`
	DateLastRead  int64  `json:"date_last_read"`
	DateLastDrill int64  `json:"date_last_drill"`
	Definitions   string `json:"definitions"`
	DrillType     int    `json:"drill_type"`
	DateLastWrong int64  `json:"date_last_wrong"`
	DateAdded     int64  `json:"date_added"`
}

type JpToken struct {
	Surface          string        `json:"surface,omitempty" bson:"surface,omitempty"`
	WordId           int64         `json:"wordId,omitempty" bson:"wordId,omitempty"`
	POS              string        `json:"pos,omitempty" bson:"pos"`
	POS_1            string        `json:"pos1,omitempty" bson:"pos1"`
	POS_2            string        `json:"pos2,omitempty" bson:"pos2"`
	POS_3            string        `json:"pos3,omitempty" bson:"pos3"`
	InflectionalType string        `json:"inflectionalType,omitempty" bson:"inflectionalType"`
	InflectionalForm string        `json:"inflectionalForm,omitempty" bson:"inflectionalForm"`
	BaseForm         string        `json:"baseForm,omitempty" bson:"baseForm"`
	Reading          string        `json:"reading,omitempty" bson:"reading"`
	Pronunciation    string        `json:"pronunciation,omitempty" bson:"pronunciation"`
	Entries          []JMDictEntry `json:"entries,omitempty" bson:"entries,omitempty"`
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

type JMDictEntry struct {
	XMLName               *xml.Name          `xml:"entry" bson:"xmlname,omitempty" json:"xmlname,omitempty"`
	ID                    primitive.ObjectID `bson:"_id, omitempty"`
	Ent_seq               string             `xml:"ent_seq" bson:"sequence_number,omitempty" json:"sequence_number,omitempty"`
	Senses                []JMDictSense      `xml:"sense" bson:"senses,omitempty" json:"senses,omitempty"`
	Readings              []JMDictR_ele      `xml:"r_ele" bson:"readings,omitempty" json:"readings,omitempty"`
	KanjiSpellings        []JMDictK_ele      `xml:"k_ele" bson:"kanji_spellings,omitempty" json:"kanji_spellings,omitempty"`
	ShortestKanjiSpelling int
	ShortestReading       int
}

type JMDictSense struct {
	Stagk   []string        `xml:"stagk" bson:"restricted_to_kanji_spellings,omitempty" json:"restricted_to_kanji_spellings,omitempty"` //  indicate that the sense is restricted to the lexeme represented by the keb
	Stagr   []string        `xml:"stagr" bson:"restricted_to_readings,omitempty" json:"restricted_to_readings,omitempty"`               //  indicate that the sense is restricted to the lexeme represented by the reb
	Pos     []string        `xml:"pos" bson:"parts_of_speech,omitempty" json:"parts_of_speech,omitempty"`                               // part of speech
	Ant     []string        `xml:"ant" bson:"antonyms,omitempty" json:"antonyms,omitempty"`                                             // ref to another entry which is an antonym of the current entry/sense
	Gloss   []JMDictGloss   `xml:"gloss" bson:"glosses,omitempty" json:"glosses,omitempty"`
	Misc    []string        `xml:"misc" bson:"misc,omitempty" json:"misc,omitempty"`
	Dial    []string        `xml:"dial" bson:"dialects,omitempty" json:"dialects,omitempty"` // associated with regional dialects in Japanese, the entity code for that dialect, e.g. ksb for Kansaiben.
	Example []JMDictExample `xml:"example" bson:"examples,omitempty" json:"examples,omitempty"`
	Xref    []string        `xml:"xref" bson:"related_words,omitempty" json:"related_words,omitempty"`
	Lsource []JMDictLsource `xml:"lsource" bson:"source_languages,omitempty" json:"source_languages,omitempty"` // source language(s) of a loan-word/gairaigo
	Field   []string        `xml:"field" bson:"applications,omitempty" json:"applications,omitempty"`           // Information about the field of application of the entry/sense.
	S_inf   []string        `xml:"s_inf" bson:"information,omitempty" json:"information,omitempty"`
}

type JMDictExample struct {
	Ex_srce *JMDictEx_srce  `xml:"ex_srce" bson:"source,omitempty" json:"source,omitempty"`
	Ex_text string          `xml:"ex_text" bson:"text,omitempty" json:"text,omitempty"`
	Ex_sent []JMDictEx_sent `xml:"ex_sent" bson:"sentence,omitempty" json:"sentence,omitempty"`
}

// reading element
type JMDictR_ele struct {
	Reading    string `xml:"reb" bson:"reading,omitempty" json:"reading,omitempty"`
	Re_nokanji string `xml:"re_nokanji" bson:"no_kanji,omitempty" json:"no_kanji,omitempty"`
	/* indicates that the reb, while associated with the keb,
	cannot be regarded as a true reading of the kanji. It is
	typically used for words such as foreign place names,
	gairaigo which can be in kanji or katakana, etc. */
	Re_restr []string `xml:"re_restr" bson:"restrictions,omitempty" json:"restrictions,omitempty"` // reading only applies to a subset of the keb elements in the entry
	Re_inf   []string `xml:"re_inf" bson:"information,omitempty" json:"information,omitempty"`     // denotes orthography, e.g. okurigana irregularity
	Re_pri   []string `xml:"re_pri" bson:"priority,omitempty" json:"priority,omitempty"`           // relative priority (see schema)
	Pitch    string   `bson:"pitch,omitempty" json:"pitch,omitempty"`
}

// kanji element
type JMDictK_ele struct {
	KanjiSpelling string   `xml:"keb" bson:"kanji_spelling,omitempty" json:"kanji_spelling,omitempty"`
	Ke_inf        []string `xml:"ke_inf" bson:"information,omitempty" json:"information,omitempty"` // denotes orthography, e.g. okurigana irregularity
	Ke_pri        []string `xml:"ke_pri" bson:"priority,omitempty" json:"priority,omitempty"`       // relative priority (see schema)
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
