var cardsDiv = document.getElementById('cards');
var answerDiv = document.getElementById('answer');
var drillButton = document.getElementById('drill_button');
var doneButton = document.getElementById('done_button');
var drillComlpeteDiv = document.getElementById('drill_complete');
var kanjiResultsDiv = document.getElementById('kanji_results');
var definitionsDiv = document.getElementById('definitions');

const COOLDOWN_TIME = 60 * 60 * 3 // number of seconds

var drillSet = null;
var answeredSet = [];

drillButton.onclick = function (evt) {
    fetch('drill', {
        method: 'GET', // or 'PUT'
        headers: {
            'Content-Type': 'application/json',
        }
    }).then((response) => response.json())
        .then((data) => {
            console.log('Success:', data);
            shuffle(data);
            drillSet = data;
            answeredSet = [];
            displayWords();
        })
        .catch((error) => {
            console.error('Error:', error);
        });
};

function displayWords() {
    html = '';

    for (let word of drillSet) {
        if (!word.answered) {
            html += `<div class="drill_word ${word.wrong ? 'wrong': ''}">${word.base_form}</div>`;
        }        
    }
    
    for (let word of answeredSet) {
        if (word.answered) {
            html += `<div class="drill_word ${word.wrong ? 'wrong': ''} answered">${word.base_form}</div>`;
        }
    }
    
    cardsDiv.innerHTML = html;
    kanjiResultsDiv.innerHTML = '';
    definitionsDiv.innerHTML = '';
}

document.body.onkeydown = async function (evt) {
    if (drillSet && drillSet.length > 0) {
        var word = drillSet[0];
        if (evt.code === 'KeyS') {
            evt.preventDefault();
            showWord(word);
        } else if ((evt.code === 'KeyR') && evt.altKey) {  // todo reset drill    
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
            let unixtime = Math.floor(Date.now() / 1000); // in seconds
            if ((unixtime - word.date_last_drill > COOLDOWN_TIME) && word.countdown > 0) {
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

function showWord(word) {
    getKanji(word.base_form); // might as well get all possibly relevant kanji
    let defs = JSON.parse(word.definitions);
    if (defs && typeof defs === 'object') {
        html = '';
        for (let def of defs) {
            console.log(def);
            html += displayEntry(def);
        }
        definitionsDiv.innerHTML = html;
    }       
}

document.body.onload = function (evt) {
    console.log('on page load');
    drillButton.onclick();
};