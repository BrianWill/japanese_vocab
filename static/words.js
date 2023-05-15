var cardsDiv = document.getElementById('cards');
var answerDiv = document.getElementById('answer');
var drillInfoH = document.getElementById('drill_info');
var drillCountInput = document.getElementById('drill_count');
var drillCountInputText = document.getElementById('drill_count_text');
var drillTypeSelect = document.getElementById('drill_type');
var drillWrongSelect = document.getElementById('drill_wrong');
var drillStorySelect = document.getElementById('drill_story');
var doneButton = document.getElementById('done_button');
var drillComlpeteDiv = document.getElementById('drill_complete');
var kanjiResultsDiv = document.getElementById('kanji_results');
var definitionsDiv = document.getElementById('definitions');
var ignoreCountdownCheckbox = document.getElementById('ignore_countdown_checkbox');

const COOLDOWN_TIME = 60 * 60 * 3 // number of seconds

var drillSet = null;
var answeredSet = [];

function newDrill() {
    var storyIds = [];
    for (let option of drillStorySelect.options) {
        if (option.selected) {
            storyIds.push(parseInt(option.value))
        }
    }
    console.log('storyIds: ', storyIds);

    fetch('words', {
        method: 'POST', // or 'PUT'
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({
            count: parseInt(drillCountInput.value),
            drill_type: drillTypeSelect.value,
            wrong: parseInt(drillWrongSelect.value),
            ignore_cooldown: ignoreCountdownCheckbox.checked,
            storyIds: storyIds
        })
    }).then((response) => response.json())
        .then((data) => {
            console.log('Drill success:', data);
            drillComlpeteDiv.style.display = 'none';
            shuffle(data.words);
            drillSet = data.words;
            answeredSet = [];
            drillInfoH.innerHTML = `${data.wordAllCount} words (${data.wordOffCooldownCount} 
                    active words off cooldown); ${data.wordMatchCount} words matching filter`;
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
drillWrongSelect.onchange = newDrill;
ignoreCountdownCheckbox.onchange = newDrill;
drillStorySelect.onchange = newDrill;

function displayWords() {
    function wordInfo(word, answered) {
        return `<div class="drill_word ${word.wrong ? 'wrong' : ''} ${word.answered ? 'answered' : ''}">
                    <div class="base_form">${word.base_form}</div>
                    <div class="countdown">${word.countdown}</div>
                    <div class="drill_count">${word.drill_count}</div>
                </div>`;
    }

    html = `<h3 id="current_drill_count">${drillSet.length} words of ${drillSet.length + answeredSet.length}</h3>`;

    for (let word of drillSet) {
        if (!word.answered) {
            html += wordInfo(word, false);
        }
    }

    for (let word of answeredSet) {
        if (word.answered) {
            html += wordInfo(word, true);
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
        } else if (evt.code === 'Minus') {  // dec countdown
            evt.preventDefault();
            if (drillSet && drillSet[0]) {
                var word = drillSet[0];
                word.countdown--;
                updateWord(word);
                displayWords();
            }
        } else if (evt.code === 'Equal') {  // inc countdown
            evt.preventDefault();
            if (drillSet && drillSet[0]) {
                var word = drillSet[0];
                word.countdown++;
                updateWord(word);
                displayWords();
            }
        } else if (evt.code === 'Backspace') {  // set countdown to 0
            evt.preventDefault();
            if (drillSet && drillSet[0]) {
                var word = drillSet[0];
                word.countdown = 0;
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

function updateStoryDrillList(stories, storyId) {
    html = `<option value="0">all stories</option>
            <option value="-1" ${storyId == -1 ? 'selected' : ''}>all stories in progress</option>`;
    for (let story of stories) {
        if (storyId === story.id) {
            html += `<option value="${story.id}" selected>${story.title}</option>`;
        } else {
            html += `<option value="${story.id}">${story.title}</option>`;
        }
    }
    drillStorySelect.innerHTML = html;
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

            var url = new URL(window.location.href);
            var storyId = parseInt(url.searchParams.get("storyId") || undefined);
            updateStoryDrillList(data, storyId);

            drillCountInputText.innerHTML = drillCountInput.value;
            newDrill();
        })
        .catch((error) => {
            console.error('Error:', error);
        });
};