package main

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	// vlc "japanese_vocab_cmdline/vlc_control"
	"fmt"

	"github.com/ikawaha/kagome-dict/ipa"
	"github.com/ikawaha/kagome/v2/tokenizer"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	openai "github.com/sashabaranov/go-openai"
	_ "modernc.org/sqlite"
)

var startTime time.Time = time.Now()

var (
	boldStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("#000")).Bold(true)
	redStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Width(20)
	greenStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Width(20)
	cyanStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("6")).Width(20)
	greyStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("7")).Width(20)
	yellowBoldStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("0")).Bold(true)
)

var tok *tokenizer.Tokenizer

func main() {
	var err error
	tok, err = tokenizer.New(ipa.Dict(), tokenizer.OmitBosEos())
	if err != nil {
		panic(err)
	}

	vocabList, err := LoadVocabCSVFile("add_words.csv")
	if err != nil {
		log.Fatal(err)
	}

	// // Optionally save back
	// err = SaveVocabCSVFile("vocab_out.csv", vocabList)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	db, err := InitVocabDB()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// upsert all words into the db
	err = db.UpsertAll(vocabList)
	if err != nil {
		log.Fatal(err)
	}
	vocabTable := table.New(
		table.WithColumns([]table.Column{
			{Title: "Word", Width: 20},
			{Title: "Kana", Width: 20},
			{Title: "Definition", Width: 30},
		}),
		table.WithRows([]table.Row{
			{"", "", ""},
		}),
		table.WithFocused(true),
		table.WithHeight(0),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("240")).
		Bold(false)
	vocabTable.SetStyles(s)
	vocabTable.Blur()

	wordList := list.New(nil, list.NewDefaultDelegate(), 0, 0)
	wordList.Title = "Words in text:"

	p := tea.NewProgram(
		MainModel{
			drillModel: DrillModel{
				DB:         db,
				CurrentIdx: 0,
				VocabTable: vocabTable,
			},
			extractModel: ExtractModel{
				WordList: wordList,
			},
		},
		tea.WithAltScreen(),
	)
	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
	}
}

func (m MainModel) Init() tea.Cmd {
	return nil
}

func (m MainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// first process 'global' messages
	// (messages that are processed the same no matter which menu we're in)
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		h, v := docStyle.GetFrameSize()
		m.extractModel.WordList.SetSize(msg.Width-h, msg.Height-v)
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		}
	case ReturnToMainMsg:
		m.menuState = MAIN_SCREEN
		return m, nil
	case IOErrorMsg:
		// todo
	}

	// then process messages for the current menu
	switch m.menuState {
	case MAIN_SCREEN:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "q":
				return m, tea.Quit
			case "1":
				m.menuState = DRILL_SCREEN
				m.drillModel = DrillModel{
					DB:         m.drillModel.DB,
					Vocab:      loadVocab(m.drillModel.DB),
					VocabTable: m.drillModel.VocabTable,
				}
				return m, nil
			case "2":
				m.menuState = EXTRACT_SCREEN
				m.extractModel.IsLoaded = false
				return m, func() tea.Msg {
					vocab, err := extractVocab()
					if err != nil {
						log.Printf("failed to read file: %v", err)
						return IOErrorMsg{err: err}
					}
					return vocab
				}
			}
		}
	case DRILL_SCREEN:
		model, cmd := m.drillModel.Update(msg)
		m.drillModel = model.(DrillModel)
		return m, cmd
	case EXTRACT_SCREEN:
		model, cmd := m.extractModel.Update(msg)
		m.extractModel = model.(ExtractModel)
		return m, cmd
	}
	return m, nil
}

func (m MainModel) View() string {
	switch m.menuState {
	case DRILL_SCREEN:
		return m.drillModel.View()
	case EXTRACT_SCREEN:
		return m.extractModel.View()
	}

	return fmt.Sprintf(
		"Japanese Vocab \n\n  %s\n  %s\n  %s\n",
		"1. Drill random words",
		"2. Extract words from file",
		"q. Quit",
	)
}

func (m DrillModel) Init() tea.Cmd {
	return nil
}

func (m DrillModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.VocabTable, cmd = m.VocabTable.Update(msg)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			return m, func() tea.Msg {
				return ReturnToMainMsg{}
			}
		}
		if m.Done {
			break
		}

		v := m.Vocab[m.CurrentIdx]
		switch msg.String() {
		case "c":
			m.NumCorrect++
			if v.DrillInfo.IsWrong {
				m.NumToRepeat++
			} else {
				v.DrillCount++ // only update drill count when user didn't answer wrong this round
			}
			v.DrillInfo.IsCorrect = true
			v.DrillInfo.IsShown = true

			// scan to first word in list that isn't correct
			roundOver := true
			for i, v := range m.Vocab {
				if !v.DrillInfo.IsCorrect {
					m.CurrentIdx = i
					roundOver = false
					break
				}
			}

			if roundOver {
				if m.NumToRepeat > 0 {
					m.nextRound()
				} else {
					m.Done = true
				}
			}

			cmd = func() tea.Msg {
				err := m.DB.Update(v)
				if err != nil {
					return IOErrorMsg{err}
				}
				return nil
			}

		case "z":
			v.DrillInfo.IsWrong = true
			v.DrillInfo.IsShown = true

			// scan to first word in list that isn't correct
			for i, v := range m.Vocab {
				if !v.DrillInfo.IsCorrect && m.CurrentIdx != i { // skip current
					m.CurrentIdx = i
					break
				}
			}
		}
	}

	return m, cmd
}

func (m *DrillModel) nextRound() {
	vocab := make([]*Vocab, m.NumToRepeat)
	idx := 0
	for _, v := range m.Vocab {
		if v.DrillInfo.IsWrong {
			v.DrillInfo.IsCorrect = false
			v.DrillInfo.IsWrong = false
			v.DrillInfo.IsShown = false
			vocab[idx] = v
			idx++
		}
	}
	rand.Shuffle(len(vocab), func(i, j int) {
		vocab[i], vocab[j] = vocab[j], vocab[i]
	})
	m.Vocab = vocab
	m.CurrentIdx = 0
	m.NumCorrect = 0
	m.NumToRepeat = 0
}

func (m DrillModel) View() string {
	if len(m.Vocab) == 0 {
		return "No vocabulary loaded.\nPress q to quit."
	}

	shownWords := m.VocabTable.View()

	currentWord := ""
	if !m.Done {
		currentWord = m.Vocab[m.CurrentIdx].Word
	}

	currentWasShown := false
	wordList := ""
	for _, v := range m.Vocab {
		var marker = "\t  "
		if currentWord == v.Word {
			marker = "\t➤ "
		}
		if v.DrillInfo.IsWrong && v.DrillInfo.IsCorrect {
			wordList += marker + greyStyle.Render(v.Word) + greyStyle.Render(v.Kana) + greyStyle.Width(40).Render(v.Definition) + "\n"
		} else if v.DrillInfo.IsWrong {
			wordList += marker + redStyle.Render(v.Word) + redStyle.Render(v.Kana) + redStyle.Width(40).Render(v.Definition) + "\n"
		} else if v.DrillInfo.IsCorrect {
			wordList += marker + greenStyle.Render(v.Word) + greenStyle.Render(v.Kana) + greenStyle.Width(40).Render(v.Definition) + "\n"
		} else {
			if currentWasShown {
				wordList += "\t  ___\n"
			} else {
				wordList += "\t➤ " + yellowBoldStyle.Render(currentWord) + "\n"
			}
		}

		if currentWord == v.Word {
			currentWasShown = true
		}
	}

	if m.Done {
		return fmt.Sprintf(
			"\n\n  %s\n\n%s \n\t"+boldStyle.Render("DRILL COMPLETE!")+"\n\n\tPress q = quit",
			shownWords,
			wordList,
		)
	} else {
		return fmt.Sprintf(
			"\n\n  %s\n\n %s \n\n\n\tPress c = correct, z = incorrect, q = quit",
			shownWords,
			wordList,
		)
	}
}

func (m ExtractModel) Init() tea.Cmd {
	return nil
}

func (m ExtractModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			return m, func() tea.Msg {
				return ReturnToMainMsg{}
			}
		}
	case ExtractedVocabMsg:
		m.IsLoaded = true
		m.Text = msg.Text
		m.Words = msg.Words
		items := make([]list.Item, 0)
		for _, v := range m.Words {
			items = append(items, ExtractedWordItem{Base: v.BaseForm, Kana: v.Reading})
		}
		m.WordList.SetItems(items)
	}

	var cmd tea.Cmd
	m.WordList, cmd = m.WordList.Update(msg)
	return m, cmd
}

var docStyle = lipgloss.NewStyle().Margin(1, 2)

func (m ExtractModel) View() string {
	// if !m.IsLoaded {
	// 	return "Extracting in progress..."
	// }

	return docStyle.Render(m.WordList.View())

	// words := ""
	// for _, word := range m.Words[:10] {
	// 	words += word.BaseForm + "\n"
	// }

	// return fmt.Sprintf("done extracting:\n\n %v \n\n %v", len(m.Words), words)
}

func loadVocab(db VocabDB) []*Vocab {
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

func testChatRequest() {
	client := openai.NewClient(os.Getenv("OPENAI_API_KEY"))

	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT4oMini,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: "You are a helpful assistant.",
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: "Explain RAII in one paragraph.",
				},
			},
		},
	)
	if err != nil {
		panic(err)
	}

	fmt.Println(resp.Choices[0].Message.Content)
}

func testChatHTTPRequest() {
	reqBody := ChatRequest{
		Model: "gpt-4o-mini",
		Messages: []struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		}{
			{"system", "You are a helpful assistant."},
			{"user", "Explain RAII in one paragraph."},
		},
	}

	b, _ := json.Marshal(reqBody)

	req, _ := http.NewRequestWithContext(
		context.Background(),
		"POST",
		"https://api.openai.com/v1/chat/completions",
		bytes.NewReader(b),
	)
	req.Header.Set("Authorization", "Bearer "+os.Getenv("OPENAI_API_KEY"))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	fmt.Println(resp.Status)
}

type ChatRequest struct {
	Model    string `json:"model"`
	Messages []struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	} `json:"messages"`
}
