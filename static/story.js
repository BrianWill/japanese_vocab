var storyLines = document.getElementById('story_lines');
var wordList = document.getElementById('word_list');
var playerSpeedNumber = document.getElementById('player_speed_number');
var player = document.getElementById('video_player');
var captionsJa = document.getElementById('captions_ja');
var captionsEn = document.getElementById('captions_en');
var repetitionsInfoDiv = document.getElementById('repetitions_info');
var storyActions = document.getElementById('story_actions');

var englishCheckbox = document.getElementById('transcript_en_checkbox');
var japaneseCheckbox = document.getElementById('transcript_ja_checkbox');

var playerControls = document.getElementById('player_controls');

var story = null;
var words = null;
var wordMap = null;

var cueGuideElement = document.getElementById('captions_meter');
var cueGuideIndicator = document.getElementById('captions_meter_indicator');

const TEXT_TRACK_TIMING_ADJUSTMENT = 0.2;
const TEXT_TRACK_TIMING_PUSH_BACK_ADJUSTMENT = 10;
const PLAYBACK_ADJUSTMENT = 0.05;

const MAX_INTEGER = Math.pow(2, 52) - 1;

englishCheckbox.addEventListener('change', displaySubtitles);
japaneseCheckbox.addEventListener('change', displaySubtitles);

document.getElementById('subtitle_hide_unhighlighted').addEventListener('change', function (evt) {
    var ele = document.getElementById('caption_container');
    ele.classList.toggle('hide_unhighlighted');
});

document.getElementById('story_container').addEventListener('dblclick', function (evt) {
    if (evt.target.classList.contains('subtitle_word')) {
        evt.preventDefault();
        let baseForm = evt.target.getAttribute('base_form');
        console.log('clicked word: ', baseForm);
        if (baseForm) {
            let w = wordMap[baseForm];
            if (w) {
                w.archived = w.archived == 1 ? 0 : 1;
                updateWord(w, () => {
                    generateSubtitleHTML(story);
                    displaySubtitles();
                    displayStoryText();
                });
            }
        }
    }
});

storyLines.onwheel = function (evt) {
    evt.preventDefault();
    let scrollDelta = evt.wheelDeltaY * 2;
    storyLines.scrollTop -= scrollDelta;
};

storyLines.onclick = function (evt) {
    if (evt.target.classList.contains('subtitle_start_time')) {
        evt.preventDefault();
        let subtitleContainer = evt.target.closest('div[subtitle_index]');
        if (!subtitleContainer) {
            return;
        }
        let idx = subtitleContainer.getAttribute('subtitle_index');
        let lang = subtitleContainer.getAttribute('subtitle_lang');

        var cues = (lang == 'ja') ? story.subtitles_ja : story.subtitles_en;

        var cue = cues[idx];

        player.currentTime = cue.start_time;
        player.play();
    }
};

storyActions.onclick = function (evt) {
    if (evt.target.classList.contains('open_transcript')) {
        let lang = evt.target.classList.contains('en') ? 'en' : 'ja';
        openTranscript(story.source, story.title, lang, () => { });
        return;
    } else if (evt.target.classList.contains('reimport')) {
        if (!window.confirm("Reimport the story?")) {
            return;
        }
        snackbarMessage(`reimporting story`);
        importStory(story.source, story.title, function () {
            window.location.reload();
        });
        return;

    } else if (evt.target.classList.contains('drill_vocab')) {
        window.location.href = '/words.html?storyId=' + story.id;
    } else if (evt.target.classList.contains('log_story')) {
        evt.preventDefault();

        if (window.confirm("Log this story?")) {
            let unixtime = Math.floor(Date.now() / 1000);

            for (let logItem of story.log) {
                if ((unixtime - logItem.date) < STORY_LOG_COOLDOWN) {
                    snackbarMessage("this story has already been logged within the cooldown window");
                    return false;
                }
            }

            logStory(story.id, unixtime, function () {
                snackbarMessage(`story logged`);
            });
        }
    } else if (!evt.target.closest('#excerpts')) {
        return;
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
    } else if (evt.code === 'KeyR') {
        evt.preventDefault();
        scrollSubtitleIntoView();
    } else if (evt.code === 'KeyA') {
        evt.preventDefault();
        let newTime = timemark - 1.8;
        player.currentTime = newTime;
        displaySubtitles();
    } else if (evt.code === 'KeyD') {
        evt.preventDefault();
        let newTime = timemark + 1;
        player.currentTime = newTime;
        displaySubtitles();
    } else if (evt.code === 'KeyQ') {
        evt.preventDefault();
        let newTime = timemark - 6;
        player.currentTime = newTime;
        displaySubtitles();
    } else if (evt.code === 'KeyE') {
        evt.preventDefault();
        let newTime = timemark + 5;
        player.currentTime = newTime;
        displaySubtitles();
    } else if (evt.code === 'KeyG') {
        evt.preventDefault();
        translateCurrentSubtitle();
    } else if (evt.code === 'KeyX') {
        // jump to start of current subtitle

        let cues = null;
        if (document.getElementById('transcript_en_checkbox').checked) {
            cues = story.subtitles_en;
        }
        if (document.getElementById('transcript_ja_checkbox').checked) {
            cues = story.subtitles_ja;
        }
        if (cues === null) {
            return;
        }

        for (let index = cues.length - 1; index >= 0; index--) {
            const cue = cues[index];
            if (cue.start_time <= player.currentTime) {
                player.currentTime = cue.start_time;
                break;
            }
        }
    } else if (evt.code === 'KeyC') {
        // jump to start of next subtitle

        let cues = null;
        if (document.getElementById('transcript_en_checkbox').checked) {
            cues = story.subtitles_en;
        }
        if (document.getElementById('transcript_ja_checkbox').checked) {
            cues = story.subtitles_ja;
        }
        if (cues === null) {
            return;
        }

        for (let index = 0; index < cues.length; index++) {
            const cue = cues[index];
            if (cue.start_time > player.currentTime) {
                player.currentTime = cue.start_time;
                break;
            }
        }
    } else if (evt.code === 'KeyZ') {
        // jump to start of prior subtitle

        let cues = null;
        if (document.getElementById('transcript_en_checkbox').checked) {
            cues = story.subtitles_en;
        }
        if (document.getElementById('transcript_ja_checkbox').checked) {
            cues = story.subtitles_ja;
        }
        if (cues === null) {
            return;
        }

        for (let index = cues.length - 1; index >= 0; index--) {
            const cue = cues[index];
            if (cue.end_time < player.currentTime) {
                player.currentTime = cue.start_time;
                break;
            }
        }
    } else if (evt.code === 'KeyP' || evt.code === 'KeyS') {
        evt.preventDefault();
        currentCue = null;
        if (player.paused) {  // playing
            play();
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
                adjustTextTrackAllTimings(story.subtitles_en, adjustment);
            }

            if (japanese) {
                lang = 'Japanese';
                adjustTextTrackAllTimings(story.subtitles_ja, adjustment);
            }

            displaySubtitles();

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
            lang = 'English';
            bringForwardTextTrackTimings(story.subtitles_en, player.currentTime);
        }

        if (japanese) {
            lang = 'Japanese ';
            bringForwardTextTrackTimings(story.subtitles_ja, player.currentTime);
        }

        displaySubtitles();

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
            adjustTextTrackTimings(story.subtitles_en, TEXT_TRACK_TIMING_PUSH_BACK_ADJUSTMENT, player.currentTime);
        }

        if (japanese) {
            lang = 'Japanese ';
            adjustTextTrackTimings(story.subtitles_ja, TEXT_TRACK_TIMING_PUSH_BACK_ADJUSTMENT, player.currentTime);
        }

        displaySubtitles();

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
            displaySubtitles();
        }
    } else if (evt.code.startsWith('Space')) {
        evt.preventDefault();

        if (currentCue) {
            player.currentTime = currentCue.start_time;
            play();
        }
    }
};

function play() {
    displaySubtitles();
    player.play();
}


// returns time in seconds
function parseTimestamp(timestamp) {
    let [mins, seconds] = timestamp.split(':');
    mins = parseInt(mins);
    seconds = parseFloat(seconds);
    return mins * 60 + seconds;
}

function translateCurrentSubtitle() {
    let cues = findCues(story.subtitles_ja, player.currentTime);
    let text = '';
    for (let i = 0; i < cues.length; i++) {
        let cue = cues[i];
        text += cue.text;
    }
    console.log('translate: ', text);
    let url = `https://translate.google.com/?sl=auto&tl=en&text=${text}&op=translate`;
    var win = window.open(url, '_blank');
    win.focus();
}

function displaySubtitles() {
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

    let enSubtitles = findCues(story.subtitles_en, player.currentTime);
    let jaSubtitles = findCues(story.subtitles_ja, player.currentTime);

    function display(subs, target) {
        let html = '<div>';

        for (sub of subs) {
            html += sub.html;
        }

        html += '</div>';

        if (subs.length == 0) {
            target.style.visibility = 'hidden';
        } else {
            target.style.visibility = 'visible';
        }

        if (target.innerHTML != html) {
            console.log(subs);

            // change sub highlighting only when a new sub is current 
            // (in other words, keep the highlight of the last sub until a new one is active)
            if (subs.length > 0) {

                // unhighlight the currently highlighted subtitles
                var currentlyHighlighted = document.querySelectorAll(`#story_text .subtitle.highlight`);
                for (ele of currentlyHighlighted) {
                    ele.classList.remove('highlight');
                }

                // set highlighted subs in story text
                for (subIdx in subs) {
                    let sub = subs[subIdx];
                    var ele = document.querySelector(`#story_text [subtitle_index="${sub.idx}"]`);
                    if (ele) {
                        ele.classList.add('highlight');
                    }
                }
            }

            target.innerHTML = html;
        }

        updateCueGuide(subs);
    }

    display(enSubtitles, captionsEn);
    display(jaSubtitles, captionsJa);
}

function isSubtitleScrolledIntoView() {
    var currentlyHighlighted = document.querySelectorAll(`#story_text .subtitle.highlight`);
    if (!currentlyHighlighted) {
        return true;
    }
    var container = document.getElementById('story_lines');
    for (ele of currentlyHighlighted) {
        if (isElementVisible(ele, container)) {
            return true;
        }
        return false;
    }
    return false;
}

function scrollSubtitleIntoView() {
    var currentlyHighlighted = document.querySelectorAll(`#story_text .subtitle.highlight`);
    for (ele of currentlyHighlighted) {
        ele.scrollIntoView();
        return;
    }
}

function displayStoryInfo(story) {
    document.getElementById('story_title').innerHTML = `<a href="${story.link}">${story.title}</a> &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;<a href="#">Source: ${story.source}</a><hr>`;
}

function processStory(story, words) {
    story.date_last_rep = 0;
    for (let logItem of story.log) {
        if (logItem.date > story.date_last_rep) {
            story.date_last_rep = logItem.date
        }
    }

    story.subtitles_en = JSON.parse(story.subtitles_en);
    story.subtitles_ja = JSON.parse(story.subtitles_ja);

    wordMap = {};
    for (word of words) {
        wordMap[word.base_form] = word;
        word.count = 0;
    }

    // count word frequencies
    let idx = 0;
    for (sub of story.subtitles_ja) {
        for (word of sub.words) {
            let w = wordMap[word.base_form];
            if (w) {
                w.count++;
            }
        }
        sub.idx = idx;
        idx++;
    }

    generateSubtitleHTML(story);
}

function getWords(id) {
    fetch('words', {
        method: 'POST', // or 'PUT'
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({
            story_id: id,
            excerpt_hash: 0,
        })
    }).then((response) => response.json())
        .then((data) => {
            words = data.words;
            for (w of words) {
                w.definitions = JSON.parse(w.definitions);
            }

            console.log('loaded words', words);

            processStory(story, words);
            displayStoryInfo(story);
            displayStoryText(story);
        })
        .catch((error) => {
            console.error('Error:', error);
        });
}

const maxHighlightsPerSecond = 1;

function generateSubtitleHTML(story) {
    for (sub of story.subtitles_ja) {

        let duration = sub.end_time - sub.start_time;
        let maxHighlightWords = Math.ceil(maxHighlightsPerSecond * duration);

        // pick the words to highlight
        let highlightWords = {};
        // let highlightWordsCapped = {};
        {
            let candidateWords = {};
            for (word of sub.words) {
                let w = wordMap[word.base_form];
                if (w && w.archived == 0 && w.count > 0) {
                    candidateWords[word.base_form] = w;
                }
            }
            candidateWords = Object.values(candidateWords);

            candidateWords.sort((a, b) => {
                return a.count < b.count;
            });
            shuffle(candidateWords);

            for (let w of candidateWords) {
                highlightWords[w.base_form] = w;
            }

            // for (let w of candidateWords.slice(0, maxHighlightWords)) {
            //     highlightWordsCapped[w.base_form] = w;
            // }
        }

        let html = '<div>';
        for (word of sub.words) {
            //let _class =  'subtitle_word';
            let _class = 'subtitle_word';
            if (highlightWords[word.base_form]) {
                _class += ' highlighted';
            }
            let w = wordMap[word.base_form];
            if (w && w.archived == 1) {
                _class += ' archived';
            }
            if (w && isVerb(w)) {
                _class += ' verb';
            }
            html += `<span base_form="${word.base_form}" class="${_class}">${word.display}</span>`;
        }
        html += '</div>';
        sub.html = html;
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

                player.addEventListener("durationchange", (event) => {
                    getWords(id);

                });

                player.src = path + time;
            } else {
                getWords(id);
            }

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

function displayStoryText() {
    let subs = null;
    let lang = 'en';
    if (document.getElementById('transcript_en_checkbox').checked) {
        subs = story.subtitles_en;
    }
    if (document.getElementById('transcript_ja_checkbox').checked) {
        lang = 'ja';
        subs = story.subtitles_ja;
    }
    if (subs == null) {
        storyLines.innerHTML = "";
        return;
    }
    let html = '';
    //let html = '<div class="scroll_subtitle_into_view_msg">Scroll current subtitle into view</div>';
    for (let subIdx = 0; subIdx < subs.length; subIdx++) {
        let sub = subs[subIdx];

        let startTime = formatTrackTime(sub.start_time, 2);

        html += `<div class="subtitle" subtitle_index="${subIdx}" subtitle_lang="${lang}">
            <div class="subtitle_start"><a href="#" class="subtitle_start_time">${startTime}</a></div>
            <div class="subtitle_lines">${sub.html}</div>
        </div>`;
    }

    storyLines.innerHTML = html;
}

document.body.onload = function (evt) {
    load();
};

function load(evt) {
    var url = new URL(window.location.href);
    var storyId = parseInt(url.searchParams.get("storyId") || undefined);
    openStory(storyId);
}


function updateCueGuide(cues) {
    if (cues == null || cues.length == 0) {
        document.getElementById('captions_meter').style.visibility = 'hidden';
        return;
    }

    // because of overlap, more than one cue can be active
    let longestCue = null;
    let longestDuration = 0;
    for (let i = 0; i < cues.length; i++) {
        let cue = cues[i];
        let duration = cue.end_time - cue.start_time;
        if (duration > longestDuration) {
            longestDuration = duration;
            longestCue = cue;
        }
    }

    if (longestDuration > 0) {
        document.getElementById('captions_meter').style.visibility = 'visible';
        displayCueGuide(longestCue);
    }
}

function displayCueGuide(cue) {
    let duration = cue.end_time - cue.start_time;
    if (duration <= 0) {
        return;
    };

    // this can happen some times when seeking because of the event queue order
    if (player.currentTime < cue.start_time || player.currentTime > cue.end_time) {
        return;
    }

    const widthPercentagePointsPerSecond = 3;
    const maxWidth = 90;
    const minWidth = 5;
    let width = duration * widthPercentagePointsPerSecond;
    width = Math.min(Math.max(width, minWidth), maxWidth); // clamp 

    cueGuideElement.style.width = width + '%';

    let cueProgress = (player.currentTime - cue.start_time) / duration;
    cueGuideIndicator.style.marginLeft = (cueProgress * 100).toFixed(5) + '%';
}

player.addEventListener("seeked", (evt) => {
    displaySubtitles();
});

var intervalHandle;
const deltaTime = 0.01;  // in seconds

player.addEventListener("playing", (event) => {
    window.clearInterval(intervalHandle); // just in case
    intervalHandle = window.setInterval(function () {
        displaySubtitles();
    }, deltaTime * 1000);
});

player.addEventListener("pause", (event) => {
    window.clearInterval(intervalHandle);
});
