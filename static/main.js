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
        story.repetitions_remaining = parseInt(evt.target.value);
        updateStoryStats(story, () => { });
    }

    if (evt.target.className.includes('status_select')) {
        let storyId = parseInt(evt.target.getAttribute('story_id'));
        let story = storiesById[storyId];
        story.status = evt.target.value;
        updateStoryStats(story, () => { });
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
    updateStoryStats(story, () => { });
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
                <select class="status_select" story_id="${s.id}">
                    <option ${(s.status == 'catalog') ? 'selected' : ''} value="catalog">catalog</option>
                    <option ${(s.status == 'in progress') ? 'selected' : ''} value="in progress">in progress</option>
                    <option ${(s.status == 'backlog') ? 'selected' : ''} value="backlog">backlog</option>
                    <option ${(s.status == 'archived') ? 'selected' : ''} value="archived">archived</option>
                </select>
            </td>
            <td>
                <span title="when this story was last read">${timeSince(s.date_marked)}</span>
            </td>  
            <td>
               <input story_id="${s.id}" type="number" class="count_spinner" min="0" max="9" steps="1" value="${s.repetitions_remaining}">
            </td>
            <td>
                <select class="level_select" story_id="${s.id}" title="difficulty level of this story">
                    <option ${(s.level == 'low') ? 'selected' : ''} value="low">low</option>
                    <option ${(s.level == 'medium') ? 'selected' : ''} value="medium">medium</option>
                    <option ${(s.level == 'high') ? 'selected' : ''} value="high">high</option>
                </select>
            </td>
            <td><a class="story_title" story_id="${s.id}" href="/story.html?storyId=${s.id}">${s.title}</a></td>
            <td><span class="story_source">${s.source}</span></td>
            </tr>`;
    }

    let tableHeader = `<table class="story_table">
    <tr>
        <th title="number of additional times you intend to read this story">Status</th>
        <th>Time last read</th>
        <th title="number of additional times you intend to read this story">Remaining<br>Repetitions</th>
        <th title="difficulty level of this story">Level</th>
        <th>Title</th>
        <th>Source</th>
    </tr>`;

    let html = '';

    let source = sourceSelect.options[sourceSelect.selectedIndex].text;

    if (source == 'In Progress') {
        // in progress
        html = tableHeader;
        for (let s of stories) {
            if (s.status != 'catalog') {
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

