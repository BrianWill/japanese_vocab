var storyTitle = document.getElementById('story_title');
var tokenizedText = document.getElementById('tokenized_story');
var wordList = document.getElementById('word_list');
var definitionsDiv = document.getElementById('definitions');
var kanjiResultsDiv = document.getElementById('kanji_results');
var playerSpeedSlider = document.getElementById('player_speed_slider');


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

    noUiSlider.create(playerSpeedSlider, {
        start: [1],
        step: 0.05,
        connect: true,
        pips: {
            mode: 'count',
            values: 5,
            format: {
                // 'to' the formatted value. Receives a number.
                to: function (value) {
                    var value = (value).toLocaleString(
                        undefined, // leave undefined to use the visitor's browser 
                                   // locale or a string like 'en-US' to override it.
                        { minimumFractionDigits: 1 }
                    );
                    return value;
                },
                // 'from' the formatted value.
                // Receives a string, should return a number.
                from: function (value) {
                    return value;
                }
            }
        },
        range: {
            'min': 0.4,
            'max': 1.6
        },
        format: {
            // 'to' the formatted value. Receives a number.
            to: function (value) {
                return value;
            },
            // 'from' the formatted value.
            // Receives a string, should return a number.
            from: function (value) {
                return value;
            }
        }
    });

    function sliderUpdate(values, handle, unencoded, tap, positions, noUiSlider) {
        player.setPlaybackRate(values[0]);
    }

    fetch('/story/' + id, {
        method: 'GET', // or 'PUT'
        headers: {
            'Content-Type': 'application/json',
        }
    }).then((response) => response.json())
        .then((data) => {
            story = data;
            storyTitle.innerHTML = `<a href="${story.link}">${story.title}</a>`;
            story.tokens = JSON.parse(story.tokens);
            story.words = JSON.parse(story.words);
            for (let key in story.words) {
                let word = story.words[key];
                word.definitions = JSON.parse(word.definitions);
            }
            //console.log(`/story/${id} success:`, story);
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
                    height: '700',
                    width: '900',
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
            }

            playerSpeedSlider.noUiSlider.on('update', sliderUpdate);  // calls newDrill upon registration
        })
        .catch((error) => {
            console.error('Error:', error);
        });
}

function displayStory(story) {
    //let punctuationTokens = [' ', '。', '、'];

    let html = '<p>';
    let prior = null;
    for (let i = 0; i < story.tokens.length; i++) {
        let t = story.tokens[i];
        let posClass = '';
        if (t.surface === "。") {
            html += '。</p><p>';
        } else if (t.surface === "\n\n") {
            if (prior && prior.surface !== "。") {
                html += '</p><p>';
            }
        } else if (t.surface === "\n") {
            if (prior && prior.surface !== "。") {
                html += '</p><p>';
            }
        } else if (t.surface === " ") {
            console.log("surface was space");
            if (prior && prior.surface !== "。") {
                html += '。</p><p>';
            }
        } else {
            if ((t.pos === "動詞" && t.pos1 === "接尾") ||
                (t.pos === "助動詞") ||
                (t.surface === "で" && t.pos === "助詞" && t.pos1 === "接続助詞") ||
                (t.surface === "て" && t.pos === "助詞" && t.pos1 === "接続助詞") ||
                (t.surface === "じゃ" && t.pos === "助詞" && t.pos1 === "副助詞") ||
                (t.surface === "し" && t.pos === "動詞" && t.pos1 === "自立")) {  // auxilliary verb
                posClass = 'verb_auxiliary';
            } else if (t.pos === "動詞" && t.pos1 === "非自立") { // auxilliary verb
                posClass = 'verb_auxiliary';
            } else if ((t.pos === "助詞" && t.pos1 === "格助詞") || // case particle
                (t.pos === "助詞" && t.pos1 === "接続助詞") ||   // conjunction particle
                (t.pos === "助詞" && t.pos1 === "係助詞") || // binding particle (も　は)
                (t.pos === "助詞" && t.pos1 === "副助詞")) {  // auxiliary particle
                posClass = 'particle';
            } else if (t.pos === '副詞') {
                posClass = 'adverb';
            } else if (t.pos === "接続詞" && t.pos1 === "*") { // conjunction
                posClass = 'conjunction';
            } else if ((t.pos === "助詞" && t.pos1 === "連体化") || // connecting particle　(の)
                (t.pos === "助詞" && t.pos1 === "並立助詞")) {  // connecting particle (や)
                posClass = 'connecting_particle';
            } else if (t.pos === "形容詞") { // i-adj
                posClass = 'i_adjective pad_left';
            } else if (t.pos === "名詞" && t.pos1 === "代名詞") { // pronoun
                posClass = 'pronoun pad_left';
            } else if (t.pos === "連体詞") { // adnominal adjective
                posClass = 'admoninal_adjective pad_left';
            } else if (t.pos === "動詞") { //　verb
                posClass = 'verb pad_left';
            } else if (t.pos === "名詞" && t.pos1 === "接尾") { // noun suffix
                posClass = 'noun';
            } else if ((prior && prior.pos === "助詞" && (prior.pos1 === "連体化" || prior.pos1 === '並立助詞')) ||  // preceded by connective particle
                (prior && prior.pos === "接頭詞" && prior.pos1 === "名詞接続")) {  // preceded by prefix
                posClass = 'noun';
            } else if (t.pos === "名詞") { // noun
                posClass = 'noun';
            } else if (t.pos === "記号") { // symbol
            } else if (t.pos == "号") { // counter
                posClass = 'counter';
            } else {
                posClass = 'pad_left';
            }
            html += `<span tokenIndex="${i}" class="${posClass}">${t.surface}</span>`;
        }

        prior = t;
    }
    tokenizedText.innerHTML = html + '</p>';

}

var selectedTokenIndex = null;

tokenizedText.onmousedown = function (evt) {
    let index = evt.target.getAttribute("tokenIndex");
    if (index) {
        selectedTokenIndex = index;
        displayDefinition(index);
    }
};

function displayDefinition(index) {
    let token = story.tokens[index];
    let word = story.words[token.wordId];
    getKanji(token.baseForm + token.surface);
    html = '';
    for (let entry of word.definitions) {
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
