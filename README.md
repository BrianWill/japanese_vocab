# Japanese Input trainer

A program for [learning Japanese through input](input.md) by listening to stories and drilling their vocabulary.

![](./images/story.png)

[Intro video:](https://www.youtube.com/watch?v=cQ0Z6ZM1KyI)

[![INTRO VIDEO](https://img.youtube.com/vi/cQ0Z6ZM1KyI/0.jpg)](https://www.youtube.com/watch?v=cQ0Z6ZM1KyI)

## Running the program

For Windows, you can just clone the repo and run `app/japanese.exe`, then go to `localhost:8080` in a web browser.

For other platforms, you'll need to build the executable first:

1. [Install Go](https://go.dev/doc/install), version 1.15 or later.
1. At the command line, switch to the `app` directory.
1. If on Linux, you may need to run `sudo apt install build-essential` (or equivalent for your distribution)
1. Run `go get` to fetch the package dependencies.
1. Run `go build` to build the executable.

## Main page

The main page first displays the "stories with queued reps", then the "recent stories" (stories that have a rep logged within the last 2 weeks but no remaining queued reps), and last the full catalog of stories.

In the catalog, select a source of stories from the drop down.

## Importing stories

Stories are imported in sets called "sources", which are represented as directories directly under the "static/sources" directory. For example, the directory "static/sources/example" represents a source named "example".

In a source directory, each mp3 and mp4 file represents an individual stories. For a story named `thing.mp3` or `thing.mp4`, the English and Japanese VTT subtitle files in the same directory should be named, respectively, `thing.en.vtt` and `thing.ja.vtt`.

The import page (link next to "Catalog" on the main page) lists all detected source directories. Clicking the `import` button of a source imports or re-imports all the stories in that source. (Re-importing a story update its subtitles if its subtitle files have changed.)

## Story page

The story page displays the story's title, text content, and an audio or video player.

A story has one or more "excerpts", which represent subranges of the audio or video:

- Each excerpt of a story has a separate queue of reps. Use the + and - buttons to add and remove reps from the queue.
- The `log` button records a rep of the excerpt (and removes a rep from the excerpt's queue).
- The `vocab` button takes you to a page for drilling the words in the excerpt. (The words are extracted from the subtitles in the subrange.)

### Video and audio player hotkey controls

- `f` : toggle fullscreen
- `s` : toggle play/pause
- `d` : jump ahead less than a second
- `a` : jump back less than a second
- `e` : jump ahead a few seconds
- `q` : jump back a few seconds
- `n` : jump ahead to the next subtitle (if subtitles are available and active)
- `b` : jump back to the previous subtitle (if subtitles are available and active)
- `-` : decrease playback speed
- `+` : increase playback speed

To align the subtitles with the audio:

- `alt` + `-` : shift all subtitle timings forward by 0.2 seconds
- `alt` + `+` : shift all subtitle timings back by 0.2 seconds
- `alt` + `[` : shift all subtitles after the current timemark forward such that the next subtitle begins at the current timemark
- `alt` + `]` : shift all subtitles after the current timemark back by 10 seconds

## Drill page

![](./images/drill.png)

The words of the drill are displayed in a random order, with the current word at the top.

Hotkeys that affect the current word:

- `d`: mark the word correct
- `a`: mark the word incorrect
- `1` : toggle whether the word is archived

Once you mark all words in the list correct or incorrect, the words you marked incorrect will be reshuffled for another round of drilling.

Words in the drill list can be filtered by type: kanji characters, words spelt in katakana, ichidan verbs, or godan verbs.

The `log the answered words` button will increment the rep count of each answered word and mark them with the current date and time.
