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

function displayReps(entries) {

    entries = entries.stories;

    // sort entries by source, then by title?
    entries.sort((a, b) => {
        if (a.source == b.source) {
            return (a.title < b.title) ? -1 : (a.title > b.title) ? 1 : 0;
        } else {
            return (a.source < b.source) ? -1 : (a.source > b.source) ? 1 : 0;
        }
    });

    let html = `<h2>Current stories</h2>
            <table class="schedule_table">
            <tr class="day_row logged_row">
                <td>Source</td>
                <td>Title</td>
                <td>Completed<br>listening reps</td>
                <td>Completed<br>drill reps</td>
                <td>Time since<br>last rep</td>
                <td>Reps<br>todo</td>
            </tr>`;

    for (const entry of entries) {

        let timeLastRep = 1;

        if (entry.reps_logged) {
            for (let rep of entry.reps_logged) {
                if (rep.date > timeLastRep) {
                    timeLastRep = rep.date;
                }
            }
        }

        let todoReps = ``;
        for (let rep of entry.reps_todo) {
            if (rep == LISTENING) {
                todoReps += `<span class="listening" title="listening rep">聞</span>`;
            } else if (rep == DRILLING) {
                todoReps += `<span class="drill" title="vocabulary drill rep">語</span>`;
            }
        }

        html += `<tr>
            <td>${entry.source}</td>    
            <td><a class="story_title" story_id="${entry.id}" href="/story.html?storyId=${entry.id}">${entry.title}</a></td>
            <td>${entry.repetitions}</td>
            <td>${entry.repetitions}</td>
            <td>${timeSince(timeLastRep)}</td>
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

storyList.onclick = function (evt) {
    if (evt.target.className.includes('add_multiple_reps')) {
        var storyId = evt.target.getAttribute('story_id');
        let story = storiesById[storyId];
        addStoryReps(story.id, DEFAULT_REPS, 
            function () { 
                getReps(displayReps);
                snackbarMessage(`added reps to story: ${story.title}`);
            }
        );
    }
};

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
                <td>
                    <a href="#" story_id="${s.id}" class="add_multiple_reps" title="add several alternating listening and drill reps">queue reps</a>
                </td>
                <td><span class="story_source">${s.source}</span></td>
                <td><a class="story_title" story_id="${s.id}" href="/story.html?storyId=${s.id}">${s.title}</a></td>
                <td>${s.repetitions} reps</td>
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
