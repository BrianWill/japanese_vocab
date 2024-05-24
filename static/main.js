var scheduleDiv = document.getElementById('schedule');
var ipDiv = document.getElementById('ip');

const STORY_COOLDOWN = 60 * 60 * 24;

document.body.onload = function (evt) {
    getSchedule(displaySchedule);
    getIP((ips) => {
        let html = 'local ip: ';
        for (const ip of ips) {
            html += ip + '&nbsp;&nbsp;'
        }
        ipDiv.innerHTML = html;
    });
};


var scheduleEntries = [];

let levelNameMap = {
    1: 'Low',
    2: 'Medium',
    3: 'High',
};

scheduleDiv.onclick = function (evt) {
    evt.preventDefault();

    if (evt.target.className.includes('schedule_remove_link')) {
        var entryId = parseInt(evt.target.getAttribute('entry_id'));
        unscheduleStory(entryId, 0, () => getSchedule(displaySchedule));
    }

    if (evt.target.className.includes('schedule_add_link')) {
        var entryId = parseInt(evt.target.getAttribute('entry_id'));
        scheduleAddRep(entryId, () => getSchedule(displaySchedule));
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

        html += `<tr class="day_row logged_row">
            <td class="schedule_day">Logged in last 24 hours</td>
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

            html += `<tr class="logged_row">
                <td>${typeStr}</td>  
                <td>${entry.source}</td>    
                <td><span style="color:${color};" class="story_title">${entry.title}</span></td>
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

                html += `<tr class="day_row">
                    <td class="schedule_day">${dayStr}</td>
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
                typeStr = 'Read 📖';
            } else if (entry.type == 1) {
                typeStr = 'Listen 👂';
            } else if (entry.type == 2) {
                typeStr = 'Drill 📣';
            }

            let page = entry.type == 2 ? 'words' : 'story';
            let storyLink = `/${page}.html?storyId=${entry.story}&scheduleId=${entry.id}`;

            let hash = integerHash(entry.story + entry.source + entry.title);
            let color = randomPaletteColor(hash);

            html += `<tr>
            <td>${typeStr}</td>
            <td>${entry.source}</td>    
            <td><a style="color: ${color};" class="story_title" story_id="${entry.story}" href="${storyLink}">${entry.title}</a></td>
            <td>${entry.repetitions} rep</td>
            <td><a href="#" class="schedule_remove_link" entry_id="${entry.id}" title="remove this rep">✖️</a></td>
            <td><a href="#" class="schedule_add_link" entry_id="${entry.id}" title="add another rep of this story to the next day">➕</a></td>
            <td><a href="#" class="schedule_down_link" entry_id="${entry.id}" title="move this rep to the next day">🡳</a></td>
            <td><a href="#" class="schedule_up_link" entry_id="${entry.id}" title="move this rep to the previous day">🡱</a></td>
        </tr>`;
        }

        scheduleDiv.innerHTML = html + `</table>`;
    }
};
