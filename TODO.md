# japanese_vocab TODO

story status:

    CATALOG            (initial state; never read before)
    IN PROGRESS    (currently repeating)
        "repetitions remaining" countdown to zero: how many times to repeat the story. When reaches zero, story is put in backlog
    BACKLOG        (read before and maybe want to come back to it)
    ARCHIVED       (never want to see it again)

when a story is added into catalog, all words are put in the catalog

marking a story IN PROGRESS will mark all UNKNOWN words as IN PROGRESS and sets the drill count

when looking at a story, see list of its words sorted by status

word status:

    CATALOG         (never studied before)
    IN PROGRESS     (currently drilling)
        "repetitions remaining" countdown to zero: how many times to drill. When reaches zero, word is put in backlog
    BACKLOG         (drilled before but currently set aside)
    ARCHIVED        (never want to see it again)
    DISCARD         (word that is malformed or maybe not even a word, etc.)

a kanji is just a word entry marked as being just a kanji


MAIN page

    - shows only IN PROGRESS stories

STORY CATALOG page

    - a page that show all stories
        - sorted by source or status (toggle option at top of page)
        - also can sort by level?

STORY PAGE

    - shows text of story

STORY VOCAB PAGE

    - displays all the words in a story, sorted by status
    - option to drill the words just of that story?

DRILL page

    - words selected from IN PROGRESS set
    - shows the list of words to drill in right sidebar
    - displays current word / kanji on the left
    - hit key to mark word as drilled and move on to the next
    - in auto play mode:
        - show a word with its definition (and play audio?) for n seconds, then automatically move to the next
        - words auto drilled will be temporarily marked
        - when done with the drill, button to decrement counter for all words that were temporarily marked

IMPORTING STORIES

    - command line? or page with file open dialog?
    - add set of stories through json
    - stories are uniquely identified by source + title
        - during import, if story already exists, it is updated (except for status)


CATALOG_STORIES

    title
    source          (e.g. Nihongo Picnic)
    status          (unknown, in progress, backlog, archived)   string?
    countdown       (planned repetitions remaining)
    date
    link
    episode_number
    audio
    video
    content
    content_format
    transcript_en           (subtitle file)
    transcript_jp           (subtitle file)
    transcript_en_format    (subtitle file)
    transcript_jp_format    (subtitle file)
    words                   (json text list of ids of all the words contained in the story; empty if the story has never yet been marked IN PROGRESS)
    

WORDS

    words identified by base_form

    base_form
    drill_count                 (total number of times word has been drilled)
    drill_countdown             (times remaining to drill the word)
    date_marked                 (last time the word was drilled)
    category                    (bit flags: kanji, ichidan verb, godan verb)
    status                      (unknown, in progress, backlog, archived)
    audio                       (link to audio file that contains the word)
    audio_start                 (float in seconds)
    audio_end
    

    to remove:
        rank


story JSON format:

    title
    source          (e.g. Nihongo Picnic)
    date
    link
    episode_number
    audio
    video
    content
    content_format
    transcript_en           (subtitle file)
    transcript_jp           (subtitle file)
    transcript_en_format    (subtitle file)
    transcript_jp_format    (subtitle file)

subtitles: 
    english: https://subscene.com/   https://www.opensubtitles.org/en/search/subs   https://www.podnapisi.net
    japanese: https://kitsunekko.net/dirlist.php?dir=subtitles%2Fjapanese%2F 

how to convert .ass to .srt?

for a word, track all sentences that include the word

x nihongo picnic
x sakura
x japanese with noriko
x cj
japanese with shun (pdfs?)


ffmpeg -i [input] -c:v libx265 -an -r 24000/1001 -crf 23 -preset slow -tune animation -x265-params limit-sao=1:deblock=1,1:bframes=8:ref=6:psy-rd=1.5:psy-rdoq=2:aq-mode=3 -pix_fmt yuv420p10le [output]



use puppeteer to scrape for transcripts and meta info
use podcast-dl (https://www.npmjs.com/package/podcast-dl/v/7.0.0-async.1) to get audio files
    npx podcast-dl --url <PODCAST_RSS_URL>

autoplay drill mode
    show the word very large, play the audio, short pause before next word
        (only play cards with audio? maybe an option)

words need a drill countdown

story importer from json file

[
    {
        title: "",   
        date: "",
        episodeNumber: "",
        audio: "",   // path or url?
        video: "",   // path or url?
        link: "",    // url of source
        content: "",  // transcript
        contentFormat: "",  // "text" or "srt"
    }

]

story importer for podcasts:
    nihongo picnic
    sakura tips
    japanese with shuntod
    japanese with norico
    cj (how to enable download from patreon?)

    a way to get browse in the app? or just provide a URL / podcast number?

store audio link with timestamps
    anyway to capture from stories with youtube audio? probably not

import Enlgish and Japanese from transcript files
    display the subtitles simultenously together on the video

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

- display error when user attempts to add story with same title as existing story
  
- definition for transitive / intransitive verb pairs should always show its pair

- in absence of baseform, maybe should NOT use surface? investigate "引き出し", "飛べる", "鬼滅の" -> "滅"
    - potential form should not count as verb base form: e.g. 飛べる should be added only as 飛ぶ, not as 飛べる

- should the word id's for each story be in sorted order?

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