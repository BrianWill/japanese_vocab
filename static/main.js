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
    stories.sort((a, b) => {
        let diff = a.countdown - b.countdown;
        if (diff === 0) {
            return a.date_last_read - b.date_last_read
        }
        return diff;
    });
    
    storiesById = {};

    let header = `<table id="story_table">
    <tr>
    <th>Drill words</th>
    <th>Countdown</th>
    <th>Read count</th>
    <th>Days ago last read</th>
    <th>Days ago created</th>
    <th>Title</th>
    </tr>`;

    function storyRow(s) {
        
        return `<tr>
            <td><a story_id="${s.id}" href="/words.html?storyId=${s.id}">words</a></td>
            <td><a story_id="${s.id}" action="dec_countdown" href="#">-</a>
                <span>${s.countdown}</span>
                <a story_id="${s.id}" action="inc_countdown" href="#">+</a>
            </td>
            <td><span>${s.read_count}</span></td>
            <td><span>${timeSince(s.date_last_read * 1000)}</span></td>
            <td><span>${timeSince(s.date_added * 1000)}</span></td>
            <td><a class="story_title" story_id="${s.id}" href="/story.html?storyId=${s.id}">${s.title}</a></td>
            </tr>`;
    }

    for (let s of stories) {
        storiesById[s.id] = s;
    }
    
    let html = `<h3 alt="read count is zero">Stories in progress</h3>` + header;
    for (let s of stories) {
        if (s.countdown > 0 && s.read_count > 0) {
            html += storyRow(s);    
        }
    }
    html += '</table><h3 alt="read count is zero">Stories never read</h3>' + header;
    for (let s of stories) {
        if (s.countdown > 0 && s.read_count === 0) {
            html += storyRow(s);
        }
    }
    html += '</table><h3 alt="countdown is zero">Stories finished</h3>' + header;
    for (let s of stories) {
        if (s.countdown === 0) {
            html += storyRow(s);
        }
    }
    storyList.innerHTML = html + '</table>';
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
