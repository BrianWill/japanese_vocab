var storyList = document.getElementById('story_list');
var sourceSelect = document.getElementById('source_select');
var ipDiv = document.getElementById('ip');

document.body.onload = function (evt) {
    getStories(processCatalog);
    getReps(displayReps);
    getIP((ips) => {
        let html = 'local ip: ';
        for (const ip of ips) {
            html += ip + '&nbsp;&nbsp;'
        }
        ipDiv.innerHTML = html;
    });
};

function displayReps(stories) {

    stories = stories.stories;

    for (const s of stories) {
        s.timeLastRep = 1;
        s.listeningRepCount = 0;
        s.drillRepCount = 0;

        if (s.reps_logged) {
            for (let rep of s.reps_logged) {
                if (rep.type == LISTENING) {
                    s.listeningRepCount++;
                } else if (rep.type == DRILLING) {
                    s.drillRepCount++;
                }

                if (rep.date > s.timeLastRep) {
                    s.timeLastRep = rep.date;
                }
            }
        }
    }

    // sort entries by source, then by title?
    stories.sort((a, b) => {
        if (a.source == b.source) {
            return (a.title < b.title) ? -1 : (a.title > b.title) ? 1 : 0;
        } else {
            return (a.source < b.source) ? -1 : (a.source > b.source) ? 1 : 0;
        }
    });

    stories.sort((a, b) => {
        return (a.timeLastRep < b.timeLastRep) ? -1 : (a.timeLastRep > b.timeLastRep) ? 1 : 0;
    });

    let html = `
            <table class="schedule_table">
            <tr class="day_row logged_row">
                <td>Source</td>
                <td>Title</td>
                <td title="number of completed listening reps">Listening<br>reps</td>
                <td title="number of completed drill reps">Drill<br>reps</td>
                <td>Time since<br>last rep</td>
                <td>Queued<br>reps</td>
            </tr>`;

    for (const s of stories) {

        let todoReps = ``;
        for (let rep of s.reps_todo) {
            if (rep == LISTENING) {
                todoReps += `<span class="listening" title="listening rep">聞</span>`;
            } else if (rep == DRILLING) {
                todoReps += `<span class="drill" title="vocabulary drill rep">語</span>`;
            }
        }

        html += `<tr>
            <td>${s.source}</td>    
            <td><a class="story_title" story_id="${s.id}" href="/story.html?storyId=${s.id}">${s.title}</a></td>
            <td>${s.listeningRepCount}</td>
            <td>${s.drillRepCount}</td>
            <td>${timeSince(s.timeLastRep)}</td>
            <td class="rep_sequence">${todoReps}</td>
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
        selectOptionsHTML += `<option value="${i}">${source}</option>`
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
    let source = sourceSelect.options[sourceSelect.selectedIndex].text;

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
