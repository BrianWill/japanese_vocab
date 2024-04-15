var storyList = document.getElementById('story_list');
var sourceSelect = document.getElementById('source_select');

const STORY_COOLDOWN = 60 * 60 * 24;

document.body.onload = function (evt) {
    //getStoryList(displayStoryList);
    getCatalogStories(processStories);
};


sourceSelect.onchange = function (evt) {
    displayStories();
}

storyList.onchange = function (evt) {
    if (evt.target.className.includes('count_spinner')) {
        let storyId = parseInt(evt.target.getAttribute('story_id'));
        let story = storiesById[storyId];
        story.countdown = parseInt(evt.target.value);
        updateStoryCounts(story, () => { });
    }
};

var storiesById = {};
var storiesBySource = {};
var stories = [];

let levelNameMap = {
    1: 'Low',
    2: 'Medium',
    3: 'High',
};

storyList.onclick = function (evt) {
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

function processStories(storyData) {
    stories = storyData;
    console.log(stories);

    for (let s of stories) {
        if (s.repetitions_remaining === undefined) {
            s.repetitions_remaining = 0;
        }

        if (s.date_marked === undefined) {
            s.date_marked = 0;
        }
    }

    stories.sort((a, b) => {
        return b.date_marked - a.date_marked
    });

    storiesById = {};
    storiesBySource = {};

    for (let s of stories) {
        storiesById[s.id] = s;
        let list = storiesBySource[s.source];
        if (list === undefined) {
            list = storiesBySource[s.source] = [];
        }
        list.push(s);
    }


    let selectOptionsHTML = `<option value="in progress">In Progress</option>`;
    let i = 1;
    for (let source in storiesBySource) {
        selectOptionsHTML += `<option value="${i}">${source}</option>`
        i++;
    }
    sourceSelect.innerHTML = selectOptionsHTML;

    displayStories();
};



function displayStories() {
    function storyRow(s) {
        return `<tr>
            <td>
                <span title="when this story was last read">${timeSince(s.date_marked)}</span>
            </td>  
            <td>
               <input story_id="${s.id}" type="number" class="count_spinner" min="-1" max="9" steps="1" value="${s.repetitions_remaining}">
            </td>
            <td>
                <span story_id="${s.id}" level="${s.level}" class="level ${s.level}" title="difficulty level of this story">${s.level}</span>
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

    let html = tableHeader;

    let source = sourceSelect.options[sourceSelect.selectedIndex].text;

    if (source == 'In Progress') {
        // in progress
        for (let s of stories) {
            if (s.status != 'catalog') {
                console.log('in progress', s);
                html += storyRow(s);
            }
        }

        html += `</table>`
    } else {
        // display by source
        let list = storiesBySource[source];
        html += `<hr><h2>${source}</h2>` + tableHeader;
        for (let s of list) {
            html += storyRow(s);
        }
        html += `</table>`;
    }

    storyList.innerHTML = html;
};

