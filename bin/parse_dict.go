package main

import (
	"encoding/xml"
	"fmt"
	"regexp"

	//"io/ioutil"
	"os"
	"strconv"
	"strings"
)

func main() {
	//makeCollection()
	processCollection()
}

func processCollection() {
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

	re, err := regexp.Compile(`^\w+$`)
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
				} else {
					fmt.Println(word.Text)
				}
			}
		default:
		}
	}

	fmt.Println("done decoding. Number of words:", len(words))
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
