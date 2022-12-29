var newStoryText = document.getElementById('new_story_text');
var newStoryButton = document.getElementById('new_story_button');
var storyList = document.getElementById('story_list');
var newStoryTitle = document.getElementById('new_story_title');
var storyTitle = document.getElementById('story_title');
var storyText = document.getElementById('story');
var tokenizedText = document.getElementById('tokenized_story');
var wordList = document.getElementById('word_list');

newStoryButton.onclick = function (evt) {
    let data = {
        content: newStoryText.value,
        title: newStoryTitle.value
    };

    fetch('story', {
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
    console.log('asdf');

    fetch('stories_list', {
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
};

function updateStoryList(stories) {
    let html = '';
    for (let s of stories) {
        html += `<li><a story_id="${s._id}" href="#">${s.title}</a>&nbsp;&nbsp;&nbsp;&nbsp;<a story_id="${s._id}" action="retokenize" href="#">retokenize</a></li>`;
    }
    storyList.innerHTML = html;
};

storyList.onclick = function (evt) {
    if (evt.target.tagName == 'A') {
        var storyId = evt.target.getAttribute('story_id');

        var action = evt.target.getAttribute('action');
        if (action === 'retokenize') {
            console.log('retokenizing');
            retokenize(storyId);
        } else {
            openStory(storyId);
        }        
    }
};

function retokenize(id) {
    fetch('story_retokenize/' + id, {
        method: 'POST', // or 'PUT'
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({}),
    }).then((response) => response.json())
        .then((data) => {
            console.log('Success retokenizing:', data);
            openStory(id);
        })
        .catch((error) => {
            console.error('Error:', error);
        });
}

function openStory(id) {
    fetch('story/' + id, {
        method: 'GET', // or 'PUT'
        headers: {
            'Content-Type': 'application/json',
        }
    }).then((response) => response.json())
        .then((data) => {
            console.log('Success:', data.tokens);
            storyTitle.innerText = data.title;
            //storyText.innerText = data.content;

            let words = '';
            let punctuationTokens = [' ', '。', '、'];

            let html = '<p>';
            let prior = null;
            for (let t of data.tokens) {
                let posClass = '';
                if (t.surface === "。") {
                    html += '。</p><p>';
                } else if (t.surface === " ") {
                    if (prior && prior.surface !== "。") {
                        html += '。</p><p>';
                    }
                } else {
                    if ((t.pos === "動詞" && t.pos1 === "接尾") ||
                        (t.pos === "助動詞") ||
                        (t.surface === "で" && t.pos === "助詞" && t.pos1 === "接続助詞") ||
                        (t.surface === "て" && t.pos === "助詞" && t.pos1 === "接続助詞") ||
                        (t.surface === "じゃ" && t.pos === "助詞" && t.pos1 === "副助詞") ||
                        (t.pos === "動詞" && t.pos1 === "非自立") ||
                        (t.surface === "し" && t.pos === "動詞" && t.pos1 === "自立")) {  // auxilliary verb
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
                    } else {
                        posClass = 'pad_left';
                    }
                    html += `<span class="${posClass}">${t.surface}</span>`;
                }
                
                if (!punctuationTokens.includes(t.surface)) {
                    let baseForm = t.baseForm !== t.surface ? t.baseForm : '';
                    let pronunciation = t.pronunciation !== t.reading ? t.pronunciation : '';
                    words += `<tr class="${posClass}">
                        <td>${t.surface}</td>
                        <td>${t.reading}</td>
                        <td>${baseForm}</td>
                        <td>${t.inflectionalForm}, ${t.inflectionalType}</td>
                        <!--<td>${pronunciation}</td>-->
                        <td>${t.pos}, ${t.pos1}, ${t.pos2}, ${t.pos3}</td>
                        </tr>`;
                }
                if (t.pos4) {
                    console.log(`pos4: ${t.pos4}`);
                }
                prior = t;
            }
            tokenizedText.innerHTML = html + '</p>';
            wordList.innerHTML = words;

        })
        .catch((error) => {
            console.error('Error:', error);
        });
}

storyText.onmousedown = function (evt) {
    //console.log(document.getSelection().toString());
};

storyText.onmouseup = function (evt) {
    console.log(document.getSelection().toString());
};

storyText.onmouseleave = function (evt) {
    console.log(document.getSelection().toString());
};

// document.body.addEventListener('selectionchange', (event) => {
//     console.log('changed');
// });


// window.setInterval(function () {
//     console.log(document.getSelection().toString());
// }, 300);

