var storyList = document.getElementById('story_list');
var sourceSelect = document.getElementById('source_select');
var ipDiv = document.getElementById('ip');

const STORY_COOLDOWN = 60 * 60 * 24;

document.body.onload = function (evt) {
    getStories(processStories);
    getIP((ips) => {
        let html = 'local ip: ';
        for (const ip of ips) {
            html += ip + '&nbsp;&nbsp;'
        }
        ipDiv.innerHTML = html;
    });
};


sourceSelect.onchange = function (evt) {
    displayStories();
}

var storiesById = {};
var storiesBySource = {};
var stories = [];
var scheduleEntries = [];

storyList.onclick = function (evt) {
    if (evt.target.className.includes('schedule_link')) {
        var storyId = evt.target.getAttribute('story_id');
        let story = storiesById[storyId];
        scheduleStory(story.id, () => window.location.href = "/");
    }
};

function processStories(storyData) {
    stories = storyData;

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

    let selectOptionsHTML = ``;
    let i = 0;
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
            <td><span class="story_source">${s.source}</span></td>
            <td><a class="story_title" story_id="${s.id}" href="/story.html?storyId=${s.id}">${s.title}</a></td>
            <td>${s.repetitions} reps</td>            
            <td>
                <a href="#" story_id="${s.id}" class="schedule_link">schedule</a>
            </td>
            </tr>`;
    }

    let tableHeader = `<table class="story_table">`;

    let html = tableHeader;

    let source = sourceSelect.options[sourceSelect.selectedIndex].text;

    // display by source
    let list = storiesBySource[source];

    list.sort((a, b) => {
        return a.episode_number - b.episode_number;
    });

    for (let s of list) {
        html += storyRow(s);
    }
    html += `</table>`;

    storyList.innerHTML = html;
};


