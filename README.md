# Japanese Story Reader and Vocab Trainer

A program for reading stories (short excerpts of Japanese) and drilling their vocabulary.

![](./images/story.png)


## Running

1. [Install Go](https://go.dev/doc/install), version 1.15 or later.
1. In the `app` directory, use `go build` to make the executable.
1. Run the executable.
1. In the browser, open `localhost:8080`

## Stories

The general idea is to repeat each story you read several times over the course of a week or two, drilling its vocabulary each time before you re-read it.

Each story has:

- a read count: the total number of times the story has been read
- a countdown: the number of additional times you plan to read the story
- a timestamp: the date and time that the story was last read

The main page only displays only the stories that have a countdown greater than 0. The story "catalog" page displays all the stories and allows you to add new stories.

Click a story's title to read it. The words are highlighted based on part of speech:

- white words: nouns
- yellow: particles
- dark yellow: connective particles (such as の and と)
- red: verbs
- dark red: verb auxilliaries and the copula
- green: adverbs
- violet: i-adjectives
- blue: pronouns and determines (such as これ and 何)

(Note that the auto-generated grammatical analysis is not always 100% accurate but is generally quite good.)

Click a word in the text to get information about it and its kanji.

## Drilling

![](./images/drill.png)

From a story's page, you can click the "drill" link to drill its words. Words have a rank 1 through 4, where higher ranks have longer cooldowns:

- Rank 1 cooldown: 5 hours
- Rank 2 cooldown: 4 days
- Rank 3 cooldown: 30 days
- Rank 4 cooldown: 1000 days

Hotkeys:

- **d** marks the current word correct (moving the card to the discard pile at the bottom) and sets its timestamp
- **a** marks the current word wrong (marking it red and moving it down to the second position in the list) and sets its timestamp
- **1**: sets the selected word's rank to level 1
- **2**: sets the selected word's rank to level 2
- **3**: sets the selected word's rank to level 3
- **4**: sets the selected word's rank to level 4

Once you mark all words in the list correct or wrong, the words you marked wrong will be reshuffled. Keep answering until all words are marked correct.

Marking a word correct or wrong puts it on cooldown. By default, the drill list only includes words off cooldown, but you can choose to include words on cooldown.

Words in the drill list can also be filtered by type: kanji characters, words spelt in katakana, ichidan verbs, or godan verbs.