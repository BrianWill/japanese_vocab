# japanese_vocab TODO   

    reimport story should ignore the existing json
    
    subtitle tokenization: 
        まとめたので should include ので as part of the verb
        手触り isn't properly tokenized? The word is not in the words table for some reason
        怠け者 is not in words table
    

    drill vocab from all stories with queued reps
        limit words in drill to 50?
            user can pick the limit number?
            pick the words at random?
        bring back per-word cooldowns?
    
    mode that only shows unknown words in the subtitles
        either make the non-hinted words totally transparent or make them very faint
        pick words from non-archived set and based on count in the story
            words that occur more frequently should be prioritized
        maybe set max of hint words per subtitle
            perhaps this cap is based on the time duration of the subtitle, i.e. a target hint words per second
    
    mode that displays only (or highlights?) the verbs 
        add coloring or icons that indicate the form

    subtitle jump keys should work even when subtitles are hidden (defaults to japanese subtitle timings)

    while listening, ability to mark times
        useful for listening without subtitles and marking times when you hear something you don't understand
            can then come back to the marked times after

    bug: for new stories, the default end time of the excerpt is NaN
        (this is probably because we need to wait for the video to load before displaying the excerpts?)

    don't reconstruct the subtitle html if it hasn't changed
        (currently there is flicker if you try to select text while playing because the subtitle is being reconstructed as we play)

    getWordsFromExcerpt
        update to use json subtitles instead of vtt transcript

    openTranscript
        get rid of this once subtitle editing is done?
        or maybe keep so users can easily edit the original file?

    in app subtitle editor
        content view shows the subtitles and allows you to break/join lines and edit the start/end times
        maybe content view / subtitle editing should be a separate page
            player will be smaller
        or maybe just a mode that shows the content and makes the player much smaller + hides the subtitle overlay on the video
            the edit button on the subtitles would reveal this mode
                button next to player takes you back to normal mode
        in edit mode, buttons to shove the timings forward and back
            this would replace the current hot keys
            could allow for more fine grained adjustments than the current hotkeys
    
    a subtitle mode that highlights certain words (or only shows those words?) with translation
        e.g. only show verbs, or only show kangoz

    (session?) cookie that stores your last viewed source
        currently annoying that you have to navigate the source pulldown every time you go back to the main page

    visual timeline for editing subtitle timings

    for each word, track all stories / sentences that include the word

    button to mark current cue as continuing into next (for purpose of playing individual cues)
        insert arrow char at end of the caption?
        maybe actually show these subtitles at same time (with marker or highlight indicating which is the current for marking purposes)
    button to mark current cue as skipped (not counted for purpose of playing individual cues and cue navigation)
        skipped captions have different color background            

    furigana for katakana:
        display hiragana above katakana character
        allow user to specify which katakana characters to see hiragana for (so they only see it for the ones that give them trouble)

    hotkey to delete the current subtitle (get rid of filler, like sounds)
        instead of automatically skipping to next subtitle, add manual skip points or ranges?        

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
    
    display content below story with time-marks
        clicking timemark jumps player to timestamp

    alternate source dir format:
        main json file includes all info about every story
        or just always have one json file per story?

    playback mode where randomly selected words in subtitles are highlighted with their definitions shown in same color

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

    
    word counts per story (and per excerpt?):
        server side: given story id, return the count of unarchived and archived words
        show count for each story in current stories
        show count in story page
        show in catalog? would be expensive, but not impossible
    
    list of recently completed:
        stories with no queued reps but with a logged rep within last week

    clicking info marks should show a modal popup with the info
   
    hotkey to open the current japanese subtitle in google translate

    link to get translation of whole story content? (what is the limit?)

    audit for dead css styles

    for current subtitle, show list of all the words with their word info
        easy way to change the word status and set remaining reps
        when importing story, need to store word ids for each subtitle? how to get words? 


    deduplicate the word ids in words field of stories


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

store audio link with timestamps
    any way to capture from stories with youtube audio? probably not


- toucan-like stories:
    - use AI to replace some words with Japanese, then just add them as stories?
    - for English text story, identify the commonly reoccuring words and replace them with Japanese equivalents. (might be tricky to do accurate translation)
        - generally stick to nouns, adjectives and verbs?

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