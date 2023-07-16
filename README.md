# Japanese Vocab trainer

## Running

1. [Install Go](https://go.dev/doc/install), version 1.15 or later.
1. In the `app` directory, use `go build` to make the executable.
1. Run the executable.
1. In the browser, open `localhost:8080`

## Stories

You can add stories *via* the form at the top of the main page. For example, paste the title of a youtube video, its link, and its Japanese transcript into the form, then click the "Create Story" button.

You can set the status of each story: 

- "Current": for stories you want to focus on
- "Read": for stories you want to put aside but revisit later
- "Never Read": the initial state of new stories
- "Archive": for stories you are unlikely to revisit

Click a story's title to see its text. Through grammatical analysis, the words of the stories are highlighted based on part of speech:

- white words: nouns
- yellow: particles
- dark yellow: connective particles (such as の and と)
- red: verbs
- dark red: verb auxilliaries and the copula
- green: adverbs
- violet: i-adjectives
- blue: pronouns and determines (such as これ and 何)

Note that the grammatical analysis is not always 100% accurate but is generally quite good.

Clicking a word gives its definitions and information about its kanji.

## Drilling

You can drill the words of an individual story by clicking its "words" link. Two links at the top of the story list let you drill the words from all stories or the words from all current stories.

The right side shows the word list, with the current word at the top with a white border. The left side shows the definitions of the current word.

Word rank is displayed on the right. New words start off at Rank 4. Words can be filtered by rank, and Rank 1 words are filtered out by the default filter.

Words with lower ranks have longer cooldowns:

- Rank 4 cooldown: 3 hours
- Rank 3 cooldown: 2 days
- Rank 2 cooldown: 7 days
- Rank 1 cooldown: 30 days

Hotkeys:

- **d** marks the current word correct (moving the card to the discard pile at the bottom)
- **a** marks the current word wrong (marking it red and moving it down to the second position in the list)
- **1**, **2**, **3**, **4** set the current word's rank
- **alt-r** refreshes the drill

When you exhaust the list, the words you marked wrong will be reshuffled. Keep answering until you see the message "Round Complete", then press **alt-r** to drill more words from the story.

When you mark a word correct or wrong, it is put on cooldown for 3 hours. Words can be filtered by their cooldown status.

Words can also be filtered by type: kanji characters, words spelt in katakana, ichidan verbs, or godan verbs.