var wordSearchText = document.getElementById('word_search');
var wordResultsDiv = document.getElementById('word_results');
var kanjiResultsDiv = document.getElementById('kanji_results');
var searchButton = document.getElementById('search_button');
var searchTypeButton = document.getElementById('search_type_button');

searchButton.onclick = function (evt) {
    let word = { word: wordSearchText.value };

    fetch('word_search', {
        method: 'POST', // or 'PUT'
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(word),
    }).then((response) => response.json())
        .then((data) => {
            console.log('Success:', data);
            displayWordResults(data);
            displayKanji(data.kanji, word.word);
        })
        .catch((error) => {
            console.error('Error:', error);
        });
};


searchTypeButton.onclick = function (evt) {
    let word = { word: wordSearchText.value };

    fetch('word_type_search', {
        method: 'POST', // or 'PUT'
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(word),
    }).then((response) => response.json())
        .then((data) => {
            console.log('Success:', data);
            displayTypeResults(data, word.word);
        })
        .catch((error) => {
            console.error('Error:', error);
        });
};

wordSearchText.onkeydown = function (evt) {
    if (evt.key !== "Enter") {
        return;
    }

    let word = { word: wordSearchText.value };

    fetch('word_search', {
        method: 'POST', // or 'PUT'
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(word),
    }).then((response) => response.json())
        .then((data) => {
            console.log('Success:', data);
            displayWordResults(data);
            displayKanji(data.kanji, word.word);
        })
        .catch((error) => {
            console.error('Error:', error);
        });
};

function displayTypeResults(results, typeString) {
    let reAllKana = /^[あ-んア-ン]+$/gm;
    let reKanjiThenKana = /^[\u4E00-\u9FAF\u3400-\u4dbf][あ-んア-ン]+$/gm;
    let count = 0;

    let uniqueReadings = new Set();
    let uniqueReadingsCount = 0;
    let html = '';
    for (let entry of results.entries) {
        
        let include = false;
        if (entry.kanji_spellings && entry.kanji_spellings.length > 0) {
            include = true;
            for (let kanji of entry.kanji_spellings) {
                if (!kanji.kanji_spelling.match(reKanjiThenKana)) {
                    include = false;
                }
            }
        } else {
            for (let reading of entry.readings) {
                if (reading.reading.match(reAllKana)) {
                    include = true;
                }
            }
        }
        if (include) {
            let readings = [];
            for (let reading of entry.readings) {
                if (reading.information && reading.information.includes('irregular')) {
                    continue;
                }
                readings.push(reading);
            }
            entry.readings = readings;

            let isUnique = true;
            for (let reading of readings) {
                if (reading.reading.match(reAllKana)) {
                    if (uniqueReadings.has(reading.reading)) {
                        isUnique = false;
                    }
                }
            }
            for (let reading of readings) {
                uniqueReadings.add(reading.reading);
            }
            if (isUnique) {
                uniqueReadingsCount++;
            }

            html += displayEntry(entry);
            count++;
        }        
    }
    console.log(`unique readings: ${uniqueReadingsCount}`, uniqueReadings);
    wordResultsDiv.innerHTML = `<h2>Found ${count} of type "${typeString}":` +  html;
}

function isIrregularReading(reading) {
    
}


function displayWordResults(results) {
    let html = `<h2>Search term found at start of ${results.count_start} word${results.count_start !== 1 ? 's' : ''}`;
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