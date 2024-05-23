var storyList = document.getElementById('story_list');
var sourceSelect = document.getElementById('source_select');
var scheduleDiv = document.getElementById('schedule');
var ipDiv = document.getElementById('ip');

const STORY_COOLDOWN = 60 * 60 * 24;

document.body.onload = function (evt) {
    getStories(processStories);
    getSchedule(displaySchedule);
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

storyList.onchange = function (evt) {
    if (evt.target.className.includes('level_select')) {
        var storyId = evt.target.getAttribute('story_id');
        let story = storiesById[storyId];
        story.level = evt.target.value;
        updateStory(story, () => { });
    }
};

var storiesById = {};
var storiesBySource = {};
var stories = [];
var scheduleEntries = [];

let levelNameMap = {
    1: 'Low',
    2: 'Medium',
    3: 'High',
};

scheduleDiv.onclick = function (evt) {
    if (evt.target.className.includes('schedule_remove_link')) {
        var entryId = parseInt(evt.target.getAttribute('entry_id'));
        unscheduleStory(entryId, 0, () => getSchedule(displaySchedule));
    }

    if (evt.target.className.includes('schedule_down_link')) {
        var entryId = parseInt(evt.target.getAttribute('entry_id'));
        adjustSchedule(entryId, +1, () => getSchedule(displaySchedule));
    }

    if (evt.target.className.includes('schedule_up_link')) {
        var entryId = parseInt(evt.target.getAttribute('entry_id'));
        adjustSchedule(entryId, -1, () => getSchedule(displaySchedule));
    }
};

storyList.onclick = function (evt) {
    if (evt.target.className.includes('schedule_link')) {
        var storyId = evt.target.getAttribute('story_id');
        let story = storiesById[storyId];
        scheduleStory(story.id, () => getSchedule(displaySchedule));
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

function displaySchedule(entries) {

    let scheduleEntries = entries.schedule;
    let logEntries = entries.log;

    // sort entries by day_offset
    scheduleEntries.sort((a, b) => {
        return a.day_offset - b.day_offset;
    });

    logEntries.sort((a, b) => {
        return a.day_offset - b.day_offset;
    });

    {
        let html = `<table class="schedule_table">`;

        html += `<tr class="day_row"><td class="schedule_day">Logged in last 24 hours</td>
            <td></td>
            <td></td>
            <td></td>
            <td></td>
            <td></td>
            <td></td>
            <td></td>
            <td></td>
        </tr>`;

        for (const entry of logEntries) {

            let typeStr = '';
            if (entry.type == 0) {
                typeStr = 'Read';
            } else if (entry.type == 1) {
                typeStr = 'Listen';
            } else if (entry.type == 2) {
                typeStr = 'Drill';
            }

            let hash = integerHash(entry.story + entry.source + entry.title);
            let color = randomPaletteColor(hash);

            html += `<tr>
                <td>${typeStr}</td>  
                <td>${entry.source}</td>    
                <td><a style="color:${color};" class="story_title" story_id="${entry.story}" 
                        href="/story.html?storyId=${entry.story}">${entry.title}</a></td>
                <td>${entry.level}</td>
                <td>${entry.repetitions} rep</td>
                <td></td>
                <td></td>
                <td></td>
            </tr>`;
        }

        let currentDay = -1;

        for (const entry of scheduleEntries) {
            if (entry.day_offset > currentDay) {
                currentDay = entry.day_offset;

                let dayStr = '';
                if (currentDay == 0) {
                    dayStr = 'Today';
                } else if (currentDay == 1) {
                    dayStr = 'Tomorrow';
                } else {
                    dayStr = currentDay + ' days from now';
                }

                html += `<tr class="day_row"><td class="schedule_day">${dayStr}</td>
                    <td></td>
                    <td></td>
                    <td></td>
                    <td></td>
                    <td></td>
                    <td></td>
                    <td></td>
                    <td></td>
                </tr>`;
            }

            let typeStr = '';
            if (entry.type == 0) {
                typeStr = 'Read';
            } else if (entry.type == 1) {
                typeStr = 'Listen';
            } else if (entry.type == 2) {
                typeStr = 'Drill';
            }

            let page = entry.type == 2 ? 'words' : 'story';
            let storyLink = `/${page}.html?storyId=${entry.story}&scheduleId=${entry.id}`;

            let hash = integerHash(entry.story + entry.source + entry.title);
            let color = randomPaletteColor(hash);

            html += `<tr>
            <td>${typeStr}</td>
            <td>${entry.source}</td>    
            <td><a style="color: ${color};" class="story_title" story_id="${entry.story}" href="${storyLink}">${entry.title}</a></td>
            <td>${entry.level}</td>
            <td>${entry.repetitions} rep</td>
            <td><a href="#" class="schedule_remove_link" entry_id="${entry.id}" title="move this rep to the next day">remove</a></td>
            <td><a href="#" class="schedule_down_link" entry_id="${entry.id}" title="move this rep to the next day">down</a></td>
            <td><a href="#" class="schedule_up_link" entry_id="${entry.id}" title="move this rep to the previous day">up</a></td>
        </tr>`;
        }

        scheduleDiv.innerHTML = html + `</table>`;
    }
};

function displayStories() {
    function storyRow(s) {
        return `<tr>
            <td><span class="story_source">${s.source}</span></td>
            <td><a class="story_title" story_id="${s.id}" href="/story.html?storyId=${s.id}">${s.title}</a></td>
            <td>
                <select class="level_select" story_id="${s.id}" title="difficulty level of this story">
                    <option ${(s.level == 'low') ? 'selected' : ''} value="low">low</option>
                    <option ${(s.level == 'medium') ? 'selected' : ''} value="medium">medium</option>
                    <option ${(s.level == 'high') ? 'selected' : ''} value="high">high</option>
                </select>
            </td>
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


