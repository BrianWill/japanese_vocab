var storyList = document.getElementById('story_list');
var sourceList = document.getElementById('sources_list');
var ipDiv = document.getElementById('ip');

var storiesById = {};
var storiesBySource = {};
var stories = [];

document.body.onload = function (evt) {
    getStories(function (data) {
        processCatalog(data.stories, data.word_stats);
    });
    getIP((ips) => {
        let html = 'local ip: ';
        for (const ip of ips) {
            html += ip + '&nbsp;&nbsp;'
        }
        ipDiv.innerHTML = html;
    });
};

document.getElementById('main_sidebar').onclick = function (evt) {
    if (evt.target.classList.contains('action_recently_logged')) {
        evt.preventDefault();
        window.history.replaceState(null, null, "?");
        displayRecentlyLogged(stories);
    } else if (evt.target.classList.contains('source_li')) {
        evt.preventDefault();
        let source = evt.target.getAttribute('source');
        window.history.replaceState(null, null, `?source=${encodeURIComponent(source)}`);
        displaySourceStoryList(storiesBySource[source]);
    }
};


const TWO_WEEKS_IN_SECONDS = 60 * 60 * 24 * 7 * 2;
const TWO_MONTHS_IN_SECONDS = 60 * 60 * 24 * 7 * 8;

function displayRecentlyLogged(stories) {
    stories = stories.filter((s) => {
        if (s.date_last_rep <= 1) {
            return false;
        }

        let now = Math.floor(new Date() / 1000);
        let elapsedSeconds = now - s.date_last_rep;
        return elapsedSeconds < TWO_MONTHS_IN_SECONDS;
    });

    let html = `<h2>Stories recently logged <span class="story_recency">(last 2 months)</span></h2>
        <a href="/words.html?storyId=0">drill vocab of all recently logged stories</a><br><br>`;

    if (stories.length == 0) {
        html += `<h4 style="margin-left: 2em;">(none)</h4>`;
        storyList.innerHTML = html;
        return;
    }

    stories.sort((a, b) => {
        return (a.date_last_rep < b.date_last_rep);
    });

    html += `<table class="story_table">
    <tr class="day_row logged_row">
        <th>Source</th>
        <th>Title</th>
        <th>Time since<br>last logged</th>
        <th>Number of<br>times logged</th>
        <th>Word<br>Count</th>
        <th>Words<br>Archived</th>
    </tr>`

    for (const s of stories) {
        html += `<tr>
            <td>${s.source}</td>    
            <td><a class="story_title" story_id="${s.id}" href="/story.html?storyId=${s.id}">${s.title}</a></td>
            <td>${timeSince(s.date_last_rep)}</td>
            <td>${s.log.length}</td>
            <td>${s.word_count}</td>
            <td>${archivedWordPercentage(s)}</td>
        </tr>`;
    }

    html += `</table>`;

    storyList.innerHTML = html;
}

function archivedWordPercentage(story) {
    if (story.word_count == 0) {
        return 'n/a';
    }
    var percentage = Math.floor((story.archived_word_count / story.word_count) * 100);
    var color = 'red';
    if (percentage >= 70) {
        color = 'yellow';
    }
    if (percentage >= 85) {
        color = 'green';
    }

    return `<span class="${color}">${percentage}</span> <span class="archived_percentage ${color}"> %<span>`;
}

function processCatalog(storyData, wordStats) {
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

        s.date_last_rep = 0;
        for (let logItem of s.log) {
            if (logItem.date > s.date_last_rep) {
                s.date_last_rep = logItem.date
            }
        }
    }

    for (let source in storiesBySource) {
        let stories = storiesBySource[source];
        stories.sort((a, b) => {
            return ('' + a.title).localeCompare(b.title);
        });
    }

    let sourcesSorted = Object.keys(storiesBySource);
    sourcesSorted.sort((a, b) => {
        return a > b;
    });

    let sourcesHTML = ``;
    for (let source of sourcesSorted) {
        let count = storiesBySource[source].length;
        sourcesHTML += `<li><a href="#" source="${source}" class="source_li">${source} (${count})</a></li>`;
    }
    sourceList.innerHTML = sourcesHTML;

    var url = new URL(window.location.href);
    let source = url.searchParams.get("source");
    if (source) {
        displaySourceStoryList(storiesBySource[source]);
    } else {
        displayRecentlyLogged(stories);
    }

    displayWordStats(wordStats);
};

function displayWordStats(wordStats) {
    var total = document.querySelector('#word_stats .total');
    total.innerHTML = wordStats.total;
    var archived = document.querySelector('#word_stats .archived');
    archived.innerHTML = wordStats.archived;
}

function displaySourceStoryList(list) {
    function storyRow(s) {
        return `<tr>
            <td>${s.source}</td>    
            <td><a class="story_title" story_id="${s.id}" href="/story.html?storyId=${s.id}">${s.title}</a></td>
            <td>${timeSince(s.date_last_rep)}</td>
            <td>${s.log.length}</td>
            <td>${s.word_count}</td>
            <td>${archivedWordPercentage(s)}</td>
        </tr>`;
    }

    let tableHeader = `<table class="story_table">
     <tr class="day_row logged_row">
        <th>Source</th>
        <th>Title</th>
        <th>Time since<br>last logged</th>
        <th>Number of<br>times logged</th>
        <th>Word<br>Count</th>
        <th>Words<br>Archived</th>
    </tr>`;

    let html = tableHeader;

    list.sort((a, b) => {
        return a.episode_number - b.episode_number;
    });

    for (let s of list) {
        html += storyRow(s);
    }
    html += `</table>`;

    storyList.innerHTML = html;
};