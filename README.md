# Japanese Input trainer

A program for [learning Japanese through input](input.md) by listening to stories and drilling their vocabulary.

![](./images/story.png)


## Running the program

For Windows, you can just clone the repo and run `app/japanese.exe`, then go to `localhost:8080` in a web browser.

For other platforms, you'll need to build the executable first:

1. [Install Go](https://go.dev/doc/install), version 1.15 or later.
1. At the command line, switch to the `app` directory.
1. If on Linux, you may need to run `sudo apt install build-essential` (or equivalent for your distribution)
1. Run `go get` to fetch the package dependencies.
1. Run `go build` to build the executable.

## Importing stories

Stories are imported in sets called "sources", which are represented as directories.

In a source directory, each mp3 and mp4 file represents an individual stories. For a story named `foo.mp3` or `foo.mp4`, the English and Japanese VTT subtitle files in the same directory should be named, respectively, `foo.en.vtt` and `foo.ja.vtt`.

The import page (link next to "Catalog" on the main page) lists all detected source directories. Clicking the "import" link of a source will:

1. import all stories of the source that have not yet been imported
2. re-import the subtitles of the already imported stories

## Main page

The main page first displays the "stories with queued reps", then the "recently logged stories", and last the full catalog of stories.

In the catalog, select a source from the drop down.

## Story page

A story's page displays the title, text content, and an audio or video player.

A story can have one or more "excerpts", which represent subranges of the audio or video.

Each excerpt of a story has a separate queue of reps.

The "vocab" link takes you to a drilling page for the words in the excerpt (extracted from the subtitles in the subrange).

The "log" link records a rep of the excerpt and removes a rep from the excerpt's queue.

In an excerpt's queue:

- Clicking a rep toggles its type (between listening and drilling).
- Alt-clicking a rep inserts another rep of the same kind after it in the queue.
- Ctrl-clicking a rep removes it from the queue.

### Video and audio player hotkey controls

- `f` : toggle fullscreen
- `s` : toggle play/pause
- `d` : jump ahead ~1 seconds
- `a` : jump back ~1 seconds
- `e` : jump ahead ~5 seconds
- `q` : jump back ~5 seconds
- `-` : decrease playback speed
- `+` : increase playback speed

Because subtitle data from sources may not line up with the video or audio, these keys may be helpful:

- `alt` + `-` : shift all subtitle timings forward by 0.2 seconds
- `alt` + `+` : shift all subtitle timings back by 0.2 seconds
- `alt` + `[` : shift the timings of the first subtitle after the current timemark up to the current timemark, and shift all following subtitles forward the same amount (only does something if the current timemark is between subtitles)
- `alt` + `]` : shift the timings of the first subtitle after the current timemark back by 10 seconds, and shift all following subtitles back the same amount (only does something if the current timemark is between subtitles)

## Drill page

![](./images/drill.png)

The words of the drill are displayed in a random order, with the current word at the top.

Hotkeys that affect the current word:

- `d`: mark the word correct
- `a`: mark the word incorrect
- `1` : toggle whether the word is archived

Once you mark all words in the list correct or incorrect, the words you marked incorrect will be reshuffled for another round of drilling.

Words in the drill list can be filtered by type: kanji characters, words spelt in katakana, ichidan verbs, or godan verbs.

The "log this drill" link will record the date of the drill and increment the repetition count for the drilled words.