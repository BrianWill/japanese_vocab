# japanese_vocab TODO

Mini Lesson Compilation #5 ミニレッスン集 #5 Complete Beginner Japanese 日本語超初心者 Comprehensible Input	https://www.youtube.com/watch?v=8xsKNluTuo8
引き出しの中身 What’s in the drawers	https://cijapanese.com/whats-in-the-drawers-members-only/
日本に関する⚪︎×クイズ Japan Quiz (True or False)	https://cijapanese.com/japan-quiz-true-or-false-members-only/
学校の教科 School subjects	https://cijapanese.com/school-subjects-members-only/
夫の１日 A day in the life of my husband	https://cijapanese.com/a-day-in-the-life-of-my-husband-members-only/
Monster episode 1	https://animelon.com/video/61793101526a1a2df4e2b3c5
マリオの映画を観に行った話 We went to see the Mario movie	https://cijapanese.com/we-went-to-see-the-mario-movie-members-only/
足が弱くなって転んでけがをする子どもが増えている	https://www3.nhk.or.jp/news/easy/k10014062771000/k10014062771000.html
Spot the difference (Mother’s Day) 間違い探し（母の日）	https://cijapanese.com/spot-the-difference-mothers-day-members-only/
冬 Winter	https://cijapanese.com/winter-members-only/
消えた息子の靴下 My son’s missing socks	https://cijapanese.com/my-sons-missing-socks-members-only/
Haru: “行くんですか”跟“行きますか”有什麼不同？問句“～んですか？”的用法和意思！初-中級日語 【台灣學生最常搞錯的日語】 講中文的日本人Haru老師【#28】日文發音，中文字幕	https://www.youtube.com/watch?v=sUy9SrwReyk&list=PL2MpP9BgnjX6GQxUo5YMLdnZl5sALvZ1E&index=1
Japanese Podcast『日本語って！』第63回　歳をとると「丸くなる」？	https://www.youtube.com/watch?v=UYlHH-2AX6w
カゴの中身は何でしょう #2 Guess what’s in the basket! #2	https://cijapanese.com/guess-whats-in-the-basket-2-members-only/
忘れ物の多い息子? My son often forgets to take something to school	https://cijapanese.com/my-son-often-forgets-to-take-something-to-school-members-only/
#41 Short story / トイレの女の子《A girl in the restroom	https://www.youtube.com/watch?v=i6Mh-Dxv8TI
#4 早起（はやお）きは三文（さんもん）の徳（とく） // N3 Level	https://www.youtube.com/watch?v=eItP9NIkutM&list=PLL9Szyax7Exw0fhcn04b0DVbpCiff8i4h
#3 Short story あめだま (amedama) storytelling //N4 Level	https://youtu.be/jDBh7-a92sg
Japanese Podcast for beginners / Ep15 Common sense (Genki 1 level)	https://www.youtube.com/watch?v=Bt_ibvl-Dtk
Psycho Pass ep 1	https://animelon.com/video/57986e7c641f6457395b6b22
Death Note ep 2 (1 of 3)	https://animelon.com/video/571f0854cdec5d5697d3dd28
Death Note ep 2 (2 of 3)	https://animelon.com/video/571f0854cdec5d5697d3dd28
Death Note ep 2 (3 of 3)	https://animelon.com/video/571f0854cdec5d5697d3dd28


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