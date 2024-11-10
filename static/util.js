const REP_COOLDOWN = 60 * 60 * 18;  // 18 hours (in seconds)
const STORY_LOG_COOLDOWN = 60 * 60 * 18;  // 18 hours (in seconds)


const DRILL_CATEGORY_KATAKANA = 1;
const DRILL_CATEGORY_ICHIDAN = 2;
const DRILL_CATEGORY_GODAN_SU = 8;
const DRILL_CATEGORY_GODAN_RU = 16;
const DRILL_CATEGORY_GODAN_U = 32;
const DRILL_CATEGORY_GODAN_TSU = 64;
const DRILL_CATEGORY_GODAN_KU = 128;
const DRILL_CATEGORY_GODAN_GU = 256;
const DRILL_CATEGORY_GODAN_MU = 512;
const DRILL_CATEGORY_GODAN_BU = 1024;
const DRILL_CATEGORY_GODAN_NU = 2048;
const DRILL_CATEGORY_KANJI = 4096;
const DRILL_CATEGORY_GODAN = DRILL_CATEGORY_GODAN_SU | DRILL_CATEGORY_GODAN_RU | DRILL_CATEGORY_GODAN_U | DRILL_CATEGORY_GODAN_TSU |
	DRILL_CATEGORY_GODAN_KU | DRILL_CATEGORY_GODAN_GU | DRILL_CATEGORY_GODAN_MU | DRILL_CATEGORY_GODAN_BU | DRILL_CATEGORY_GODAN_NU;

function isVerb(word) {
    return (word.category & (DRILL_CATEGORY_GODAN | DRILL_CATEGORY_ICHIDAN)) > 0;
}

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

function updateWord(word, successFn) {
    fetch('/update_word', {
        method: 'POST', // or 'PUT'
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(word),
    }).then((response) => response.json()
    ).then((data) => {
        if (successFn) {
            successFn(data);
        }
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

function updateSubtitles(story, successFn) {
    story = {
        id: story.id,
        source: story.source,
        title: story.title,
        subtitles_en: JSON.stringify(story.subtitles_en, null, 4),
        subtitles_ja: JSON.stringify(story.subtitles_ja, null, 4)
    };
    fetch(`/update_subtitles`, {
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

function updateExcerpts(story, successFn) {
    // todo: can remove this once legacy stories all have excerpts with hashes
    for (let ex of story.excerpts) {
        if (!ex.hash) {
            ex.hash = Math.floor(Math.random() * MAX_INTEGER + 1);  // random value [1, MAX_INTEGER];
        }
    }

    fetch('/update_excerpts', {
        method: 'POST', // or 'PUT'
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({ "story_id": story.id, "excerpts": story.excerpts })
    }).then((response) => response.json())
        .then((data) => {
            console.log('updated story excerpts:', data);
            if (successFn) {
                successFn(data);
            }
        })
        .catch((error) => {
            console.error('Error:', error);
        });
}

function getSources(successFn) {
    fetch('/sources', {
        method: 'GET',
        headers: {
            'Content-Type': 'application/json',
        }
    }).then((response) => response.json())
        .then((data) => {
            console.log('Sources:', data);
            successFn(data);
        })
        .catch((error) => {
            console.error('Error:', error);
        });
}

function getIP(successFn) {
    fetch('/ip', {
        method: 'GET',
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

function logStory(storyId, date, successFn) {
    fetch('/log_story', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({"story_id": storyId, "date": date})
    }).then((response) => response.json())
        .then((data) => {
            console.log('Story logged');
            successFn(data);
        })
        .catch((error) => {
            console.error('Error:', error);
        });
}

function importSource(source, successFn) {
    fetch('/import_source', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({"source": source})
    }).then((response) => response.json())
        .then((data) => {
            console.log('Source:', data);
            successFn(data);
        })
        .catch((error) => {
            console.error('Error:', error);
        });
}

function importStory(source, title, successFn) {
    fetch('/import_story', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({"source": source, "story_title": title})
    }).then((response) => response.json())
        .then((data) => {
            console.log('Story imported:', data);
            successFn(data);
        })
        .catch((error) => {
            console.error('Error:', error);
        });
}

function openTranscript(sourceName, storyName, lang, successFn) {
    fetch('/open_transcript', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({"source_name": sourceName, "story_name": storyName, "lang": lang})
    }).then((response) => response.json())
        .then((data) => {
            console.log('Open transcript response:', data);
            successFn(data);
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


function timeSinceRep(date) {
    if (date <= 1) {
        return '';
    }

    let now = Math.floor(new Date() / 1000);
    let elapsedSeconds = now - date;
    if (elapsedSeconds > 0) {
        elapsedSeconds++; // adding one second fixes cases like "24 hours" instead of "1 day";
        let interval = elapsedSeconds / 31536000;
        if (interval > 1) {
            var val = Math.floor(interval);
            return `(${val} ${val == 1 ? 'year' : 'years'} since last rep)`;
        }
        interval = elapsedSeconds / 2592000;
        if (interval > 1) {
            var val = Math.floor(interval);
            return `(${val} ${val == 1 ? 'month' : 'months'} since last rep)`;
        }
        interval = elapsedSeconds / 86400;
        if (interval > 1) {
            var val = Math.floor(interval);
            return `(${val} ${val == 1 ? 'day' : 'days'} since last rep)`;
        }
        interval = elapsedSeconds / 3600;
        if (interval > 1) {
            var val = Math.floor(interval);
            return `(${val} ${val == 1 ? 'hour' : 'hours'} since last rep)`;
        }
        interval = elapsedSeconds / 60;
        if (interval > 1) {
            var val = Math.floor(interval);
            return `(${val} ${val == 1 ? 'minute' : 'minutes'} since last rep)`;
        }
        if (interval > 1) {
            var val = Math.floor(elapsedSeconds);
            return `(${val} ${val == 1 ? 'second' : 'seconds'} since last rep)`;
        }
    } 
    return '(last rep completed just now)';
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
function formatTrackTime(time, padding = 3, includeHours = false) {
    let seconds = Math.trunc(time);

    let fractionStr = '';
    let arr = time.toFixed(padding).split('.');
    if (arr.length > 1) {
        fractionStr = arr[1].padEnd(padding, '0');
    } else {
        fractionStr = ''.padEnd(padding, '0');
    }

    let secondsStr = String(seconds % 60).padStart(2, '0');
    let minutesStr = String(Math.trunc(seconds / 60) % 60).padStart(2, '0');

    let str = `${minutesStr}:${secondsStr}`;

    let hours = Math.trunc(seconds / (60 * 60));
    if (includeHours || hours > 0) {
        let hoursStr = String(hours).padStart(2, '0');
        str = hoursStr + ':' + str;
    }

    if (padding > 0) {
        str += '.' + fractionStr;
    }

    return str;
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
${formatTrackTime(cue.startTime, 3, true)} --> ${formatTrackTime(cue.endTime, 3, true)}
${cue.text}\n\n`;
    }

    return vtt;
}


// add adjustment to every start and end timing, but only those which overlap or come after 'time'
function adjustTextTrackTimings(cues, adjustment, time) {
    time = time || 0;
    
    // find first cue that ends after time
    let index = cues.length;
    for (let i = 0; i < cues.length; i++) {
        let cue = cues[i];
        if (cue.end_time > time) {
            index = i;
            break;
        }
    }

    // we have to work with a copy of the cues because setting startTime or endTime will reorder track.cues

    let copy = []; 
    copy.length = cues.length;
    for (let i = 0; i < cues.length; i++) {
        copy[i] = cues[i];
    }

    for (let i = index; i < copy.length; i++) {
        let cue = copy[i];
        console.log("adjustment: ", adjustment, cue.text);

        cue.start_time += adjustment;
        cue.end_time += adjustment;
    }

    for (let i = 0; i < cues.length; i++) {
        cues[i] = copy[i];
    }
}

// add adjustment to every start and end timing
function adjustTextTrackAllTimings(cues, adjustment) {
    // do nothing if first cue start time would be less than 0
    let firstCue = cues[0];
    if (firstCue.start_time + adjustment < 0) {
        return false;
    }

    // todo shouldn't need to use a copy now that we aren't using tracks

    let copy = []; 
    copy.length = cues.length;
    for (let i = 0; i < cues.length; i++) {
        copy[i] = cues[i];
    }

    for (let i = 0; i < copy.length; i++) {
        let cue = copy[i];
        cue.start_time += adjustment;
        cue.end_time += adjustment;
    }

    for (let i = 0; i < copy.length; i++) {
        copy[i] = cues[i];
    }

    return true;
}

// move next cue after time up to time, and move all cues after up the same amount
function bringForwardTextTrackTimings(cues, time) {
    let found = false;
    let index = 0;

    // find first cue which comes after time
    for (let i = 0; i < cues.length; i++) {
        let cue = cues[i];

        // do nothing if a track overlaps the time
        if (cue.start_time < time && cue.end_time > time) {
            return false;
        }

        if (cue.start_time > time) {
            index = i;
            found = true;
            break;
        }
    }

    if (!found) {
        return;
    }

    let adjustment = time - cues[index].start_time;

    console.log('adjustment', adjustment);
    
    // todo shouldn't have to do a copy anymore

    let copy = []; 
    copy.length = cues.length;
    for (let i = 0; i < cues.length; i++) {
        copy[i] = cues[i];
    }

    for (let i = index; i < copy.length; i++) {
        let cue = copy[i];
        cue.start_time += adjustment;
        cue.end_time += adjustment;
    }

    for (let i = 0; i < copy.length; i++) {
        copy[i] = cues[i];
    }

    return true;
}

// find all cues for which time is between the start and end
function findCues(cues, time) {
    let selected = [];
    for (let cue of cues) {
        if (cue.start_time <= time && time < cue.end_time) {
            selected.push(cue);
        }
    }
    return selected;
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

function setCookie(name,value,days) {
    var expires = "";
    if (days) {
        var date = new Date();
        date.setTime(date.getTime() + (days*24*60*60*1000));
        expires = "; expires=" + date.toUTCString();
    }
    document.cookie = name + "=" + (value || "")  + expires + "; path=/ ; SameSite=Lax; ";
}

function getCookie(name) {
    var nameEQ = name + "=";
    var ca = document.cookie.split(';');
    for(var i=0;i < ca.length;i++) {
        var c = ca[i];
        while (c.charAt(0)==' ') c = c.substring(1,c.length);
        if (c.indexOf(nameEQ) == 0) return c.substring(nameEQ.length,c.length);
    }
    return null;
}