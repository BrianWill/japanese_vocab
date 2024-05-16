var storyLines = document.getElementById('story_lines');
var wordList = document.getElementById('word_list');
var playerSpeedNumber = document.getElementById('player_speed_number');
var drillWordsLink = document.getElementById('drill_words_link');
var player = document.getElementById('video_player');
var trackJa = document.getElementById('track_ja');
var trackEn = document.getElementById('track_en');
var captionsJa = document.getElementById('captions_ja');
var captionsEn = document.getElementById('captions_en');
var statusSelect = document.getElementById('status_select');

var playerControls = document.getElementById('player_controls');
var markStoryLink = document.getElementById('mark_story');
var countSpinner = document.getElementById('count_spinner');

var story = null;
var selectedLineIdx = 0;
var youtubePlayer;

const TEXT_TRACK_TIMING_ADJUSTMENT = 0.25;

statusSelect.onchange = function (evt) {
    story.status = statusSelect.value;
    updateStoryInfo(story, () => { console.log('updated story info success') });
};

// only way to detect enter vs exit is whether the number of active increases (enter) or decreases (exit)

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

storyLines.onwheel = function (evt) {
    evt.preventDefault();
    let scrollDelta = evt.wheelDeltaY * 2;
    storyLines.scrollTop -= scrollDelta;
};

const STORY_MARK_COOLDOWN = 60 * 60 * 4;

markStoryLink.onclick = function (evt) {
    evt.preventDefault();
    let unixTime = Math.floor(Date.now() / 1000);
    if (story.date_marked + STORY_MARK_COOLDOWN > unixTime) {
        snackbarMessage("story cannot be marked right now because it's on cooldown");
        return;
    }
    story.date_marked = unixTime;
    story.repetitions_remaining = Math.max(story.repetitions_remaining - 1, 0);
    story.lifetime_repetitions++;
    countSpinner.value = story.repetitions_remaining;
    updateStoryInfo(story, () => {
        displayStoryInfo(story);
        snackbarMessage('marked story as read');
    });
};

var timeoutHandle = 0;

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

            let adjustment = (evt.code === 'Equal') ? TEXT_TRACK_TIMING_ADJUSTMENT : -TEXT_TRACK_TIMING_ADJUSTMENT;
            let lang = ''

            let english = document.getElementById('transcript_en_checkbox').checked;
            let japanese = document.getElementById('transcript_ja_checkbox').checked;

            if (english) {
                lang = 'English';
                adjustTextTrackTimings(trackEn.track, player.currentTime, adjustment);
                story.transcript_en = textTrackToString(trackEn.track);
                let cues = findCues(trackEn.track, player.currentTime);
                displayCues(cues, captionsEn);
            }
            
            if (japanese) {
                lang = 'Japanese';
                adjustTextTrackTimings(trackJa.track, player.currentTime, adjustment);
                story.transcript_ja = textTrackToString(trackJa.track);
                let cues = findCues(trackJa.track, player.currentTime);
                displayCues(cues, captionsJa);
            }

//            console.log(`updated ${lang} cues: ${adjustment}`);
            snackbarMessage(`updated ${lang} subtitle timings past the current mark by ${adjustment}`);

            clearTimeout(timeoutHandle);
            timeoutHandle = setTimeout(
                function() {
                    updateStoryInfo(story, () => {
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
    }
};

storyLines.onmousedown = function (evt) {
    if (evt.target.hasAttribute('word_idx_in_line')) {
        if (evt.ctrlKey) {
            let lineIdx = parseInt(evt.target.parentNode.parentNode.getAttribute('line_idx'));
            splitLine(evt.target, lineIdx);
        }
    } else if (evt.target.classList.contains('line_timestamp')) {
        evt.preventDefault();
        let lineIdx = parseInt(evt.target.parentNode.parentNode.getAttribute('line_idx'));
        if (evt.ctrlKey) {
            fetch('/story_consolidate_line', {
                method: 'POST', // or 'PUT'
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    story_id: story.id,
                    line_to_remove: lineIdx
                }),
            }).then((response) => response.json())
                .then((data) => {
                    console.log('Success:', data);
                    story.lines = data;
                    displayStory(story);
                })
                .catch((error) => {
                    console.error('Error:', error);
                });
        } else if (evt.which == 2 || evt.button == 4) {
            evt.preventDefault();
            if (evt.altKey) {
                var words = story.lines[lineIdx].words;
                let text = '';
                for (const word of words) {
                    text += word.surface;
                }
                window.open(`https://translate.google.com/?sl=auto&tl=en&text=${text}&op=translate`);
            } else {
                let marked = story.lines[lineIdx].marked == true;  // coerce undefined to bool
                setLineMark(lineIdx, !marked);
            }
        } else if (evt.altKey) {
            evt.preventDefault();
            if (youtubePlayer) {
                selectedLineIdx = lineIdx;
                let seconds = youtubePlayer.getCurrentTime();
                seconds -= 0.5;
                if (seconds < 0) {
                    seconds = 0;
                }
                setTimestamp(selectedLineIdx, seconds);
            } else if (player) {
                selectedLineIdx = lineIdx;
                let seconds = roundToHalfSecond(player.currentTime);
                seconds -= 0.5;
                if (seconds < 0) {
                    seconds = 0;
                }
                console.log(seconds);
                setTimestamp(selectedLineIdx, seconds);
            }
        } else {
            evt.preventDefault();
            if (youtubePlayer) {
                selectedLineIdx = lineIdx;
                let seconds = parseTimestamp(evt.target.innerHTML);
                youtubePlayer.seekTo(seconds);
                youtubePlayer.playVideo();
            } else if (player) {
                selectedLineIdx = lineIdx;
                let seconds = parseTimestamp(evt.target.innerHTML);
                player.currentTime = seconds;
                player.play();
            }
        }
    }
};


countSpinner.onchange = function (evt) {
    story.repetitions_remaining = parseInt(evt.target.value);
    updateStoryInfo(story, () => { });
};

// returns time in seconds
function parseTimestamp(timestamp) {
    let [mins, seconds] = timestamp.split(':');
    mins = parseInt(mins);
    seconds = parseFloat(seconds);
    return mins * 60 + seconds;
}

function setTimestamp(lineIdx, timestamp) {
    fetch('/story_set_timestamp', {
        method: 'POST', // or 'PUT'
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({
            story_id: story.id,
            line_idx: lineIdx,
            timestamp: timestamp,
        }),
    }).then((response) => response.json())
        .then((data) => {
            console.log('Success:', data);
            story.lines = data;
            displayStory(story);
        })
        .catch((error) => {
            console.error('Error:', error);
        });
}

function setLineMark(lineIdx, marked) {
    fetch('/story_set_mark', {
        method: 'POST', // or 'PUT'
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({
            story_id: story.id,
            line_idx: lineIdx,
            marked: marked,
        }),
    }).then((response) => response.json())
        .then((data) => {
            console.log('Success:', data);
            story.lines = data;
            displayStory(story);
        })
        .catch((error) => {
            console.error('Error:', error);
        });
}


function displayStoryInfo(story) {
    drillWordsLink.setAttribute('href', `/words.html?storyId=${story.id}`);
    document.getElementById('story_title').innerHTML = `<a href="${story.link}">${story.title}</a><hr>`;
    document.getElementById('source_info').innerText = 'Source: ' + story.source;
    document.getElementById('date_info').innerText = story.date;
    document.getElementById('lifetime_repetitions_info').innerText = 'Times repeated: ' + story.lifetime_repetitions;
    console.log(story);

    countSpinner.value = story.repetitions_remaining;
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

            statusSelect.value = story.status;

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

            if (story.audio && story.audio.endsWith('.mp4')) {
                // todo why did I create a new video element?
                const video = document.createElement("video");
                player.replaceWith(video);
                player = video;
                player.setAttribute('id', 'video_player');
                player.setAttribute('controls', 'true');

                player.src = '/audio/' + story.audio;
                player.style.display = 'block';
            } else if (story.video) {
                player.style.display = 'block';
                player.setAttribute('type', 'video/mp4');
                player.src = '/sources/' + story.source + "/" + story.video;

                console.log("src", player.src);
            } else if (story.audio) {
                player.style.display = 'block';
                player.src = '/audio/' + story.audio;
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

function displayStory(story) {
    var lines = story.content.split('\n').filter(x => x);  // filter out blank lines

    let html = '';
    for (let i = 0; i < lines.length; i++) {
        html += `<div>${lines[i]}</div>`
    }

    storyLines.innerHTML = html;
}

function isOffCooldown(rank, dateMarked, unixTime) {
    let timeSinceLastDrill = unixTime - dateMarked;
    return timeSinceLastDrill > cooldownsByRank[rank];
}

var selectedWordBaseForm = null;

function splitLine(target, lineIdx) {
    let timestamp = parseTimestamp(story.lines[lineIdx].timestamp);
    if (youtubePlayer) {
        timestamp = youtubePlayer.getCurrentTime();
        timestamp -= 0.5;
        if (timestamp < 0) {
            timestamp = 0;
        }
    } else if (player) {
        timestamp = player.currentTime;
        timestamp -= 0.5;
        if (timestamp < 0) {
            timestamp = 0;
        }
    }
    let wordIdx = parseInt(target.getAttribute('word_idx_in_line'));
    fetch('/story_split_line', {
        method: 'POST', // or 'PUT'
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({
            story_id: story.id,
            line_to_split: lineIdx,
            timestamp: timestamp,
            word_idx: wordIdx
        }),
    }).then((response) => response.json())
        .then((data) => {
            console.log('Success:', data);
            story.lines = data;
            displayStory(story);
        })
        .catch((error) => {
            console.error('Error:', error);
        });
}

function roundToHalfSecond(seconds) {
    return Math.round(seconds * 2) / 2;
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