var storyTitle = document.getElementById('story_title');
var tokenizedText = document.getElementById('tokenized_story');
var wordList = document.getElementById('word_list');
var definitionsDiv = document.getElementById('definitions');
var kanjiResultsDiv = document.getElementById('kanji_results');
var playerSpeedNumber = document.getElementById('player_speed_number');


var story = null;

tokenizedText.onwheel = function (evt) {
    if (evt.wheelDeltaY < 0) {
        if (tokenizedText.scrollTop >= tokenizedText.scrollTopMax) {
            evt.preventDefault();
        }
    } else {
        if (tokenizedText.scrollTop <= 0) {
            evt.preventDefault();
        }
    }
};

document.body.onkeydown = async function (evt) {
    if (evt.ctrlKey) {
        return;
    }
    //console.log(evt);
    let timemark = player.getCurrentTime();
    
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
    } else if (evt.code === 'KeyC') {
        evt.preventDefault();
        tokenizedText.classList.toggle('show_rank');    
    } else if (evt.code === 'Space') { 
        evt.preventDefault();
        if (selectedWordBaseForm) {
            let wordInfo = story.word_info[selectedWordBaseForm];
            updateWord({
                base_form: selectedWordBaseForm,
                date_last_drill: Math.floor(Date.now() / 1000), 
                rank: wordInfo.rank,
            }, true);
        }
    } else if (evt.code.startsWith('Digit')) {
        evt.preventDefault();
        let digit = parseInt(evt.code.slice(-1));
        if (!evt.altKey) {
            if (digit < 1 || digit > 4) {
                return;
            }
            if (selectedWordBaseForm) {
                updateWord({
                    base_form: selectedWordBaseForm,
                    date_last_drill: Math.floor(Date.now() / 1000),
                    rank: digit,
                });
            }
        } else {
            if (!player) { return; }
            let duration = player.getDuration();
            player.seekTo(duration * (digit / 10), true);
        }
    }
};

function updateWordInfo(word) {
    let wordSpans = tokenizedText.querySelectorAll(`span[baseform="${word.base_form}"]`);
    console.log('updating word info', word.base_form, word.rank, word.date_last_drill, 'found spans', wordSpans.length);
    for (let span of wordSpans) {
        span.classList.remove('rank1', 'rank2', 'rank3', 'rank4', 'offcooldown');
        span.classList.add('rank' + word.rank);
    }

    var wordInfo = story.word_info[word.base_form];
    wordInfo.rank = word.rank;
    wordInfo.date_last_drill = word.date_last_drill;
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
                    height: '300',
                    width: '500',
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

const DRILL_COOLDOWN_RANK_4 = 60 * 60 * 24 * 1000; // 1000 days in seconds
const DRILL_COOLDOWN_RANK_3 = 60 * 60 * 24 * 40;  // 40 days in seconds
const DRILL_COOLDOWN_RANK_2 = 60 * 60 * 24 * 5;   // 5 days in seconds
const DRILL_COOLDOWN_RANK_1 = 60 * 60 * 5;        // 5 hours in second

const cooldownsByRank = [0, DRILL_COOLDOWN_RANK_1, DRILL_COOLDOWN_RANK_2, DRILL_COOLDOWN_RANK_3, DRILL_COOLDOWN_RANK_4];

function displayStory(story) {
    //let punctuationTokens = [' ', '。', '、'];

    let html = '';

    let unixTime = Math.floor(Date.now() / 1000);

    for (let line of story.lines) {
        html += `<p><a class="line_timestamp">${line.timestamp}</a>`;
        for (let word of line.words) {
            let wordinfo = story.word_info[word.baseform];
            if (word.id) {
                let timeSinceLastDrill = unixTime - wordinfo.date_last_drill;
                let offCooldown = timeSinceLastDrill > cooldownsByRank[wordinfo.rank];
                html += `<span wordId="${word.id || ''}" baseform="${word.baseform || ''}" 
                    class="lineword rank${wordinfo.rank} ${offCooldown ? 'offcooldown' : ''} ${word.pos || ''}">${word.surface}</span>`;
            } else {
                html += `<span class="lineword nonword">${word.surface}</span>`;
            }            
        }
        html += '</p>'
    }

    tokenizedText.innerHTML = html;
}

var selectedWordBaseForm = null;


tokenizedText.onmousedown = function (evt) {
    //console.log(evt.target);
    if (evt.target.hasAttribute('baseform')) {
        let baseform = evt.target.getAttribute('baseform');
        selectedWordBaseForm = baseform;

        console.log('baseform', baseform);
        let surface = evt.target.innerHTML;
        displayDefinition(baseform, surface);    
    } else if (evt.target.classList.contains('line_timestamp')) {
        let [mins, seconds] = evt.target.innerHTML.split(':');
        mins = parseInt(mins);
        seconds = parseInt(seconds);
        console.log("timestamp", mins, seconds);
        player.seekTo(mins * 60 + seconds);
        player.playVideo();
    }    
};

function displayDefinition(baseform, surface) {
    getKanji(baseform + surface);
    html = '';
    let wordInfo = story.word_info[baseform];
    for (let entry of wordInfo.definitions) {
        html += displayEntry(entry);
    }
    definitionsDiv.innerHTML = html;
}

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
    let playerControlsDiv = document.getElementById('player_controls');
    playerControlsDiv.style.display = 'block';
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
