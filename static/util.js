const READING = 0;
const LISTENING = 1;
const DRILLING = 2;

const DEFAULT_REPS = [
    LISTENING, DRILLING,
    LISTENING, DRILLING,
    LISTENING, DRILLING,
    LISTENING
];

const REP_COOLDOWN = 60 * 60 * 18;  // 18 hours (in seconds)

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

function updateStory(story, successFn) {
    story = {
        id: story.id,
        source: story.source,
        title: story.title,
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

function getStories(successFn) {
    fetch('/stories', {
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


function getReps(successFn) {
    fetch('/reps', {
        method: 'GET', // or 'PUT'
        headers: {
            'Content-Type': 'application/json',
        }
    }).then((response) => response.json())
        .then((data) => {
            console.log('Reps:', data);
            if (successFn) {
                successFn(data);
            }
        })
        .catch((error) => {
            console.error('Error:', error);
        });
}

function updateReps(story, successFn) {
    fetch('/update_reps', {
        method: 'POST', // or 'PUT'
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({ "story_id": story.id, "reps_logged": story.reps_logged, "reps_todo": story.reps_todo})
    }).then((response) => response.json())
        .then((data) => {
            console.log('Update reps:', data);
            if (successFn) {
                successFn(data);
            }
        })
        .catch((error) => {
            console.error('Error:', error);
        });
}

function incWords(words, successFn) {
    fetch('/inc_words', {
        method: 'POST', // or 'PUT'
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({"words": words})
    }).then((response) => response.json())
        .then((data) => {
            console.log('Update reps:', data);
            if (successFn) {
                successFn(data);
            }
        })
        .catch((error) => {
            console.error('Error:', error);
        });
}

function addStoryReps(storyId, reps, successFn) {
    fetch('/add_reps', {
        method: 'POST', // or 'PUT'
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({ "story_id": storyId, "reps": reps })
    }).then((response) => response.json())
        .then((data) => {
            console.log('Res added to story:', data);
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

function getIP(successFn) {
    fetch('/ip', {
        method: 'GET', // or 'PUT'
        headers: {
            'Content-Type': 'application/json',
        }
    }).then((response) => response.json())
        .then((data) => {
            console.log('IP:', data);
            successFn(data);
        })
        .catch((error) => {
            console.error('Error:', error);
        });
}

function adjustSchedule(entryId, adjustment, successFn) {
    fetch('/schedule_adjust', {
        method: 'POST',
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

function scheduleAddRep(entryId, successFn) {
    fetch('/schedule_add', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({ "id": entryId })
    }).then((response) => response.json())
        .then((data) => {
            console.log('Log:', data);
            if (data.status != "success") {
                snackbarMessage("new rep not added because the next day already has a rep of this story");
                return;
            }
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
        method: 'POST',
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
function formatTrackTime(time, noFraction) {
    let seconds = Math.trunc(time);

    let fractionStr = '000';
    let arr = String(time).split('.');
    if (arr.length > 1) {
        fractionStr = arr[1].substring(0, 3).padEnd(3, '0');
    }

    let secondsStr = String(seconds % 60).padStart(2, '0');
    let minutesStr = String(Math.trunc(seconds / 60) % 60).padStart(2, '0');
    let hoursStr = String(Math.trunc(seconds / (60 * 60))).padStart(2, '0');

    if (noFraction) {
        return `${hoursStr}:${minutesStr}:${secondsStr}`;
    }

    return `${hoursStr}:${minutesStr}:${secondsStr}.${fractionStr}`;
}

function textTrackToString(track, startTime, endTime) {
    if (startTime === undefined) {
        startTime = Number.MIN_VALUE;
    }
    if (endTime === undefined) {
        endTime = Number.MAX_VALUE;
    }

    let vtt = 'WEBVTT\n\n';

    for (let cue of track.cues) {
        if (cue.startTime > endTime || cue.endTime < startTime) {
            continue;
        }

        vtt += `${cue.id}
${formatTrackTime(cue.startTime)} --> ${formatTrackTime(cue.endTime)}
${cue.text}\n\n`;
    }

    return vtt;
}


// add adjustment to every start and end timing, but only those which overlap or come after 'time'
function adjustTextTrackTimings(track, time, adjustment) {
    let index = track.cues.length;

    const minMargin = 0.5;

    // find first track that overlap 'time'
    for (let i = 0; i < track.cues.length; i++) {
        let cue = track.cues[i];

        if (cue.endTime > time) {
            // return false if adjusting the cue would make it overlap the prior cue
            let prior = track.cues[i - 1];
            let adjustedStart = cue.startTime + adjustment;
            let excessOverlap = prior && adjustedStart < (prior.startTime + minMargin);
            if (excessOverlap || adjustedStart < 0) {
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

// add adjustment to every start and end timing
function adjustTextTrackAllTimings(track, adjustment) {
    // do nothing if first cue start time would be less than 0
    let firstCue = track.cues[0];
    if (firstCue.startTime + adjustment < 0) {
        return false;
    }

    for (let i = index; i < track.cues.length; i++) {
        let cue = track.cues[i];
        cue.startTime += adjustment;
        cue.endTime += adjustment;
    }

    return true;
}

// move next cue after time up to time, and move all cues after up the same amount
function bringForwardTextTrackTimings(track, time) {
    let index = track.cues.length;

    // find first cue which comes after time
    for (let i = 0; i < track.cues.length; i++) {
        let cue = track.cues[i];

        // do nothing if a track overlaps the time
        if (cue.startTime < time && cue.endTime > time) {
            return false;
        }

        if (cue.startTime > time) {
            index = i;
            break;
        }
    }

    let adjustment = track.cues[i].startTime - time;

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

function integerHash(str) {
    let hash = 0;
    str.split('').forEach(char => {
        hash = char.charCodeAt(0) + ((hash << 5) - hash)
    });
    return hash;
}

var colorPalette = ['#c7522a', '#e5c185', '#fbf2c4', '#74a892', "#d9042b",
    "#730220", "#03658c", "#f29f05", "#f27b50", "#c7522a", "#e5c185",
    "#f0daa5", "#fbf2c4", "#b8cdab", "#74a892", "#008585"
];

function randomPaletteColor(hash) {
    let idx = Math.abs(hash) % colorPalette.length;
    return colorPalette[idx];
}

function insertRep(story, repIdx) {
    let type = story.reps_todo[repIdx]; 
    story.reps_todo.splice(repIdx, 0, type);
    
	updateReps(story, function(data) {
        displayStoryInfo(story);
        snackbarMessage("removed a rep");
    });
}

function deleteRep(story, repIdx) {
    story.reps_todo.splice(repIdx, 1);
    
	updateReps(story, function(data) {
        displayStoryInfo(story);
        snackbarMessage("removed a rep");
    });
}

function toggleRepType(story, repIdx) {
    let priorType = story.reps_todo[repIdx];
    if (priorType == LISTENING) {
        story.reps_todo[repIdx] = DRILLING;
    } else if (priorType == DRILLING) {
        story.reps_todo[repIdx] = LISTENING;
    }
    
	updateReps(story, function(data) {
        displayStoryInfo(story);
        snackbarMessage("toggled type of a queued rep");
    });
}

function logRep(story, type, successFn) {
    let unixtime = Math.floor(Date.now() / 1000);
    let cooldownTime = unixtime - REP_COOLDOWN;

    if (story.reps_logged) {
        for (let rep of story.reps_logged) {
            if (rep.type == type && rep.date > cooldownTime) {
                snackbarMessage("story has already been logged within the cooldown window");
                return;
            }
        }
    }

    // remove first rep matching the type
    let i = 0;
    for (let repType of story.reps_todo) {
        if (repType == type) {
            story.reps_todo.splice(i, 1);
            story.reps_logged.push({ "date" : unixtime, "type": type});
            updateReps(story, successFn);
            return;
        }
        i++;
    }

    snackbarMessage(`no ${type == LISTENING ? 'listening' : 'drill'} reps are currently queued for this story`);
}