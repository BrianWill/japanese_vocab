var cardsDiv = document.getElementById('cards');
var answerDiv = document.getElementById('answer');
var drillInfoH = document.getElementById('drill_info');
var drillCountInput = document.getElementById('drill_count');
var drillCountInputText = document.getElementById('drill_count_text');
var drillTypeSelect = document.getElementById('drill_type');
var doneButton = document.getElementById('done_button');
var drillComlpeteDiv = document.getElementById('drill_complete');
var kanjiResultsDiv = document.getElementById('kanji_results');
var definitionsDiv = document.getElementById('definitions');
var drillFilterSelect = document.getElementById('drill_filter');

const COOLDOWN_TIME = 60 * 60 * 3 // number of seconds

var drillSet = null;
var answeredSet = [];
var storyIds = {};
var stories;

function newDrill() {

    let ids = Object.keys(storyIds).map(x => parseInt(x));

    fetch('words', {
        method: 'POST', // or 'PUT'
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({
            count: parseInt(drillCountInput.value),
            drill_type: drillTypeSelect.value,
            drill_filter: drillFilterSelect.value,
            storyIds: ids
        })
    }).then((response) => response.json())
        .then((data) => {
            console.log('Drill success:', data);
            drillComlpeteDiv.style.display = 'none';
            shuffle(data.words);
            drillSet = data.words;
            answeredSet = [];

            var titles = [];
            
            if (ids.includes(-1)) {
                titles.push('ALL CURRENT STORIES');
            }
            if (ids.includes(0)) {
                titles.push('ALL STORIES');
            }
            for (let story of stories) {
                if (ids.includes(story.id)) {
                    titles.push(story.title);
                }
            }

            drillInfoH.innerHTML = `<h3>${titles.join(', ')}</h3>${data.wordMatchCount} words<br>
            ${data.wordsOnCooldownCount} on cooldown<br>
            ${data.wordsZeroCountdown} with zero countdown`;
            displayWords();
        })
        .catch((error) => {
            console.error('Error:', error);
        });
}

drillCountInput.oninput = function (evt) {
    drillCountInputText.innerHTML = drillCountInput.value;
};

drillCountInput.onchange = newDrill;
drillTypeSelect.onchange = newDrill;
drillFilterSelect.onchange = newDrill;

function displayWords() {
    function wordInfo(word, idx, answered) {
        return `<div index="${idx}" class="drill_word ${word.wrong ? 'wrong' : ''} ${word.answered ? 'answered' : ''}">
                    <div class="base_form">${word.base_form}</div>
                    <div class="countdown">${word.countdown}</div>
                    <div class="countdown_max">${word.countdown_max}</div>
                </div>`;
    }

    html = `<h3 id="current_drill_count">${drillSet.length} words of ${drillSet.length + answeredSet.length}</h3>`;

    idx = 0;
    for (let word of drillSet) {
        if (!word.answered) {
            html += wordInfo(word, idx, false);
            idx++;
        }
    }

    for (let word of answeredSet) {
        if (word.answered) {
            html += wordInfo(word, idx, true);
            idx++;
        }
    }

    cardsDiv.innerHTML = html;

    if (drillSet[0]) {
        loadWordDefinition(drillSet[0])
    }
}

document.body.onkeydown = async function (evt) {
    if (evt.ctrlKey) {
        return;
    }
    if (drillSet && drillSet.length > 0) {
        var word = drillSet[0];
        //console.log(evt.code);
        if (evt.code === 'KeyS') {
            evt.preventDefault();
            // showWord();
        } else if (evt.code === 'KeyW') {  // dec countdown
            evt.preventDefault();
            if (drillSet && drillSet[0]) {
                var word = drillSet[0];
                word.countdown = 2;
                word.countdown_max = 2;
                updateWord(word);
                drillSet.shift();
                answeredSet.unshift(word);
                if (drillSet.length === 0) {
                    nextRound();
                }
                displayWords();
            }
        } else if (evt.code === 'KeyE') {  // dec countdown
            evt.preventDefault();
            if (drillSet && drillSet[0]) {
                var word = drillSet[0];
                word.countdown = 5;
                word.countdown_max = 5;
                updateWord(word);
                drillSet.shift();
                answeredSet.unshift(word);
                if (drillSet.length === 0) {
                    nextRound();
                }
                displayWords();
            }
        } else if (evt.code === 'KeyQ') {  // dec countdown
            evt.preventDefault();
            if (drillSet && drillSet[0]) {
                var word = drillSet[0];
                word.countdown = 0;
                word.countdown_max = 0;
                word.answered = true;
                updateWord(word);
                drillSet.shift();
                answeredSet.unshift(word);
                if (drillSet.length === 0) {
                    nextRound();
                }
                displayWords();
            }
        } else if ((evt.code === 'KeyR') && evt.altKey) {  // todo reset drill    
        } else if (evt.code === 'KeyA') {  // mark wrong and swap top two words
            evt.preventDefault();
            word.wrong = true;
            if (drillSet.length > 1) {
                [drillSet[0], drillSet[1]] = [drillSet[1], drillSet[0]];
            }
            let unixtime = Math.floor(Date.now() / 1000); // in seconds
            if ((unixtime - word.date_last_drill > COOLDOWN_TIME) &&
                (unixtime - word.date_last_wrong > COOLDOWN_TIME)) {
                word.date_last_wrong = unixtime;
                word.countdown--;
                word.drill_count++;
                updateWord(word);
            }
            displayWords();
        } else if (evt.code === 'KeyD') {  // mark answered
            evt.preventDefault();
            word.answered = true;
            let unixtime = Math.floor(Date.now() / 1000); // in seconds
            if ((unixtime - word.date_last_drill > COOLDOWN_TIME) &&
                (unixtime - word.date_last_wrong > COOLDOWN_TIME) &&
                word.countdown > 0) {
                word.date_last_drill = unixtime;
                word.countdown--;
                word.drill_count++;
                updateWord(word);
            }
            drillSet.shift();
            answeredSet.unshift(word);
            if (drillSet.length === 0) {
                nextRound();
            }
            displayWords();
        }
    }
};


cardsDiv.onclick = function (evt) {
    evt.preventDefault();
    var idx = parseInt(evt.target.getAttribute('index'));
    if (idx && idx < drillSet.length - 1) {
        console.log("clicked card", idx);
        var front = drillSet.slice(0, idx);
        var back = drillSet.slice(idx);
        drillSet = back;
        answeredSet = front.concat(answeredSet);
        for (let word of answeredSet) {
            word.answered = true;
        }
        displayWords();
    }
};

function nextRound() {
    let temp = [];
    for (let word of answeredSet) {
        if (word.wrong) {
            word.wrong = false;
            word.answered = false;
            drillSet.push(word);
        } else {
            temp.push(word);
        }
    }
    answeredSet = temp;

    if (drillSet.length === 0) {
        drillComlpeteDiv.style.display = 'block';
        return;
    }

    shuffle(drillSet);
}

function loadWordDefinition(word) {
    getKanji(word.base_form); // might as well get all possibly relevant kanji
    let defs = JSON.parse(word.definitions);
    if (defs && typeof defs === 'object') {
        html = '';
        for (let def of defs) {
            html += displayEntry(def);
        }
        definitionsDiv.innerHTML = html;
    }
}

function showWord() {
    kanjiResultsDiv.style.visibility = 'visible';
    definitionsDiv.style.visibility = 'visible';
}

document.body.onload = function (evt) {
    console.log('on page load');

    fetch('/stories_list', {
        method: 'GET', // or 'PUT'
        headers: {
            'Content-Type': 'application/json',
        }
    }).then((response) => response.json())
        .then((data) => {
            console.log('Stories list success:', data);

            storyIds = {};
            var url = new URL(window.location.href);
            if (url.searchParams && url.searchParams.has("storyId")) {
                let storyIdsList = url.searchParams.get("storyId").split(',');
                for (let idStr of storyIdsList) {
                    var id = parseInt(idStr.trim());
                    storyIds[id] = true;
                }
            }
            console.log('story ids', storyIds);

            drillCountInputText.innerHTML = drillCountInput.value;
            stories = data;
            newDrill();
        })
        .catch((error) => {
            console.error('Error:', error);
        });
};