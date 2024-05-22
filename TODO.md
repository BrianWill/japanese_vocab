# japanese_vocab TODO

daily schedule system:

    ability in story mode to mark the start time and end time of a video, then extract that to a new story
        gets added to the same source
        writes out a json file for the sub-story
            (so that it can be imported with the rest of the source)
            the transcript data is stored directly in the json file
                transcript_ja
                transcript_en
        use modal popup to pick the name?
            name is same as the parent story + title of your choosing in modal
            popup reports the start time and end time
        option in app to delete these stories (along with the json file)?
            or maybe just option to delete any story in catalog?

    add video_start time and video_end time to 
    append to src url: #t=[starttime][,endtime]
        <source src=http://techslides.com/demos/sample-videos/small.webm#t=2,3 type=video/webm>

    hotkey to open the current japanese subtitle in google translate

    ability to modify a rep's type (in main menu via select box?)

    ability to add an additional rep for a story in the schedule?
        add button on rep that creates a new rep of same type the next day (or none if it the next day already has a rep of that story)

    in app menu for importing / reimporting sources

    undo system for the schedule/log
        upon every change, just store a full copy?
            jsonify in a "backup" table

        what about undoing increments of story and word reps?

    cleanup or remove the "content" display at bottom of story
        for audio stories, shrink the video element

    when logging a rep, display message if the rep had already been logged (or some other error?)
        or just grey out the link when story is already loggged
        or if scheduleId is not defined, log link does not appear
        upon success of logging, redirect to main page?

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

- reload entries bson only when needed by request rather than keeping in memory?
    - strip out unneeded parts of dictionaries

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