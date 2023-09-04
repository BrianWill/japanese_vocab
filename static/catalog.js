
var storyList = document.getElementById('story_list');

var newStoryText = document.getElementById('new_story_text');
var newStoryButton = document.getElementById('new_story_button');
var newStoryTitle = document.getElementById('new_story_title');
var newStoryLink = document.getElementById('new_story_link');

document.body.onload = function (evt) {
    getStoryList();
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
        })
        .catch((error) => {
            console.error('Error retokenizing:', error);
        });
}


var storiesById = {};

function displayStoryList(stories) {
    stories.sort((a, b) => {
        return b.date_added - a.date_added
    });

    function storyRow(s) {
        let button = '';
        return `<tr>
            <td>
                <a action="schedule" story_id="${s.id}" href="#">schedule</a>
            </td>
            <td><a class="story_title status${s.status}" story_id="${s.id}" href="/story.html?storyId=${s.id}">${s.title}</a></td>
            </tr>`;
    }


    let html = `<table class="story_table">`;

    storiesById = {};

    for (let s of stories) {
        storiesById[s.id] = s;
        html += storyRow(s);
    }

    storyList.innerHTML = html + '</table>';
};


storyList.onclick = function (evt) {
    console.log(evt.target);

    if (evt.target.tagName == 'A') {
        var storyId = evt.target.getAttribute('story_id');
        var action = evt.target.getAttribute('action');
        let story = storiesById[storyId];
        switch (action) {
            case 'schedule':
                evt.preventDefault();
                scheduleStory(storyId);
                break;
            case 'add':
                evt.preventDefault();
                story.status = 3;
                updateStory(story, true);
                break;
            case 'remove':
                evt.preventDefault();
                story.status = 0;
                updateStory(story, true);
                break;
        }
    }
};

function scheduleStory(storyId) {
    fetch(`/schedule_story/${storyId}`, {
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

function displayStorySchedule() {
    // intentionally blank
}

