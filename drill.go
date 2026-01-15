package main

import (
	"fmt"
	"math/rand"

	tea "github.com/charmbracelet/bubbletea"
)

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
