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
            displayKanji(data.kanji);
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
            displayKanji(data.kanji);
        })
        .catch((error) => {
            console.error('Error:', error);
        });
};

function displayWordResults(results) {
    html = '';

    // let excessResults = results.count_start - results.entries_start.length;
    // excessResults += results.count_mid - results.entries_mid.length;
    // if (excessResults > 0) {
    //     html += `<h3><a>${excessResults} more results</a></h3>`;
    // }

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