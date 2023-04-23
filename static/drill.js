var cardsDiv = document.getElementById('cards');
var answerDiv = document.getElementById('answer');
var drillButton = document.getElementById('drill_button');

drillButton.onclick = function (evt) {
    fetch('drill', {
        method: 'GET', // or 'PUT'
        headers: {
            'Content-Type': 'application/json',
        }
    }).then((response) => response.json())
        .then((data) => {
            console.log('Success:', data);
            displayWords(data);
            //displayKanji(data.kanji);
        })
        .catch((error) => {
            console.error('Error:', error);
        });
};

function displayWords(words) {
    html = '';

    for (let word of words) {
        html += `<div class="drill_word">${word.base_form}</div>`;
    }    
    
    cardsDiv.innerHTML = html;
}



document.body.onload = function (evt) {
    console.log('on page load');
};