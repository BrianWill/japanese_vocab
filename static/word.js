var wordSearchText = document.getElementById('word_search');
var wordResultsDiv = document.getElementById('word_results');
var kanjiResultsDiv = document.getElementById('kanji_results');
var searchButton = document.getElementById('search_button');

searchButton.onclick = function (evt) {
    let data = { word: wordSearchText.value };

    fetch('word_search', {
        method: 'POST', // or 'PUT'
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(data),
    }).then((response) => response.json())
        .then((data) => {
            console.log('Success:', data);
            displayWordResults(data);
            displayKanjiResults(data);
        })
        .catch((error) => {
            console.error('Error:', error);
        });
};

wordSearchText.onkeydown = function (evt) {
    if (evt.key !== "Enter") {
        return;
    }

    let data = { word: wordSearchText.value };

    fetch('word_search', {
        method: 'POST', // or 'PUT'
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(data),
    }).then((response) => response.json())
        .then((data) => {
            console.log('Success:', data);
            displayWordResults(data);
            displayKanjiResults(data);
        })
        .catch((error) => {
            console.error('Error:', error);
        });
};

function displayKanjiResults(results) {
    html = '';

    if (!results.kanji || results.kanji.length === 0) {
        kanjiResultsDiv.innerHTML = `<h2>No kanji in search term.</h2>`;
        return;
    }

    for (let k of results.kanji) {       
        for (let group of k.readingmeaning.group) {
            onyomi = group.reading.filter(x => x.type === 'ja_on').map(x => `<span class="kanji_reading">${x.value}</span>`);
            kunyomi = group.reading.filter(x => x.type === 'ja_kun').map(x => `<span class="kanji_reading">${x.value}</span>`);

            var meanings = group.meaning.filter(x => !x.language).map(x => x.value);

            var misc = `strokes: ${k.misc.stroke_count}<span class="misc_filler">&nbsp;&nbsp;&nbsp;&nbsp;</span>frequency: ${k.misc.frequency}`;

            html += `<div class="kanji">
                    <div>
                    <span class="literal">${k.literal}</span>
                    <span class="onyomi_readings">${onyomi.join('')}</span>
                    <span class="kunyomi_readings">${kunyomi.join('')}</span>
                    </div>
                    <div class="kanji_meanings">${meanings.join(';  ')}</div>
                    <div class="kanji_misc">${misc}</div>
                    </div>`;
        }
    }

    kanjiResultsDiv.innerHTML = html;
}

function displayWordResults(results) {
    html = '';

    let excessResults = results.count_start - results.entries_start.length;
    excessResults += results.count_mid - results.entries_mid.length;
    if (excessResults > 0) {
        html += `<h3><a>${excessResults} more results</a></h3>`;
    }

    html += `<h2>Search term found at start of ${results.count_start} word${results.count_start !== 1 ? 's' : ''}`;
    if (results.count_start !== results.entries_start.length) {
        html += ` (displaying ${results.entries_start.length})</h2>`;
    } else {
        html += `</h2>`;
    }
    html += `<h2>Search term found in interior of ${results.count_mid} word${results.count_mid > 1 ? 's' : ''} <a id="interior_link" href="#interior_results">▾</a>`;
    if (results.count_mid !== results.entries_mid.length) {
        html += ` (displaying ${results.entries_mid.length})</h2>`;
    } else {
        html += `</h2>`;
    }

    for (var entry of results.entries_start) {
        html += displayEntry(entry);
    }

    if (results.count_mid > 0) {
        html += `<h2 id="interior_results">Search term found in the interior of the word:</h2>`;
    }

    for (var entry of results.entries_mid) {
        html += displayEntry(entry);
    }

    wordResultsDiv.innerHTML = html;
}

function displayEntry(entry) {
    html = `<div class="entry"><div class="word">`;

    html += '<div class="readings">'
    for (var reading of entry.readings || []) {
        html += `<div class="reading">${reading.reading}</div>`;
    }

    html += '</div><div class="kanji_spellings">'
    for (var kanji of entry.kanji_spellings || []) {
        html += `<div class="kanji_spelling">${kanji.kanji_spelling}</div>`;
    }

    html += `</div></div><div class="senses">`;

    for (var sense of entry.senses || []) {
        html += `<div class="sense">
            <span class="pos">${sense.parts_of_speech.join(', ')}</span>
            <span class="glosses">${sense.glosses.map(x => x.value).join('; ')}</span>
        </div>`;
    }

    return html + `</div></div><hr>`;

}

document.body.onload = function (evt) {
    console.log('on page load');

    // fetch('stories_list', {
    //     method: 'GET', // or 'PUT'
    //     headers: {
    //         'Content-Type': 'application/json',
    //     }
    // }).then((response) => response.json())
    //     .then((data) => {
    //         console.log('Success:', data);
    //         updateStoryList(data);
    //     })
    //     .catch((error) => {
    //         console.error('Error:', error);
    //     });
};