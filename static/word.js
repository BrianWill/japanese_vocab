var wordSearchText = document.getElementById('word_search');
var wordResultsDiv = document.getElementById('word_results');
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
        })
        .catch((error) => {
            console.error('Error:', error);
        });
};

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