# japanese_vocab TODO

- kanji drills
    - add kanji to sqlite
    - track kanji by story the same way we track words
        - each story has list of kanji ids
    - "drill count" for number of times you've drilled the kanji directly
    - "encounter count" for number of times you've dilled a word containing the kanji

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
    -use AI to replace some words with Japanese, then just add them as stories?
    - for English text story, identify the commonly reoccuring words and replace them with Japanese equivalents. (might be tricky to do accurate translation)
        - generally stick to nouns, adjectives and verbs?

- When baking definitions and kanji in tokens, the spelling / reading that is used in the token should be top (display others in smaller text below).
    - distinguish between readings/spellings/definitions that user has encountered from others
    - prioritize showing definition of word as used in context + "core" definition

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