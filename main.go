package main

import (
	"log"
	"math/rand"
	"strings"
	"time"

	// vlc "japanese_vocab_cmdline/vlc_control"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	_ "modernc.org/sqlite"
)

var startTime time.Time = time.Now()

var (
	redStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	greenStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	cyanStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
	greyStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("7"))
	yellowBoldStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("0")).Bold(true)
)

func main() {
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

	p := tea.NewProgram(
		MainModel{
			drillModel: DrillModel{
				DB:         db,
				CurrentIdx: 0,
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
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		}
	case ReturnToMainMenu:
		m.menuState = MAIN_MENU
		return m, nil
	case IOErrorMsg:
		// todo
	}

	// then process messages for if we're in the main menu
	switch m.menuState {
	case MAIN_MENU:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "q":
				return m, tea.Quit
			case "1":
				m.menuState = DRILL_MENU
				m.drillModel = DrillModel{
					DB:    m.drillModel.DB,
					Vocab: loadVocab(m.drillModel.DB),
				}
				return m, nil
			}
		}
	// lastly process messages for active submenu
	case DRILL_MENU:
		// gotcha:
		// return m.drillModel.Update(msg) // returns DrillModel instead of expected MainModel
		// (if we return DrillModel instead of MainModel, then Bubbletea calls next root Update on DrillModel, which we don't want)

		model, cmd := m.drillModel.Update(msg)
		m.drillModel = model.(DrillModel)
		return m, cmd
	}
	return m, nil
}

func (m MainModel) View() string {
	if m.menuState == DRILL_MENU {
		return m.drillModel.View()
	}

	return fmt.Sprintf(
		"Japanese Vocab \n\n  %s\n  %s\n",
		"1. Drill random words",
		"q. Quit",
	)
}

func (m DrillModel) Init() tea.Cmd {
	return nil
}

type ReturnToMainMenu struct{}

func (m DrillModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		v := m.Vocab[m.CurrentIdx]
		switch msg.String() {
		case "q":
			return m, func() tea.Msg {
				return ReturnToMainMenu{}
			}
		}
		if m.Done {
			break
		}
		isLast := m.CurrentIdx+1 == len(m.Vocab)
		switch msg.String() {
		case "c":
			v.DrillCount++
			m.NumCorrect++
			if v.DrillInfo.IsWrong {
				m.NumToRepeat++
				m.AnsweredMarkersStr += "☒ "
			} else {
				m.AnsweredMarkersStr += "■ "
			}
			v.DrillInfo.IsCorrect = true

			m.PriorVocab = v

			if isLast {
				m.Done = true
			} else {
				m.CurrentIdx++
			}

			return m, func() tea.Msg {
				err := m.DB.Update(v)
				if err != nil {
					return IOErrorMsg{err}
				}
				return nil
			}

		case "z":
			v.DrillInfo.IsWrong = true
			m.PriorVocab = v

			if !isLast {
				// swap with next but currentIndex stays the same
				m.Vocab[m.CurrentIdx], m.Vocab[m.CurrentIdx+1] = m.Vocab[m.CurrentIdx+1], m.Vocab[m.CurrentIdx]
			}
		}
	}

	return m, nil
}

func (m DrillModel) View() string {
	if len(m.Vocab) == 0 {
		return "No vocabulary loaded.\nPress q to quit."
	}

	v := m.Vocab[m.CurrentIdx]

	// countStr := strings.Repeat("■ ", m.NumCorrect-m.NumToRepeat) +
	// 	strings.Repeat("☒ ", m.NumToRepeat) + strings.Repeat("□ ", len(m.Vocab)-m.NumCorrect)

	countStr := m.AnsweredMarkersStr + strings.Repeat("□ ", len(m.Vocab)-m.NumCorrect)

	priorStr := ""
	if m.PriorVocab != nil {
		var word string
		if m.PriorVocab.DrillInfo.IsWrong && m.PriorVocab.DrillInfo.IsCorrect {
			word = cyanStyle.Render(m.PriorVocab.Word)
		} else if m.PriorVocab.DrillInfo.IsWrong {
			word = redStyle.Render(m.PriorVocab.Word)
		} else if m.PriorVocab.DrillInfo.IsCorrect {
			word = greenStyle.Render(m.PriorVocab.Word)
		}

		priorStr = fmt.Sprintf(
			"\n\n Prior: \n\n  \t%s\t\t  %s\t\t  %s\n\n\n ",
			word,
			m.PriorVocab.Kana,
			m.PriorVocab.Definition,
		)
	}

	if m.Done {
		return fmt.Sprintf(
			"\n\n  %s \n\n Press q = quit\n\n DRILL COMPLETE! \n\n \t%s\n\n ",
			countStr,
			priorStr,
		)
	} else {
		return fmt.Sprintf(
			"\n\n  %s \n\n Press c = correct, z = incorrect, q = quit\n\n \t%s\n\n  \n\n %s",
			countStr,
			yellowBoldStyle.Render(v.Word),
			priorStr,
		)
	}
}

func loadVocab(db VocabDB) []*Vocab {
	allVocab, err := db.ListAll(true)
	if err != nil {
		log.Fatalln(err)
	}

	vocab, err := PickRandomN(allVocab, 12)
	if err != nil {
		log.Fatal(err)
	}
	rand.Shuffle(len(vocab), func(i, j int) {
		vocab[i], vocab[j] = vocab[j], vocab[i]
	})

	return vocab
}
