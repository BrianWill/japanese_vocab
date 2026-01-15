package main

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"regexp"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ikawaha/kagome/v2/tokenizer"
)

type ExtractModel struct {
	IsLoaded bool
	Text     string
	WordList list.Model
	Vocab    []*Vocab
	VocabDB  DB
	Keys     ExtractKeysMap
}

type ExtractedVocabMsg struct {
	Text string
	// Vocab []*Vocab
	Vocab []*Vocab
}

type ExtractedWordItem struct {
	Vocab *Vocab
}

func (i ExtractedWordItem) FilterValue() string { return i.Vocab.Word }
func (i ExtractedWordItem) Title() string       { return i.Vocab.Word }
func (i ExtractedWordItem) Description() string { return i.Vocab.Kana }

type ExtractKeysMap struct {
	Enabled  key.Binding
	Disabled key.Binding
	Invalid  key.Binding
	Archived key.Binding
}

func newExtractKeysMap() ExtractKeysMap {
	return ExtractKeysMap{
		Enabled: key.NewBinding(
			key.WithKeys("e"),
			key.WithHelp("e", "enable"),
		),
		Disabled: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "disable"),
		),
		Invalid: key.NewBinding(
			key.WithKeys("i"),
			key.WithHelp("i", "invalidate"),
		),
		Archived: key.NewBinding(
			key.WithKeys("a"),
			key.WithHelp("a", "archive"),
		),
	}
}

func (k ExtractKeysMap) ShortHelp() []key.Binding {
	return []key.Binding{
		k.Enabled,
		k.Disabled,
		k.Invalid,
		k.Archived,
	}
}

func (k ExtractKeysMap) FullHelp() []key.Binding {
	return []key.Binding{
		k.Enabled,
		k.Disabled,
		k.Invalid,
		k.Archived,
	}
}

func (m ExtractModel) Init() tea.Cmd {
	return nil
}

type ExtractedVocabUpdated struct {
	Vocab Vocab
	Index int
}

func (m ExtractModel) markWord(status VocabStatus) func() tea.Msg {
	selected := m.WordList.SelectedItem().(ExtractedWordItem)
	index := m.WordList.GlobalIndex()
	vocab := *selected.Vocab // to be proper, our command should read a copy, not the original
	vocab.Status = status
	return func() tea.Msg {
		err := m.VocabDB.Update(&vocab)
		if err != nil {
			return IOErrorMsg{err}
		}
		return ExtractedVocabUpdated{Vocab: vocab, Index: index} // dummy msg to trigger a refresh
	}
}

func (m ExtractModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.Keys.Enabled):
			return m, m.markWord(VOCAB_STATUS_ENABLED)
		case key.Matches(msg, m.Keys.Disabled):
			return m, m.markWord(VOCAB_STATUS_DISABLED)
		case key.Matches(msg, m.Keys.Invalid):
			return m, m.markWord(VOCAB_STATUS_INVALID)
		case key.Matches(msg, m.Keys.Archived):
			return m, m.markWord(VOCAB_STATUS_ARCHIVED)
		case key.Matches(msg, m.WordList.KeyMap.Quit):
			return m, func() tea.Msg {
				return ReturnToMainMsg{}
			}
		}
	case ExtractedVocabUpdated:
		item := m.WordList.Items()[msg.Index].(ExtractedWordItem)
		item.Vocab = &msg.Vocab
		m.WordList.SetItem(msg.Index, item)
	case ExtractedVocabMsg:
		m.IsLoaded = true
		m.Text = msg.Text
		m.Vocab = msg.Vocab
		items := make([]list.Item, 0)
		for _, v := range m.Vocab {
			items = append(items, ExtractedWordItem{Vocab: v})
		}
		m.WordList.SetItems(items)
	}

	var cmd tea.Cmd
	m.WordList, cmd = m.WordList.Update(msg)
	return m, cmd
}

var docStyle = lipgloss.NewStyle().Margin(1, 2)

func (m ExtractModel) View() string {
	if !m.IsLoaded {
		return "Extracting in progress..."
	}

	return docStyle.Render(m.WordList.View())
}

func loadVocab(db DB) []*Vocab {
	allVocab, err := db.ListAll(true)
	if err != nil {
		log.Fatalln(err)
	}

	vocab, err := PickRandomN(allVocab, 10)
	if err != nil {
		log.Fatal(err)
	}
	rand.Shuffle(len(vocab), func(i, j int) {
		vocab[i], vocab[j] = vocab[j], vocab[i]
	})

	return vocab
}

type ExtractWordItemDelegate struct{}

func (ExtractWordItemDelegate) Height() int  { return 1 } // number of lines per item
func (ExtractWordItemDelegate) Spacing() int { return 1 } // space between items
func (ExtractWordItemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	return nil // no special updates
}
func (ExtractWordItemDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	v := item.(ExtractedWordItem)

	selected := "  " // marker for selection
	if index == m.Index() {
		selected = "> "
	}

	status := "enabled"
	switch v.Vocab.Status {
	case VOCAB_STATUS_ARCHIVED:
		status = "archived"
	case VOCAB_STATUS_INVALID:
		status = "invalid"
	case VOCAB_STATUS_DISABLED:
		status = "disabled"
	}

	fmt.Fprintf(w, "%s%s%s%s", selected,
		columnWidthStyle.Render(v.Title()),
		columnWidthStyle.Render(v.Description()),
		columnWidthStyle.Render(status))
}

func extractVocab(db DB) (ExtractedVocabMsg, error) {
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

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	vocab, err := db.InsertOrGetVocabBatch(ctx, tokens)
	if err != nil {
		return ExtractedVocabMsg{}, err
	}

	filteredVocab := make([]*Vocab, 0, len(vocab))
	for _, v := range vocab {
		// filter out words that are VOCAB_STATUS_INVALID or VOCAB_STATUS_ARCHIVED
		if v.Status == VOCAB_STATUS_INVALID || v.Status == VOCAB_STATUS_ARCHIVED {
			continue
		}
		// filter out words that are VOCAB_STATUS_ENABLED if their DrillTodo is > 0
		if v.Status == VOCAB_STATUS_ENABLED && v.DrillTodo > 0 {
			continue
		}
		filteredVocab = append(filteredVocab, v)
	}

	return ExtractedVocabMsg{Text: text, Vocab: filteredVocab}, nil
}

func tokenize(content string) ([]*JpToken, error) {
	analyzerTokens := tok.Analyze(content, tokenizer.Normal)
	tokens := make([]*JpToken, 0, len(analyzerTokens))

	tokenMap := make(map[string]struct{})

	for _, t := range analyzerTokens {
		features := t.Features()

		if !ContainsJapanese(t.Surface) {
			continue
		}

		if _, ok := tokenMap[t.Surface]; ok {
			continue
		}
		tokenMap[t.Surface] = struct{}{}

		var token *JpToken
		if len(features) < 9 {
			token = &JpToken{
				Surface: t.Surface,
				POS:     features[0],
				POS_1:   features[1],
			}
		} else {
			token = &JpToken{
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

		if token.BaseForm == "" {
			token.BaseForm = token.Surface
		}
		tokens = append(tokens, token)
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
