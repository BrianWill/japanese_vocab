var newStoryText = document.getElementById('new_story_text');
var newStoryButton = document.getElementById('new_story_button');
var storyList = document.getElementById('story_list');
var newStoryTitle = document.getElementById('new_story_title');
var newStoryLink = document.getElementById('new_story_link');
var storyTitle = document.getElementById('story_title');
var wordList = document.getElementById('word_list');
var definitionsDiv = document.getElementById('definitions');
var kanjiResultsDiv = document.getElementById('kanji_results');

document.body.onload = function (evt) {
    getStoryList();

    if (window.location.pathname.startsWith('/read/')) {
        let storyId = window.location.pathname.substring(6);
        console.log("opening story with id " + storyId);
        openStory(storyId);
    }
};

newStoryButton.onclick = function (evt) {
    let data = {
        content: newStoryText.value,
        title: newStoryTitle.value,
        link: newStoryLink.value
    };

    newStoryText.value = '';
    newStoryTitle.value = '';
    newStoryLink.value = '';

    fetch('/create_story', {
        method: 'POST', // or 'PUT'
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(data),
    }).then((response) => response.json())
        .then((data) => {
            console.log('Success:', data);
            getStoryList();
        })
        .catch((error) => {
            console.error('Error:', error);
        });
};

function retokenizeStory(story) {
    fetch('/retokenize_story', {
        method: 'POST', // or 'PUT'
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({ id: story.id }),
    }).then((response) => response.json())
        .then((data) => {
            getStoryList();
            console.log('Success retokenizing:', data);
        })
        .catch((error) => {
            console.error('Error retokenizing:', error);
        });
}

function openStory(id) {
    fetch('/story/' + id, {
        method: 'GET', // or 'PUT'
        headers: {
            'Content-Type': 'application/json',
        }
    }).then((response) => response.json())
        .then((data) => {
            story = data;
            storyTitle.innerText = data.title;
            story.tokens = JSON.parse(story.tokens);
            console.log(`/story/${id} success:`, story);
            displayStory(data);
        })
        .catch((error) => {
            console.error('Error:', error);
        });
}


resetStoryWordCountdowns = function (storyId) {
    let data = { id: parseInt(storyId) };

    fetch('/story_reset_countdowns', {
        method: 'POST', // or 'PUT'
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(data),
    }).then((response) => response.json())
        .then((data) => {
            console.log('Success:', data);
            getStoryList();
        })
        .catch((error) => {
            console.error('Error:', error);
        });
};


storyList.onclick = function (evt) {
    if (evt.target.tagName == 'A') {
        var storyId = evt.target.getAttribute('story_id');
        let story = storiesById[storyId];

        var action = evt.target.getAttribute('action');
        switch (action) {
            case 'drill':
                break;
            case 'reset_story_word_countdowns':
                evt.preventDefault();
                resetStoryWordCountdowns(storyId);
                break;
            case 'drill_in_progress':
                break;
            case 'retokenize':
                console.log('story to retokenize', storiesById[storyId]);
                retokenizeStory(story);
                break;
            default:
                break;
        }
    }
};

storyList.onchange = function (evt) {
    var storyId = evt.target.getAttribute('story_id');
    let story = storiesById[storyId];
    story.status = parseInt(evt.target.value);
    console.log("changed", parseInt(evt.target.value));
    updateStory(story, true);
};

