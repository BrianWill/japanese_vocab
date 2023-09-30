# Japanese Input trainer

A program for [learning Japanese through input](input.md) by reading stories and drilling their vocabulary.

![](./images/story.png)


## Running the program

1. [Install Go](https://go.dev/doc/install), version 1.15 or later.
1. At the command line, in the `app` directory, use `go build` to make the executable.
1. Run the executable.
1. In a web browser, go to `localhost:8080`

## Stories

Each story has:

- a title
- a read count: the total number of times the story has been read
- a countdown: the number of additional times you plan to read the story
- a timestamp: the date and time that the story was last read
- a link (optional): URL from which the story originates
- an audio file (optional): an mp3, mp4, or m4a file path relative to `static/audio/`

## Main page

The main page displays all the stories and allows you to add new stories. A checkbox toggles whether to show only stories with a countdown greater than 0.

## Story page

A story's page display's its title and text. 

A link under the title takes you to a page for drilling the words and kanji in the story.

Another link under the title marks the story as read (which sets its timestamp, increments its read count, and decrements its countdown).

### Youtube and audio player

If the story link is a youtube video link, the video will show in an embedded player in the top corner.
Otherwise, if the story audio path is not empty, the story page will instead have an audio player in the top corner.

The youtube and audio players can be controlled via hotkeys:

- `s``: toggle play/pause
- `d``: jump ahead ~2 seconds
- `a``: jump back ~2 seconds
- `e``: jump ahead ~5 seconds
- `q``: jump back ~5 seconds

The words of the story text are highlighted based on part of speech:

- white words: nouns
- yellow: particles
- dark yellow: connective particles (such as の and と)
- red: verbs
- dark red: verb auxilliaries and the copula
- green: adverbs
- violet: i-adjectives
- blue: pronouns and determiners (such as これ and 何)

(Note that the auto-generated grammatical analysis is not always 100% accurate but is generally quite good.)

Clicking a word in the text selects it and displays its defintions and any kanji it contains. Some hotkeys affect the selected word:

- `1`, `2`, `3`, and `4`: set the rank of the word to 1, 2, 3, or 4 
- `space`: sets the word's drill timestamp (see below)

### Story line timestamps

Each line of the story has a timestamp:

- clicking a timestamp jumps the player (if present) to that timestamp
- alt-clicking a line's timestamp will set it to the player's current time.
- middle-clicking a line's timestamp toggles whether the line is marked, changing the timestamp's color, which is useful for tracking interseting lines in a story.

Clicking a timestamp also selects the line, and you can adjust the timestamp of the selected line:

- `-` (minus) decrements the line's timestamp by half a second
- `+` (plus) increments the line's timestamp by half a second

Lines can be split and joined:

- ctrl-clicking a line's timestamp merges the line to the end of the preceding line
- ctrl-clicking a word in a line splits that word and the rest of the line into a new line; the new line's timestamp is set to the player's current time.

## Drilling page

![](./images/drill.png)

Each word has a rank of 1, 2, 3, or 4 and a timestamp denoting when the word was last drilled. Words are considered "on cooldown" if they have been drilled within the cooldown window, which depends upon rank:

- Rank 1 cooldown: 5 hours
- Rank 2 cooldown: 4 days
- Rank 3 cooldown: 30 days
- Rank 4 cooldown: 1000 days

By default, only "on cooldown" are not included in the drill list, but this can be controlled 

Some hotkeys affect the word at the top of the list:

- `d`: sets the word's drill timestamp and marks it correct (which moves the card to the discard pile at the bottom)
- `a`: sets the word's drill timestamp and marks it incorreect (which marks it red and moves it down to the second position in the list)
- `1`, `2`, `3`, and `4`: set the rank of the word to 1, 2, 3, or 4 

Once you mark all words in the list correct or incorrect, the words you marked incorrect will be reshuffled for another round of drilling.

Words in the drill list can also be filtered by type: kanji characters, words spelt in katakana, ichidan verbs, or godan verbs.