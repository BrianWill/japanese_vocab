const DRILL_COOLDOWN_RANK_4 = 60 * 60 * 24 * 1000; // 1000 days in seconds
const DRILL_COOLDOWN_RANK_3 = 60 * 60 * 24 * 30;  // 30 days in seconds
const DRILL_COOLDOWN_RANK_2 = 60 * 60 * 24 * 4;   // 4 days in seconds
const DRILL_COOLDOWN_RANK_1 = 60 * 60 * 5;        // 5 hours in second
const cooldownsByRank = [0, DRILL_COOLDOWN_RANK_1, DRILL_COOLDOWN_RANK_2, DRILL_COOLDOWN_RANK_3, DRILL_COOLDOWN_RANK_4];

const STORY_STATUS_CURRENT = 2;
const STORY_STATUS_NEVER_READ = 1;
const STORY_STATUS_ARCHIVE = 0;

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

function displayKanji(kanjiDefs, word) {
    html = '';

    if (!kanjiDefs || kanjiDefs.length === 0) {
        kanjiResultsDiv.innerHTML = '';
        return;
    }


    for (let def of kanjiDefs) {
        for (let group of def.readingmeaning.group) {
            onyomi = group.reading.filter(x => x.type === 'ja_on').map(x => `<span class="kanji_reading">${x.value}</span>`);
            kunyomi = group.reading.filter(x => x.type === 'ja_kun').map(x => `<span class="kanji_reading">${x.value}</span>`);

            var meanings = group.meaning.filter(x => !x.language).map(x => x.value);

            var misc = '';
            if (def.misc.stroke_count) {
                misc += `<span class="strokes">strokes: ${def.misc.stroke_count}</span>`;
            }
            if (def.misc.frequency) {
                misc += `<span class="frequency">frequency: ${def.misc.frequency}</span>`;
            }

            html += `<div class="kanji">
                            <div>
                            <span class="literal">${def.literal}</span>
                            <div><span class="onyomi_readings">${onyomi.join('')}</span></div>
                            <div><span class="kunyomi_readings">${kunyomi.join('')}</span></div>
                            </div>
                            <div class="kanji_meanings">${meanings.join(';  &nbsp;&nbsp;')}</div>
                            <div class="kanji_misc">${misc}</div>
                            </div>`;
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
        for (let i in data) {
            data[i] = JSON.parse(data[i]);
        }
        //console.log('kanji response:', data);
        displayKanji(data, str);
    }).catch((error) => {
        console.error('Error:', error);
    });
}

function updateWord(word, marking) {
    fetch('/update_word', {
        method: 'POST', // or 'PUT'
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(word),
    }).then((response) => response.json()
    ).then((data) => {
        // if (marking) {
        //     snackbarMessage(`word <span class="snackbar_word">${data.base_form}</span> marked as reviewed`);
        // } else {
        //     snackbarMessage(`word <span class="snackbar_word">${data.base_form}</span> set to archived: ${data.archived}`);
        // }
    }).catch((error) => {
        console.error('Error:', error);
    });
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

function updateStoryInfo(story, successFn) {
    story = {
        id: story.id,
        level: story.level,
        date_marked: story.date_marked,
        repetitions: story.repetitions,
        archived: story.archived,
        transcript_ja: story.transcript_ja,
        transcript_en: story.transcript_en
    };
    fetch(`/update_story_info`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(story),
    }).then((response) => response.json())
        .then((data) => {
            if (successFn) {
                successFn(data);
            }
        })
        .catch((error) => {
            console.error('Error marking story:', error);
        });
}


function getStoryList(successFn) {
    fetch('/stories_list', {
        method: 'GET', // or 'PUT'
        headers: {
            'Content-Type': 'application/json',
        }
    }).then((response) => response.json())
        .then((data) => {
            console.log('Stories list success:', data);
            if (successFn) {
                successFn(data);
            }
        })
        .catch((error) => {
            console.error('Error:', error);
        });
}

function getCatalogStories(successFn) {
    fetch('/catalog_stories', {
        method: 'GET', // or 'PUT'
        headers: {
            'Content-Type': 'application/json',
        }
    }).then((response) => response.json())
        .then((data) => {
            console.log('Stories list success:', data);
            if (successFn) {
                successFn(data);
            }
        })
        .catch((error) => {
            console.error('Error:', error);
        });
}

function getSchedule(successFn) {
    fetch('/schedule', {
        method: 'GET', // or 'PUT'
        headers: {
            'Content-Type': 'application/json',
        }
    }).then((response) => response.json())
        .then((data) => {
            console.log('Schedule:', data);
            if (successFn) {
                successFn(data);
            }
        })
        .catch((error) => {
            console.error('Error:', error);
        });
}

function scheduleStory(storyId, successFn) {
    fetch('/schedule_story', {
        method: 'POST', // or 'PUT'
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({ "story": storyId })
    }).then((response) => response.json())
        .then((data) => {
            console.log('Story scheduled:', data);
            if (successFn) {
                successFn(data);
            }
        })
        .catch((error) => {
            console.error('Error:', error);
        });
}

function unscheduleStory(entryId, storyId, successFn) {
    fetch('/unschedule_story', {
        method: 'POST', // or 'PUT'
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({ "story": storyId, "id": entryId })
    }).then((response) => response.json())
        .then((data) => {
            console.log('Story scheduled:', data);
            if (successFn) {
                successFn(data);
            }
        })
        .catch((error) => {
            console.error('Error:', error);
        });
}

function getLog(successFn) {
    fetch('/log', {
        method: 'GET', // or 'PUT'
        headers: {
            'Content-Type': 'application/json',
        }
    }).then((response) => response.json())
        .then((data) => {
            console.log('Log:', data);
            successFn(data);
        })
        .catch((error) => {
            console.error('Error:', error);
        });
}

function adjustSchedule(entryId, adjustment, successFn) {
    fetch('/schedule_adjust', {
        method: 'POST', // or 'PUT'
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({ "offset_adjustment": adjustment, "id": entryId })
    }).then((response) => response.json())
        .then((data) => {
            console.log('Log:', data);
            if (successFn) {
                successFn(data);
            }
        })
        .catch((error) => {
            console.error('Error:', error);
        });
}

function logStory(entryId, storyId, wordIds, successFn) {
    fetch('/log_story', {
        method: 'POST', // or 'PUT'
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({ "story": storyId, "id": entryId, "words": wordIds })
    }).then((response) => response.json())
        .then((data) => {
            console.log('Log:', data);
            if (successFn) {
                successFn(data);
            }
        })
        .catch((error) => {
            console.error('Error:', error);
        });
}

function timeSince(date) {
    if (date <= 1) {
        return 'never';
    }

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
    return 'just now';
}

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

// todo test with negative time
function formatTrackTime(time) {
    let seconds = Math.trunc(time);

    let fractionStr = '000';
    let arr = String(time).split('.');
    if (arr.length > 1) {
        fractionStr = arr[1].substring(0, 3).padEnd(3, '0');
    }

    let secondsStr = String(seconds % 60).padStart(2, '0');
    let minutesStr = String(Math.trunc(seconds / 60) % 60).padStart(2, '0');
    let hoursStr = String(Math.trunc(seconds / (60 * 60))).padStart(2, '0');

    return `${hoursStr}:${minutesStr}:${secondsStr}.${fractionStr}`;
}

function textTrackToString(track) {
    let vtt = 'WEBVTT\n\n';

    for (let cue of track.cues) {
        vtt += `${cue.id}
${formatTrackTime(cue.startTime)} --> ${formatTrackTime(cue.endTime)}
${cue.text}\n\n`;
    }

    return vtt;
}


// add adjustment to every start and end timing, but only those which overlap or come after 'time'
function adjustTextTrackTimings(track, time, adjustment) {
    let index = track.cues.length;

    // find first track that overlap afterTime
    for (let i = 0; i < track.cues.length; i++) {
        let cue = track.cues[i];

        if (cue.endTime > time) {
            // return false if adjusting the cue would make it overlap the prior cue
            let prior = track.cues[i - 1];
            let adjustedStart = cue.startTime + adjustment;
            if (adjustedStart < prior.endTime || adjustedStart < 0) {
                return false;
            }

            index = i;
            break;
        }
    }

    for (let i = index; i < track.cues.length; i++) {
        let cue = track.cues[i];
        cue.startTime += adjustment;
        cue.endTime += adjustment;
    }

    return true;
}

// find all cues for which time is between the start and end
function findCues(track, time) {
    let cues = [];
    for (let cue of track.cues) {
        if (cue.startTime <= time && time <= cue.endTime) {
            cues.push(cue);
        }
    }
    return cues;
}


function stringToColor(str) {
    let hash = 0;
    str.split('').forEach(char => {
        hash = char.charCodeAt(0) + ((hash << 5) - hash)
    });
    let vals = []
    for (let i = 0; i < 3; i++) {
        const value = (hash >> (i * 8)) & 0xff;
        vals[i] = value;
    }
    let h = vals[0] * (360 / 256);
    let s = vals[1] * (50 / 256) + 50;  // scale to value between 50 and 100
    let l = vals[2] * (50 / 256) + 50;  // scale to value between 50 and 100

    return hslToHex(h, s, l);
}

function integerHash(str) {
    let hash = 0;
    str.split('').forEach(char => {
        hash = char.charCodeAt(0) + ((hash << 5) - hash)
    });
    return hash;
}

function randomInteger(max) {
    return Math.floor(Math.random() * (max + 1));
}

function hslToHex(h, s, l) {
    l /= 100;
    const a = s * Math.min(l, 1 - l) / 100;
    const f = n => {
        const k = (n + h / 30) % 12;
        const color = l - a * Math.max(Math.min(k - 3, 9 - k, 1), -1);
        return Math.round(255 * color).toString(16).padStart(2, '0');   // convert to Hex and prefix "0" if needed
    };
    return `#${f(0)}${f(8)}${f(4)}`;
}

var colorPalette = ['#c7522a', '#e5c185', '#fbf2c4', '#74a892', "#d9042b", 
    "#730220", "#03658c", "#f29f05", "#f27b50", "#c7522a", "#e5c185", 
    "#f0daa5", "#fbf2c4", "#b8cdab", "#74a892", "#008585" 
];

function randomPaletteColor(hash) {
    let idx = Math.abs(hash) % colorPalette.length;
    return colorPalette[idx];
}