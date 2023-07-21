package main

import (
	"bufio"
	"encoding/xml"
	"fmt"
	"regexp"

	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	//"github.com/ikawaha/kagome-dict/ipa"
	"github.com/ikawaha/kagome/v2/tokenizer"

	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	//"github.com/go-xmlfmt/xmlfmt"

	//"io/ioutil"
	"os"
	"strconv"
	"strings"
)

var client *mongo.Client
var db *mongo.Database
var wiktionaryCollection *mongo.Collection
var dictCollection *mongo.Collection

var tok *tokenizer.Tokenizer

const SQL_FILE = "../ja.db"

func main() {
	var err error

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err = mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))

	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()

	db = client.Database("JapaneseEnglish")

	dumpStoriesEndpoint()
	//updatePitch()
}

type StorySql struct {
	ID      int64  `json:"id,omitempty"`
	State   string `json:"state,omitempty"`
	Words   string `json:"words,omitempty"`
	Content string `json:"content,omitempty"`
	Title   string `json:"title,omitempty"`
	Link    string `json:"link,omitempty"`
	Tokens  string `json:"tokens,omitempty"`
}

type StoryList struct {
	Stories []StorySql `json:"stories,omitempty"`
}

func dumpStoriesEndpoint() {
	sqldb, err := sql.Open("sqlite3", SQL_FILE)
	if err != nil {
		panic(err)
	}
	defer sqldb.Close()

	rows, err := sqldb.Query(`SELECT id, state, content, title, link FROM stories WHERE user = $1;`, 0)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	var stories []StorySql
	for rows.Next() {
		var story StorySql
		if err := rows.Scan(&story.ID, &story.State, &story.Content, &story.Title, &story.Link); err != nil {
			panic(err)
		}
		stories = append(stories, story)
	}

	bytes, err := bson.Marshal(StoryList{Stories: stories})
	if err != nil {
		panic(err)
	}

	err = os.WriteFile("../stories_temp.bson", bytes, 0644)
	if err != nil {
		panic(err)
	}
}

func updatePitch() {
	fmt.Println("updating pitch")

	jmdictCollection := db.Collection("jmdict")

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	err := client.Ping(ctx, readpref.Primary())
	if err != nil {
		panic(err)
	}

	lines := LinesInFile(`./accents.txt`)
	nKanjiWords := 0
	nNoKanjiWords := 0
	updatedDocs := make(map[primitive.ObjectID]map[int]int)
	nNoMatch := 0
	nCollisions := 0

	for _, line := range lines[117000:] {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		words := strings.Fields(line)

		var query primitive.D
		var kanji string
		var reading string
		var pitch string

		if len(words) == 3 {
			kanji = words[0]
			reading = words[1]
			pitch = words[2]
			query = bson.D{
				{"kanji_spellings.kanji_spelling", kanji},
				{"readings.reading", reading}}
			nKanjiWords++
		} else {
			reading = words[0]
			pitch = words[1]
			query = bson.D{
				{"kanji_spellings", bson.D{{"$exists", false}}},
				{"readings.reading", reading}}
			nNoKanjiWords++
		}

		cursor, err := jmdictCollection.Find(ctx, query)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer cursor.Close(ctx)

		entries := make([]JMDictEntry, 0)
		for cursor.Next(ctx) {
			var entry JMDictEntry
			cursor.Decode(&entry)
			entries = append(entries, entry)
		}

		if len(entries) == 0 {
			nNoMatch++
			fmt.Println("no match: ", kanji, reading, pitch)
			continue
		}

		for _, entry := range entries {
			for idx, ele := range entry.R_ele {
				if ele.Reb == reading {
					//fmt.Println(entry.ID, kanji, reading, pitch)
					_, err := jmdictCollection.UpdateByID(ctx, entry.ID,
						bson.D{{"$set", bson.D{{"readings." + strconv.Itoa(idx) + ".pitch", pitch}}}})
					if err != nil {
						fmt.Println(err)
						return
					}

					if obj, ok := updatedDocs[entry.ID]; ok {
						if val, ok := obj[idx]; ok {
							obj[idx] = val + 1
							fmt.Println("collision: ", kanji, reading)
							if val == 1 {
								nCollisions++
							}
						}
					} else {
						updatedDocs[entry.ID] = make(map[int]int)
						updatedDocs[entry.ID][idx] = 1
					}
					break
				}
			}
		}
	}
	fmt.Println(nKanjiWords, nNoKanjiWords)
	fmt.Printf("no matches: %v, collisions: %v", nNoMatch, nCollisions)
}

func LinesInFile(fileName string) []string {
	f, err := os.Open(fileName)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	defer f.Close()
	// Create new Scanner.
	scanner := bufio.NewScanner(f)
	result := []string{}
	// Use Scan.
	for scanner.Scan() {
		line := scanner.Text()
		// Append line to result.
		result = append(result, line)
	}
	return result
}

func parseKanjiDict() {
	dictCollection = db.Collection("kanjidict")

	xmlFile, err := os.Open("./kanjidic2.xml")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Successfully Opened xml")
	// defer the closing of our xmlFile so that we can parse it later on
	defer xmlFile.Close()

	characters := make([]KanjiCharacter, 0)

	decoder := xml.NewDecoder(xmlFile)

	for {
		// Read tokens from the XML document in a stream.
		t, _ := decoder.Token()
		if t == nil {
			break
		}
		// Inspect the type of the token just read.
		switch ele := t.(type) {
		case xml.StartElement:
			if ele.Name.Local == "character" {
				var char KanjiCharacter
				err = decoder.DecodeElement(&char, &ele)
				if err != nil {
					panic(err)
				}
				characters = append(characters, char)
			}
		default:
		}
	}

	fmt.Println("num characters: ", len(characters))

	// outputFile, err := os.Create("./kanjidict_output.xml")
	// if err != nil {
	// 	fmt.Printf("opening output file error: %v\n", err)
	// }
	// fmt.Println("Successfully Opened output xml")
	// // defer the closing of our xmlFile so that we can parse it later on
	// defer outputFile.Close()

	// for _, ch := range characters {

	// 	output, err := xml.MarshalIndent(ch, "", "    ")
	// 	if err != nil {
	// 		fmt.Printf("marshall error: %v\n", err)
	// 	}

	// 	_, err = outputFile.Write(output)
	// 	if err != nil {
	// 		fmt.Printf("write file error: %v\n", err)
	// 	}
	// 	_, err = outputFile.Write([]byte{'\n'})
	// 	if err != nil {
	// 		fmt.Printf("write file error: %v\n", err)
	// 	}
	// }

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		panic(err)
	}

	//entries = entries[:100000]
	characters_ := make([]interface{}, len(characters))

	for i, entry := range characters {
		//entry.Sense = nil
		entry.XMLName = nil
		characters_[i] = characters[i]
	}

	//_, err := dictCollection.InsertOne(ctx, entry)
	_, err = dictCollection.InsertMany(ctx, characters_)
	if err != nil {
		fmt.Println(err)
		return
	}

}

func parseJmdict() {
	dictCollection = db.Collection("jmdict")

	xmlFile, err := os.Open("./JMdict_test.xml")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Successfully Opened xml")
	// defer the closing of our xmlFile so that we can parse it later on
	defer xmlFile.Close()

	entries := make([]JMDictEntry, 0)

	decoder := xml.NewDecoder(xmlFile)

	// re, err := regexp.Compile(`^[\w\s]+$`)
	// if err != nil {
	// 	panic(err)
	// }

	for {
		// Read tokens from the XML document in a stream.
		t, _ := decoder.Token()
		if t == nil {
			break
		}
		// Inspect the type of the token just read.
		switch ele := t.(type) {
		case xml.StartElement:
			if ele.Name.Local == "entry" {
				var entry JMDictEntry
				err = decoder.DecodeElement(&entry, &ele)
				if err != nil {
					panic(err)
				}

				entries = append(entries, entry)
			}
		default:
		}
	}

	fmt.Println("done decoding. Number of entries:", len(entries))

	// outputFile, err := os.Create("./JMdict_test_output.xml")
	// if err != nil {
	// 	fmt.Printf("opening output file error: %v\n", err)
	// }
	// fmt.Println("Successfully Opened output xml")
	// // defer the closing of our xmlFile so that we can parse it later on
	// defer outputFile.Close()

	// for _, e := range entries {

	// 	output, err := xml.MarshalIndent(e, "", "    ")
	// 	if err != nil {
	// 		fmt.Printf("marshall error: %v\n", err)
	// 	}

	// 	_, err = outputFile.Write(output)
	// 	if err != nil {
	// 		fmt.Printf("write file error: %v\n", err)
	// 	}
	// 	_, err = outputFile.Write([]byte{'\n'})
	// 	if err != nil {
	// 		fmt.Printf("write file error: %v\n", err)
	// 	}
	// }

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		panic(err)
	}

	//entries = entries[:100000]
	entries_ := make([]interface{}, len(entries))

	for i, entry := range entries {
		//entry.Sense = nil
		entry.XMLName = nil
		entries_[i] = entries[i]
	}

	//_, err := dictCollection.InsertOne(ctx, entry)
	_, err = dictCollection.InsertMany(ctx, entries_)
	if err != nil {
		fmt.Println(err)
		return
	}

}
func parseWiktionary() {

	wiktionaryCollection = db.Collection("wiktionary")

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	err := client.Ping(ctx, readpref.Primary())
	if err != nil {
		panic(err)
	}
	//makeCollection()

	words := processCollection()

	for _, w := range words {
		err := createWord(w)
		if err != nil {
			fmt.Println(err)
			break
		}
	}
}

func createWord(word Word) error {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	_, err := wiktionaryCollection.InsertOne(ctx, word)
	return err
}

func processCollection() []Word {
	// Open our xmlFile
	xmlFile, err := os.Open("../../enwiktionary-collection.xml")
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("Successfully Opened xml")
	// defer the closing of our xmlFile so that we can parse it later on
	defer xmlFile.Close()

	words := make([]Word, 0)

	decoder := xml.NewDecoder(xmlFile)

	re, err := regexp.Compile(`^[\w\s]+$`)
	if err != nil {
		panic(err)
	}

	for {
		// Read tokens from the XML document in a stream.
		t, _ := decoder.Token()
		if t == nil {
			break
		}
		// Inspect the type of the token just read.
		switch se := t.(type) {
		case xml.StartElement:
			if se.Name.Local == "word" {
				var word Word
				decoder.DecodeElement(&word, &se)

				if !re.MatchString(word.Text) {
					words = append(words, word)
					//fmt.Println(word.Text)
				} else {
					//fmt.Println(word.Text)
				}
			}
		default:
		}
	}

	fmt.Println("done decoding. Number of words:", len(words))

	return words
}

func makeCollection() {
	// Open our xmlFile
	xmlFile, err := os.Open("../../enwiktionary-20221101-pages-meta-current.xml")
	//xmlFile, err := os.Open("./text.xml")
	// if we os.Open returns an error then handle it
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("Successfully Opened xml")
	// defer the closing of our xmlFile so that we can parse it later on
	defer xmlFile.Close()

	words := make([]Word, 0)

	decoder := xml.NewDecoder(xmlFile)

	for {
		// Read tokens from the XML document in a stream.
		t, _ := decoder.Token()
		if t == nil {
			break
		}
		// Inspect the type of the token just read.
		switch se := t.(type) {
		case xml.StartElement:
			if se.Name.Local == "page" {
				var page Page
				decoder.DecodeElement(&page, &se)
				if page.Namespace == 0 && strings.Contains(page.Revision.Text, "==Japanese==") {
					words = append(words, Word{page.Title, page.Revision.Text})
					fmt.Println("Number of pages:", strconv.Itoa(len(words)), page.Title)
				}
			}
		default:
		}
	}

	outputFile, err := os.Create("../../enwiktionary-collection.xml")
	if err != nil {
		fmt.Println(err)
	}

	defer outputFile.Close()

	encoder := xml.NewEncoder(outputFile)

	encoder.Indent("", "  ")
	if err := encoder.Encode(WordCollection{words}); err != nil {
		panic(err)
	}

	fmt.Println("done encoding")
}

func parseRevisionText(text string) Word {
	fmt.Println(text)
	var word Word
	return word
}

type WordCollection struct {
	Words []Word `xml:"word"`
}

type Word struct {
	Text       string `xml:"text"`
	Definition string `xml:"definition"`
}

type Entry struct {
	R_ele R_ele `xml:"r_ele"`
	Sense Sense `xml:"sense"`
}

type R_ele struct {
	Reb string `xml:"reb"`
}

type Sense struct {
}

// 	<ent_seq>1073210</ent_seq>
// <r_ele>
// <reb>スポーツ</reb>
// <re_pri>gai1</re_pri>
// <re_pri>ichi1</re_pri>
// </r_ele>
// <sense>
// <pos>noun</pos>
// <pos>adjective-no</pos>
// <gloss>sport</gloss>
// <gloss>sports</gloss>
// <example>
// <ex_srce exsrc_type="tat">200846</ex_srce>
// <ex_text>スポーツ</ex_text>
// <ex_sent xml:lang="jpn">ところであなたはどんなスポーツが好きですか。</ex_sent>
// <ex_sent xml:lang="eng">Well, what sports do you like?</ex_sent>
// </example>
// </sense>

type Revision struct {
	/*
	   	<id>1692687</id>
	         <timestamp>2022-08-20T11:55:00Z</timestamp>
	         <contributor>
	           <username>Mtodo</username>
	           <id>329</id>
	         </contributor>
	         <comment>[[权现]]を繁体字化</comment>
	         <model>wikitext</model>
	         <format>text/x-wiki</format>
	         <text
	*/
	Id        int    `xml:"id"`
	Timestamp string `xml:"timestamp"`
	Format    string `xml:"format"`
	Text      string `xml:"text"`
}

type Page struct {
	Title     string   `xml:"title"`
	Namespace int      `xml:"ns"`
	Id        int      `xml:"id"`
	Revision  Revision `xml:"revision"`
}

type MediaWiki struct {
	Siteinfo string `xml:"siteinfo"`
	Pages    []Page `xml:"page"`
}

// JMDict xml format

type JMDict struct {
	XMLName xml.Name      `xml:"JMDict"`
	Entry   []JMDictEntry `xml:"entry"`
}

type JMDictEntry struct {
	XMLName *xml.Name          `xml:"entry" bson:"xmlname,omitempty"`
	ID      primitive.ObjectID `bson:"_id, omitempty"`
	Ent_seq string             `xml:"ent_seq" bson:"sequence_number,omitempty"`
	Sense   []JMDictSense      `xml:"sense" bson:"senses,omitempty"`
	R_ele   []JMDictR_ele      `xml:"r_ele" bson:"readings,omitempty"`
	K_ele   []JMDictK_ele      `xml:"k_ele" bson:"kanji_spellings,omitempty"`
}

type JMDictSense struct {
	Stagk   []string        `xml:"stagk" bson:"restricted_to_kanji_spellings,omitempty"` //  indicate that the sense is restricted to the lexeme represented by the keb
	Stagr   []string        `xml:"stagr" bson:"restricted_to_readings,omitempty"`        //  indicate that the sense is restricted to the lexeme represented by the reb
	Pos     []string        `xml:"pos" bson:"parts_of_speech,omitempty"`                 // part of speech
	Ant     []string        `xml:"ant" bson:"antonyms,omitempty"`                        // ref to another entry which is an antonym of the current entry/sense
	Gloss   []JMDictGloss   `xml:"gloss" bson:"glosses,omitempty"`
	Misc    []string        `xml:"misc" bson:"misc,omitempty"`
	Dial    []string        `xml:"dial" bson:"dialects,omitempty"` // associated with regional dialects in Japanese, the entity code for that dialect, e.g. ksb for Kansaiben.
	Example []JMDictExample `xml:"example" bson:"examples,omitempty"`
	Xref    []string        `xml:"xref" bson:"related_words,omitempty"`
	Lsource []JMDictLsource `xml:"lsource" bson:"source_languages,omitempty"` // source language(s) of a loan-word/gairaigo
	Field   []string        `xml:"field" bson:"applications,omitempty"`       // Information about the field of application of the entry/sense.
	S_inf   []string        `xml:"s_inf" bson:"information,omitempty"`
}

type JMDictExample struct {
	Ex_srce *JMDictEx_srce  `xml:"ex_srce" bson:"source,omitempty"`
	Ex_text string          `xml:"ex_text" bson:"text,omitempty"`
	Ex_sent []JMDictEx_sent `xml:"ex_sent" bson:"sentence,omitempty"`
}

// reading element
type JMDictR_ele struct {
	Reb        string `xml:"reb" bson:"reading,omitempty"`
	Re_nokanji string `xml:"re_nokanji" bson:"no_kanji,omitempty"`
	/* indicates that the reb, while associated with the keb,
	cannot be regarded as a true reading of the kanji. It is
	typically used for words such as foreign place names,
	gairaigo which can be in kanji or katakana, etc. */
	Re_restr []string `xml:"re_restr" bson:"restrictions,omitempty"` // reading only applies to a subset of the keb elements in the entry
	Re_inf   []string `xml:"re_inf" bson:"information,omitempty"`    // denotes orthography, e.g. okurigana irregularity
	Re_pri   []string `xml:"re_pri" bson:"priority,omitempty"`       // relative priority (see schema)
	Pitch    string   `bson:"pitch,omitempty"`
}

// kanji element
type JMDictK_ele struct {
	Keb    string   `xml:"keb" bson:"kanji_spelling,omitempty"`
	Ke_inf []string `xml:"ke_inf" bson:"information,omitempty"` // denotes orthography, e.g. okurigana irregularity
	Ke_pri []string `xml:"ke_pri" bson:"priority,omitempty"`    // relative priority (see schema)
}

type JMDictEx_srce struct {
	Exsrc_type string `xml:"exsrc_type,attr,omitempty" bson:"source_type,omitempty"`
	Value      string `xml:",chardata" bson:"value,omitempty"`
}

type JMDictEx_sent struct {
	Lang  string `xml:"xml:lang,attr,omitempty" bson:"language,omitempty"`
	Value string `xml:",chardata" bson:"value,omitempty"`
}

type JMDictLsource struct {
	Lang  string `xml:"xml:lang,attr,omitempty" bson:"language,omitempty"`
	Value string `xml:",chardata" bson:"value,omitempty"`
}

type JMDictGloss struct {
	Lang   string `xml:"xml:lang,attr,omitempty" bson:"language,omitempty"`
	G_type string `xml:"g_type,attr,omitempty" bson:"type,omitempty"`   // gloss is of a particular type, e.g. "lit" (literal), "fig" (figurative), "expl" (explanation).
	G_gend string `xml:"g_gend,attr,omitempty" bson:"gender,omitempty"` //  gender of the gloss (typically a noun in the target language)
	Value  string `xml:",chardata" bson:"value,omitempty"`
}

// Kanji dicttionary

type KanjiDict struct {
	XMLName    xml.Name         `xml:"kanjidic2"`
	Characters []KanjiCharacter `xml:"character"`
}

type KanjiCharacter struct {
	XMLName *xml.Name `xml:"character"`
	Literal string    `xml:"literal"`
	// Codepoint      []KanjiCodePoint      `xml:"cp_value"`
	Radical        *KanjiRadical        `xml:"radical"`
	Misc           *KanjiMisc           `xml:"misc"`
	ReadingMeaning *KanjiReadingMeaning `xml:"reading_meaning,omitempty"`
}

// type KanjiCodePoint struct {
// 	Type  string `xml:"cp_type,attr,omitempty"`
// 	Value string `xml:",chardata"`
// }

type KanjiRadical struct {
	Values []KanjiRadicalValue `xml:"rad_value"`
}

type KanjiRadicalValue struct {
	Type  string `xml:"rad_type,attr,omitempty"`
	Value string `xml:",chardata"`
}

type KanjiMisc struct {
	Frequency   *int              `xml:"freq,omitempty"`
	StrokeCount *int              `xml:"stroke_count,omitempty"`
	Grade       *int              `xml:"grade,omitempty"`
	JLPT        *int              `xml:"jlpt,omitempty"`
	Variant     *KanjiMiscVariant `xml:"variant,omitempty"`
}

type KanjiFrequency struct {
	Value string `xml:",chardata"`
}

type KanjiMiscVariant struct {
	Type  string `xml:"var_type,attr,omitempty"`
	Value string `xml:",chardata"`
}

type KanjiReadingMeaning struct {
	Group  []KanjiRMGroup `xml:"rmgroup,omitempty"`
	Nanori []string       `xml:"nanori.omitempty"`
}

type KanjiRMGroup struct {
	Reading []KanjiReading `xml:"reading,omitempty"`
	Meaning []KanjiMeaning `xml:"meaning,omitempty"`
}

type KanjiReading struct {
	Value string `xml:",chardata"`
	Type  string `xml:"r_type,attr,omitempty"`
}

type KanjiMeaning struct {
	Value    string `xml:",chardata"`
	Language string `xml:"m_lang,attr,omitempty"`
}
