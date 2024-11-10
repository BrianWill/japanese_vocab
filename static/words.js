var cardsDiv = document.getElementById('cards');
var answerDiv = document.getElementById('answer');
var drillTitleH = document.getElementById('drill_title');
var drillInfoH = document.getElementById('drill_info');
var categorySelect = document.getElementById('category');
var doneButton = document.getElementById('done_button');
var drillComlpeteDiv = document.getElementById('drill_complete');
var kanjiResultsDiv = document.getElementById('kanji_results');
var definitionsDiv = document.getElementById('definitions');
var archivedSelect = document.getElementById('archived_select');
var logLink = document.getElementById('log_link');

var drillSet = [];
var answeredSet = [];
var words;

var story;

archivedSelect.onchange = function (evt) {
    newDrill();
};

function newDrill() {
    let includeNotArchived = false;
    let includeArchived = false;
    for (let option of archivedSelect.selectedOptions) {
        switch (option.value) {
            case 'unarchived':
                includeNotArchived = true;
                break;
            case 'archived':
                includeArchived = true;
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

    drillSet = [];
    for (let word of words) {
        let isOther = (word.category & DRILL_ALL) == 0;
        let isCategoryMatch = (word.category & categoryMask) != 0;
        let categoryFilter = isCategoryMatch || (includeOther && isOther);

        let statusFilter = (includeNotArchived && word.archived == 0) ||
            (includeArchived && word.archived == 1);

        if (!categoryFilter || !statusFilter) {
            continue;
        }

        word.answered = false;
        drillSet.push(word);
    }

    drillComlpeteDiv.style.display = 'none';
    shuffle(drillSet);
    answeredSet = [];

    drillInfoH.innerHTML = `${words.length} words in story`;
    displayWords();
}

categorySelect.onchange = newDrill;

function displayWords() {
    function wordInfo(word, idx, answered) {
        let archived = word.archived ? 'archived' : '';
        return `<div index="${idx}" class="drill_word ${word.wrong ? 'wrong' : ''} ${word.answered ? 'answered' : ''}">
                    <div class="base_form">${word.base_form}</div>
                    <div class="rank ">
                        ${archived}<br>
                        ${word.repetitions} reps
                    </div>
                </div>`;
    }

    html = `<h3 id="current_repetitions">${drillSet.length} words of ${drillSet.length + answeredSet.length}</h3>`;

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
        if (evt.code === 'KeyA') {  // mark wrong and swap top two words
            evt.preventDefault();
            word.wrong = true;
            if (drillSet.length > 1) {
                [drillSet[0], drillSet[1]] = [drillSet[1], drillSet[0]];
            }
            displayWords();
        } else if (evt.code === 'KeyD') {  // mark answered
            evt.preventDefault();
            word.answered = true;
            drillSet.shift();
            answeredSet.unshift(word);
            if (drillSet.length === 0) {
                nextRound();
            }
            displayWords();
        } else if (evt.code === 'Digit1') {
            evt.preventDefault();
            if (word.archived == 0) {
                word.archived = 1;
            } else if (word.archived == 1) {
                word.archived = 0;
            }
            updateWord(word);
            displayWords();
        }
    }
};

cardsDiv.onclick = function (evt) {
    evt.preventDefault();
    let ele = evt.target.closest('.drill_word');
    if (!ele) {
        return;
    }
    var idx = parseInt(ele.getAttribute('index'));
    if (idx && idx < drillSet.length) {
        //console.log("clicked card", idx); 
        var front = drillSet.slice(0, idx);
        var back = drillSet.slice(idx + 1);
        drillSet = [drillSet[idx]].concat(front, back);
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

logLink.onclick = function (evt) {
    evt.preventDefault();

    let unixTime = Math.floor(Date.now() / 1000);

    let wordIds = [];
    for (const word of answeredSet) {
        if ((unixTime - word.date_last_rep) > REP_COOLDOWN) {
            wordIds.push(word.id);
        }
    }

    if (wordIds.length== 0) {
        snackbarMessage(`no reps logged: all answered words are on cooldown`);
        return;
    }

    incWords(wordIds, function () {
        load();
        snackbarMessage(`reps logged for ${wordIds.length} words`);
    });    
}

function load(evt) {
    console.log('on page load');

    var url = new URL(window.location.href);
    let storyId = parseInt(url.searchParams.get("storyId"));

    fetch('words', {
        method: 'POST', // or 'PUT'
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({
            story_id: storyId,
        })
    }).then((response) => response.json())
        .then((data) => {
            words = data.words;
            console.log(words);
            for (w of words) {
                w.definitions = JSON.parse(w.definitions);
            }

            drillTitleH.innerHTML = `
                <a href="story.html?storyId=${storyId}">${data.story_title}</a>
                <br><hr>${data.story_source}<br>`;
            newDrill();
        })
        .catch((error) => {
            console.error('Error:', error);
        });

    fetch('/story/' + storyId, {
        method: 'GET', // or 'PUT'
        headers: {
            'Content-Type': 'application/json',
        }
    }).then((response) => response.json())
        .then((data) => {
            story = data;
        });
}

document.body.onload = function (evt) {
    load(evt);
};