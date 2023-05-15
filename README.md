# Japanese Vocab trainer

## Running

1. [Install Go](https://go.dev/doc/install), version 1.15 or later.
1. In the `app` directory, use `go build` to make the executable.
1. Run the executable.
1. In the browser, open `localhost:8080`

## Main page

The top has a form a form to add new stories, and below has a list of your existing stories. Clicking a story opens it in a new page.

## Story page

The right side shows the text of the story. Click a word to display its definitions on the left side. Ctrl-click a word to increase its drill countdown (up to max of 5).

## Drills page

The right side shows the drill list, and the left side shows the definitions of the current drill word (the word with a white border at the top of the list).

Hotkeys:

- 'd' mark the current word correct (moving the card to the discard pile at the bottom)
- 'a' mark the current word wrong (marking it red and moving it down to the second position in the list)
- '-' decrement the current word's countdown
- '=' increment the current word's countdown
- 'backspace' set the current word's countdown to 0

The controls at the top allow you to filter the list of drill words. From left-to-right:

- the number of words in the list
- select words recently added to the word bank
- select words recently marked wrong
- select words of certain types: katakana, ichidan verbs, godan verbs
- select words from specific stories
- select words on cooldown (words that have recently been marked correct)


// ichidan: 904
// verb-godan-su: 459
// verb-godan-ru: 701
// verb-godan-u: 156
// verb-godan-tsu: 26
// verb-godan-gu: 56
// verb-godan-ku: 233
// verb-godan-nu: 3
// verb-godan-mu: 178
// verb-godan-bu: 21
// verb-godan-aru: 4    // this explains kudasai! it's noun form of kudasaru