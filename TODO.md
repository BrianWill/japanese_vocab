# japanese_vocab TODO

- enforce unqiueness of story title
- button to clear the create story form
- the temp text of the inputs should be greyed out and disappear when the user selects the box
- async story creation:
    - immediately insert story and link and content
    - reset the form when story is received and display message confirming it was added and is being tokenized
    - isTokenized flag
    - browser automatically requests async request to tokenize
        - story tokenization should be idempotent?
        - flash message when story has fininshed tokenizing

- in tokenization, should distinguish between paragraphs and sentences. Provide an option to separate sentences to separate lines or not?

- definition for transitive / intransitive verb pairs should always show its pair

- in story, definition shows drill stats for word and hotkey let's you modify its counters

- drill filter: words that have 0 drills (maybe replace filter for recently added with filter for max number of times drilled)

- option to drill all wrong words, regardless of cooldown

- in absence of baseform, maybe should NOT use surface? investigate "引き出し", "飛べる", "鬼滅の" -> "滅"
    - potential form should not count as verb base form: e.g. 飛べる should be added only as 飛ぶ, not as 飛べる

- revisit word cooldowns:
    - what should policy be for words marked wrong?
    - words marked wrong within "wrong cooldown"  window should not be marked correct (but they should still get a tally?)

- option to "disable" all words unique to that story?
    - alternavitely, perhaps just have ability to mark stories and filter drill words unique to stories with certain markings?
        - or do we just filter for words belonging to stories that have certain countdowns? e.g. words from all stories with countdown of 2 or lower
    - better to disable rather than remove because we don't want to lose drill history & counts
    - checkbox next to story that toggles enable/disable of its words
    - only words that are unique to that story are disabled
    - requires re-processing all stories when new stories are added / removed?
        - probably should just track which stories each word belongs to
    - useful for adding stories that we want to get to later but aren't ready to start reading / drilling
    
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
- make "add story" look nicer
- click word in drill list to sort it to top/active (sometimes curious and want to skip ahead to see def of word down the list)
- better hotkeys for drilling (cursor keys?)

- stories have countdown and cooldowns
    - story list can filter out zero countdown stories
    - sort smallest non-zero countdown to top

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
- roman letters in word search to search by definition (might be very slow; would text indexing help?)

- Track encountered words / kanji. (sqlite?)
- Drilling for words / kanji. Filter drill sets for encountered, for specific stories, for common features (e.g. 'godan verbs ending in つ')

- Add pitch info to entries in mongo.
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


x Map each story token to a dictionary entry. (map to specific meaning?)
X When clicking/hovering on each word in a story, display the definition and kanji.
x Result entries should be sorted on server in order of fewest characters in spelling and/or reading