var storyTitle = document.getElementById('story_title');
var tokenizedStory = document.getElementById('tokenized_story');
var wordList = document.getElementById('word_list');
var definitionsDiv = document.getElementById('definitions');
var kanjiResultsDiv = document.getElementById('kanji_results');
var playerSpeedNumber = document.getElementById('player_speed_number');
var drillWordsLink = document.getElementById('drill_words_link');
var highlightMessage = document.getElementById('highlight_message');
var logStoryLink = document.getElementById('log_story_link');
var audioPlayer = document.getElementById('audio_player');

var story = null;
var selectedLineIdx = 0;
var videoPlayer;

tokenizedStory.onwheel = function (evt) {
    if (evt.wheelDeltaY < 0) {
        if (tokenizedStory.scrollTop >= tokenizedStory.scrollTopMax) {
            evt.preventDefault();
        }
    } else {
        if (tokenizedStory.scrollTop <= 0) {
            evt.preventDefault();
        }
    }
};

document.body.onkeydown = async function (evt) {
    if (evt.ctrlKey) {
        return;
    }
    //console.log(evt);

    if (videoPlayer) {
        let timemark = videoPlayer.getCurrentTime();
        if (evt.code === 'KeyA') {
            evt.preventDefault();
            videoPlayer.seekTo(timemark - 1.7, true);
        } else if (evt.code === 'KeyD') {
            evt.preventDefault();
            videoPlayer.seekTo(timemark + 1.2, true);
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
            audioPlayer.currentTime = timemark - 1.7;
        } else if (evt.code === 'KeyD') {
            evt.preventDefault();
            audioPlayer.currentTime = timemark + 1.2;
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
        evt.preventDefault();
        tokenizedStory.classList.toggle('highlight_all_words');
        if (tokenizedStory.classList.contains('highlight_all_words')) {
            highlightMessage.innerHTML = 'Highlighting all rank 1-3 words';
        } else {
            highlightMessage.innerHTML = 'Highlighting only the rank 1-3 words off cooldown';
        }
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
    if (evt.target.hasAttribute('baseform')) {
        evt.preventDefault();
        if (evt.ctrlKey) {
            splitLine(evt.target);
        } else {
            let baseform = evt.target.getAttribute('baseform');
            selectedWordBaseForm = baseform;
            console.log('baseform', baseform);
            let surface = evt.target.innerHTML;
            displayDefinition(baseform, surface);
        }

    } else if (evt.target.classList.contains('line_timestamp')) {
        evt.preventDefault();
        if (evt.ctrlKey) {
            let lineIdx = parseInt(evt.target.parentNode.getAttribute('line_idx'));
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
        } else if (evt.altKey) {
            if (videoPlayer) {
                selectedLineIdx = parseInt(evt.target.parentNode.getAttribute('line_idx'));
                let seconds = videoPlayer.getCurrentTime();
                seconds -= 0.5;
                if (seconds < 0) {
                    seconds = 0;
                }
                setTimestamp(selectedLineIdx, seconds);
            } else if (audioPlayer) {
                selectedLineIdx = parseInt(evt.target.parentNode.getAttribute('line_idx'));
                let seconds = roundToHalfSecond(audioPlayer.currentTime);
                seconds -= 0.5;
                if (seconds < 0) {
                    seconds = 0;
                }
                setTimestamp(selectedLineIdx, seconds);
            }
        } else {
            if (videoPlayer) {
                selectedLineIdx = parseInt(evt.target.parentNode.getAttribute('line_idx'));
                let seconds = parseTimestamp(evt.target.innerHTML);
                videoPlayer.seekTo(seconds);
                videoPlayer.playVideo();
            } else if (audioPlayer) {
                selectedLineIdx = parseInt(evt.target.parentNode.getAttribute('line_idx'));
                let seconds = parseTimestamp(evt.target.innerHTML);
                audioPlayer.currentTime = seconds;
                audioPlayer.play();
            }
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
            } else if (story.audio !== '') {
                audioPlayer.style.display = 'block';
                audioPlayer.src = '/audio/' + story.audio;
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
    //let punctuationTokens = [' ', '。', '、'];

    let html = '';

    let unixTime = Math.floor(Date.now() / 1000);

    for (let idx in story.lines) {
        let line = story.lines[idx];
        html += `<p line_idx="${idx}"><a class="line_timestamp">${line.timestamp}</a>`;
        for (let wordIdx in line.words) {
            let word = line.words[wordIdx];
            let wordinfo = story.word_info[word.baseform];
            if (word.id) {
                let offCooldown = isOffCooldown(wordinfo.rank, wordinfo.date_marked, unixTime);
                html += `<span word_idx_in_line="${wordIdx}" word_id="${word.id || ''}" baseform="${word.baseform || ''}" 
                    class="lineword rank${wordinfo.rank} ${offCooldown ? 'offcooldown' : ''} ${word.pos || ''}">${word.surface}</span>`;
            } else {
                html += `<span class="lineword nonword">${word.surface}</span>`;
            }
        }
        html += '</p>'
    }

    tokenizedStory.innerHTML = html;
}

function isOffCooldown(rank, dateMarked, unixTime) {
    let timeSinceLastDrill = unixTime - dateMarked;
    return timeSinceLastDrill > cooldownsByRank[rank];
}

var selectedWordBaseForm = null;

function splitLine(target) {
    let lineIdx = parseInt(target.parentNode.getAttribute('line_idx'));
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

logStoryLink.onclick = function (evt) {
    evt.preventDefault();
    addLogEvent(story.id);
};

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
