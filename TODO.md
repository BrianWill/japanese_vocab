
parse text from file to find new words
    user picks words
    

when enabling a word, if it doesn't already have a definition or kana, use chat API to get definition and kana

create table for kanji
for every kanji in every word, add the kanji to the kanji table
    (use chat to get definitions of the kanji)


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

    

add/arcive csv files from directories instead of from a single file
- (easier to maintain separate vocab list files)



track word part of speech, esp for verbs (godan ichidan)

timestamp for when a word was added and when it was last drilled
- also when it was archived?