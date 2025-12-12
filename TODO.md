
drill menu:
    option to toggle display of kanji defs and pronunciations

tcell (ncurses like ui)

flashcard drill
- pick random words from unarchived vocab

main menu:
    - drill that picks with bias towards recently added words
    - add words
        from csv file
    - mark words to archive

After each drill, show list of words suggested to archive based on their lifetime drill count


 
video playback:
    control vlc or another video player from the app
        open / play video, seek to timestamp
            VLC RC interface or HTTP interace
            launch VLC with option: `--extraintf rc` or `--extraintf http`
        try mpv player

display timer indicating length of the drill session
    clock that counts up since start of the program
    

add/arcive csv files from directories instead of from a single file
- (easier to maintain separate vocab list files)

injest japanese words from text file
- interactive mode that lets me pick the words to add
- generate vocab list csv file from the text

track word part of speech, esp for verbs (godan ichidan)

timestamp for when a word was added and when it was last drilled
- also when it was archived?