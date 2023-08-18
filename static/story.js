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
    console.log(evt.code);
    if (player) {
        if (evt.code === 'KeyA') {
            evt.preventDefault();
            var timemark = player.getCurrentTime();
            player.seekTo(timemark - 1.7, true);
            //player.playVideo();
            // showWord();
        } else if (evt.code === 'KeyD') { 
            evt.preventDefault();
            var timemark = player.getCurrentTime();
            player.seekTo(timemark + 1.2, true);
        } else if (evt.code === 'KeyQ') {
            evt.preventDefault();
            var timemark = player.getCurrentTime();
            player.seekTo(timemark - 5, true);
            //player.playVideo();
            // showWord();
        } else if (evt.code === 'KeyE') { 
            evt.preventDefault();
            var timemark = player.getCurrentTime();
            player.seekTo(timemark + 4, true);
        } else if (evt.code === 'Space') { 
            evt.preventDefault();
            var state = player.getPlayerState();
            if (state === 1) {  // playing
                player.pauseVideo();
            } else {
                player.playVideo();
            }
        } else if (evt.code.startsWith('Digit')) {
            var digit = parseInt(evt.code.slice(-1));
            var duration =player.getDuration();
            player.seekTo(duration * (digit / 10), true);
        }
    }
};

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

function displayStory(story) {
    //let punctuationTokens = [' ', '。', '、'];

    let html = '';

    for (let line of story.lines) {
        html += `<p><a class="line_timestamp">${line.timestamp}</a>`;
        for (let word of line.words) {
            html += `<span wordId="${word.id || ''}" baseform="${word.baseform || ''}" class="lineword ${word.pos || ''}">${word.surface}</span>`;
            // if (!word.id) {
            //     console.log(`no id: ${word.surface} __ ${word.pos}`);
            // }
        }
        html += '</p>'
    }

    tokenizedText.innerHTML = html;
}

var selectedTokenIndex = null;


tokenizedText.onmousedown = function (evt) {
    console.log(evt.target);
    if (!evt.target.hasAttribute('baseform')) {
        return;
    }
    let baseform = evt.target.getAttribute('baseform');
    let surface = evt.target.innerHTML;
    displayDefinition(baseform, surface);
    
};

function displayDefinition(baseform, surface) {
    getKanji(baseform + surface);
    html = '';
    for (let entry of story.definitions[baseform]) {
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
