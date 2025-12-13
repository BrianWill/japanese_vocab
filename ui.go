package main

import (
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/gdamore/tcell/v3"
)

var magentaStyle tcell.Style = tcell.StyleDefault.Foreground(tcell.ColorDarkMagenta).Background(tcell.ColorBlack)
var whiteStyle = tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorBlack)
var cyanStyle = tcell.StyleDefault.Foreground(tcell.ColorDarkCyan).Background(tcell.ColorBlack)
var greenStyle = tcell.StyleDefault.Foreground(tcell.ColorGreen).Background(tcell.ColorBlack)
var yellowStyle = tcell.StyleDefault.Foreground(tcell.ColorYellow).Background(tcell.ColorBlack)
var greyStyle = tcell.StyleDefault.Foreground(tcell.ColorDarkGrey).Background(tcell.ColorBlack)
var redStyle = tcell.StyleDefault.Foreground(tcell.ColorRed).Background(tcell.ColorBlack)

func drawTextWrap(s tcell.Screen, x1, y1, x2, y2 int, style tcell.Style, text string) {
	row := y1
	col := x1
	var width int
	for text != "" {
		text, width = s.Put(col, row, text, style)
		col += width
		if col >= x2 {
			row++
			col = x1
		}
		if row > y2 {
			break
		}
		if width == 0 {
			// incomplete grapheme at end of string
			break
		}
	}
}

func drawBox(s tcell.Screen, x1, y1, x2, y2 int, style tcell.Style, text string) {
	if y2 < y1 {
		y1, y2 = y2, y1
	}
	if x2 < x1 {
		x1, x2 = x2, x1
	}

	// Fill background
	for row := y1; row <= y2; row++ {
		for col := x1; col <= x2; col++ {
			s.Put(col, row, " ", style)
		}
	}

	// Draw borders
	for col := x1; col <= x2; col++ {
		s.Put(col, y1, string(tcell.RuneHLine), style)
		s.Put(col, y2, string(tcell.RuneHLine), style)
	}
	for row := y1 + 1; row < y2; row++ {
		s.Put(x1, row, string(tcell.RuneVLine), style)
		s.Put(x2, row, string(tcell.RuneVLine), style)
	}

	// Only draw corners if necessary
	if y1 != y2 && x1 != x2 {
		s.Put(x1, y1, string(tcell.RuneULCorner), style)
		s.Put(x2, y1, string(tcell.RuneURCorner), style)
		s.Put(x1, y2, string(tcell.RuneLLCorner), style)
		s.Put(x2, y2, string(tcell.RuneLRCorner), style)
	}

	drawTextWrap(s, x1+1, y1+1, x2-1, y2-1, style, text)
}

func printTextAtCell(s tcell.Screen, x, y int, text string, style tcell.Style) {
	for _, rune := range text {
		s.SetContent(x, y, rune, nil, style)
		x++
	}
}

func printMainMenu(screen tcell.Screen) {
	screen.Clear()

	printTextAtCell(screen, 2, 1, "JAPANESE VOCAB TRAINER", magentaStyle)
	printTextAtCell(screen, 2, 3, "d.  Drill random words", whiteStyle)
	printTextAtCell(screen, 2, 4, "l.  List all words", whiteStyle)
	printTextAtCell(screen, 2, 6, "(ESC to exit)", cyanStyle)

	screen.Show()
}

func printDrillScreen(screen tcell.Screen, vocab []*Vocab, currentIdx int) {
	screen.Clear()

	elapsedTime := time.Since(startTime)
	minutes := int(elapsedTime / time.Minute)
	seconds := int((elapsedTime % time.Minute) / time.Second)

	const (
		firstCol  = 2
		secondCol = 6
		thirdCol  = 22
		fourthCol = 40
	)

	header := fmt.Sprintf("DRILL WORDS (elpased time %d min %02d sec):", minutes, seconds)
	printTextAtCell(screen, firstCol, 1, header, magentaStyle)

	currentVocab := vocab[currentIdx]

	line := 2

	// print unanswered and wrong
	for _, v := range vocab {
		line++

		prefix := ""
		color := greyStyle
		def := v.Definition

		if v.DrillInfo.Correct {
			prefix = "◯"
			if v.DrillInfo.Wrong {
				// correct but not on first try
				color = cyanStyle
			} else {
				// correct on first try
				color = greenStyle
			}
		} else {
			if v.DrillInfo.Wrong {
				prefix = "🗙"
				if v == currentVocab {
					color = yellowStyle
				} else {
					color = redStyle
				}
			} else {
				// not yet answered
				def = ""
				if v == currentVocab {
					prefix = "-"
					color = yellowStyle
				}
			}
		}

		printTextAtCell(screen, firstCol, line, prefix, color)
		printTextAtCell(screen, secondCol, line, v.Word, color)
		printTextAtCell(screen, thirdCol, line, v.Kana, color)
		printTextAtCell(screen, fourthCol, line, def, color)
	}

	line++
	printTextAtCell(screen, firstCol, line, "a.  Mark yellow word as wrong", whiteStyle)
	line++
	printTextAtCell(screen, firstCol, line, "d.  Mark yellow word as correct", whiteStyle)
	line++
	printTextAtCell(screen, firstCol, line, "q.  Back to main menu", whiteStyle)

	screen.Show()
}

func tuiLoop(db VocabDB) {
	defStyle := tcell.StyleDefault.Background(tcell.ColorReset).Foreground(tcell.ColorReset)

	screen, err := tcell.NewScreen()
	if err != nil {
		log.Fatalf("%+v", err)
	}
	if err := screen.Init(); err != nil {
		log.Fatalf("%+v", err)
	}
	screen.SetStyle(defStyle)
	screen.EnablePaste()

	quit := func() {
		maybePanic := recover()
		screen.Fini()
		if maybePanic != nil {
			panic(maybePanic)
		}
	}
	defer quit()

	appState := AppState{CurrentScreen: SCREEN_MAIN}

	for {
		switch appState.CurrentScreen {
		case SCREEN_MAIN:

			printMainMenu(screen)

			ev := <-screen.EventQ()

			switch ev := ev.(type) {
			case *tcell.EventResize:
				screen.Sync()
			case *tcell.EventKey:

				key := strings.ToLower(ev.Str())

				if ev.Key() == tcell.KeyEscape || ev.Key() == tcell.KeyCtrlC {
					return
				} else if ev.Key() == tcell.KeyCtrlL {
					screen.Sync()
				} else if key == "d" { // start drill
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

					appState.CurrentScreen = SCREEN_DRILL
					appState.Drill = DrillState{Vocab: vocab}
				}
			}
		case SCREEN_DRILL:

			printDrillScreen(screen, appState.Drill.Vocab, appState.Drill.CurrentIdx)

			ev := <-screen.EventQ()

			switch ev := ev.(type) {
			case *tcell.EventResize:
				screen.Sync()
			case *tcell.EventKey:
				key := strings.ToLower(ev.Str())

				drill := &appState.Drill

				if ev.Key() == tcell.KeyEscape || ev.Key() == tcell.KeyCtrlC {
					return
				} else if ev.Key() == tcell.KeyCtrlL {
					screen.Sync()
				} else if key == "q" {
					appState.CurrentScreen = SCREEN_MAIN
				} else if key == "a" {
					processAnswer(drill, false, db)
				} else if key == "d" {
					processAnswer(drill, true, db)
				}
				if drill.Done {
					appState.CurrentScreen = SCREEN_MAIN
				}
			}
		}
	}
}

func processAnswer(drill *DrillState, markCorrect bool, db VocabDB) {
	currentVocab := drill.Vocab[drill.CurrentIdx]

	if markCorrect {
		currentVocab.DrillInfo.Correct = true
		drill.NumCorrect++
	} else {
		currentVocab.DrillInfo.Wrong = true
	}

	// if next round or drill is done
	if drill.NumCorrect == len(drill.Vocab) {
		newVocab := make([]*Vocab, 0)

		for _, v := range drill.Vocab {
			if v.DrillInfo.Wrong {
				v.DrillInfo.Wrong = false
				v.DrillInfo.Correct = false
				newVocab = append(newVocab, v)
			} else {
				// update drill counts in db
				v.DrillCount++
				db.Update(v)
			}
		}

		if len(newVocab) == 0 { // drill is done
			drill.Done = true
			return
		}

		// another round
		rand.Shuffle(len(newVocab), func(i, j int) {
			newVocab[i], newVocab[j] = newVocab[j], newVocab[i]
		})
		drill.Vocab = newVocab
		drill.CurrentIdx = 0
		drill.NumCorrect = 0
		return
	}

	// round continues, so we set CurrentIdx...

	// pick the first item that is not correct and isn't at currentIdx
	// (for case of one remaining, currentIdx remains unchanged)
	for i, v := range drill.Vocab {
		if i != drill.CurrentIdx && !v.DrillInfo.Correct {
			drill.CurrentIdx = i
			break
		}
	}
}
