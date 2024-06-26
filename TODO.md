# japanese_vocab TODO

daily schedule system:
    
    alternate source dir format:
        main json file includes all info about every story
        or just always have one json file per story?

    after "stories with queued reps", show stories with recent logged reps (but no queued reps)

    maybe don't need drill/listening distinction in queue?
        queue could just be a count of listening reps:
            norm is that you drill words right before doing a listening rep
                only logging a listening rep

    in app page for importing / reimporting sources
        page lists the sources
        for each source, show number of in directory, show number already imported
        link to import individual source
        link to import all sources
            

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

    (get rid of rep count in DB and just sum up the logged reps)
        display separate rep count for listening and drills
    
    way to set start and end points for the current story:
        ctrl plus `[` `]` ?

    hotkey to open the current japanese subtitle in google translate

    link to get translation of whole story content? (what is the limit?)


    convert podcast content into transcripts?   

    audit for dead code

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


for a word, track all sentences that include the word


ffmpeg -i [input] -c:a copy -c:v libx265 -an -r 24000/1001 -crf 23 -preset slow -tune animation -x265-params limit-sao=1:deblock=1,1:bframes=8:ref=6:psy-rd=1.5:psy-rdoq=2:aq-mode=3 -pix_fmt yuv420p10le [output]


use puppeteer to scrape for transcripts and meta info
use podcast-dl (https://www.npmjs.com/package/podcast-dl/v/7.0.0-async.1) to get audio files
    npx podcast-dl --url <PODCAST_RSS_URL>

store audio link with timestamps
    any way to capture from stories with youtube audio? probably not

oscilloscope for selecting audio range for a word

<audio id="audio" src="test.mp3"></audio>
<script type="text/javascript">
    var context = new webkitAudioContext;
    var el = document.getElementById('audio');
    var source = context.createMediaElementSource(el);
    source.connect(context.destination);
    el.play();
</script>




- highlighting for proper names
    - filter out of vocab? or just make it a word category you can filter for?
  
- definition for transitive / intransitive verb pairs should always show its pair

- in absence of baseform, maybe should NOT use surface? investigate "引き出し", "飛べる", "鬼滅の" -> "滅"
    - potential form should not count as verb base form: e.g. 飛べる should be added only as 飛ぶ, not as 飛べる

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