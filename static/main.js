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
var queuedStoriesDiv = document.getElementById('queued_stories');
var balanceQueueLink = document.getElementById('balance_queue_link');

document.body.onload = function (evt) {
    getStoryList();
    getLogEvents();
    getQueuedStories();

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
                markQueuedStory(logId, storyId);
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
            console.log('Rebalanced queue', data);
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
            console.log('Added a log event');
            getLogEvents();
        })
        .catch((error) => {
            console.error('Error:', error);
        });
}


function removeQueuedStory(logId) {
    fetch(`/remove_queued_story/${logId}`, {
        method: 'GET',
        headers: {
            'Content-Type': 'application/json',
        }
    }).then((response) => response.json())
        .then((data) => {
            console.log('remove a queued story');
            getQueuedStories();
        })
        .catch((error) => {
            console.error('Error:', error);
        });
}

function scheduleStory(storyId) {
    fetch(`/schedule_story/${storyId}`, {
        method: 'GET',
        headers: {
            'Content-Type': 'application/json',
        }
    }).then((response) => response.json())
        .then((data) => {
            console.log('scechduled a story');
            getQueuedStories();
        })
        .catch((error) => {
            console.error('Error:', error);
        });
}

function markQueuedStory(logId, storyId) {
    fetch(`/mark_queued_story/${logId}/${storyId}`, {
        method: 'GET',
        headers: {
            'Content-Type': 'application/json',
        }
    }).then((response) => response.json())
        .then((data) => {
            console.log('moved queud story entry to the log');
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

function displayQueuedStories(queuedStories) {
    let html = `<table class="log_table">`;
    
    for (let qs of queuedStories) {
        html += `<tr>
            <td><a action="remove" log_id="${qs.id}" href="#">remove</a></td>
            <td><a action="log" log_id="${qs.id}" story_id="${qs.story_id}"  href="#">log</a></td>
            <td>in ${qs.days_from_now} days</td>
            <td><a story_id="${qs.story_id}" href="/story.html?storyId=${qs.story_id}">${qs.title}</a></td>
            </tr>`;        
    }

    queuedStoriesDiv.innerHTML = html + '</table>';
}

function updateLogEvents(data) {
    let html = `<table class="log_table">`;
    
    for (let d of data.logEvents) {
        html += `<tr>
            <td><a action="remove" log_id="${d.id}" href="#">remove log entry</a></td>
            <td>${timeSince(d.date)}</td>
            <td><a story_id="${d.story_id}" href="/story.html?storyId=${d.story_id}">${d.title}</a></td>
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

function getQueuedStories(id) {
    fetch('/get_enqueued_stories', {
        method: 'GET', // or 'PUT'
        headers: {
            'Content-Type': 'application/json',
        }
    }).then((response) => response.json())
        .then((data) => {
            console.log(`success retrieving enqueued stories`, data);
            displayQueuedStories(data);
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

storyList.onclick = function (evt) {
    if (evt.target.tagName == 'A') {
        var storyId = evt.target.getAttribute('story_id');
        var action = evt.target.getAttribute('action');
        switch (action) {
            case 'schedule':
                evt.preventDefault();
                scheduleStory(storyId);
                break;
        }
    }
};

