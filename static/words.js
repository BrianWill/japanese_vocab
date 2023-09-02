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
var rankSlider = document.getElementById('rank_slider');

const COOLDOWN_TIME = 60 * 60 * 3 // number of seconds

var drillSet = null;
var answeredSet = [];
var stories;
var words;
var wordInfoMap;

function newDrill() {
    let includeOffCooldown = true;
    let includeOnCooldown = true;
    switch (filterSelect.value) {
        case 'off':
            includeOnCooldown = false;
            break;
        case 'on':
            includeOffCooldown = false;
            break;
    }

    let categoryMask = getCategoryMask(categorySelect.value);
    let [minRank, maxRank] = rankSlider.noUiSlider.get();
    let unixTime = Math.floor(Date.now() / 1000);

    let countsByRank = [0, 0, 0, 0, 0];
    let onCooldownCountsByRank = [0, 0, 0, 0, 0];

    drillSet = [];
    for (let word of words) {
        let wordInfo = wordInfoMap[word.base_form];
        let offcooldown = (unixTime - wordInfo.date_marked) > cooldownsByRank[wordInfo.rank];
        countsByRank[wordInfo.rank]++;
        if (offcooldown) {
            onCooldownCountsByRank[wordInfo.rank]++;
        }

        // filter
        if (categorySelect.value === 'all' || (word.category & categoryMask) != 0) {
            let wordInfo = wordInfoMap[word.base_form];
            if (wordInfo.rank >= minRank && wordInfo.rank <= maxRank) {
                if ((includeOffCooldown && includeOnCooldown) ||
                    (includeOffCooldown && offcooldown) ||
                    (includeOnCooldown && !offcooldown)
                ) {
                    word.answered = false;
                    drillSet.push(word);
                }
            }
        }
    }

    drillComlpeteDiv.style.display = 'none';
    shuffle(drillSet);
    answeredSet = [];

    drillInfoH.innerHTML = `
                        ${words.length} words in story &nbsp;&nbsp;&nbsp;
                        <span class="rank_number">Rank 1:</span> ${countsByRank[1]} words <span class="cooldown">(${onCooldownCountsByRank[1]} off cooldown)</span> &nbsp;&nbsp;&nbsp;
                        <span class="rank_number">Rank 2:</span> ${countsByRank[2]} words <span class="cooldown">(${onCooldownCountsByRank[2]} off cooldown)</span> &nbsp;&nbsp;&nbsp;
                        <span class="rank_number">Rank 3:</span> ${countsByRank[3]} words <span class="cooldown">(${onCooldownCountsByRank[3]} off cooldown)</span> &nbsp;&nbsp;&nbsp;
                        <span class="rank_number">Rank 4:</span> ${countsByRank[4]} words <span class="cooldown">(${onCooldownCountsByRank[4]} off cooldown)</span>`;
    displayWords();
}

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

function getCategoryMask(category) {
    switch (category) {
        case 'kanji':
            return DRILL_CATEGORY_KANJI;
        case 'katakana':
            return DRILL_CATEGORY_KATAKANA;
        case 'godan':
            return DRILL_CATEGORY_GODAN;
        case 'ichidan':
            return DRILL_CATEGORY_ICHIDAN;
    }
    return -1;
}

categorySelect.onchange = newDrill;
filterSelect.onchange = newDrill;

function displayWords() {
    function wordInfo(word, idx, answered) {
        return `<div index="${idx}" class="drill_word ${word.wrong ? 'wrong' : ''} ${word.answered ? 'answered' : ''}">
                    <div class="base_form">${word.base_form}</div>
                    <div class="rank rank${word.rank}"><span>rank</span> ${word.rank}</div>
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
        loadWordDefinition(drillSet[0].base_form)
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
        //console.log(evt.code);
        if (evt.code === 'KeyS') {
            evt.preventDefault();
            // showWord();
        } else if (evt.code === 'Digit1') {
            evt.preventDefault();
            if (drillSet && drillSet[0]) {
                var word = drillSet[0];
                word.rank = 1;
                updateWord(word, wordInfoMap);
                displayWords();
            }
        } else if (evt.code === 'Digit2') {
            evt.preventDefault();
            if (drillSet && drillSet[0]) {
                var word = drillSet[0];
                word.rank = 2;
                updateWord(word, wordInfoMap);
                displayWords();
            }
        } else if (evt.code === 'Digit3') {
            evt.preventDefault();
            if (drillSet && drillSet[0]) {
                var word = drillSet[0];
                word.rank = 3;
                updateWord(word, wordInfoMap);
                displayWords();
            }
        } else if (evt.code === 'Digit4') {
            evt.preventDefault();
            if (drillSet && drillSet[0]) {
                var word = drillSet[0];
                word.rank = 4;
                updateWord(word, wordInfoMap);
                displayWords();
            }
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
                word.drill_count++;
                updateWord(word, wordInfoMap);
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

function loadWordDefinition(baseForm) {
    getKanji(baseForm); // get all possibly relevant kanji

    let wordInfo = wordInfoMap[baseForm];
    if (wordInfo) {
        let defs = wordInfo.definitions;
        html = '';
        if (defs) {
            for (let def of defs) {
                html += displayEntry(def);
            }
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

    noUiSlider.create(rankSlider, {
        start: [1, 4],
        step: 1,
        connect: true,
        pips: {
            mode: 'count',
            values: 4,
            density: 1,
            stepped: true
        },
        range: {
            'min': 1,
            'max': 4
        },
        format: {
            // 'to' the formatted value. Receives a number.
            to: function (value) {
                return value;
            },
            // 'from' the formatted value.
            // Receives a string, should return a number.
            from: function (value) {
                return value;
            }
        }
    });

    function sliderUpdate(values, handle, unencoded, tap, positions, noUiSlider) {
        newDrill();
    }

    fetch('/stories_list', {
        method: 'GET', // or 'PUT'
        headers: {
            'Content-Type': 'application/json',
        }
    }).then((response) => response.json())
        .then((data) => {
            //console.log('Stories list success:', data);
            stories = data;

            storyIds = {};
            var url = new URL(window.location.href);
            if (url.searchParams && url.searchParams.has("storyId")) {
                let storyIdsList = url.searchParams.get("storyId").split(',');
                for (let idStr of storyIdsList) {
                    var id = parseInt(idStr.trim());
                    storyIds[id] = true;
                }
            }

            var titles = [];

            let ids = Object.keys(storyIds).map(x => parseInt(x));
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

            drillTitleH.innerHTML = `<h3>${titles.join(', ')}</h3>`;

            fetch('words', {
                method: 'POST', // or 'PUT'
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    story_ids: ids,
                })
            }).then((response) => response.json())
                .then((data) => {
                    //console.log('Words retrieved:', data);
                    words = data.words;
                    wordInfoMap = data.wordInfoMap;

                    rankSlider.noUiSlider.on('update', sliderUpdate);  // calls newDrill upon registration
                })
                .catch((error) => {
                    console.error('Error:', error);
                });
        })
        .catch((error) => {
            console.error('Error:', error);
        });
};