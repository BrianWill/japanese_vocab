# japanese_vocab TODO   
    
    drill words:
        choose only from the recent stories? 
            mostly from?
            or favor selection from words that appear most often across the recent stories?
    
    in story list, show count of kanji in the story? 
        number of unique kanji total throughout the text
            compute this at import time?
            new column in stories table
        also report how many archived vs unarchived?

    test mac and linux builds

    allow for splitting subtitle on punctuation marks
        maybe allow for everything but spaces?

    drill page:
        option to sort words in kana order
        option to show full definition of each word with each word (instead of just showing def for the current word)                 

    mark words for drills?

    audio/podcast mode:
        use text-to-speech to generate mp3's with word definitions inserted (text-to-speech):
        alternatively, a listening app that shows you definitions for highlighted words in current sentence            

    word dictionary should have rank for how common each word is
        e.g. foo is the nth most common word in the language
        could then do word highlighting and drilling that filters words based on how common they are

    seems to be an import issue for stories that have periods in the name

    openTranscript
        get rid of this once subtitle editing is done?
        or maybe keep so users can easily edit the original file?

    make player controls into a better help pop-up
    
    check why 置けません is highlighted in red (issue with potential form?)

    check why 釣られてしまいました is not one word (釣ら is separate from れてしまいました)
        
    when displaying both english and japanese subtitles, highlighting should match up key corresponding words
   
    for highlighted words, show kana and definitions
        on story page, show them below the video
        but need very short definitions (and hiragana spelling)...

    drilling:
        bring back word cooldown filtering?
        maybe user marks words that they want to drill
            mark by default expires after set period (2 weeks?)
            user can filter drill for all marked words
                the marked words also have a cooldown so you don't drill them too frequently before they expire?
        option to pick the most frequent words in a story?
            or list in story shows most frequent words, and user adds them manually
        every time a word is added, the user is notified how many words are in their current set
        bring back per-word cooldowns?

    subtitle tokenization (might be fixed now? it seems new words being added to the db): 
        まとめたので should not include ので as part of the verb
        手触り isn't properly tokenized? The word is not in the words table for some reason
        怠け者 is not in words table

    for highlighted verbs, use coloring to indicate the form
        or maybe use icons?
        also apply to non-highlighted verbs too?

    in-app subtitle editor
        visual timeline for editing subtitle timings
        content view shows the subtitles and allows you to break/join lines and edit the start/end times
        maybe content view / subtitle editing should be a separate page
            player will be smaller
        or maybe just a mode that shows the content and makes the player much smaller + hides the subtitle overlay on the video
            the edit button on the subtitles would reveal this mode
                button next to player takes you back to normal mode
        in edit mode, buttons to shove the timings forward and back
            this would replace the current hot keys
            could allow for more fine grained adjustments than the current hotkeys
        hotkey to delete the current subtitle (get rid of filler, like sounds)
            instead of automatically skipping to next subtitle, add manual skip points or ranges?        

    button to mark current cue as continuing into next (for purpose of playing individual cues)
        insert arrow char at end of the caption?
        maybe actually show these subtitles at same time (with marker or highlight indicating which is the current for marking purposes)
    button to mark current cue as skipped (not counted for purpose of playing individual cues and cue navigation)
        skipped captions have different color background            

    import link should just import all sources instead of taking you to another page
        refresh main page when done

    test clean install on other computer (mac and windows)

    fix word kanji field
        many seem to be missing kanji info
            maybe import error from earlier version?
            or is this a bug in current importing?

    generate transcript file from text file
        e.g. foo.ja.txt
        treat every sentence as its own subtitle
            just make up timing: space them out by a few seconds in order

    add a kana drilling page
        allow to pick exactly which characters to include?
        or fixed sets of similar characters?

        no typing, just reveal of sound?
            tap key to reveal one by one or reveal for all characters?
            
        scrape list of katakana words from the dictionary
            for a katakana word, use altering colors to indicate correspondence of romaji to kana
            order the words by number of characters

        exclude the oddball characters that never come up? or just go by words

    import stories with just katakana words?
        https://www.youtube.com/watch?v=F8tu5CeVWDM
        https://www.youtube.com/watch?v=hrjV4VuDfiU
    

    clicking info marks should show a modal popup with the info

    link to get translation of whole story content? (what is the limit?)

    audit for dead css styles

    drill auto play mode
        - show a word with its definition (and play audio?) for n seconds, then automatically move to the next
        - words auto drilled will be temporarily marked
        - when done with the drill, button to decrement counter for all words that were temporarily marked
        - show the word very large, play the audio, short pause before next word
            (only play cards with audio? maybe an option)

    subtitles: 
        english: https://subscene.com/   https://www.opensubtitles.org/en/search/subs   https://www.podnapisi.net
        japanese: https://kitsunekko.net/dirlist.php?dir=subtitles%2Fjapanese%2F 


    - highlighting for proper names
        - filter out of vocab? or just make it a word category you can filter for?
    
    - definition for transitive / intransitive verb pairs should always show its pair

    - in absence of baseform, maybe should NOT use surface? investigate "引き出し", "飛べる", "鬼滅の" -> "滅"
        - potential form should not count as verb base form: e.g. 飛べる should be added only as 飛ぶ, not as 飛べる


cross compile:

    GOOS=darwin GOARCH=amd64 go build -o japanese_mac
    GOOS=linux GOARCH=amd64 go build -o japanese_linux

cut video: 
    
    ffmpeg -i input.mp4 -ss 00:05:10 -to 00:15:30 -c:v copy -c:a copy output2.mp4

convert all mkv to mp4 in dir:
    
    for f in *.mkv; do ffmpeg -i "$f" -acodec copy -b:v 1500k -maxrate 3000k "${f%.*}.mp4"; done

convert all mp4 to mp3 in dir:

    for f in *.mp4; do ffmpeg -i "$f" "${f%.*}.mp3"; done


    ffmpeg -i [input] -c:a copy -c:v libx265 -an -r 24000/1001 -crf 23 -preset slow -tune animation -x265-params limit-sao=1:deblock=1,1:bframes=8:ref=6:psy-rd=1.5:psy-rdoq=2:aq-mode=3 -pix_fmt yuv420p10le [output]


use puppeteer to scrape for transcripts and meta info
use podcast-dl (https://www.npmjs.com/package/podcast-dl/v/7.0.0-async.1) to get audio files
    npx podcast-dl --url <PODCAST_RSS_URL>


- When encountering a new compound word, should include parts as related words.

- Readings should include spaces at border between kanji: e.g. 最近稼働 is given reading "さい きん か どう". Unfortunately, this info is not in the entries, so would have to infer from possible readings of the kanji. In some cases this is not fully determinable: e.g. for kanji spelling AB, might have possible readings "xy z" but also "x yz". (maybe just display such cases with special highlight, e.g. "xyz" in red indicates that it should be split but the split point is ambiguous)
- Use priority to star the preferred spellings / readings.
- Display "other forms". Can we get the frequency of use for various forms from kanshudo?
- Mark old readings/spellings.
- Related words, kanji, info:
    x Show all component kanji
    - Mark ellided -i and -u sounds. (What about cases like 室 where the shi is part of the preceding syllabal?)
    - All forms of a verb, with pitch
    - Corresponding verb/noun (e.g. 読む / 読み)
    - Words with similar definitions that you've already encountered.
    - Find all homonyms with same pitch.
    - Find all homonyms with different pitch.
    - Find all homonyms from all possible verb forms