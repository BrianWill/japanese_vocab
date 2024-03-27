var storyList = document.getElementById('story_list');
var newStoryText = document.getElementById('new_story_text');
var newStoryButton = document.getElementById('new_story_button');
var newStoryTitle = document.getElementById('new_story_title');
var newStoryLink = document.getElementById('new_story_link');
var newStoryAudio = document.getElementById('new_story_audio');

const STORY_COOLDOWN = 60 * 60 * 24;

document.body.onload = function (evt) {
    getStoryList(displayStoryList);
};

newStoryButton.onclick = function (evt) {
    let data = {
        content: newStoryText.value,
        title: newStoryTitle.value,
        link: newStoryLink.value,
        audio: newStoryAudio.value,
    };

    newStoryText.value = '';
    newStoryTitle.value = '';
    newStoryLink.value = '';
    newStoryAudio.value = '';

    fetch('/create_story', {
        method: 'POST', // or 'PUT'
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(data),
    }).then((response) => response.json())
        .then((data) => {
            getStoryList(displayStoryList);
        })
        .catch((error) => {
            console.error('Error:', error);
        });
};

storyList.onchange = function (evt) {
    if (evt.target.className.includes('count_spinner')) {
        let storyId = parseInt(evt.target.getAttribute('story_id'));
        let story = storiesById[storyId];
        story.countdown = parseInt(evt.target.value);
        updateStoryCounts(story, () => { });
    }
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
            getStoryList(displayStoryList);
        })
        .catch((error) => {
            console.error('Error retokenizing:', error);
        });
}


var storiesById = {};

let levelNameMap = {
    1: 'Low',
    2: 'Medium',
    3: 'High',
};

storyList.onclick = function(evt) {
    if (!evt.target.classList.contains('level')) {
        return;
    }
    evt.preventDefault();
    var storyId = evt.target.getAttribute('story_id');
    var level = parseInt(evt.target.getAttribute('level'));

    let maxLevel = Object.keys(levelNameMap).length;
    
    let newLevel = level + 1;
    if (newLevel > maxLevel) {
        newLevel = 1;
    }

    evt.target.setAttribute('level', newLevel);
    evt.target.className = `level ${levelNameMap[newLevel]}`;
    evt.target.innerText = levelNameMap[newLevel];

    console.log(`some level ${level} newlevel ${newLevel} id ${storyId}`);

    // update db
    let story = storiesById[storyId];
    story.level = newLevel;
    updateStoryCounts(story, () => { });
};

function displayStoryList(stories) {
    stories.sort((a, b) => {
        if (a.date_last_read === b.date_last_read) {
            return b.date_added - a.date_added
        }
        return b.date_last_read - a.date_last_read
    });


    function storyRow(s) {
        return `<tr>
            <td>
                <span title="when this story was last read">${timeSince(s.date_last_read)}</span>
            </td>  
            <td>
               <input story_id="${s.id}" type="number" class="count_spinner" min="-1" max="9" steps="1" value="${s.countdown}">
            </td>
            <td>
                <span story_id="${s.id}" level="${s.level}" class="level ${levelNameMap[s.level]}" title="difficulty level of this story">${levelNameMap[s.level]}</span>
            </td>
            <td><a class="story_title" story_id="${s.id}" href="/story.html?storyId=${s.id}">${s.title}</a></td>
            </tr>`;
    }

    let tableHeader = `<table class="story_table">
    <tr>
        <th>Time last read</th>
        <th title="number of additional times you intend to read this story">Countdown</th>
        <th title="difficulty level of this story">Level</th>
        <th>Title</th>
    </tr>`;


    let html = `
        <h3><a class="drill_link" title="vocab and kanji from stories with a countdown greater than zero" href="words.html?set=current">drill countdown &gt; 0</a>
        <a class="drill_link" title="vocab and kanji from stories with a countdown equal to zero" href="words.html?set=active">drill countdown = 0</a>
        <a class="drill_link" title="vocab and kanji from stories with a countdown less than zero" href="words.html?set=archived">drill countdown &lt; 0</a></h3>` 
        + tableHeader;

    storiesById = {};

    for (let s of stories) {
        storiesById[s.id] = s;
        if (s.countdown > 0) {
            html += storyRow(s);
        }
    }

    html += `</table> <hr>` + tableHeader;

    for (let s of stories) {
        storiesById[s.id] = s;
        if (s.countdown == 0) {
            html += storyRow(s);
        }
    }

    html += `</table> <hr>` + tableHeader;

    for (let s of stories) {
        storiesById[s.id] = s;
        if (s.countdown < 0) {
            html += storyRow(s);
        }
    }

    storyList.innerHTML = html + '</table>';
};
