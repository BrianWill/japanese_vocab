var storyTitle = document.getElementById('story_title');
var storiesDiv = document.getElementById('stories');

document.body.onload = function (evt) {
    getStoryList(displayStoryList);
};



storiesDiv.onclick = function(evt) {
    if (evt.target.tagName == 'A') {
        var logId = evt.target.getAttribute('log_id');
        var storyId = evt.target.getAttribute('story_id');
        var action = evt.target.getAttribute('action');
        switch (action) {
            case 'log':
                evt.preventDefault();
                logScheduledStory(logId, storyId);
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
            displayStory(data);
        })
        .catch((error) => {
            console.error('Error:', error);
        });
}

const STORY_COOLDOWN = 60 * 60 * 24;  

function displayStoryList(stories) {
    stories.sort((a, b) => {
        return a.date_last_read - b.date_last_read
    });

    function storyRow(s) {
        return `<tr>
            <td>
                <span title="when this story was last read">${timeSince(s.date_last_read)}</span>
            </td>    
            <td>
               <span title="number of times left to read this story">${s.countdown}</span>
            </td>
            <td>
                <span title="number of times this story has been read">${s.read_count}</span>
            </td>
            <td><a class="story_title" story_id="${s.id}" href="/story.html?storyId=${s.id}">${s.title}</a></td>
            </tr>`;
    }


    let html = `<table class="story_table">
        <tr>
            <th title="when this story was last read">Time last read</th>    
            <th>Countdown</th>
            <th title="number of times this story has been read">Total reads</th>
            <th>Title</th>
        </tr>`;

    storiesById = {};

    for (let s of stories) {
        storiesById[s.id] = s;
        if (s.countdown > 0) {
            html += storyRow(s);
        }        
    }

    storiesDiv.innerHTML = html + '</table>';
};

