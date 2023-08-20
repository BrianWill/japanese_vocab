var newStoryText = document.getElementById('new_story_text');
var newStoryButton = document.getElementById('new_story_button');
var storyList = document.getElementById('story_list');
var newStoryTitle = document.getElementById('new_story_title');
var newStoryLink = document.getElementById('new_story_link');
var storyTitle = document.getElementById('story_title');
var wordList = document.getElementById('word_list');
var definitionsDiv = document.getElementById('definitions');
var kanjiResultsDiv = document.getElementById('kanji_results');
var logEvents = document.getElementById('log_events');

document.body.onload = function (evt) {
    getStoryList();

    if (window.location.pathname.startsWith('/read/')) {
        let storyId = window.location.pathname.substring(6);
        console.log("opening story with id " + storyId);
        openStory(storyId);
    }
};

logEvents.onclick = function(evt) {
    if (evt.target.tagName == 'A') {
        var logId = evt.target.getAttribute('log_id');
        var action = evt.target.getAttribute('action');
        switch (action) {
            case 'remove':
                evt.preventDefault();
                removeLogEvent(logId);
                break;
        }
    }
};

function removeLogEvent(logId) {
    fetch(`/remove_log_event/${logId}`, {
        method: 'GET',
        headers: {
            'Content-Type': 'application/json',
        }
    }).then((response) => response.json())
        .then((data) => {
            console.log('Added a log event');
            getLogEvents();
        })
        .catch((error) => {
            console.error('Error:', error);
        });
}

function getLogEvents() {
    const WEEK_IN_SECONDS = 60 * 60 * 24 * 7;
    var timeOfLastWeek = new Date() - WEEK_IN_SECONDS;

    fetch('/log_events/' + timeOfLastWeek, {
        method: 'GET',
        headers: {
            'Content-Type': 'application/json',
        }
    }).then((response) => response.json())
        .then((data) => {
            updateLogEvents(data);
        })
        .catch((error) => {
            console.error('Error:', error);
        });
}

function updateLogEvents(data) {
    let html = `<table class="log_table">`;
    var storyLookup = storiesById || {};

    
    for (let d of data.logEvents) {
        html += `<tr>
            <td><a action="remove" log_id="${d.id}" href="#">remove log entry</a></td>
            <td>${timeSince(d.date)}</td>
            <td><a story_id="${d.story_id}" href="/story.html?storyId=${d.story_id}">${storyLookup[d.story_id].title}</a></td>
            </tr>`;        
    }

    logEvents.innerHTML = html + '</table>';
}

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

storyList.onchange = function (evt) {
    var storyId = evt.target.getAttribute('story_id');
    let story = storiesById[storyId];
    story.status = parseInt(evt.target.value);
    console.log("changed", parseInt(evt.target.value));
    updateStory(story, true);
};

