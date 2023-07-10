var storyTitle = document.getElementById('story_title');
var tokenizedText = document.getElementById('tokenized_story');
var wordList = document.getElementById('word_list');
var definitionsDiv = document.getElementById('definitions');
var kanjiResultsDiv = document.getElementById('kanji_results');

var story = null;

document.body.onload = function (evt) {
    var url = new URL(window.location.href);
    var storyId = parseInt(url.searchParams.get("storyId") || undefined);
    openStory(storyId);
};

tokenizedText.onwheel = function (evt) {
    if (evt.wheelDeltaY < 0) {
        if (tokenizedText.scrollTop >= tokenizedText.scrollTopMax) {
            evt.preventDefault();
        }
    } else {
        if (tokenizedText.scrollTop <= 0) {
            evt.preventDefault();
        }
    }
};

const STORY_READ_COOLDOWN = 60 * 60 * 16; // 16 hours


function openStory(id) {
    fetch('/story/' + id, {
        method: 'GET', // or 'PUT'
        headers: {
            'Content-Type': 'application/json',
        }
    }).then((response) => response.json())
        .then((data) => {
            story = data;
            storyTitle.innerHTML = `<a href="${story.link}">${story.title}</a>`;
            story.tokens = JSON.parse(story.tokens);
            story.words = JSON.parse(story.words);
            for (let key in story.words) {
                let word = story.words[key];
                word.definitions = JSON.parse(word.definitions);
            }
            //console.log(`/story/${id} success:`, story);
            displayStory(data);
        })
        .catch((error) => {
            console.error('Error:', error);
        });
}

function displayStory(story) {
    let words = '';
    let punctuationTokens = [' ', '。', '、'];

    let html = '<p>';
    let prior = null;
    for (let i = 0; i < story.tokens.length; i++) {
        let t = story.tokens[i];
        let posClass = '';
        if (t.surface === "。") {
            html += '。</p><p>';
        } else if (t.surface === "\n\n") {
            if (prior && prior.surface !== "。") {
                html += '</p><p>';
            }
        } else if (t.surface === "\n") {
            if (prior && prior.surface !== "。") {
                html += '</p><p>';
            }
        } else if (t.surface === " ") {
            console.log("surface was space");
            if (prior && prior.surface !== "。") {
                html += '。</p><p>';
            }
        } else {
            if ((t.pos === "動詞" && t.pos1 === "接尾") ||
                (t.pos === "助動詞") ||
                (t.surface === "で" && t.pos === "助詞" && t.pos1 === "接続助詞") ||
                (t.surface === "て" && t.pos === "助詞" && t.pos1 === "接続助詞") ||
                (t.surface === "じゃ" && t.pos === "助詞" && t.pos1 === "副助詞") ||
                (t.surface === "し" && t.pos === "動詞" && t.pos1 === "自立")) {  // auxilliary verb
                posClass = 'verb_auxiliary';
            } else if (t.pos === "動詞" && t.pos1 === "非自立") { // auxilliary verb
                posClass = 'verb_auxiliary';
            } else if ((t.pos === "助詞" && t.pos1 === "格助詞") || // case particle
                (t.pos === "助詞" && t.pos1 === "接続助詞") ||   // conjunction particle
                (t.pos === "助詞" && t.pos1 === "係助詞") || // binding particle (も　は)
                (t.pos === "助詞" && t.pos1 === "副助詞")) {  // auxiliary particle
                posClass = 'particle';
            } else if (t.pos === '副詞') {
                posClass = 'adverb';
            } else if (t.pos === "接続詞" && t.pos1 === "*") { // conjunction
                posClass = 'conjunction';
            } else if ((t.pos === "助詞" && t.pos1 === "連体化") || // connecting particle　(の)
                (t.pos === "助詞" && t.pos1 === "並立助詞")) {  // connecting particle (や)
                posClass = 'connecting_particle';
            } else if (t.pos === "形容詞") { // i-adj
                posClass = 'i_adjective pad_left';
            } else if (t.pos === "名詞" && t.pos1 === "代名詞") { // pronoun
                posClass = 'pronoun pad_left';
            } else if (t.pos === "連体詞") { // adnominal adjective
                posClass = 'admoninal_adjective pad_left';
            } else if (t.pos === "動詞") { //　verb
                posClass = 'verb pad_left';
            } else if (t.pos === "名詞" && t.pos1 === "接尾") { // noun suffix
                posClass = 'noun';
            } else if ((prior && prior.pos === "助詞" && (prior.pos1 === "連体化" || prior.pos1 === '並立助詞')) ||  // preceded by connective particle
                (prior && prior.pos === "接頭詞" && prior.pos1 === "名詞接続")) {  // preceded by prefix
                posClass = 'noun';
            } else if (t.pos === "名詞") { // noun
                posClass = 'noun';
            } else if (t.pos === "記号") { // symbol
            } else if (t.pos == "号") { // counter
                posClass = 'counter';
            } else {
                posClass = 'pad_left';
            }
            html += `<span tokenIndex="${i}" class="${posClass}">${t.surface}</span>`;
        }

        prior = t;
    }
    tokenizedText.innerHTML = html + '</p>';

}

var selectedTokenIndex = null;

const MAX_WORD_COUNTDOWN = 5;

tokenizedText.onmousedown = function (evt) {
    let index = evt.target.getAttribute("tokenIndex");
    if (index) {
        selectedTokenIndex = index;
        
        if (evt.ctrlKey) {  // inc countdown of the word
            let token = story.tokens[index];
            let word = story.words[token.wordId];
            if (word.countdown < word.countdown_max) {
                word.countdown = word.countdown_max;
                updateWord(word);
            }
        }   

        displayDefinition(index);
    }
};

function displayDefinition(index) {
    let token = story.tokens[index];
    let word = story.words[token.wordId];
    getKanji(token.baseForm + token.surface);
    html = '';
    for (let entry of word.definitions) {
        html += displayEntry(entry);
    }
    definitionsDiv.innerHTML = html;
}