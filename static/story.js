var storyLines = document.getElementById('story_lines');
var wordList = document.getElementById('word_list');
var playerSpeedNumber = document.getElementById('player_speed_number');
var drillWordsLink = document.getElementById('drill_words_link');
var player = document.getElementById('video_player');
var trackJa = document.getElementById('track_ja');
var trackEn = document.getElementById('track_en');
var captionsJa = document.getElementById('captions_ja');
var captionsEn = document.getElementById('captions_en');
var logStoryLink = document.getElementById('mark_story');
var repetitionsInfoDiv = document.getElementById('repetitions_info');

var playerControls = document.getElementById('player_controls');

var story = null;
var selectedLineIdx = 0;

var markedStartTime = 0;
var markedEndTime = 0;

const TEXT_TRACK_TIMING_ADJUSTMENT = 0.2;
const TEXT_TRACK_TIMING_PUSH_BACK_ADJUSTMENT = 10;
const PLAYBACK_ADJUSTMENT = 0.05;

trackJa.track.addEventListener('cuechange', displayCurrentCues);
trackEn.track.addEventListener('cuechange', displayCurrentCues);

logStoryLink.onclick = function (evt) {
    evt.preventDefault();
    logRep(story, LISTENING, () => window.location.href = '/');
};

storyLines.onwheel = function (evt) {
    evt.preventDefault();
    let scrollDelta = evt.wheelDeltaY * 2;
    storyLines.scrollTop -= scrollDelta;
};

repetitionsInfoDiv.onclick = function (evt) {
    evt.preventDefault();

    if (evt.target.classList.contains('rep')) {
        evt.preventDefault();
        let repIdx = parseInt(evt.target.getAttribute('repIdx'));

        if (evt.altKey) {
            insertRep(story, repIdx);
        } else if (evt.ctrlKey) {
            deleteRep(story, repIdx);
        } else {
            toggleRepType(story, repIdx);
        }
    } else if (evt.target.classList.contains('add_reps_link')) {
        addStoryReps(story.id, DEFAULT_REPS,
            function () {
                story.reps_todo = DEFAULT_REPS;
                displayReps(story);
                snackbarMessage(`added reps to story: ${story.title}`);
            }
        );
    }
};

var subtitleAdjustTimeoutHandle = 0;

document.body.onkeydown = async function (evt) {
    if (evt.ctrlKey) {
        return;
    }
    //console.log(evt);
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
                    updateStory(story, () => {
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
                updateStory(story, () => {
                    snackbarMessage(`saved updates to subtitle timings`);
                });
            },
            3000
        );
    } else if (evt.code === 'BracketRight' && evt.altKey) {
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
            adjustTextTrackTimings(trackEn.track, player.currentTime, adjustment);
            story.transcript_en = textTrackToString(trackEn.track);
            let cues = findCues(trackEn.track, player.currentTime);
            displayCues(cues, captionsEn);
        }

        if (japanese) {
            lang = 'Japanese ';
            adjustTextTrackTimings(trackJa.track, player.currentTime, adjustment);
            story.transcript_ja = textTrackToString(trackJa.track);
            let cues = findCues(trackJa.track, player.currentTime);
            displayCues(cues, captionsJa);
        }

        snackbarMessage(`updated ${lang} subtitle timings past the current mark by ${adjustment} seconds`);

        clearTimeout(subtitleAdjustTimeoutHandle);
        subtitleAdjustTimeoutHandle = setTimeout(
            function () {
                updateStory(story, () => {
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
    } else if (evt.code === 'BracketLeft') {
        evt.preventDefault();
        markedStartTime = Math.trunc(player.currentTime);
        snackbarMessage('subrange start marker set to current position');
    } else if (evt.code === 'BracketRight') {
        evt.preventDefault();
        markedEndTime = Math.trunc(player.currentTime);
        snackbarMessage('subrange end marker set to current position');
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


function displaySubranges(story) {
    function repsHTML(subrange, idx, duration, isActive) {
        let timeLastRep = 1;
        let listeningRepCount = 0;
        let drillRepCount = 0;

        if (subrange.reps_logged) {
            for (let rep of subrange.reps_logged) {
                if (rep.type == LISTENING) {
                    listeningRepCount++;
                } else if (rep.type == DRILLING) {
                    drillRepCount++;
                }

                if (rep.date > timeLastRep) {
                    timeLastRep = rep.date;
                }
            }
        }

        let todoReps;
        if (subrange.reps_todo.length == 0) {
            todoReps = `Queued reps: <a class="add_reps_link" href="#" title="add reps">add reps</a>`
        } else {
            todoReps = `Queued reps: `;
            let i = 0;
            for (let rep of subrange.reps_todo) {
                if (rep == LISTENING) {
                    todoReps += `<span class="listening rep" repIdx="${i}" title="listening rep">聞</span>`;
                } else if (rep == DRILLING) {
                    todoReps += `<span class="drill rep" repIdx="${i}" title="vocabulary drill rep">語</span>`;
                }
                i++;
            }
            todoReps += `<span class="info_symbol" title="red = listening; yellow = vocabulary drill; click to toggle type; alt-click to insert another rep; ctrl-click to remove a rep">ⓘ</span>`;
        }

        startTime = subrange.start_time || 0;
        endTime = subrange.end_time || duration;

        let html = `<div subrange_idx="${idx}">
            <hr>
            <a href="#" title="set the start time">${formatTrackTime(startTime, true)}</a>-<a href="#" title="set the end time">${formatTrackTime(endTime, true)}</a>
            <a class="drill_subrange" href="#" title="Log this excerpt">log</a>
            <a class="drill_subrange" href="#" title="Drill the words of this excerpt">drill</a>
            <a class="delete_subrange" href="#" title="Remove this excerpt">remove</a>
            <br>
            <span>Time since last rep: ${timeSince(timeLastRep)}<span><br>
            <span>Completed reps: ${listeningRepCount} listening, ${drillRepCount} drilling</span><br>
            ${todoReps}<br>`;

        return html + '</div>';
    }

    let html = `Excerpts:
    <a class="add_subrange" href="#" title="Add a new excerpt">add</a>
    <a class="sort_subrange" href="#" title="Reorder the excerpts by start time">reorder</a>`;

    let activeIdx = 0;
    for (idx in story.subranges) {
        html += repsHTML(story.subranges[idx], idx, player.duration, idx == activeIdx);
    }

    if (story.subranges.length == 0) {
        html = `<a class="add_subrange" href="#">Create a subrange</a>`;
    }

    document.getElementById('subranges').innerHTML = html;
}

function displayStoryInfo(story) {
    drillWordsLink.setAttribute('href', `/words.html?storyId=${story.id}`);
    document.getElementById('story_title').innerHTML = `<a href="${story.link}">${story.title}</a><hr>`;
    document.getElementById('source_info').innerText = 'Source: ' + story.source;
    if (story.date) {
        document.getElementById('date_info').innerText = story.date;
    }

    displayReps(story);
    displaySubranges(story);
}

function displayReps(story) {
    let timeLastRep = 1;
    let listeningRepCount = 0;
    let drillRepCount = 0;

    if (story.reps_logged) {
        for (let rep of story.reps_logged) {
            if (rep.type == LISTENING) {
                listeningRepCount++;
            } else if (rep.type == DRILLING) {
                drillRepCount++;
            }

            if (rep.date > timeLastRep) {
                timeLastRep = rep.date;
            }
        }
    }

    let todoReps;
    if (story.reps_todo.length == 0) {
        todoReps = `<a class="add_reps_link">Add reps to queue</a>`
    } else {
        todoReps = `Queued reps: `;
        let i = 0;
        for (let rep of story.reps_todo) {
            if (rep == LISTENING) {
                todoReps += `<span class="listening rep" repIdx="${i}" title="listening rep">聞</span>`;
            } else if (rep == DRILLING) {
                todoReps += `<span class="drill rep" repIdx="${i}" title="vocabulary drill rep">語</span>`;
            }
            i++;
        }
        todoReps += `<span class="info_symbol" title="red = listening; yellow = vocabulary drill; click to toggle type; alt-click to insert another rep; ctrl-click to remove a rep">ⓘ</span>`;
    }

    document.getElementById('repetitions_info').innerHTML = `Listening reps: ${listeningRepCount}<br>
        Drilling reps: ${drillRepCount}<br>
        Time since last rep: ${timeSince(timeLastRep)}<br>
        ${todoReps}`;
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

            story.reps_logged = story.reps_logged || [];
            story.reps_todo = story.reps_todo || [];

            displayStoryContent(data);

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

            setTimeout(() => displayStoryInfo(data), 10);  // timeout because needs the player src to load first
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
    var url = new URL(window.location.href);
    var storyId = parseInt(url.searchParams.get("storyId") || undefined);
    openStory(storyId);
};