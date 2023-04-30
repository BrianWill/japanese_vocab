var newStoryText = document.getElementById('new_story_text');
var newStoryButton = document.getElementById('new_story_button');
var storyList = document.getElementById('story_list');
var newStoryTitle = document.getElementById('new_story_title');
var newStoryLink = document.getElementById('new_story_link');
var storyTitle = document.getElementById('story_title');
var tokenizedText = document.getElementById('tokenized_story');
var wordList = document.getElementById('word_list');
var definitionsDiv = document.getElementById('definitions');
var kanjiResultsDiv = document.getElementById('kanji_results');
var addWordsButton = document.getElementById('add_words_button');

var story = null;

newStoryButton.onclick = function (evt) {
    let data = {
        content: newStoryText.value,
        title: newStoryTitle.value,
        link: newStoryLink.value
    };

    fetch('/create_story', {
        method: 'POST', // or 'PUT'
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(data),
    }).then((response) => response.json())
        .then((data) => {
            console.log('Success:', data);
        })
        .catch((error) => {
            console.error('Error:', error);
        });
};


document.body.onload = function (evt) {
    getStoryList();

    if (window.location.pathname.startsWith('/read/')) {
        let storyId = window.location.pathname.substring(6);
        console.log("opening story with id " + storyId);
        openStory(storyId);
    }
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


function updateStoryList(stories) {
    let html = '';

    html += `<h3>Active stories</h3>`
    for (let s of stories) {
        if (s.state !== 'active') {
            continue;
        }
        html += `<li>
            <a story_id="${s.id}" action="drill" href="#">drill</a>&nbsp;&nbsp;
            <a story_id="${s.id}" action="mark_inactive" href="#">mark inactive</a>&nbsp;&nbsp;
            <a story_id="${s.id}" action="mark_unread" href="#">mark unread</a>&nbsp;&nbsp;
            <a story_id="${s.id}" href="#">${s.title}</a>
            </li>`;
    }

    html += `<h3>Inactive stories</h3>`
    for (let s of stories) {
        if (s.state !== 'inactive') {
            continue;
        }
        html += `<li>
            <a story_id="${s.id}" action="drill" href="#">drill</a>&nbsp;&nbsp;
            <a story_id="${s.id}" action="mark_active" href="#">make active</a>&nbsp;&nbsp;
            <a story_id="${s.id}" action="mark_unread" href="#">mark unread</a>&nbsp;&nbsp;
            <a story_id="${s.id}" href="#">${s.title}</a>
            </li>`;
    }

    html += `<h3>Unread stories</h3>`
    for (let s of stories) {
        if (s.state !== 'unread') {
            continue;
        }
        html += `<li>
            <a story_id="${s.id}" action="drill" href="#">drill</a>&nbsp;&nbsp;
            <a story_id="${s.id}" action="mark_active" href="#">mark active</a>&nbsp;&nbsp;
            <a story_id="${s.id}" href="#">${s.title}</a>
            </li>`;
    }
    storyList.innerHTML = html;
};

storyList.onclick = function (evt) {
    if (evt.target.tagName == 'A') {
        evt.preventDefault();
        var storyId = evt.target.getAttribute('story_id');

        var action = evt.target.getAttribute('action');
        switch (action) {
            case 'drill':
                window.location.href = `/drill.html?storyId=${storyId}`;
                break;
            case 'mark_inactive':
                markStory(storyId, 'inactive');
                break;
            case 'mark_unread':
                markStory(storyId, 'unread');
                break;
            case 'mark_active':
                markStory(storyId, 'active');
                break;
            default:
                openStory(storyId);
                break;
        }
    }
};


function getStoryList() {
    fetch('/stories_list', {
        method: 'GET', // or 'PUT'
        headers: {
            'Content-Type': 'application/json',
        }
    }).then((response) => response.json())
        .then((data) => {
            console.log('Success:', data);
            updateStoryList(data);
        })
        .catch((error) => {
            console.error('Error:', error);
        });
}

function markStory(id, action) {
    fetch(`/mark/${action}/${id}`, {
        method: 'GET',
        headers: {
            'Content-Type': 'application/json',
        }
    }).then((response) => response.json())
        .then((data) => {
            console.log(`Success ${action}:`, data);
            getStoryList();
        })
        .catch((error) => {
            console.error('Error marking story:', error);
        });
}

function openStory(id) {
    fetch('/story/' + id, {
        method: 'GET', // or 'PUT'
        headers: {
            'Content-Type': 'application/json',
        }
    }).then((response) => response.json())
        .then((data) => {
            console.log('Success:', data);
            story = data;
            storyTitle.innerText = data.title;
            story.tokens = JSON.parse(story.tokens);
            story.words = JSON.parse(story.words);
            displayStory(data);
        })
        .catch((error) => {
            console.error('Error:', error);
        });
}

var wordSet = [];

function displayStory(story) {
    let words = '';
    let punctuationTokens = [' ', '。', '、'];

    wordSet = [];
    let html = '<p>';
    let prior = null;
    for (let i = 0; i < story.tokens.length; i++) {
        let t = story.tokens[i];
        let posClass = '';
        if (t.surface === "。") {
            html += '。</p><p>';
        } else if (t.surface === " ") {
            if (prior && prior.surface !== "。") {
                html += '。</p><p>';
            }
        } else {
            t.drillable = true;
            if ((t.pos === "動詞" && t.pos1 === "接尾") ||
                (t.pos === "助動詞") ||
                (t.surface === "で" && t.pos === "助詞" && t.pos1 === "接続助詞") ||
                (t.surface === "て" && t.pos === "助詞" && t.pos1 === "接続助詞") ||
                (t.surface === "じゃ" && t.pos === "助詞" && t.pos1 === "副助詞") ||
                (t.surface === "し" && t.pos === "動詞" && t.pos1 === "自立")) {  // auxilliary verb
                posClass = 'verb_auxiliary';
                t.drillable = false;
            } else if (t.pos === "動詞" && t.pos1 === "非自立") { // auxilliary verb
                posClass = 'verb_auxiliary';
            } else if ((t.pos === "助詞" && t.pos1 === "格助詞") || // case particle
                (t.pos === "助詞" && t.pos1 === "接続助詞") ||   // conjunction particle
                (t.pos === "助詞" && t.pos1 === "係助詞") || // binding particle (も　は)
                (t.pos === "助詞" && t.pos1 === "副助詞")) {  // auxiliary particle
                posClass = 'particle';
                t.drillable = false;
            } else if (t.pos === '副詞') {
                posClass = 'adverb';
            } else if (t.pos === "接続詞" && t.pos1 === "*") { // conjunction
                posClass = 'conjunction';
            } else if ((t.pos === "助詞" && t.pos1 === "連体化") || // connecting particle　(の)
                (t.pos === "助詞" && t.pos1 === "並立助詞")) {  // connecting particle (や)
                posClass = 'connecting_particle';
                t.drillable = false;
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
                t.drillable = false;
            } else {
                posClass = 'pad_left';
            }
            html += `<span tokenIndex="${i}" class="${posClass}">${t.surface}</span>`;
        }

        if (t.drillable && !punctuationTokens.includes(t.surface)) {
            //let baseForm = t.baseForm !== t.surface ? t.baseForm : '';
            wordSet.push(t);
            let pronunciation = t.pronunciation !== t.reading ? t.pronunciation : '';
            words += `<tr class="${posClass}">
                <td>${t.baseForm || t.surface}</td>
                <td>${t.inflectionalForm}, ${t.inflectionalType}</td>
                <!--<td>${pronunciation}</td>-->
                <td>${t.pos}, ${t.pos1}, ${t.pos2}, ${t.pos3}</td>
                </tr>`;
        }

        // if (!t.drillable && !punctuationTokens.includes(t.surface)) {
        //     //let baseForm = t.baseForm !== t.surface ? t.baseForm : '';
        //     let pronunciation = t.pronunciation !== t.reading ? t.pronunciation : '';
        //     words += `<tr style="color: white; background-color: black;" class="${posClass}">
        //         <td>${t.surface}</td>
        //         <td>${t.reading}</td>
        //         <td>${t.baseForm}</td>
        //         <td>${t.inflectionalForm}, ${t.inflectionalType}</td>
        //         <!--<td>${pronunciation}</td>-->
        //         <td>${t.pos}, ${t.pos1}, ${t.pos2}, ${t.pos3}</td>
        //         </tr>`;
        // }
        prior = t;
    }
    tokenizedText.innerHTML = html + '</p>';
    wordList.innerHTML = words;

    console.log(wordSet);
}

var selectedTokenIndex = null;

// tokenizedText.onmouseover = function (evt) {
//     var index = evt.target.getAttribute("tokenIndex");
//     if (index) {
//         displayDefinition(index);
//     } else {
//         if (selectedTokenIndex !== null) {
//             displayDefinition(selectedTokenIndex);
//         }
//     }
// };

tokenizedText.onmousedown = function (evt) {
    var index = evt.target.getAttribute("tokenIndex");
    if (index) {
        selectedTokenIndex = index;
        displayDefinition(index);
        if (evt.ctrlKey) {
            addWord();
        }
    }
};

addWordsButton.onclick = function (evt) {
    fetch('/add_words', {
        method: 'POST', // or 'PUT'
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(wordSet),
    }).then((response) => response.json())
        .then((data) => {
            console.log('Success:', data);
        })
        .catch((error) => {
            console.error('Error:', error);
        });
};

function addWord() {
    if (selectedTokenIndex === null) {
        return;
    }
    var token = story.tokens[selectedTokenIndex];

    fetch('/add_word', {
        method: 'POST', // or 'PUT'
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(token),
    }).then((response) => response.json())
        .then((data) => {
            console.log('Success:', data);
        })
        .catch((error) => {
            console.error('Error:', error);
        });
}

function displayDefinition(index) {
    var token = story.tokens[index];
    getKanji(token.baseForm + token.surface); // might as well get all possibly relevant kanji
    html = '';
    for (let entry of token.entries) {
        html += displayEntry(entry);
    }
    definitionsDiv.innerHTML = html;
}

// storyText.onmouseup = function (evt) {
//     console.log(document.getSelection().toString());
// };

// storyText.onmouseleave = function (evt) {
//     console.log(document.getSelection().toString());
// };

// document.body.addEventListener('selectionchange', (event) => {
//     console.log('changed');
// });


// window.setInterval(function () {
//     console.log(document.getSelection().toString());
// }, 300);

