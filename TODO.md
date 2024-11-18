# japanese_vocab TODO   

    make player controls into a better help pop-up     
    
    add subtitle offset field to videos
        if user adjusts the subtitles such that first subtitle timemark goes under 0, then this offset becomes non-zero
            so each subtitles timemarks are their stored timemarks - the offset
                subtitles in story_text will be displayed with this offset factored in
                subtitles with end timemarks below 0 will still be shown in the story text
                    but you can't jump to a timemark below 0 (doing so just jumps to 0)
    
    check why 置けません is highlighted in red (issue with potential form?)

    check why 釣られてしまいました is not one word (釣ら is separate from れてしまいました)
        
    when displaying both english and japanese subtitles, highlighting should match up key corresponding words
   
    for highlighted words, show kana and definitions
        on story page, show them below the video
        but need very short definitions (and hiragana spelling)...

    drilling options:
        limit words in drill to 50?
            user can pick the limit number?
            pick the words at random?x
        or pick range of stories to include
        or pick set of individually selected stories?
        maybe user marks non-archived words?
            mark by default expires after set period (2 weeks?)
            user drills all marked words
                the marked words also have a cooldown so you don't drill them too frequently before they expire?
        option to pick the most frequent words in a story?
            or list in story shows most frequent words, and user adds them manually
        every time a word is added, the user is notified how many words are in their current set
        bring back per-word cooldowns?

    debug mode for importing:
        for story content:
            show spaces between each word
            if word baseform is not in db, mark word with an icon
            on hover over word, show all the info from the analyzer

    subtitle tokenization (might be fixed now? it seems new words being added to the db): 
        まとめたので should not include ので as part of the verb
        手触り isn't properly tokenized? The word is not in the words table for some reason
        怠け者 is not in words table

        investigate why a number of words seem to be missing from the word database
    
    feature to analyze the story catalog to find stories with right balance of known and unknown words
        maybe identify appropriate excerpts by looking at run of subtitles
        for each word, track all stories / sentences that include the word

    for highlighted verbs, use coloring to indicate the form
        or maybe use icons?
        also apply to non-highlighted verbs too?

    subtitle jump keys should work even when subtitles are hidden (defaults to japanese subtitle timings)

    while listening, ability to mark times
        useful for listening without subtitles and marking times when you hear something you don't understand
            can then come back to the marked times after

    openTranscript
        get rid of this once subtitle editing is done?
        or maybe keep so users can easily edit the original file?

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

    
    word counts per story (and per excerpt?):
        server side: given story id, return the count of unarchived and archived words
        show count for each story in current stories
        show count in story page
        show in catalog? would be expensive, but not impossible

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