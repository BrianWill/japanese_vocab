var newStoryText = document.getElementById('new_story_text');
var newStoryButton = document.getElementById('new_story_button');
var storyList = document.getElementById('story_list');
var newStoryTitle = document.getElementById('new_story_title');
var newStoryLink = document.getElementById('new_story_link');
var storyTitle = document.getElementById('story_title');
var wordList = document.getElementById('word_list');
var definitionsDiv = document.getElementById('definitions');
var kanjiResultsDiv = document.getElementById('kanji_results');

newStoryButton.onclick = function (evt) {
    let data = {
        content: newStoryText.value,
        title: newStoryTitle.value,
        link: newStoryLink.value
    };

    fetch('/create_story', {
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
    getStoryList();

    if (window.location.pathname.startsWith('/read/')) {
        let storyId = window.location.pathname.substring(6);
        console.log("opening story with id " + storyId);
        openStory(storyId);
    }
};

storiesById = {};

function updateStoryList(stories) {
    stories.sort((a, b) => a.countdown - b.countdown);
    storiesById = {};
    let html = ``
    for (let s of stories) {
        storiesById[s.id] = s;
        html += `<li>
            <a story_id="${s.id}" href="/drill.html?storyId=${s.id}">drill</a>&nbsp;&nbsp;
            <a story_id="${s.id}" action="dec_countdown" href="#">-1</a>&nbsp;&nbsp;
            <span>${s.countdown}</span>&nbsp;&nbsp;
            <a story_id="${s.id}" action="inc_countdown" href="#">+1</a>&nbsp;&nbsp;
            <a story_id="${s.id}" href="/story.html?storyId=${s.id}">${s.title}</a>
            </li>`;
    }
    storyList.innerHTML = html;
};

storyList.onclick = function (evt) {
    if (evt.target.tagName == 'A') {
        var storyId = evt.target.getAttribute('story_id');
        let story = storiesById[storyId];

        var action = evt.target.getAttribute('action');
        switch (action) {
            case 'drill':
                break;
            case 'inc_countdown':
                evt.preventDefault();
                story.countdown++;
                updateStory(story);
                break;
            case 'dec_countdown':
                evt.preventDefault();
                if (story.countdown > 0) {
                    story.countdown--;
                    updateStory(story);
                }
                break;
            default:
                break;
        }
    }
};

function getStoryList() {
    fetch('/stories_list', {
        method: 'GET', // or 'PUT'
        headers: {
            'Content-Type': 'application/json',
        }
    }).then((response) => response.json())
        .then((data) => {
            console.log('Stories list success:', data);
            updateStoryList(data);
        })
        .catch((error) => {
            console.error('Error:', error);
        });
}

function updateStory(story) {
    fetch(`/update_story`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(story),
    }).then((response) => response.json())
        .then((data) => {
            console.log(`Success update_story:`, data);
            getStoryList();
        })
        .catch((error) => {
            console.error('Error marking story:', error);
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
