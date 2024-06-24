package main

import (
	"encoding/xml"
)

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

// JMDict xml format
type JMDict struct {
	XMLName xml.Name      `xml:"JMDict"`
	Entries []JMDictEntry `xml:"entry" json:"entries"`
}

type JMDictEntry struct {
	Senses                []JMDictSense `xml:"sense" bson:"senses,omitempty" json:"senses,omitempty"`
	Readings              []JMDictR_ele `xml:"r_ele" bson:"readings,omitempty" json:"readings,omitempty"`
	KanjiSpellings        []JMDictK_ele `xml:"k_ele" bson:"kanji_spellings,omitempty" json:"kanji_spellings,omitempty"`
	ShortestKanjiSpelling int
	ShortestReading       int
}

type JMDictSense struct {
	Pos   []string      `xml:"pos" bson:"parts_of_speech,omitempty" json:"parts_of_speech,omitempty"` // part of speech
	Gloss []JMDictGloss `xml:"gloss" bson:"glosses,omitempty" json:"glosses,omitempty"`
}

// reading element
type JMDictR_ele struct {
	Reading string `xml:"reb" bson:"reading,omitempty" json:"reading,omitempty"`
	//Re_nokanji string `xml:"re_nokanji" bson:"no_kanji,omitempty" json:"no_kanji,omitempty"`
	/* indicates that the reb, while associated with the keb,
	cannot be regarded as a true reading of the kanji. It is
	typically used for words such as foreign place names,
	gairaigo which can be in kanji or katakana, etc. */
	Re_inf []string `xml:"re_inf" bson:"information,omitempty" json:"information,omitempty"` // denotes orthography, e.g. okurigana irregularity
	Pitch  string   `bson:"pitch,omitempty" json:"pitch,omitempty"`
}

// kanji element
type JMDictK_ele struct {
	KanjiSpelling string `xml:"keb" bson:"kanji_spelling,omitempty" json:"kanji_spelling,omitempty"`
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
	XMLName        *xml.Name            `xml:"character" json:"xmlname,omitempty"`
	Literal        string               `xml:"literal" json:"literal,omitempty"`
	Radical        *KanjiRadical        `xml:"radical,omitempty" json:"radical,omitempty"`
	Misc           *KanjiMisc           `xml:"misc,omitempty" json:"misc,omitempty"`
	ReadingMeaning *KanjiReadingMeaning `xml:"reading_meaning,omitempty" json:"readingmeaning,omitempty"`
}

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
