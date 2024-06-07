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
var subStoryLink = document.getElementById('sub_story');
var repetitionsInfoDiv = document.getElementById('repetitions_info');

var playerControls = document.getElementById('player_controls');

var story = null;
var selectedLineIdx = 0;
var youtubePlayer;

var markedStartTime = 0;
var markedEndTime = 0;

const TEXT_TRACK_TIMING_ADJUSTMENT = 0.2;
const TEXT_TRACK_TIMING_PUSH_BACK_ADJUSTMENT = 10;
const PLAYBACK_ADJUSTMENT = 0.05;

trackJa.track.addEventListener('cuechange', displayCurrentCues);
trackEn.track.addEventListener('cuechange', displayCurrentCues);

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


subStoryLink.onclick = function (evt) {
    evt.preventDefault();
    let msg = `Create new story from subrange: ${formatTrackTime(markedStartTime)} to ${formatTrackTime(markedEndTime)} of ${story.source} - ${story.title}.
(Use the [ and ] keys to set start and end time markers while playing.)`;

    if (markedStartTime < 0 || markedEndTime < markedStartTime) {
        snackbarMessage("end time must be greater than start time");
        return
    }

    let title = window.prompt(msg, 'new title');

    if (title) {
        let storyInfo = {
            "parent_story": story.id,
            "title": story.title + " - " + title,
            "start_time": markedStartTime,
            "end_time": markedEndTime,
            "transcript_ja": textTrackToString(trackJa.track, markedStartTime, markedEndTime),
            "transcript_en": textTrackToString(trackEn.track, markedStartTime, markedEndTime)
        };

        createSubrangeStory(storyInfo, () => {
            snackbarMessage("subrange story has been created");
        });
    }
}

logStoryLink.onclick = function (evt) {
    evt.preventDefault();
    logRep(story, LISTENING, () => displayStoryInfo(story));
};

storyLines.onwheel = function (evt) {
    evt.preventDefault();
    let scrollDelta = evt.wheelDeltaY * 2;
    storyLines.scrollTop -= scrollDelta;
};

repetitionsInfoDiv.onclick = function (evt) {
    evt.preventDefault();
    
    if (evt.target.className.includes('rep')) {
        evt.preventDefault();
        let repIdx = parseInt(evt.target.getAttribute('repIdx'));

        if (evt.altKey) {
            insertRep(story, repIdx);
        } else if (evt.ctrlKey) {
            deleteRep(story, repIdx);
        } else {
            toggleRepType(story, repIdx);
        }
    }
};

var subtitleAdjustTimeoutHandle = 0;

document.body.onkeydown = async function (evt) {
    if (evt.ctrlKey) {
        return;
    }
    //console.log(evt);

    if (youtubePlayer) {
        let timemark = youtubePlayer.getCurrentTime();
        if (evt.code === 'KeyA') {
            evt.preventDefault();
            youtubePlayer.seekTo(timemark - 2.1, true);
        } else if (evt.code === 'KeyD') {
            evt.preventDefault();
            youtubePlayer.seekTo(timemark + 1.5, true);
        } else if (evt.code === 'KeyQ') {
            evt.preventDefault();
            youtubePlayer.seekTo(timemark - 5, true);
        } else if (evt.code === 'KeyE') {
            evt.preventDefault();
            youtubePlayer.seekTo(timemark + 4, true);
        } else if (evt.code === 'KeyP' || evt.code === 'KeyS') {
            evt.preventDefault();
            let state = youtubePlayer.getPlayerState();
            if (state === 1) {  // playing
                youtubePlayer.pauseVideo();
            } else {
                youtubePlayer.playVideo();
            }
        } else if (evt.code.startsWith('Digit')) {
            if (evt.altKey) {
                evt.preventDefault();
                let digit = parseInt(evt.code.slice(-1));
                let duration = youtubePlayer.getDuration();
                youtubePlayer.seekTo(duration * (digit / 10), true);
            }
        }
    } else if (player) {
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
            player.currentTime = timemark - 1.2;
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
    }
};

// returns time in seconds
function parseTimestamp(timestamp) {
    let [mins, seconds] = timestamp.split(':');
    mins = parseInt(mins);
    seconds = parseFloat(seconds);
    return mins * 60 + seconds;
}


function displayStoryInfo(story) {
    drillWordsLink.setAttribute('href', `/words.html?storyId=${story.id}`);
    document.getElementById('story_title').innerHTML = `<a href="${story.link}">${story.title}</a><hr>`;
    document.getElementById('source_info').innerText = 'Source: ' + story.source;
    if (story.date) {
        document.getElementById('date_info').innerText = story.date;
    }

    let time = '';
    if (story.start_time != undefined && story.end_time != undefined) {
        time = `Start time: ${formatTrackTime(story.start_time, true)}<br> 
            End time: ${formatTrackTime(story.end_time, true)}<br>`;
    }

    document.getElementById('time_info').innerHTML = `${time}`;
    displayReps(story);
}

function displayReps(story) {
    let timeLastRep = 1;
    if (story.reps_logged) {
        for (let rep of story.reps_logged) {
            if (rep.date > timeLastRep) {
                timeLastRep = rep.date;
            }
        }
    }

    let todoReps = ``;
    let i = 0;
    for (let rep of story.reps_todo) {
        if (rep == LISTENING) {
            todoReps += `<span class="listening rep" repIdx="${i}" title="listening rep">聞</span>`;
        } else if (rep == DRILLING) {
            todoReps += `<span class="drill rep" repIdx="${i}" title="vocabulary drill rep">語</span>`;
        }
        i++;
    }

    let loggedReps = ``;

    document.getElementById('repetitions_info').innerHTML = `Times repeated: ${story.repetitions}<br>
        Time since last rep: ${timeSince(timeLastRep)}<br>
        Queued reps: ${todoReps}`;
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

            displayStoryInfo(data);
            displayStory(data);

            let youtubeId = null;

            if (data.link && data.link.startsWith('https://www.youtube.com/watch?v=')) {
                youtubeId = data.link.split('https://www.youtube.com/watch?v=')[1];
            }

            if (data.link && data.link.startsWith('https://youtu.be/')) {
                youtubeId = data.link.split('https://youtu.be/')[1];
            }

            if (youtubeId) {
                youtubePlayer = new YT.Player('player', {

                    videoId: youtubeId,

                    playerVars: {
                        'playsinline': 1,
                        'cc_lang_pref': 'ja',
                        'disablekb': 1
                    },
                    events: {
                        'onReady': onPlayerReady,
                        'onStateChange': onPlayerStateChange,
                        //'onPlaybackRateChange': onPlaybackRateChange
                    }
                });

                playerControls.style.display = 'inline';

                return;
            }

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
        })
        .catch((error) => {
            console.error('Error:', error);
        });
}


playerSpeedNumber.onchange = function (evt) {
    if (youtubePlayer) {
        youtubePlayer.setPlaybackRate(parseFloat(playerSpeedNumber.value));
    } else if (player) {
        player.playbackRate = parseFloat(playerSpeedNumber.value);
    }
}

function adjustPlaybackSpeed(adjustment) {
    let newSpeed = parseFloat(playerSpeedNumber.value) + adjustment
    playerSpeedNumber.value = newSpeed.toFixed(2);

    if (youtubePlayer) {
        youtubePlayer.setPlaybackRate(newSpeed);
    } else if (player) {
        player.playbackRate = newSpeed;
    }
}

function displayStory(story) {
    var lines = story.content.split('\n').filter(x => x);  // filter out blank lines

    let html = '';
    for (let i = 0; i < lines.length; i++) {
        html += `<div>${lines[i]}</div>`
    }

    storyLines.innerHTML = html;
}

// loads the IFrame Player API code asynchronously.
var tag = document.createElement('script');
tag.src = "https://www.youtube.com/iframe_api";
var firstScriptTag = document.getElementsByTagName('script')[0];
firstScriptTag.parentNode.insertBefore(tag, firstScriptTag);


function onYouTubeIframeAPIReady() {
    var url = new URL(window.location.href);
    var storyId = parseInt(url.searchParams.get("storyId") || undefined);
    openStory(storyId);
}

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
};