# japanese_vocab TODO

use puppeteer to scrape for transcripts and meta info
use podcast-dl (https://www.npmjs.com/package/podcast-dl/v/7.0.0-async.1) to get audio files

autoplay drill mode
    show the word very large, play the audio, short pause before next word
        (only play cards with audio? maybe an option)

words need a drill countdown

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