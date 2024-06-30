var storyLines = document.getElementById('story_lines');
var wordList = document.getElementById('word_list');
var playerSpeedNumber = document.getElementById('player_speed_number');
var player = document.getElementById('video_player');
var trackJa = document.getElementById('track_ja');
var trackEn = document.getElementById('track_en');
var captionsJa = document.getElementById('captions_ja');
var captionsEn = document.getElementById('captions_en');
var repetitionsInfoDiv = document.getElementById('repetitions_info');
var storyActions = document.getElementById('story_actions');

var playerControls = document.getElementById('player_controls');

var story = null;
var selectedLineIdx = 0;

const TEXT_TRACK_TIMING_ADJUSTMENT = 0.2;
const TEXT_TRACK_TIMING_PUSH_BACK_ADJUSTMENT = 10;
const PLAYBACK_ADJUSTMENT = 0.05;

const MAX_INTEGER = Math.pow(2, 52) - 1;

trackJa.track.addEventListener('cuechange', displayCurrentCues);
trackEn.track.addEventListener('cuechange', displayCurrentCues);

storyLines.onwheel = function (evt) {
    evt.preventDefault();
    let scrollDelta = evt.wheelDeltaY * 2;
    storyLines.scrollTop -= scrollDelta;
};

storyActions.onclick = function (evt) {
    let container = evt.target.closest('#excerpts');
    if (!container) {
        return;
    }

    if (evt.target.classList.contains('add_excerpt')) {
        evt.preventDefault();
        let hash = Math.floor(Math.random() * MAX_INTEGER + 1);  // random value [1, MAX_INTEGER];
        story.excerpts.push({ "start_time": 0, "end_time": player.duration, "reps_logged": [], "reps_todo": [], "hash": hash });
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
        player.play();
        displayCurrentCues();

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
        player.currentTime = timemark - 1.8;
        displayCurrentCues();
    } else if (evt.code === 'KeyD') {
        evt.preventDefault();
        player.currentTime = timemark + 1;
        displayCurrentCues();
    } else if (evt.code === 'KeyQ') {
        evt.preventDefault();
        player.currentTime = timemark - 5;
        displayCurrentCues();
    } else if (evt.code === 'KeyE') {
        evt.preventDefault();
        player.currentTime = timemark + 4;
        displayCurrentCues();
    } else if (evt.code === 'KeyB') {
        evt.preventDefault();
        // jump to start of previous subtitle

        let english = document.getElementById('transcript_en_checkbox').checked;
        let japanese = document.getElementById('transcript_ja_checkbox').checked;

        if (!english && !japanese) {
            return;
        }

        let prevEn = -1;
        if (english) {
            for (let cue of trackEn.track.cues) {
                if (cue.endTime >= player.currentTime) {
                    break;
                }
                prevEn = cue.startTime;
            }
        }

        let prevJa = -1;
        if (japanese) {
            for (let cue of trackJa.track.cues) {
                if (cue.endTime >= player.currentTime) {
                    break;
                }
                prevJa = cue.startTime;
            }
        }

        let next = Math.max(prevEn, prevJa);

        if (next == -1) {
            return;
        }

        player.currentTime = next;
        displayCurrentCues();
    } else if (evt.code === 'KeyN') {
        evt.preventDefault();
        // jump to start of next subtitle

        let english = document.getElementById('transcript_en_checkbox').checked;
        let japanese = document.getElementById('transcript_ja_checkbox').checked;

        if (!english && !japanese) {
            return;
        }

        let prevEn = -1;
        if (english) {
            for (let cue of trackEn.track.cues) {
                if (cue.startTime > player.currentTime) {
                    prevEn = cue.startTime;
                    break;
                }
            }
        }

        let prevJa = -1;
        if (japanese) {
            for (let cue of trackJa.track.cues) {
                if (cue.startTime > player.currentTime) {
                    prevJa = cue.startTime;
                    break;
                }
            }
        }

        let next = Math.min(prevEn, prevJa);
        if (prevEn == -1 && prevJa == -1) {
            return;
        } else if (prevEn == -1) {
            next = prevJa;
        } else if (prevJa == -1) {
            next = prevEn;
        }

        console.log("jump to: ", next);

        player.currentTime = next;
        displayCurrentCues();
    } else if (evt.code === 'KeyP' || evt.code === 'KeyS') {
        evt.preventDefault();
        if (player.paused) {  // playing
            player.play();
            displayCurrentCues();
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
                adjustTextTrackAllTimings(trackEn.track, adjustment);
                story.transcript_en = textTrackToString(trackEn.track);
                let cues = findCues(trackEn.track, player.currentTime);
                displayCues(cues, captionsEn);
            }

            if (japanese) {
                lang = 'Japanese';
                adjustTextTrackAllTimings(trackJa.track, adjustment);
                story.transcript_ja = textTrackToString(trackJa.track);
                let cues = findCues(trackJa.track, player.currentTime);
                displayCues(cues, captionsJa);
            }

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
            lang = 'English ';
            bringForwardTextTrackTimings(trackEn.track, player.currentTime);
            story.transcript_en = textTrackToString(trackEn.track);
            let cues = findCues(trackEn.track, player.currentTime);
            displayCues(cues, captionsEn);
        }

        if (japanese) {
            lang = 'Japanese ';
            bringForwardTextTrackTimings(trackJa.track, player.currentTime);
            story.transcript_ja = textTrackToString(trackJa.track);
            let cues = findCues(trackJa.track, player.currentTime);
            displayCues(cues, captionsJa);
        }

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
            adjustTextTrackTimings(trackEn.track, TEXT_TRACK_TIMING_PUSH_BACK_ADJUSTMENT, player.currentTime);
            story.transcript_en = textTrackToString(trackEn.track);
            let cues = findCues(trackEn.track, player.currentTime);
            displayCues(cues, captionsEn);
        }

        if (japanese) {
            lang = 'Japanese ';
            adjustTextTrackTimings(trackJa.track, TEXT_TRACK_TIMING_PUSH_BACK_ADJUSTMENT, player.currentTime);
            story.transcript_ja = textTrackToString(trackJa.track);
            let cues = findCues(trackJa.track, player.currentTime);
            displayCues(cues, captionsJa);
        }

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
            displayCurrentCues();
        }
    }
};

// returns time in seconds
function parseTimestamp(timestamp) {
    let [mins, seconds] = timestamp.split(':');
    mins = parseInt(mins);
    seconds = parseFloat(seconds);
    return mins * 60 + seconds;
}

function displayCurrentCues() {
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

    displayCues(trackEn.track.activeCues, captionsEn);
    displayCues(trackJa.track.activeCues, captionsJa);
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
            <a class="start_time" href="#" title="ctrl-click to set the start time">${formatTrackTime(excerpt.start_time, true)}</a>-<a class="end_time" href="#" title="ctrl-click to set the end time">${formatTrackTime(excerpt.end_time, true)}</a>
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

                player.onloadeddata = function (evt) {
                    processStory(story);
                    displayStoryInfo(story);
                    displayStoryContent(story);
                };

                player.src = path + time;

                console.log("src", player.src);
            }

            if (story.transcript_en) {
                trackEn.src = `data:text/plain;charset=utf-8,` + encodeURIComponent(story.transcript_en);
            }

            if (story.transcript_ja) {
                trackJa.src = `data:text/plain;charset=utf-8,` + encodeURIComponent(story.transcript_ja);
            }

            trackEn.track.mode = 'hidden';
            trackJa.track.mode = 'hidden';

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

function displayStoryContent(story) {
    var lines = story.content.split('\n').filter(x => x);  // filter out blank lines

    let html = '';
    for (let i = 0; i < lines.length; i++) {
        html += `<div>${lines[i]}</div>`
    }

    storyLines.innerHTML = html;
}

// loads the IFrame Player API code asynchronously.
// var tag = document.createElement('script');
// tag.src = "https://www.youtube.com/iframe_api";
// var firstScriptTag = document.getElementsByTagName('script')[0];
// firstScriptTag.parentNode.insertBefore(tag, firstScriptTag);

// function onYouTubeIframeAPIReady() {
//     var url = new URL(window.location.href);
//     var storyId = parseInt(url.searchParams.get("storyId") || undefined);
//     openStory(storyId);
// }

function onPlayerReady(event) {
    event.target.playVideo();
    console.log('starting video');
}

var done = false;
function onPlayerStateChange(event) {

}

function onPlaybackRateChange(val) {
    console.log('changed rate', val);
}

document.body.onload = function (evt) {
    load();
};

function load(evt) {
    var url = new URL(window.location.href);
    var storyId = parseInt(url.searchParams.get("storyId") || undefined);
    openStory(storyId);
}