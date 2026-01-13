package main

import (
	"database/sql"
	"fmt"
	"os"
	"regexp"
	"time"

	"github.com/ikawaha/kagome/v2/tokenizer"
)

func extractVocab() (ExtractedVocabMsg, error) {
	// read file
	data, err := os.ReadFile("words.txt")
	if err != nil {
		return ExtractedVocabMsg{}, err
	}
	text := string(data)

	tokens, err := tokenize(text)
	if err != nil {
		return ExtractedVocabMsg{}, err
	}

	return ExtractedVocabMsg{Text: text, Words: tokens}, nil
}

func tokenize(content string) ([]*JpToken, error) {
	analyzerTokens := tok.Analyze(content, tokenizer.Normal)
	tokens := make([]*JpToken, len(analyzerTokens))

	for i, t := range analyzerTokens {
		features := t.Features()
		if len(features) < 9 {

			tokens[i] = &JpToken{
				Surface: t.Surface,
				POS:     features[0],
				POS_1:   features[1],
			}
		} else {
			tokens[i] = &JpToken{
				Surface:          t.Surface,
				POS:              features[0],
				POS_1:            features[1],
				POS_2:            features[2],
				POS_3:            features[3],
				InflectionalType: features[4],
				InflectionalForm: features[5],
				BaseForm:         features[6],
				Reading:          features[7],
				Pronunciation:    features[8],
			}
		}
		if tokens[i].BaseForm == "" {
			tokens[i].BaseForm = tokens[i].Surface
		}
	}

	return tokens, nil
}

func getTokenPOS(token *JpToken, priorToken *JpToken) string {
	if token.Surface == "。" {
		return ""
	} else if token.Surface == "\n\n" {
		return ""
	} else if token.Surface == "\n" {
		return ""
	} else if token.Surface == " " {
		return ""
	}

	if (token.POS == "動詞" && token.POS_1 == "接尾") ||
		(token.POS == "助動詞") ||
		(token.Surface == "で" && token.POS == "助詞" && token.POS_1 == "接続助詞") ||
		(token.Surface == "て" && token.POS == "助詞" && token.POS_1 == "接続助詞") ||
		(token.Surface == "じゃ" && token.POS == "助詞" && token.POS_1 == "副助詞") ||
		(token.Surface == "し" && token.POS == "動詞" && token.POS_1 == "自立") { // auxilliary verb
		return "verb_auxiliary"
	} else if token.POS == "動詞" && token.POS_1 == "非自立" { // auxilliary verb
		return "verb_auxiliary"
	} else if (token.POS == "助詞" && token.POS_1 == "格助詞") || // case particle
		(token.POS == "助詞" && token.POS_1 == "接続助詞") || // conjunction particle
		(token.POS == "助詞" && token.POS_1 == "係助詞") || // binding particle (も　は)
		(token.POS == "助詞" && token.POS_1 == "副助詞") { // auxiliary particle
		return "particle"
	} else if token.POS == "副詞" {
		return "adverb"
	} else if token.POS == "接続詞" && token.POS_1 == "*" { // conjunction
		return "conjunction"
	} else if (token.POS == "助詞" && token.POS_1 == "連体化") || // connecting particle　(の)
		(token.POS == "助詞" && token.POS_1 == "並立助詞") { // connecting particle (や)
		return "connecting_particle"
	} else if token.POS == "形容詞" { // i-adj
		return "i_adjective pad_left"
	} else if token.POS == "名詞" && token.POS_1 == "代名詞" { // pronoun
		return "pronoun pad_left"
	} else if token.POS == "連体詞" { // adnominal adjective
		return "admoninal_adjective pad_left"
	} else if token.POS == "動詞" { //　verb
		return "verb pad_left"
	} else if token.POS == "名詞" && token.POS_1 == "接尾" { // noun suffix
		return "noun"
	} else if (priorToken.POS == "助詞" && (priorToken.POS_1 == "連体化" || priorToken.POS_1 == "並立助詞")) || // preceded by connective particle
		(priorToken.POS == "接頭詞" && priorToken.POS_1 == "名詞接続") { // preceded by prefix
		return "noun"
	} else if token.POS == "名詞" { // noun
		return "noun"
	} else if token.POS == "記号" { // symbol
		return ""
	} else if token.POS == "号" { // counter
		return "counter"
	} else {
		return ""
	}
}

func addWords(tokens []*JpToken, kanjiSet []string, sqldb *sql.DB) (wordIds []int64, newWordCount int, err error) {
	var reHasKanji = regexp.MustCompile(`[\x{4E00}-\x{9FAF}]`)
	var reHasKatakana = regexp.MustCompile(`[ア-ン]`)
	var reHasKana = regexp.MustCompile(`[ア-ンァ-ヴぁ-ゔ]`)

	newWordCount = 0
	unixtime := time.Now().Unix()

	keptTokens := make([]*JpToken, 0)

	// filter out tokens that have no part of speech
	for i, token := range tokens {
		priorToken := &JpToken{}
		if i > 0 {
			priorToken = tokens[i-1]
		}

		pos := getTokenPOS(token, priorToken)
		if pos == "" {
			continue
		}

		keptTokens = append(keptTokens, token)
	}

	words := make(map[string]bool)
	for _, t := range keptTokens {
		words[t.BaseForm] = true
	}

	for _, k := range kanjiSet {
		words[k] = true
	}

	wordIdsMap := make(map[int64]bool)

	for baseForm := range words {
		hasKatakana := len(reHasKatakana.FindStringIndex(baseForm)) > 0
		hasKana := len(reHasKana.FindStringIndex(baseForm)) > 0
		hasKanji := len(reHasKanji.FindStringIndex(baseForm)) > 0
		isKanji := len([]rune(baseForm)) == 1 && reHasKanji.FindStringIndex(baseForm) != nil

		if !hasKana && !hasKanji { // not a valid word
			continue
		}

		category := 0

		// has katakana
		if hasKatakana {
			category |= DRILL_CATEGORY_KATAKANA
		}

		if isKanji {
			category |= DRILL_CATEGORY_KANJI
		}

		var id int64
		err = sqldb.QueryRow(`SELECT id FROM words WHERE base_form = $1`, baseForm).Scan(&id)
		// if error other than no rows
		if err != nil && err != sql.ErrNoRows {
			return nil, 0, err
		}
		// if no error and
		if err == nil {
			wordIdsMap[id] = true
			continue
		}

		insertResult, err := sqldb.Exec(`INSERT INTO words (base_form, 
			date_added, category, repetitions, date_last_rep, archived) 
			VALUES($1, $2, $3, $4, $5, $6, $7, $8);`,
			baseForm, unixtime, category, 0, 1, 0)
		if err != nil {
			return nil, 0, fmt.Errorf("failure to insert word: " + err.Error())
		}

		id, err = insertResult.LastInsertId()
		if err != nil {
			return nil, 0, fmt.Errorf("failure to get id of inserted word: " + err.Error())
		}

		fmt.Println("inserted word: ", baseForm, id)

		newWordCount++
		wordIdsMap[id] = true
	}

	wordIds = make([]int64, 0, len(wordIdsMap))
	for id := range wordIdsMap {
		wordIds = append(wordIds, id)
	}

	return wordIds, newWordCount, nil
}
