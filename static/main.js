var storyTitle = document.getElementById('story_title');
var wordList = document.getElementById('word_list');
var definitionsDiv = document.getElementById('definitions');
var kanjiResultsDiv = document.getElementById('kanji_results');
var logEvents = document.getElementById('log_events');
var queuedStoriesDiv = document.getElementById('queued_stories');
var balanceQueueLink = document.getElementById('balance_queue_link');

document.body.onload = function (evt) {
    getLogEvents();
    getQueuedStories();
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

queuedStoriesDiv.onclick = function(evt) {
    if (evt.target.tagName == 'A') {
        var logId = evt.target.getAttribute('log_id');
        var storyId = evt.target.getAttribute('story_id');
        var action = evt.target.getAttribute('action');
        switch (action) {
            case 'remove':
                evt.preventDefault();
                removeQueuedStory(logId);
                break;
            case 'log':
                evt.preventDefault();
                logScheduledStory(logId, storyId);
                break;
        }
    }
};

balanceQueueLink.onclick = function(evt) {
    fetch(`/balance_queue`, {
        method: 'GET',
        headers: {
            'Content-Type': 'application/json',
        }
    }).then((response) => response.json())
        .then((data) => {
            getQueuedStories();
        })
        .catch((error) => {
            console.error('Error:', error);
        });
};

function removeLogEvent(logId) {
    fetch(`/remove_log_event/${logId}`, {
        method: 'GET',
        headers: {
            'Content-Type': 'application/json',
        }
    }).then((response) => response.json())
        .then((data) => {
            getLogEvents();
        })
        .catch((error) => {
            console.error('Error:', error);
        });
}


function removeQueuedStory(logId) {
    fetch(`/remove_scheduled_story/${logId}`, {
        method: 'GET',
        headers: {
            'Content-Type': 'application/json',
        }
    }).then((response) => response.json())
        .then((data) => {
            getQueuedStories();
        })
        .catch((error) => {
            console.error('Error:', error);
        });
}


function logScheduledStory(logId, storyId) {
    fetch(`/log_scheduled_story/${logId}/${storyId}`, {
        method: 'GET',
        headers: {
            'Content-Type': 'application/json',
        }
    }).then((response) => response.json())
        .then((data) => {
            getQueuedStories();
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

function displayStorySchedule(scheduledStories) {
    let html = `<table class="log_table">`;
    
    for (let qs of scheduledStories) {
        html += `<tr>
            <td><a action="remove" log_id="${qs.id}" href="#">remove</a></td>
            <td><a action="log" log_id="${qs.id}" story_id="${qs.story_id}"  href="#">log</a></td>
            <td>${timeSince(qs.date)}</td>
            <td><a story_id="${qs.story_id}" href="/story.html?storyId=${qs.story_id}">${qs.title}</a></td>
            </tr>`;        
    }

    queuedStoriesDiv.innerHTML = html + '</table>';
}

function updateLogEvents(data) {
    let html = `<table class="log_table">`;
    
    for (let d of data.logEvents) {
        html += `<tr>
            <td><a action="remove" log_id="${d.id}" href="#">remove</a></td>
            <td>${timeSince(d.date)}</td>
            <td><a story_id="${d.story_id}" href="/story.html?storyId=${d.story_id}">${d.title}</a></td>
            </tr>`;        
    }

    logEvents.innerHTML = html + '</table>';
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
            displayStory(data);
        })
        .catch((error) => {
            console.error('Error:', error);
        });
}

