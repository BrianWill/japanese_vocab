# japanese_vocab TODO

- track verbs by initial sound

- in story list, icon indicating if story is on cooldown

- kanji drills
    - add kanji to sqlite
    - track kanji by story the same way we track words
        - each story has list of kanji ids
    - "drill count" for number of times you've drilled the kanji directly
    - "encounter count" for number of times you've dilled a word containing the kanji

- in story, definition shows drill stats for word



- highlighting for proper names

    - immediately insert story and link and content
    - reset the form when story is received and display message confirming it was added and is being tokenized
    - isTokenized flag
    - display error when user attempts to add story with same title as existing story
    - browser automatically requests async request to tokenize
        - story tokenization should be idempotent?
        - flash message when story has fininshed tokenizing

- in tokenization, should distinguish between paragraphs and sentences. Provide an option to separate sentences to separate lines or not?

- definition for transitive / intransitive verb pairs should always show its pair

- in absence of baseform, maybe should NOT use surface? investigate "引き出し", "飛べる", "鬼滅の" -> "滅"
    - potential form should not count as verb base form: e.g. 飛べる should be added only as 飛ぶ, not as 飛べる


    
- option to delete stories (and optionally remove all associated words--but only the words unique to that story?)
- should the word id's for each story be in sorted order?

- keep list of timestamps for every time word is reviewed and another list for every time you answer wrong? Could be useful stats at some point
    - probably only need last couple dozen rather than complete history

- story display: split into p tags on \n and \n\n
    - add paragraph for ! and ？
    - at story tokenization time, add 。 when \n is not preceeded by end punctuation.
    - need to add paragraph markers into the token stream?
    - number the paragraphs?

- reload entries bson only when needed by request rather than keeping in memory?
    - strip out unneeded parts of dictionaries

- improve handling of people/place names in stories
    - special highlighting?
    - filter out of vocab? or just make it a word category you can filter for?

- toucan-like stories:
    -use AI to replace some words with Japanese, then just add them as stories?
    - for English text story, identify the commonly reoccuring words and replace them with Japanese equivalents. (might be tricky to do accurate translation)
        - generally stick to nouns, adjectives and verbs?

- When baking definitions and kanji in tokens, the spelling / reading that is used in the token should be top (display others in smaller text below).
    - distinguish between readings/spellings/definitions that user has encountered from others
    - prioritize showing definition of word as used in context + "core" definition

- When encountering a new compound word, should include parts as related words.
- Audio / video sync with text. (time mark per token? If no time mark, then search back through tokens for nearest prior time mark.)
- mousehover over word displays reading of word and its definition below the line (anything that doesn't fit in the line is cut off)

- sort kanji results to order of kanji as they appear in the word

- Readings should include spaces at border between kanji: e.g. 最近稼働 is given reading "さい きん か どう". Unfortunately, this info is not in the entries, so would have to infer from possible readings of the kanji. In some cases this is not fully determinable: e.g. for kanji spelling AB, might have possible readings "xy z" but also "x yz". (maybe just display such cases with special highlight, e.g. "xyz" in red indicates that it should be split but the split point is ambiguous)
- Readings should display pitch in style of https://www.gavo.t.u-tokyo.ac.jp/ojad/eng/pages/home
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