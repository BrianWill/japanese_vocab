const DRILL_COOLDOWN_RANK_4 = 60 * 60 * 24 * 1000; // 1000 days in seconds
const DRILL_COOLDOWN_RANK_3 = 60 * 60 * 24 * 30;  // 30 days in seconds
const DRILL_COOLDOWN_RANK_2 = 60 * 60 * 24 * 4;   // 4 days in seconds
const DRILL_COOLDOWN_RANK_1 = 60 * 60 * 5;        // 5 hours in second
const cooldownsByRank = [0, DRILL_COOLDOWN_RANK_1, DRILL_COOLDOWN_RANK_2, DRILL_COOLDOWN_RANK_3, DRILL_COOLDOWN_RANK_4];

function splitOnHighPitch(str, pitch) {
    let [downPitch, upPitch] = pitch;
    //console.log(`downpitch ${downPitch}, up pitch ${upPitch}`);

    if (downPitch === 0) {
        return ['', '', str];
    }
    let mora = [];
    let s = new Set(['ゅ', 'ょ', 'ゃ', 'ャ', 'ュ', 'ョ']);
    let chars = str.split('');
    for (let i = 0; i < chars.length; i++) {
        if (s.has(chars[i + 1])) {
            mora.push(chars[i] + chars[i + 1]);
            i++;
        } else {
            mora.push(chars[i]);
        }
    }
    return [
        mora.slice(0, downPitch - 1).join(''),
        mora.slice(downPitch - 1, downPitch).join(''),
        mora.slice(downPitch).join('')
    ];
}

function displayKanji(kanji, word) {
    html = '';

    if (!kanji || kanji.length === 0) {
        kanjiResultsDiv.innerHTML = '';
        return;
    }

    for (let ch of new Set(word.split(''))) {
        for (let k of kanji) {
            if (k.literal === ch) {
                for (let group of k.readingmeaning.group) {
                    onyomi = group.reading.filter(x => x.type === 'ja_on').map(x => `<span class="kanji_reading">${x.value}</span>`);
                    kunyomi = group.reading.filter(x => x.type === 'ja_kun').map(x => `<span class="kanji_reading">${x.value}</span>`);

                    var meanings = group.meaning.filter(x => !x.language).map(x => x.value);

                    var misc = '';
                    if (k.misc.stroke_count) {
                        misc += `<span class="strokes">strokes: ${k.misc.stroke_count}</span>`;
                    }
                    if (k.misc.frequency) {
                        misc += `<span class="frequency">frequency: ${k.misc.frequency}</span>`;
                    }

                    html += `<div class="kanji">
                            <div>
                            <span class="literal">${k.literal}</span>
                            <div><span class="onyomi_readings">${onyomi.join('')}</span></div>
                            <div><span class="kunyomi_readings">${kunyomi.join('')}</span></div>
                            </div>
                            <div class="kanji_meanings">${meanings.join(';  &nbsp;&nbsp;')}</div>
                            <div class="kanji_misc">${misc}</div>
                            </div>`;
                }
            }
        }
    }



    kanjiResultsDiv.innerHTML = html;
}

function getKanji(str) {
    fetch('/kanji', {
        method: 'POST', // or 'PUT'
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(str),
    }).then((response) => response.json()
    ).then((data) => {
        displayKanji(data.kanji, str);
    }).catch((error) => {
        console.error('Error:', error);
    });
}

function updateWord(word, wordInfoMap, marking) {
    fetch('/update_word', {
        method: 'POST', // or 'PUT'
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(word),
    }).then((response) => response.json()
    ).then((data) => {
        if (marking) {
            snackbarMessage(`word <span class="snackbar_word">${data.base_form}</span> marked as reviewed`);
        } else {
            snackbarMessage(`word <span class="snackbar_word">${data.base_form}</span> set to rank ${data.rank}`);
        }
        updateWordInfo(data, wordInfoMap, marking);
    }).catch((error) => {
        console.error('Error:', error);
    });
}

function updateWordInfo(word, wordInfoMap, marking) {
    let tokenizedStory = document.getElementById('tokenized_story');
    if (tokenizedStory) {
        let wordSpans = tokenizedStory.querySelectorAll(`span[baseform="${word.base_form}"]`);
        console.log('updating word info', word.base_form, word.rank, word.date_marked, 'found spans', wordSpans.length);
        let unixTime = Math.floor(Date.now() / 1000);
        for (let span of wordSpans) {
            span.classList.remove('rank1', 'rank2', 'rank3', 'rank4');
            span.classList.add('rank' + word.rank);
            if (marking) {
                span.classList.remove('offcooldown');
            }
        }
    }

    var wordInfo = wordInfoMap[word.base_form];
    wordInfo.rank = word.rank;
    wordInfo.date_marked = word.date_marked;
}

var snackebarTimeoutHandle = null;

function snackbarMessage(msg) {
    // Get the snackbar DIV
    var el = document.getElementById("snackbar");

    // Add the "show" class to DIV
    el.classList.add("show");
    el.innerHTML = msg;

    clearTimeout(snackebarTimeoutHandle);

    // After 3 seconds, remove the show class from DIV
    snackebarTimeoutHandle = setTimeout(function () {
        el.classList.remove('show');
    }, 3000);
}

function shuffle(array) {
    if (array.length < 2) {
        return;
    }

    let idx = array.length;
    while (idx != 0) {
        // Pick a remaining element.
        let randIdx = Math.floor(Math.random() * idx);
        idx--;

        // And swap it with the current element.
        [array[idx], array[randIdx]] = [array[randIdx], array[idx]];
    }
}


function displayEntry(entry) {
    let readings = '';
    for (var r of entry.readings || []) {
        if (r.pitch) {
            let pitch = r.pitch.split(',').map(x => parseInt(x));
            let parts = splitOnHighPitch(r.reading, pitch);
            readings += `<span class="reading">${parts[0]}<span class="high_pitch">${parts[1]}</span>${parts[2]}</span>`;
        } else {
            readings += `<span class="reading unknown_pitch">${r.reading}﹖</span>`;
        }
    }

    let kenjiSpellings = '';
    for (var k of entry.kanji_spellings || []) {
        kenjiSpellings += `<span class="kanji_spelling">${k.kanji_spelling}</span>`;
    }

    let senses = '';
    for (var s of entry.senses || []) {
        let pos = s.parts_of_speech.map(x => `<span class="pos">${x}</span>`);
        senses += `<span class="sense">
            <span>${pos.join(' ')}</span>
            <span class="glosses">${s.glosses.map(x => x.value).join('; &nbsp;&nbsp;')}</span>
        </span>`;
    }

    return `<div class="entry">
                <div class="word">
                    <div class="readings">${readings}</div>
                    <div class="kanji_spellings">${kenjiSpellings}</div>
                    <div class="senses">${senses}</div>
                </div>
            </div>`;
}

function updateStory(story, refreshList) {
    let temp = { ...story };
    delete temp.content;
    delete temp.tokens;
    delete temp.link;
    delete temp.title;
    delete temp.words;
    fetch(`/update_story`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(temp),
    }).then((response) => response.json())
        .then((data) => {
            console.log(`Success update_story:`, data);
            if (refreshList) {
                getStoryList();
            }
        })
        .catch((error) => {
            console.error('Error marking story:', error);
        });
}

function getStoryList() {
    fetch('/stories_list', {
        method: 'GET', // or 'PUT'
        headers: {
            'Content-Type': 'application/json',
        }
    }).then((response) => response.json())
        .then((data) => {
            console.log('Stories list success:', data);
            updateStoryList(data);
        })
        .catch((error) => {
            console.error('Error:', error);
        });
}


function timeSince(date) {
    let now = Math.floor(new Date() / 1000);
    let elapsedSeconds = now - date;
    if (elapsedSeconds > 0) {
        elapsedSeconds++; // adding one second fixes cases like "24 hours" instead of "1 day";
        let interval = elapsedSeconds / 31536000;
        if (interval > 1) {
            var val = Math.floor(interval);
            return `${val} ${val == 1 ? 'year' : 'years'} ago`;
        }
        interval = elapsedSeconds / 2592000;
        if (interval > 1) {
            var val = Math.floor(interval);
            return `${val} ${val == 1 ? 'month' : 'months'} ago`;
        }
        interval = elapsedSeconds / 86400;
        if (interval > 1) {
            var val = Math.floor(interval);
            return `${val} ${val == 1 ? 'day' : 'days'} ago`;
        }
        interval = elapsedSeconds / 3600;
        if (interval > 1) {
            var val = Math.floor(interval);
            return `${val} ${val == 1 ? 'hour' : 'hours'} ago`;
        }
        interval = elapsedSeconds / 60;
        if (interval > 1) {
            var val = Math.floor(interval);
            return `${val} ${val == 1 ? 'minute' : 'minutes'} ago`;
        }
        if (interval > 1) {
            var val = Math.floor(elapsedSeconds);
            return `${val} ${val == 1 ? 'second' : 'seconds'} ago`;
        }
    } else {
        elapsedSeconds = - elapsedSeconds;
        elapsedSeconds++;  // adding one second fixes cases like "24 hours" instead of "1 day";
        let interval = elapsedSeconds / 31536000;
        if (interval > 1) {
            var val = Math.floor(interval);
            return `in ${val} ${val == 1 ? 'year' : 'years'}`;
        }
        interval = elapsedSeconds / 2592000;
        if (interval > 1) {
            var val = Math.floor(interval);
            return `in ${val} ${val == 1 ? 'month' : 'months'}`;
        }
        interval = elapsedSeconds / 86400;
        if (interval > 1) {
            var val = Math.floor(interval);
            return `in ${val} ${val == 1 ? 'day' : 'days'}`;
        }
        interval = elapsedSeconds / 3600;
        if (interval > 1) {
            var val = Math.floor(interval);
            return `in ${val} ${val == 1 ? 'hour' : 'hours'}`;
        }
        interval = elapsedSeconds / 60;
        if (interval > 1) {
            var val = Math.floor(interval);
            return `in ${val} ${val == 1 ? 'minute' : 'minutes'}`;
        }
        if (interval > 1) {
            var val = Math.floor(elapsedSeconds);
            return `in ${val} ${val == 1 ? 'second' : 'seconds'}`;
        }
    }
    return 'now';
}

const DRILL_ALL_CURRENT = -1;
const DRILL_ALL = 0;

var storiesById = {};

function updateStoryList(stories) {
    stories.sort((a, b) => {
        let diff = b.status - a.status;
        if (diff === 0) {
            return b.date_added - a.date_added
        }
        return diff;
    });

    storiesById = {};

    function storyRow(s) {
        return `<tr>
            <td>
                <select name="status" class="status_select" story_id="${s.id}">
                    <option value="3" ${s.status === 3 ? 'selected' : ''}>Current</option>
                    <option value="1" ${s.status === 1 ? 'selected' : ''}>Never read</option>
                    <option value="0" ${s.status === 0 ? 'selected' : ''}>Archive</option>
                </select>
            </td>
            <td>
                <a action="schedule" story_id="${s.id}" href="#">schedule</a>
            </td>
            <td><a class="story_title status${s.status}" story_id="${s.id}" href="/story.html?storyId=${s.id}">${s.title}</a></td>
            </tr>`;
    }

    for (let s of stories) {
        storiesById[s.id] = s;
    }

    let html = `<table class="story_table">
            <tr>
                <td class="story_table_section" colspan="6">STORIES <span><a action="drill_all" href="/words.html">words of all stories</a>&nbsp;
                <a action="drill_current" href="/words.html?storyId=${DRILL_ALL_CURRENT}">words of all current stories</a></span>
                </td>
            </tr>`;
    for (let s of stories) {
        html += storyRow(s);
    }
    storyList.innerHTML = html + '</table>';
};

function addLogEvent(storyId) {
    fetch(`/add_log_event/${storyId}`, {
        method: 'GET',
        headers: {
            'Content-Type': 'application/json',
        }
    }).then((response) => response.json())
        .then((data) => {
            console.log(data);
            snackbarMessage(data.message);
        })
        .catch((error) => {
            console.error('Error:', error);
        });

}

