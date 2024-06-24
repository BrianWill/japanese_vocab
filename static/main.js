var storyList = document.getElementById('story_list');
var sourceSelect = document.getElementById('source_select');
var ipDiv = document.getElementById('ip');

document.body.onload = function (evt) {
    getStories(function (stories) {
        processCatalog(stories);
        displayCurrent(stories);
    });
    getIP((ips) => {
        let html = 'local ip: ';
        for (const ip of ips) {
            html += ip + '&nbsp;&nbsp;'
        }
        ipDiv.innerHTML = html;
    });
};

function displayCurrent(stories) {
    stories = stories.filter((s) => s.has_reps_todo) ;

    // sort entries by source, then by title?
    stories.sort((a, b) => {
        if (a.source == b.source) {
            return (a.title < b.title) ? -1 : (a.title > b.title) ? 1 : 0;
        } else {
            return (a.source < b.source) ? -1 : (a.source > b.source) ? 1 : 0;
        }
    });

    stories.sort((a, b) => {
        return (a.date_last_rep < b.date_last_rep) ? -1 : (a.date_last_rep > b.date_last_rep) ? 1 : 0;
    });

    let html = `
            <table class="schedule_table">
            <tr class="day_row logged_row">
                <td>Source</td>
                <td>Title</td>
                <td>Time since<br>last rep</td>
            </tr>`;

    // todo get count of todo reps across all excerpts

    for (const s of stories) {
        html += `<tr>
            <td>${s.source}</td>    
            <td><a class="story_title" story_id="${s.id}" href="/story.html?storyId=${s.id}">${s.title}</a></td>
            <td>${timeSince(s.date_last_rep)}</td>
        </tr>`;
    }

    html += `</table>`;

    document.getElementById('reps').innerHTML = html;
};


/* CATALOG */

sourceSelect.onchange = function (evt) {
    displayCatalog();
}

var storiesById = {};
var storiesBySource = {};
var stories = [];

function processCatalog(storyData) {
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
        let count = storiesBySource[source].length;
        selectOptionsHTML += `<option source="${source}" value="${i}">${source} (${count})</option>`
        i++;
    }
    sourceSelect.innerHTML = selectOptionsHTML;

    displayCatalog();
};

function displayCatalog() {
    function storyRow(s) {
        return `<tr>
                <td><span class="story_source">${s.source}</span></td>
                <td><a class="story_title" story_id="${s.id}" href="/story.html?storyId=${s.id}">${s.title}</a></td>
            </tr>`;
    }

    let tableHeader = `<table class="story_table">`;
    let html = tableHeader;
    let source = sourceSelect.options[sourceSelect.selectedIndex].getAttribute('source');

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
