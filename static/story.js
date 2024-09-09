var storyLines = document.getElementById('story_lines');
var wordList = document.getElementById('word_list');
var playerSpeedNumber = document.getElementById('player_speed_number');
var player = document.getElementById('video_player');
var captionsJa = document.getElementById('captions_ja');
var captionsEn = document.getElementById('captions_en');
var repetitionsInfoDiv = document.getElementById('repetitions_info');
var storyActions = document.getElementById('story_actions');

var englishCheckbox = document.getElementById('transcript_en_checkbox');
var japaneseCheckbox = document.getElementById('transcript_ja_checkbox');

var playerControls = document.getElementById('player_controls');

var story = null;

var cueGuideElement = document.getElementById('captions_meter');
var cueGuideIndicator = document.getElementById('captions_meter_indicator');

const TEXT_TRACK_TIMING_ADJUSTMENT = 0.2;
const TEXT_TRACK_TIMING_PUSH_BACK_ADJUSTMENT = 10;
const PLAYBACK_ADJUSTMENT = 0.05;

const MAX_INTEGER = Math.pow(2, 52) - 1;

englishCheckbox.addEventListener('change', displaySubtitles);
japaneseCheckbox.addEventListener('change', displaySubtitles);

storyLines.onwheel = function (evt) {
    evt.preventDefault();
    let scrollDelta = evt.wheelDeltaY * 2;
    storyLines.scrollTop -= scrollDelta;
};

storyLines.onclick = function (evt) {
    if (evt.target.classList.contains('subtitle_start_time')) {
        evt.preventDefault();
        let subtitleContainer = evt.target.closest('div[subtitle_index]');
        if (!subtitleContainer) {
            return;
        }
        let idx = subtitleContainer.getAttribute('subtitle_index');
        let lang = subtitleContainer.getAttribute('subtitle_lang');

        var cues = (lang == 'ja') ? story.subtitles_ja : story.subtitles_en;

        var cue = cues[idx];

        player.currentTime = cue.start_time;

    }
};

storyActions.onclick = function (evt) {
    if (evt.target.classList.contains('open_transcript')) {
        let lang = evt.target.classList.contains('en') ? 'en' : 'ja';
        openTranscript(story.source, story.title, lang, () => { });
        return;
    } else if (evt.target.classList.contains('reimport')) {
        if (!window.confirm("Reimport the story?")) {
            return;
        }
        snackbarMessage(`reimporting story`);
        importStory(story.source, story.title, function () {
            window.location.reload();
        });
        return;
    } else if (!evt.target.closest('#excerpts')) {
        return;
    }

    if (evt.target.classList.contains('add_excerpt')) {
        evt.preventDefault();
        let hash = Math.floor(Math.random() * MAX_INTEGER + 1);  // random value [1, MAX_INTEGER];
        story.excerpts.push({ "start_time": 0, "end_time": player.duration, "reps_logged": [], "reps_todo": 0, "hash": hash });
        updateExcerpts(story,
            function () {
                displayStoryInfo(story);
                snackbarMessage(`added excerpt`);
            }
        );
    } else if (evt.target.classList.contains('sort_excerpts')) {
        evt.preventDefault();
        story.excerpts.sort((a, b) => {
            if (a.start_time < b.start_time) {
                return -1;
            } else if (a.start_time > b.start_time) {
                return +1;
            }

            // use end time as secondary criterea
            if (a.end_time < b.end_time) {
                return -1;
            } else if (a.end_time > b.end_time) {
                return +1;
            }

            return 0;
        });
        updateExcerpts(story,
            function () {
                displayStoryInfo(story);
                snackbarMessage(`reordered the excerpts by start time`);
            }
        );
    }

    container = evt.target.closest('div[excerpt_idx]');
    if (!container) {
        return;
    }

    let excerptIdx = parseInt(container.getAttribute('excerpt_idx'));
    if (excerptIdx > story.excerpts.length - 1) {
        console.log("invalid excerpt idx");
    }
    let excerpt = story.excerpts[excerptIdx];

    if (evt.target.classList.contains('add_rep_link')) {
        evt.preventDefault();
        excerpt.reps_todo = excerpt.reps_todo || 0;
        excerpt.reps_todo++;

        updateExcerpts(story,
            function () {
                displayStoryInfo(story);
                snackbarMessage(`one rep added to queue of excerpt`);
            }
        );
    } else if (evt.target.classList.contains('remove_rep_link')) {
        evt.preventDefault();
        excerpt.reps_todo = excerpt.reps_todo || 0;
        if (excerpt.reps_todo == 0) {
            snackbarMessage(`excerpt already has no reps`);
            return;
        }
        excerpt.reps_todo--;

        updateExcerpts(story,
            function () {
                displayStoryInfo(story);
                snackbarMessage(`one rep removed from queue of excerpt`);
            }
        );
    } else if (evt.target.classList.contains('play_excerpt')) {
        let start = Math.trunc(excerpt.start_time);
        let end = Math.trunc(excerpt.end_time);
        if (excerpt.end_time == 0) {
            end = Math.trunc(player.duration);
        }
        let time = `#t=${start},${end}`;
        let path = '/sources/' + story.source + "/" + story.video;
        player.src = path + time;
        // setting src resets the playbackRate, so must set it again
        player.playbackRate = parseFloat(document.getElementById('player_speed_number').value);
        play();
        displaySubtitles();

    } else if (evt.target.classList.contains('start_time')) {
        evt.preventDefault();

        if (!window.confirm("Set the start time of the excerpt?")) {
            return;
        }

        let time = player.currentTime;
        excerpt.start_time = time;
        updateExcerpts(story,
            function () {
                displayStoryInfo(story);
                snackbarMessage(`set start time of excerpt to ${formatTrackTime(time)}`);
            }
        );
    } else if (evt.target.classList.contains('end_time')) {
        evt.preventDefault();

        if (!window.confirm("Set the end time of the excerpt?")) {
            return;
        }

        let time = player.currentTime;
        excerpt.end_time = time;
        updateExcerpts(story,
            function () {
                displayStoryInfo(story);
                snackbarMessage(`set end time of excerpt to ${formatTrackTime(time)}`);
            }
        );
    } else if (evt.target.classList.contains('delete_excerpt')) {
        evt.preventDefault();
        if (story.excerpts.length == 1) {
            snackbarMessage(`cannot remove the only excerpt`);
            return;
        }

        if (window.confirm("Do you really want to remove the excerpt?")) {
            story.excerpts.splice(excerptIdx, 1);
            updateExcerpts(story,
                function () {
                    displayStoryInfo(story);
                    snackbarMessage(`removed excerpt`);
                }
            );
        }
    } else if (evt.target.classList.contains('log_excerpt')) {
        evt.preventDefault();

        if (window.confirm("Log this excerpt?")) {
            if (logRep(excerpt)) {
                updateExcerpts(story, function () {
                    load();
                    snackbarMessage(`rep of excerpt logged`);
                });
            }
        }
    }
};

var subtitleAdjustTimeoutHandle = 0;

document.body.onkeydown = async function (evt) {
    if (evt.ctrlKey) {
        return;
    }
    // console.log(evt, 'code', evt.code);
    if (!player) {
        return;
    }


    let timemark = player.currentTime;
    if (evt.code === 'KeyF') {
        evt.preventDefault();
        var playerDiv = document.getElementById('player');
        if (document.fullscreenElement == null) {
            playerDiv.requestFullscreen();
        } else {
            document.exitFullscreen();
        }
    } else if (evt.code === 'KeyA') {
        evt.preventDefault();
        let newTime = timemark - 1.8;
        player.currentTime = newTime;
        displaySubtitles();
    } else if (evt.code === 'KeyD') {
        evt.preventDefault();
        let newTime = timemark + 1;
        player.currentTime = newTime;
        displaySubtitles();
    } else if (evt.code === 'KeyQ') {
        evt.preventDefault();
        let newTime = timemark - 6;
        player.currentTime = newTime;
        displaySubtitles();
    } else if (evt.code === 'KeyE') {
        evt.preventDefault();
        let newTime = timemark + 5;
        player.currentTime = newTime;
        displaySubtitles();
    } else if (evt.code === 'KeyX') {
        // jump to start of current subtitle

        let cues = null;
        if (document.getElementById('transcript_en_checkbox').checked) {
            cues = story.subtitles_en;
        }
        if (document.getElementById('transcript_ja_checkbox').checked) {
            cues = story.subtitles_ja;
        }
        if (cues === null) {
            return;
        }

        for (let index = cues.length - 1; index >= 0; index--) {
            const cue = cues[index];
            if (cue.start_time <= player.currentTime) {
                player.currentTime = cue.start_time;
                break;
            }
        }
    } else if (evt.code === 'KeyC') {
        // jump to start of next subtitle

        let cues = null;
        if (document.getElementById('transcript_en_checkbox').checked) {
            cues = story.subtitles_en;
        }
        if (document.getElementById('transcript_ja_checkbox').checked) {
            cues = story.subtitles_ja;
        }
        if (cues === null) {
            return;
        }

        for (let index = 0; index < cues.length; index++) {
            const cue = cues[index];
            if (cue.start_time > player.currentTime) {
                player.currentTime = cue.start_time;
                break;
            }
        }
    } else if (evt.code === 'KeyZ') {
        // jump to start of prior subtitle

        let cues = null;
        if (document.getElementById('transcript_en_checkbox').checked) {
            cues = story.subtitles_en;
        }
        if (document.getElementById('transcript_ja_checkbox').checked) {
            cues = story.subtitles_ja;
        }
        if (cues === null) {
            return;
        }

        for (let index = cues.length - 1; index >= 0; index--) {
            const cue = cues[index];
            if (cue.end_time < player.currentTime) {
                player.currentTime = cue.start_time;
                break;
            }
        }
    } else if (evt.code === 'KeyP' || evt.code === 'KeyS') {
        evt.preventDefault();
        currentCue = null;
        if (player.paused) {  // playing
            play();
        } else {
            player.pause();
        }
    } else if (evt.code === 'Equal' || evt.code === 'Minus') {
        evt.preventDefault();

        if (evt.altKey) {
            let adjustment = (evt.code === 'Equal') ? TEXT_TRACK_TIMING_ADJUSTMENT : -TEXT_TRACK_TIMING_ADJUSTMENT;
            let lang = 'English and Japanese';

            let english = document.getElementById('transcript_en_checkbox').checked;
            let japanese = document.getElementById('transcript_ja_checkbox').checked;

            if (!english && !japanese) {
                return;
            }

            if (english) {
                lang = 'English';
                adjustTextTrackAllTimings(story.subtitles_en, adjustment);
            }

            if (japanese) {
                lang = 'Japanese';
                adjustTextTrackAllTimings(story.subtitles_ja, adjustment);
            }

            displaySubtitles();

            snackbarMessage(`updated ${lang} subtitle timings by ${adjustment} seconds`);

            clearTimeout(subtitleAdjustTimeoutHandle);
            subtitleAdjustTimeoutHandle = setTimeout(
                function () {
                    updateSubtitles(story, () => {
                        snackbarMessage(`saved updates to subtitle timings`);
                    });
                },
                3000
            );
        } else {
            let adjustment = (evt.code === 'Equal') ? PLAYBACK_ADJUSTMENT : -PLAYBACK_ADJUSTMENT;
            adjustPlaybackSpeed(adjustment);
        }
    } else if (evt.code === 'BracketLeft' && evt.altKey) {
        evt.preventDefault();
        let adjustment = TEXT_TRACK_TIMING_PUSH_BACK_ADJUSTMENT;
        let lang = 'English and Japanese';

        let english = document.getElementById('transcript_en_checkbox').checked;
        let japanese = document.getElementById('transcript_ja_checkbox').checked;

        if (!english && !japanese) {
            return;
        }

        if (english) {
            lang = 'English';
            bringForwardTextTrackTimings(story.subtitles_en, player.currentTime);
        }

        if (japanese) {
            lang = 'Japanese ';
            bringForwardTextTrackTimings(story.subtitles_ja, player.currentTime);
        }

        displaySubtitles();

        snackbarMessage(`updated ${lang} subtitle timings past the current mark by ${adjustment} seconds`);

        clearTimeout(subtitleAdjustTimeoutHandle);
        subtitleAdjustTimeoutHandle = setTimeout(
            function () {
                updateSubtitles(story, () => {
                    snackbarMessage(`saved updates to subtitle timings`);
                });
            },
            3000
        );
    } else if (evt.code === 'BracketRight' && evt.altKey) {
        evt.preventDefault();
        let lang = 'English and Japanese';

        let english = document.getElementById('transcript_en_checkbox').checked;
        let japanese = document.getElementById('transcript_ja_checkbox').checked;

        if (!english && !japanese) {
            return;
        }

        if (english) {
            lang = 'English ';
            adjustTextTrackTimings(story.subtitles_en, TEXT_TRACK_TIMING_PUSH_BACK_ADJUSTMENT, player.currentTime);
        }

        if (japanese) {
            lang = 'Japanese ';
            adjustTextTrackTimings(story.subtitles_ja, TEXT_TRACK_TIMING_PUSH_BACK_ADJUSTMENT, player.currentTime);
        }

        displaySubtitles();

        snackbarMessage(`updated ${lang} subtitle timings past the current mark by ${TEXT_TRACK_TIMING_PUSH_BACK_ADJUSTMENT} seconds`);

        clearTimeout(subtitleAdjustTimeoutHandle);
        subtitleAdjustTimeoutHandle = setTimeout(
            function () {
                updateSubtitles(story, () => {
                    snackbarMessage(`saved updates to subtitle timings`);
                });
            },
            3000
        );
    } else if (evt.code.startsWith('Digit')) {
        if (evt.altKey) {
            evt.preventDefault();
            let digit = parseInt(evt.code.slice(-1));
            let duration = player.duration;
            player.currentTime = duration * (digit / 10);
            displaySubtitles();
        }
    } else if (evt.code.startsWith('Space')) {
        evt.preventDefault();

        if (currentCue) {
            player.currentTime = currentCue.start_time;
            play();
        }
    }
};

function play() {
    displaySubtitles();
    player.play();
}


// returns time in seconds
function parseTimestamp(timestamp) {
    let [mins, seconds] = timestamp.split(':');
    mins = parseInt(mins);
    seconds = parseFloat(seconds);
    return mins * 60 + seconds;
}

function displaySubtitles() {
    if (document.getElementById('transcript_en_checkbox').checked) {
        captionsEn.style.display = 'flex';
    } else {
        captionsEn.style.display = 'none';
    }

    if (document.getElementById('transcript_ja_checkbox').checked) {
        captionsJa.style.display = 'flex';
    } else {
        captionsJa.style.display = 'none';
    }

    let enSubtitles = findCues(story.subtitles_en, player.currentTime);
    let jaSubtitles = findCues(story.subtitles_ja, player.currentTime);

    displayCues(enSubtitles, captionsEn);
    displayCues(jaSubtitles, captionsJa);
}

function displayCues(cues, target) {
    let html = '';

    // because of overlap, more than one cue can be active
    for (let i = 0; i < cues.length; i++) {
        let cue = cues[i];
        let lines = cue.text.split('\n');
        for (let line of lines) {
            html += `<div>${line}</div>`;
        }
    }

    if (cues.length == 0) {
        target.style.visibility = 'hidden';
    } else {
        target.style.visibility = 'visible';
    }

    updateCueGuide(cues);

    target.innerHTML = html;
}


function displayExcerpts(story) {
    function repsHTML(excerpt, excerptIdx) {
        let timeLastRep = 1;
        let listeningRepCount = 0;

        if (excerpt.reps_logged) {
            for (let rep of excerpt.reps_logged) {
                listeningRepCount++;

                if (rep.date > timeLastRep) {
                    timeLastRep = rep.date;
                }
            }
        }

        let todoReps = `Queued reps: <a class="remove_rep_link" href="#" title="remove a rep">−</a>
            <a class="add_rep_link" href="#" title="add a rep">＋</a>`;
        for (let i = 0; i < excerpt.reps_todo; i++) {
            todoReps += `<span class="listening rep" title="rep">⭯</span>`;
        }

        let html = `<div excerpt_idx="${excerptIdx}">
            <hr>
            <minidenticon-svg username="seed${excerpt.hash}"></minidenticon-svg>
            <a class="play_excerpt" href="#" title="play the excerpt">play</a>
            <a class="start_time" href="#" title="click to set the start time">${formatTrackTime(excerpt.start_time, 0)}</a>-<a class="end_time" href="#" title="click to set the end time">${formatTrackTime(excerpt.end_time, 0)}</a>
            <a class="drill_excerpt" href="words.html?storyId=${story.id}&excerptHash=${excerpt.hash}" title="Drill the vocab of this excerpt">vocab</a>
            <a class="delete_excerpt" href="#" title="Remove this excerpt">remove</a>
            <br>
            <span>Completed reps: ${listeningRepCount} &nbsp;&nbsp; ${timeSinceRep(timeLastRep)}</span><br>
            ${todoReps} <a class="log_excerpt" href="#" title="Log a rep for this excerpt">log</a><br>`;

        return html + '</div>';
    }

    let html = `Excerpts:
    <a class="add_excerpt" href="#" title="Add a new excerpt">add excerpt</a>
    <a class="sort_excerpts" href="#" title="Reorder the excerpts by start time">reorder excerpts</a>`;

    for (idx in story.excerpts) {
        html += repsHTML(story.excerpts[idx], idx);
    }

    document.getElementById('excerpts').innerHTML = html;
}

function displayStoryInfo(story) {
    document.getElementById('story_title').innerHTML = `<a href="${story.link}">${story.title}</a><hr>`;
    document.getElementById('source_info').innerText = 'Source: ' + story.source;
    if (story.date) {
        document.getElementById('date_info').innerText = story.date;
    }

    displayExcerpts(story);
}

function processStory(story) {
    story.reps_logged = story.reps_logged || [];
    story.reps_todo = story.reps_todo || [];

    for (let excerpt of story.excerpts) {
        excerpt.start_time = excerpt.start_time || 0;
        excerpt.end_time = excerpt.end_time || player.duration;
    }

    story.subtitles_en = JSON.parse(story.subtitles_en);
    story.subtitles_ja = JSON.parse(story.subtitles_ja);
}

function openStory(id) {
    fetch('/story/' + id, {
        method: 'GET', // or 'PUT'
        headers: {
            'Content-Type': 'application/json',
        }
    }).then((response) => response.json())
        .then((data) => {
            story = data;

            processStory(story);
            displayStoryInfo(story);

            if (story.video) {
                player.style.display = 'block';

                let path = '/sources/' + story.source + "/" + story.video;
                if (!path.endsWith('mp4')) {
                    player.style.height = '60px';
                }

                player.setAttribute('type', 'video/mp4');
                
                let time = '';
                if (story.end_time > 0) {
                    time = `#t=${Math.trunc(story.start_time)},${Math.trunc(story.end_time)}`;
                }
                player.src = path + time;

                //console.log("src", player.src);
            }

            displayStoryContent(story);

            playerControls.style.display = 'inline';
        })
        .catch((error) => {
            console.error('Error:', error);
        });
}


playerSpeedNumber.onchange = function (evt) {
    if (player) {
        player.playbackRate = parseFloat(playerSpeedNumber.value);
    }
}

function adjustPlaybackSpeed(adjustment) {
    let newSpeed = parseFloat(playerSpeedNumber.value) + adjustment
    playerSpeedNumber.value = newSpeed.toFixed(2);

    if (player) {
        player.playbackRate = newSpeed;
    }
}

function displayStoryContent() {
    let cues = null;
    let lang = 'en';
    if (document.getElementById('transcript_en_checkbox').checked) {
        cues = story.subtitles_en;
    }
    if (document.getElementById('transcript_ja_checkbox').checked) {
        lang = 'ja';
        cues = story.subtitles_ja;
    }
    if (cues == null) {
        storyLines.innerHTML = "";
        return;
    }

    let html = '';
    //console.log(cues);
    for (let cueIndex = 0; cueIndex < cues.length; cueIndex++) {
        let cue = cues[cueIndex];

        let lineSpans = ``;

        if (cue.text) {
            let lines = cue.text.split('\n');
            for (let line of lines) {
                lineSpans += `<span class="subtitle_line">${line}</span>`;
            }
        } else {
            console.log('cue with no text', cue);
        }

        let startTime = formatTrackTime(cue.start_time, 2);

        html += `<div class="subtitle" subtitle_index="${cueIndex}" subtitle_lang="${lang}">
            <div class="subtitle_start"><a href="#" class="subtitle_start_time">${startTime}</a></div>
            <div class="subtitle_lines">${lineSpans}</div>
        </div>`;
    }

    storyLines.innerHTML = html;
}

document.body.onload = function (evt) {
    load();
};

function load(evt) {
    var url = new URL(window.location.href);
    var storyId = parseInt(url.searchParams.get("storyId") || undefined);
    openStory(storyId);
}


function updateCueGuide(cues) {
    if (cues == null || cues.length == 0) {
        document.getElementById('captions_meter').style.visibility = 'hidden';
        return;
    }

    // because of overlap, more than one cue can be active
    let longestCue = null;
    let longestDuration = 0;
    for (let i = 0; i < cues.length; i++) {
        let cue = cues[i];
        let duration = cue.end_time - cue.start_time;
        if (duration > longestDuration) {
            longestDuration = duration;
            longestCue = cue;
        }
    }

    if (longestDuration > 0) {
        document.getElementById('captions_meter').style.visibility = 'visible';
        displayCueGuide(longestCue);
    }
}

function displayCueGuide(cue) {
    let duration = cue.end_time - cue.start_time;
    if (duration <= 0) {
        return;
    };

    // this can happen some times when seeking because of the event queue order
    if (player.currentTime < cue.start_time || player.currentTime > cue.end_time) {
        return;
    }

    const widthPercentagePointsPerSecond = 3;
    const maxWidth = 90;
    const minWidth = 5;
    let width = duration * widthPercentagePointsPerSecond;
    width = Math.min(Math.max(width, minWidth), maxWidth); // clamp 

    cueGuideElement.style.width = width + '%';

    let cueProgress = (player.currentTime - cue.start_time) / duration;
    cueGuideIndicator.style.marginLeft = (cueProgress * 100).toFixed(5) + '%';
}

player.addEventListener("seeked", (evt) => {
    displaySubtitles();
});

var intervalHandle;
const deltaTime = 0.01;  // in seconds

player.addEventListener("playing", (event) => {
    window.clearInterval(intervalHandle); // just in case
    intervalHandle = window.setInterval(function () {
        displaySubtitles();
    }, deltaTime * 1000);
});

player.addEventListener("pause", (event) => {
    window.clearInterval(intervalHandle);
});
