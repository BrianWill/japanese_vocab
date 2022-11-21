var newStoryText = document.getElementById('new_story_text');
var newStoryButton = document.getElementById('new_story_button');
var storyList = document.getElementById('story_list');
var newStoryTitle = document.getElementById('new_story_title');
var storyTitle = document.getElementById('story_title');
var storyText = document.getElementById('story');
var tokenizedText = document.getElementById('tokenized_story');

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
    html = '';
    for (let s of stories) {
        html += `<li><a story_id="${s._id}" href="#">${s.title}</a></li>`;
    }
    storyList.innerHTML = html;
};

storyList.onclick = function (evt) {
    if (evt.target.tagName == 'A') {
        var storyId = evt.target.getAttribute('story_id');
        console.log(storyId);
        openStory(storyId)
    }
};

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

            html = '';
            let prior = null;
            for (let t of data.tokens) {
                if (t.surface === "。") {
                    html += '。<br>';
                } else if (t.surface === " ") {
                    if (prior && prior.surface !== "。") {
                        html += '。<br>';
                    }
                } else if (t.pos === "助詞" && t.pos1 === "格助詞") { // case particle
                    html += `<span class="case_particle">${t.surface}</span>`;
                } else if (t.pos === "助詞" && t.pos1 === "連体化") { // connecting particle
                    html += `<span class="connecting_particle">${t.surface}</span>`;
                } else if (t.pos === "助詞" && t.pos1 === "係助詞") { // binding particle (も　は)
                    html += `<span class="binding_particle">${t.surface}</span>`;
                } else if (t.pos === "助詞" && t.pos1 === "副助詞") { // auxiliary particle
                    html += `<span class="auxiliary_particle">${t.surface}</span>`;
                } else if (t.pos === "接続詞" && t.pos1 === "*") { // conjunction
                    html += `<span class="conjunction">${t.surface}</span>`;
                } else if (t.pos === "形容詞") { // i-adj
                    html += `<span class="i_adjective">${t.surface}</span>`;
                } else if (t.pos === "名詞" && t.pos1 === "代名詞") { // pronoun
                    html += `<span class="pronoun">${t.surface}</span>`;
                } else if ((t.pos === "動詞" && t.pos1 === "接尾") ||
                    (t.pos === "助動詞") ||
                    (t.surface === "で" && t.pos === "助詞" && t.pos1 === "接続助詞") ||
                    (t.surface === "て" && t.pos === "助詞" && t.pos1 === "接続助詞") ||
                    (t.pos === "動詞" && t.pos1 === "非自立") ||
                    (t.surface === "し" && t.pos === "動詞" && t.pos1 === "自立")) {  // auxilliary verb
                    html += `<span class="verb_auxiliary">${t.surface}</span>`;
                } else if (t.pos === "動詞") { //　verb
                    html += `<span class="verb">${t.surface}</span>`;
                } else if ((prior && prior.pos === "助詞" && prior.pos1 === "連体化") ||  // preceded by connective particle
                        (prior &&  prior.pos === "接頭詞" && prior.pos1 === "名詞接続")) {  // preceded by prefix
                        html += `<span class="">${t.surface}</span>`;
                    
                } else {
                    html += `<span class="pad_left">${t.surface}</span>`;
                }
                prior = t;
            }
            tokenizedText.innerHTML = html;

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

