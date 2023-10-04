var storyTitle = document.getElementById('story_title');
var tokenizedStory = document.getElementById('tokenized_story');
var wordList = document.getElementById('word_list');
var definitionsDiv = document.getElementById('definitions');
var kanjiResultsDiv = document.getElementById('kanji_results');
var playerSpeedNumber = document.getElementById('player_speed_number');
var drillWordsLink = document.getElementById('drill_words_link');
var highlightLink = document.getElementById('highlight_message');
var audioPlayer = document.getElementById('audio_player');
var playerControls = document.getElementById('player_controls');
var markStoryLink = document.getElementById('mark_story');
var deleteStoryLink = document.getElementById('delete_story');
var countSpinner = document.getElementById('count_spinner');

var story = null;
var selectedLineIdx = 0;
var videoPlayer;

tokenizedStory.onwheel = function (evt) {
    evt.preventDefault();
    let scrollDelta = evt.wheelDeltaY * 2;
    tokenizedStory.scrollTop -= scrollDelta;
};


highlightLink.onclick = toggleHighlight;

function toggleHighlight(evt) {
    evt.preventDefault();
    tokenizedStory.classList.toggle('highlight_all_words');
    if (tokenizedStory.classList.contains('highlight_all_words')) {
        highlightLink.innerHTML = 'Highlighting all rank 1-3 words';
    } else {
        highlightLink.innerHTML = 'Highlighting only the rank 1-3 words off cooldown';
    }
}

const STORY_MARK_COOLDOWN = 60 * 60 * 4;

markStoryLink.onclick = function (evt) {
    evt.preventDefault();
    let unixTime = Math.floor(Date.now() / 1000);
    if (story.date_last_read + STORY_MARK_COOLDOWN > unixTime) {
        snackbarMessage("story cannot be marked right now because it's on cooldown");
        return;
    }
    story.date_last_read = unixTime;
    story.countdown = Math.max(story.countdown - 1, 0);
    story.read_count++;
    countSpinner.value = story.countdown;
    updateStoryCounts(story, () => {
        snackbarMessage('marked story as read');
    });
};

deleteStoryLink.onclick = function (evt) {
    evt.preventDefault();

    if (!confirm('Do you want to delete this story?')) {
        return;
    }

    let url = new URL(window.location.href);
    let storyId = parseInt(url.searchParams.get("storyId"));
    let data = { ID: storyId };   

    console.log("deleting story", data.ID);

    fetch('/delete_story', {
        method: 'DELETE',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(data),
    }).then((response) => {
        if (response.status === 200)
            window.location.replace('/');
        })
        .catch((error) => {
            console.error('Error:', error);
        });
}

document.body.onkeydown = async function (evt) {
    if (evt.ctrlKey) {
        return;
    }
    //console.log(evt);

    if (videoPlayer) {
        let timemark = videoPlayer.getCurrentTime();
        if (evt.code === 'KeyA') {
            evt.preventDefault();
            videoPlayer.seekTo(timemark - 2.1, true);
        } else if (evt.code === 'KeyD') {
            evt.preventDefault();
            videoPlayer.seekTo(timemark + 1.5, true);
        } else if (evt.code === 'KeyQ') {
            evt.preventDefault();
            videoPlayer.seekTo(timemark - 5, true);
        } else if (evt.code === 'KeyE') {
            evt.preventDefault();
            videoPlayer.seekTo(timemark + 4, true);
        } else if (evt.code === 'KeyP' || evt.code === 'KeyS') {
            evt.preventDefault();
            let state = videoPlayer.getPlayerState();
            if (state === 1) {  // playing
                videoPlayer.pauseVideo();
            } else {
                videoPlayer.playVideo();
            }
        } else if (evt.code.startsWith('Digit')) {
            if (evt.altKey) {
                evt.preventDefault();
                let digit = parseInt(evt.code.slice(-1));
                let duration = videoPlayer.getDuration();
                videoPlayer.seekTo(duration * (digit / 10), true);
            }
        }
    } else if (audioPlayer) {
        let timemark = audioPlayer.currentTime;
        if (evt.code === 'KeyA') {
            evt.preventDefault();
            audioPlayer.currentTime = timemark - 2.1;
        } else if (evt.code === 'KeyD') {
            evt.preventDefault();
            audioPlayer.currentTime = timemark + 1.5;
        } else if (evt.code === 'KeyQ') {
            evt.preventDefault();
            audioPlayer.currentTime = timemark - 5;
        } else if (evt.code === 'KeyE') {
            evt.preventDefault();
            audioPlayer.currentTime = timemark + 4;
        } else if (evt.code === 'KeyP' || evt.code === 'KeyS') {
            evt.preventDefault();
            if (audioPlayer.paused) {  // playing
                audioPlayer.play();
            } else {
                audioPlayer.pause();
            }
        } else if (evt.code.startsWith('Digit')) {
            if (evt.altKey) {
                evt.preventDefault();
                let digit = parseInt(evt.code.slice(-1));
                let duration = audioPlayer.duration;
                audioPlayer.currentTime = duration * (digit / 10);
            }
        }
    }

    if (evt.code === 'KeyC') {
        toggleHighlight(evt);
    } else if (evt.code === 'Space') {
        evt.preventDefault();
        if (selectedWordBaseForm) {
            let wordInfo = story.word_info[selectedWordBaseForm];
            updateWord({
                base_form: selectedWordBaseForm,
                date_marked: Math.floor(Date.now() / 1000),
                rank: wordInfo.rank,
            }, story.word_info, true);
        }
    } else if (evt.code === 'KeyM') {
        evt.preventDefault();
        let marked = story.lines[selectedLineIdx].marked == true;  // coerce undefined to bool
        setLineMark(selectedLineIdx, !marked);
    } else if (evt.code === 'Minus') {
        evt.preventDefault();
        let timestamp = story.lines[selectedLineIdx].timestamp;
        let seconds = parseTimestamp(timestamp) - 0.5;
        if (seconds < 0) {
            return;
        }
        setTimestamp(selectedLineIdx, seconds);
    } else if (evt.code === 'Equal') {
        evt.preventDefault();
        let timestamp = story.lines[selectedLineIdx].timestamp;
        let seconds = parseTimestamp(timestamp) + 0.5;
        setTimestamp(selectedLineIdx, seconds);
    } else if (evt.code.startsWith('Digit')) {
        evt.preventDefault();
        let digit = parseInt(evt.code.slice(-1));
        if (!evt.altKey) {
            if (digit < 1 || digit > 4) {
                return;
            }
            if (selectedWordBaseForm) {
                let wordInfo = story.word_info[selectedWordBaseForm];
                updateWord({
                    base_form: selectedWordBaseForm,
                    date_marked: wordInfo.date_marked,
                    rank: digit,
                }, story.word_info);
            }
        }
    }
};

tokenizedStory.onmousedown = function (evt) {
    if (evt.target.hasAttribute('word_idx_in_line')) {
        if (evt.ctrlKey) {
            let lineIdx = parseInt(evt.target.parentNode.parentNode.getAttribute('line_idx'));
            splitLine(evt.target, lineIdx);
        } else {
            let baseform = evt.target.getAttribute('baseform');
            selectedWordBaseForm = baseform;
            console.log('baseform', baseform);
            let surface = evt.target.innerHTML;
            displayDefinition(baseform, surface);
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
            if (videoPlayer) {
                selectedLineIdx = lineIdx;
                let seconds = videoPlayer.getCurrentTime();
                seconds -= 0.5;
                if (seconds < 0) {
                    seconds = 0;
                }
                setTimestamp(selectedLineIdx, seconds);
            } else if (audioPlayer) {
                selectedLineIdx = lineIdx;
                let seconds = roundToHalfSecond(audioPlayer.currentTime);
                seconds -= 0.5;
                if (seconds < 0) {
                    seconds = 0;
                }
                console.log(seconds);
                setTimestamp(selectedLineIdx, seconds);
            }
        } else {
            evt.preventDefault();
            if (videoPlayer) {
                selectedLineIdx = lineIdx;
                let seconds = parseTimestamp(evt.target.innerHTML);
                videoPlayer.seekTo(seconds);
                videoPlayer.playVideo();
            } else if (audioPlayer) {
                selectedLineIdx = lineIdx;
                let seconds = parseTimestamp(evt.target.innerHTML);
                audioPlayer.currentTime = seconds;
                audioPlayer.play();
            }


            // navigator.clipboard.writeText(text)
            //     .then(() => {
            //         console.log('Text copied to clipboard: ' + text);
            //     })
            //     .catch((error) => {
            //         console.error('Error copying text to clipboard:', error);
            //     });
        }
    }
};


countSpinner.onchange = function (evt) {
    story.countdown = parseInt(evt.target.value);
    updateStoryCounts(story, () => { });
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


function openStory(id) {
    fetch('/story/' + id, {
        method: 'GET', // or 'PUT'
        headers: {
            'Content-Type': 'application/json',
        }
    }).then((response) => response.json())
        .then((data) => {
            story = data;
            drillWordsLink.setAttribute('href', `/words.html?storyId=${story.id}`);
            storyTitle.innerHTML = `<a href="${story.link}">${story.title}</a>`;
            console.log(story);

            countSpinner.value = story.countdown;

            displayStory(data);

            let youtubeId = null;

            if (data.link.startsWith('https://www.youtube.com/watch?v=')) {
                youtubeId = data.link.split('https://www.youtube.com/watch?v=')[1];
            }

            if (data.link.startsWith('https://youtu.be/')) {
                youtubeId = data.link.split('https://youtu.be/')[1];
            }

            if (youtubeId) {
                videoPlayer = new YT.Player('player', {

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
            } else if (story.audio) {
                audioPlayer.style.display = 'block';
                audioPlayer.src = '/audio/' + story.audio;
            }

            if (videoPlayer || audioPlayer) {
                playerControls.style.display = 'inline';
            }
        })
        .catch((error) => {
            console.error('Error:', error);
        });
}


playerSpeedNumber.onchange = function (evt) {
    if (videoPlayer) {
        videoPlayer.setPlaybackRate(parseFloat(playerSpeedNumber.value));
    } else if (audioPlayer) {
        audioPlayer.playbackRate = parseFloat(playerSpeedNumber.value);
    }
}

function displayStory(story) {
    let html = '<table id="lines_table">';

    let unixTime = Math.floor(Date.now() / 1000);

    for (let idx in story.lines) {
        let line = story.lines[idx];
        html += `<tr line_idx="${idx}"><td class="line_timestamp_container"><a class="line_timestamp ${line.marked ? 'marked_line' : ''}">${line.timestamp}</a></td><td>`;
        for (let wordIdx in line.words) {
            let word = line.words[wordIdx];
            let wordinfo = story.word_info[word.baseform];
            if (word.id) {
                let offCooldown = isOffCooldown(wordinfo.rank, wordinfo.date_marked, unixTime);
                html += `<span word_idx_in_line="${wordIdx}" word_id="${word.id || ''}" baseform="${word.baseform || ''}" 
                    class="lineword rank${wordinfo.rank} ${offCooldown ? 'offcooldown' : ''} ${word.pos || ''}">${word.surface}</span>`;
            } else {
                html += `<span word_idx_in_line="${wordIdx}" class="lineword nonword">${word.surface}</span>`;
            }
        }
        html += '</td></tr>'
    }

    tokenizedStory.innerHTML = html + '</table>';
}

function isOffCooldown(rank, dateMarked, unixTime) {
    let timeSinceLastDrill = unixTime - dateMarked;
    return timeSinceLastDrill > cooldownsByRank[rank];
}

var selectedWordBaseForm = null;

function splitLine(target, lineIdx) {
    let timestamp = parseTimestamp(story.lines[lineIdx].timestamp);
    if (videoPlayer) {
        timestamp = videoPlayer.getCurrentTime();
        timestamp -= 0.5;
        if (timestamp < 0) {
            timestamp = 0;
        }
    } else if (audioPlayer) {
        timestamp = audioPlayer.currentTime;
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

function displayDefinition(baseform, surface) {
    getKanji(baseform + surface);
    html = '';
    let wordInfo = story.word_info[baseform];
    if (wordInfo && wordInfo.definitions) {
        for (let entry of wordInfo.definitions) {
            html += displayEntry(entry);
        }
    }
    definitionsDiv.innerHTML = html;
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
