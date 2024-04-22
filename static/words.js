var cardsDiv = document.getElementById('cards');
var answerDiv = document.getElementById('answer');
var drillTitleH = document.getElementById('drill_title');
var drillInfoH = document.getElementById('drill_info');
var categorySelect = document.getElementById('category');
var doneButton = document.getElementById('done_button');
var drillComlpeteDiv = document.getElementById('drill_complete');
var kanjiResultsDiv = document.getElementById('kanji_results');
var filterSelect = document.getElementById('filter_select')
var definitionsDiv = document.getElementById('definitions');
var statusSelect = document.getElementById('status_select');

const COOLDOWN_TIME = 60 * 60 * 3 // 3 hours (in seconds)

const DRILL_CATEGORY_KATAKANA = 1;
const DRILL_CATEGORY_ICHIDAN = 2;
const DRILL_CATEGORY_GODAN_SU = 8;
const DRILL_CATEGORY_GODAN_RU = 16;
const DRILL_CATEGORY_GODAN_U = 32;
const DRILL_CATEGORY_GODAN_TSU = 64;
const DRILL_CATEGORY_GODAN_KU = 128;
const DRILL_CATEGORY_GODAN_GU = 256;
const DRILL_CATEGORY_GODAN_MU = 512;
const DRILL_CATEGORY_GODAN_BU = 1024;
const DRILL_CATEGORY_GODAN_NU = 2048;
const DRILL_CATEGORY_KANJI = 4096;
const DRILL_CATEGORY_GODAN = DRILL_CATEGORY_GODAN_SU | DRILL_CATEGORY_GODAN_RU | DRILL_CATEGORY_GODAN_U | DRILL_CATEGORY_GODAN_TSU |
    DRILL_CATEGORY_GODAN_KU | DRILL_CATEGORY_GODAN_GU | DRILL_CATEGORY_GODAN_MU | DRILL_CATEGORY_GODAN_BU | DRILL_CATEGORY_GODAN_NU;
const DRILL_ALL = DRILL_CATEGORY_GODAN | DRILL_CATEGORY_ICHIDAN | DRILL_CATEGORY_KANJI | DRILL_CATEGORY_KATAKANA;

var drillSet = [];
var answeredSet = [];
var words;

statusSelect.onchange = function (evt) {
    newDrill();
};

function newDrill() {
    let includeCatalog = false;
    let includeInProgress = false;
    let includeBacklog = false;
    let includeArchived = false;
    for (let option of statusSelect.selectedOptions) {
        switch (option.value) {
            case 'catalog':
                includeCatalog = true;
                break;
            case 'in progress':
                includeInProgress = true;
                break;
            case 'backlog':
                includeBacklog = true;
                break;
            case 'archived':
                includeArchived = true;
                break;
        }
    }

    let includeOffCooldown = false;
    let includeOnCooldown = false;
    for (let option of filterSelect.selectedOptions) {
        switch (option.value) {
            case 'off':
                includeOffCooldown = true;
                break;
            case 'on':
                includeOnCooldown = true;
                break;
        }
    }

    let categoryMask = 0;
    let includeOther = false;
    for (let option of categorySelect.selectedOptions) {
        switch (option.value) {
            case 'kanji':
                categoryMask |= DRILL_CATEGORY_KANJI;
                break;
            case 'katakana':
                categoryMask |= DRILL_CATEGORY_KATAKANA;
                break;
            case 'godan':
                categoryMask |= DRILL_CATEGORY_GODAN;
                break;
            case 'ichidan':
                categoryMask |= DRILL_CATEGORY_ICHIDAN;
                break;
            case 'other':
                includeOther = true;
                break;
        }
    }

    let unixTime = Math.floor(Date.now() / 1000);

    let countOffCooldown = 0;

    drillSet = [];
    for (let word of words) {
        let offcooldown = (unixTime - word.date_marked) > COOLDOWN_TIME;
        if (offcooldown) {
            countOffCooldown++;
        }

        let isOther = (word.category & DRILL_ALL) == 0;
        let isCategoryMatch = (word.category & categoryMask) != 0;
        let categoryFilter = isCategoryMatch || (includeOther && isOther);

        let cooldownFilter = (includeOffCooldown && includeOnCooldown) ||
            (includeOffCooldown && offcooldown) ||
            (includeOnCooldown && !offcooldown);

        let statusFilter = (includeCatalog && word.status == 'catalog') ||
            (includeInProgress && word.status == 'in progress') ||
            (includeBacklog && word.status == 'backlog') ||
            (includeArchived && word.status == 'archived');

        if (!categoryFilter || !cooldownFilter || !statusFilter) {
            continue;
        }

        word.answered = false;
        drillSet.push(word);
    }

    drillComlpeteDiv.style.display = 'none';
    shuffle(drillSet);
    answeredSet = [];

    drillInfoH.innerHTML = `${words.length} words in story <br>
                        ${countOffCooldown} words off cooldown`;
    displayWords();
}

categorySelect.onchange = newDrill;
filterSelect.onchange = newDrill;

function displayWords() {
    function wordInfo(word, idx, answered) {
        return `<div index="${idx}" class="drill_word ${word.wrong ? 'wrong' : ''} ${word.answered ? 'answered' : ''}">
                    <div class="base_form">${word.base_form}</div>
                    <div class="rank rank${word.status}">${word.status}</div>
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
    if ((evt.code === 'KeyR') && evt.altKey) {
        console.log('alt r');
        newDrill();
    }
    if (drillSet && drillSet.length > 0) {
        var word = drillSet[0];
        let unixtime = Math.floor(Date.now() / 1000); // in seconds
        if (evt.code === 'KeyS') {
            evt.preventDefault();
            // showWord();
        } else if (evt.code === 'KeyA') {  // mark wrong and swap top two words
            evt.preventDefault();
            word.wrong = true;
            if (drillSet.length > 1) {
                [drillSet[0], drillSet[1]] = [drillSet[1], drillSet[0]];
            }
            displayWords();
        } else if (evt.code === 'KeyD') {  // mark answered
            evt.preventDefault();
            word.answered = true;
            if (unixtime - word.date_marked > COOLDOWN_TIME) {
                word.date_marked = unixtime;
                word.drill_countdown--;
                if (word.drill_countdown < 0) {
                    word.drill_countdown = 0;
                }
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
    let ele = evt.target.closest('.drill_word');
    var idx = parseInt(ele.getAttribute('index'));
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
    getKanji(word.base_form); // get all possibly relevant kanji

    html = '';
    if (word.definitions) {
        for (let def of word.definitions) {
            html += displayEntry(def);
        }
    }
    definitionsDiv.innerHTML = html;
}

function showWord() {
    kanjiResultsDiv.style.visibility = 'visible';
    definitionsDiv.style.visibility = 'visible';
}

document.body.onload = function (evt) {
    console.log('on page load');

    var url = new URL(window.location.href);
    let storyId = parseInt(url.searchParams.get("storyId"));
    let set = url.searchParams.get("set");

    fetch('words', {
        method: 'POST', // or 'PUT'
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({
            story_id: storyId,
            set: set,
        })
    }).then((response) => response.json())
        .then((data) => {
            words = data.words;
            for (w of words) {
                w.definitions = JSON.parse(w.definitions);
            }
            drillTitleH.innerHTML = `${data.story_source}<br><hr><a href="${data.story_link}">${data.story_title}</a>`;
            //`<a href="/story.html?storyId=${storyId}"> ${data.story_title}</a>`;
            newDrill();
        })
        .catch((error) => {
            console.error('Error:', error);
        });
};