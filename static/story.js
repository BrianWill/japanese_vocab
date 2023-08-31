var storyTitle = document.getElementById('story_title');
var tokenizedStory = document.getElementById('tokenized_story');
var wordList = document.getElementById('word_list');
var definitionsDiv = document.getElementById('definitions');
var kanjiResultsDiv = document.getElementById('kanji_results');
var playerSpeedNumber = document.getElementById('player_speed_number');
var drillWordsLink = document.getElementById('drill_words_link');
var highlightMessage = document.getElementById('highlight_message');
var logStoryLink = document.getElementById('log_story_link');

var story = null;
var selectedLineIdx = 0;

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
    console.log(evt);
    let timemark = 0;
    if (player) {
        timemark = player.getCurrentTime();
    }

    if (evt.code === 'KeyA') {
        evt.preventDefault();
        if (!player) { return; }
        player.seekTo(timemark - 1.7, true);
    } else if (evt.code === 'KeyD') {
        evt.preventDefault();
        if (!player) { return; }
        player.seekTo(timemark + 1.2, true);
    } else if (evt.code === 'KeyQ') {
        evt.preventDefault();
        if (!player) { return; }
        player.seekTo(timemark - 5, true);
    } else if (evt.code === 'KeyE') {
        evt.preventDefault();
        if (!player) { return; }
        player.seekTo(timemark + 4, true);
    } else if (evt.code === 'KeyP' || evt.code === 'KeyS') {
        evt.preventDefault();
        if (!player) { return; }
        let state = player.getPlayerState();
        if (state === 1) {  // playing
            player.pauseVideo();
        } else {
            player.playVideo();
        }
    } else if (evt.code === 'Minus') {
        evt.preventDefault();
        if (!player) { return; }
        let timestamp = story.lines[selectedLineIdx].timestamp;
        let seconds = parseTimestamp(timestamp) - 0.5;
        if (seconds < 0) {
            return;
        }
        setTimestamp(selectedLineIdx, seconds);
    } else if (evt.code === 'Equal') {
        evt.preventDefault();
        if (!player) { return; }
        let timestamp = story.lines[selectedLineIdx].timestamp;
        let seconds = parseTimestamp(timestamp) + 0.5;
        setTimestamp(selectedLineIdx, seconds);
    } else if (evt.code === 'KeyC') {
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
        } else {
            if (!player) { return; }
            let duration = player.getDuration();
            player.seekTo(duration * (digit / 10), true);
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
                player = new YT.Player('player', {

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

                playerSpeedNumber.onchange = function (evt) {
                    player.setPlaybackRate(parseFloat(playerSpeedNumber.value));
                }
            }


        })
        .catch((error) => {
            console.error('Error:', error);
        });
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
    let timestamp = player.getCurrentTime();
    let lineIdx = parseInt(target.parentNode.getAttribute('line_idx'));
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


tokenizedStory.onmousedown = function (evt) {
    if (evt.target.hasAttribute('baseform')) {
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
        } else {
            selectedLineIdx = parseInt(evt.target.parentNode.getAttribute('line_idx'));
            let seconds = parseTimestamp(evt.target.innerHTML);
            player.seekTo(seconds);
            player.playVideo();
        }
    }
};

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

var player;
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
function stopVideo() {
    player.stopVideo();
}

function onPlaybackRateChange(val) {
    console.log('changed rate', val);
}
