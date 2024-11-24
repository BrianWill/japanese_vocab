var storyLines = document.getElementById('story_lines');
var wordList = document.getElementById('word_list');
var playerSpeedNumber = document.getElementById('player_speed_number');
var player = document.getElementById('video_player');
var captionsJa = document.getElementById('captions_ja');
var captionsEn = document.getElementById('captions_en');
var storyActions = document.getElementById('story_actions');

var englishCheckbox = document.getElementById('subtitles_en_checkbox');
var japaneseCheckbox = document.getElementById('subtitles_ja_checkbox');

var playerControls = document.getElementById('player_controls');

var story = null;
var words = null;
var wordMap = null;

var subGuideLineElement = document.getElementById('captions_meter');
var subGuideLineIndicator = document.getElementById('captions_meter_indicator');

const TEXT_TRACK_TIMING_ADJUSTMENT = 0.2;
const PLAYBACK_SPEED_ADJUSTMENT = 0.05;

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
                    snackbarMessage(`Word "${baseForm}" ${w.archived == 1 ? 'archived' : 'unarchived'}`);
                    generateAllSubtitleHTML(story);

                    // must call displayStoryText() before displaySubtitles() because displaySubtitles()
                    // marks the current line in the story text:
                    displayStoryText();
                    displaySubtitles();
                });
            }
        }
    }
});

document.getElementById('story_container').addEventListener('click', function (evt) {
    if (evt.target.classList.contains('subtitle_word')) {
        evt.preventDefault();
        
        let wordIdx = parseInt(evt.target.getAttribute('word_idx'));
        console.log('word idx: ', wordIdx, evt.ctrlKey);

        let baseForm = evt.target.getAttribute('base_form');

        // split subtitle 
        if (baseForm && evt.ctrlKey) {
            
            if (!window.confirm(`Split this subtitle, starting at ${evt.target.textContent}?`)) {
                return;
            }

            var subtitleEle = evt.target.closest('.subtitle');
            let subtitleIdx = parseInt(subtitleEle.getAttribute('subtitle_index'));

            splitSubtitle(story.subtitles_ja, subtitleIdx, wordIdx);
        }
    }
});

document.getElementById('subtitle_actions').addEventListener('click', function (evt) {
    if (evt.target.classList.contains('shift_neg_1')) {
        evt.preventDefault();
        shiftSubtitleTimings(-1);
    } else if (evt.target.classList.contains('shift_pos_1')) {
        evt.preventDefault();
        shiftSubtitleTimings(1);
    } else if (evt.target.classList.contains('shift_neg_fraction')) {
        evt.preventDefault();
        shiftSubtitleTimings(-0.1);
    } else if (evt.target.classList.contains('shift_pos_fraction')) {
        evt.preventDefault();
        shiftSubtitleTimings(0.1);
    } else if (evt.target.classList.contains('shift_back')) {
        evt.preventDefault();
        editSubtitles(adjustAllSubtitlesAfter, `shifted all subtitles past the current mark down by ${TEXT_TRACK_TIMING_PUSH_BACK_ADJUSTMENT} seconds`)
    } else if (evt.target.classList.contains('shift_forward')) {
        evt.preventDefault();
        editSubtitles(bringForwardSubtitles, `shifted all subtitles past the current timemark up to the current time mark`)
    } else if (evt.target.classList.contains('extend')) {
        evt.preventDefault();
        editSubtitles(extendSubtitle, 'extended the duration of the current subtitle');
    } else if (evt.target.classList.contains('truncate')) {
        evt.preventDefault();
        editSubtitles(truncateSubtitle, 'trucated the duration of the current subtitle');
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

        var sub = cues[idx];

        player.currentTime = sub.start_time + story.subtitles_ja_offset;
        player.play();
    }
};

storyActions.onclick = function (evt) {
    if (evt.target.classList.contains('open_transcript')) {
        evt.preventDefault();
        let lang = evt.target.classList.contains('en') ? 'en' : 'ja';
        openTranscript(story.source, story.title, lang, () => { });
        return;
    } else if (evt.target.classList.contains('reimport')) {
        evt.preventDefault();
        if (!window.confirm("Reimport the story?")) {
            return;
        }
        snackbarMessage(`reimporting story`);
        importStory(story.source, story.title, function () {
            load();
            snackbarMessage(`story was reimported`);
        });
        return;

    } else if (evt.target.classList.contains('drill_vocab')) {
        evt.preventDefault();
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
    } else if (evt.code === 'KeyA') {
        evt.preventDefault();
        let newTime = timemark - 1.8;
        player.currentTime = newTime;
    } else if (evt.code === 'KeyD') {
        evt.preventDefault();
        let newTime = timemark + 1;
        player.currentTime = newTime;
    } else if (evt.code === 'KeyQ') {
        evt.preventDefault();
        let newTime = timemark - 6;
        player.currentTime = newTime;
    } else if (evt.code === 'KeyE') {
        evt.preventDefault();
        let newTime = timemark + 5;
        player.currentTime = newTime;
    } else if (evt.code === 'KeyG') {
        evt.preventDefault();
        translateCurrentSubtitle();
    } else if (evt.code === 'KeyX') {
        // jump to start of current subtitle

        let subs = null;
        if (document.getElementById('subtitles_en_checkbox').checked) {
            subs = story.subtitles_en;
        }
        if (document.getElementById('subtitles_ja_checkbox').checked) {
            subs = story.subtitles_ja;
        }
        if (subs === null) {
            return;
        }

        let time = player.currentTime - story.subtitles_ja_offset;

        for (let index = subs.length - 1; index >= 0; index--) {
            let sub = subs[index];
            if (sub.start_time <= time) {
                player.currentTime = Math.max(sub.start_time + story.subtitles_ja_offset, 0);
                break;
            }
        }
    } else if (evt.code === 'KeyC') {
        // jump to start of next subtitle

        let subs = null;
        if (document.getElementById('subtitles_en_checkbox').checked) {
            subs = story.subtitles_en;
        }
        if (document.getElementById('subtitles_ja_checkbox').checked) {
            subs = story.subtitles_ja;
        }
        if (subs === null) {
            return;
        }

        let time = player.currentTime - story.subtitles_ja_offset;

        for (let index = 0; index < subs.length; index++) {
            let sub = subs[index];
            if (sub.start_time > time) {
                player.currentTime = Math.max(sub.start_time + story.subtitles_ja_offset, 0);
                break;
            }
        }
    } else if (evt.code === 'KeyZ') {
        // jump to start of prior subtitle

        let subs = null;
        if (document.getElementById('subtitles_en_checkbox').checked) {
            subs = story.subtitles_en;
        }
        if (document.getElementById('subtitles_ja_checkbox').checked) {
            subs = story.subtitles_ja;
        }
        if (subs === null) {
            return;
        }

        let time = player.currentTime - story.subtitles_ja_offset;

        for (let index = subs.length - 1; index >= 0; index--) {
            let sub = subs[index];
            if (sub.end_time < time) {
                player.currentTime = Math.max(sub.start_time + story.subtitles_ja_offset, 0);
                break;
            }
        }
    } else if (evt.code === 'KeyP' || evt.code === 'KeyS') {
        evt.preventDefault();
        if (player.paused) {  // playing
            play();
        } else {
            player.pause();
        }
    } else if (evt.code === 'Equal' || evt.code === 'Minus') {
        evt.preventDefault();
        let adjustment = (evt.code === 'Equal') ? PLAYBACK_SPEED_ADJUSTMENT : -PLAYBACK_SPEED_ADJUSTMENT;
        adjustPlaybackSpeed(adjustment);
    }
};

function shiftSubtitleTimings(shift) {
    let lang = 'English and Japanese';

    let english = document.getElementById('subtitles_en_checkbox').checked;
    let japanese = document.getElementById('subtitles_ja_checkbox').checked;

    if (!english && !japanese) {
        return;
    }

    if (english) {
        lang = 'English';
        story.subtitles_en_offset += shift;
        story.subtitles_en_offset = parseFloat(story.subtitles_en_offset.toFixed(3));
    }

    if (japanese) {
        lang = 'Japanese';
        story.subtitles_ja_offset += shift;
        story.subtitles_ja_offset = parseFloat(story.subtitles_ja_offset.toFixed(3));
    }

    displayStoryText();
    displaySubtitles();

    snackbarMessage(`updated ${lang} subtitle timings by ${shift} seconds`);

    clearTimeout(subtitleAdjustTimeoutHandle);
    subtitleAdjustTimeoutHandle = setTimeout(
        function () {
            updateSubtitles(story, () => {
                snackbarMessage(`saved updates to subtitle timings`);
            });
        },
        3000
    );
}

function editSubtitles(fn, message) {
    let english = document.getElementById('subtitles_en_checkbox').checked;
    let japanese = document.getElementById('subtitles_ja_checkbox').checked;

    if (!english && !japanese) {
        return;
    }

    if (english) {
        fn(story.subtitles_en, player.currentTime - story.subtitles_en_offset);
    }

    if (japanese) {
        fn(story.subtitles_ja, player.currentTime - story.subtitles_ja_offset);
    }

    displayStoryText();
    displaySubtitles();

    snackbarMessage(message);

    clearTimeout(subtitleAdjustTimeoutHandle);
    subtitleAdjustTimeoutHandle = setTimeout(
        function () {
            updateSubtitles(story, () => {
                snackbarMessage(`saved updates to subtitle timings`);
            });
        },
        3000
    );
}

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
    let cues = findSubsAtTimemark(story.subtitles_ja, player.currentTime - story.subtitles_ja_offset);
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

function displaySubtitles(afterSeek) {
    if (document.getElementById('subtitles_en_checkbox').checked) {
        captionsEn.style.display = 'flex';
    } else {
        captionsEn.style.display = 'none';
    }

    if (document.getElementById('subtitles_ja_checkbox').checked) {
        captionsJa.style.display = 'flex';
    } else {
        captionsJa.style.display = 'none';
    }

    let time = player.currentTime;
    let enSubtitles = findSubsAtTimemark(story.subtitles_en, time - story.subtitles_en_offset);
    let jaSubtitles = findSubsAtTimemark(story.subtitles_ja, time - story.subtitles_ja_offset);

    function display(time, subs, target, updateStoryText) {
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

        // change which sub is highlighted in story text when a new sub is current
        if (updateStoryText && (target.innerHTML != html || afterSeek)) {
            if (subs.length == 0) {
                subs = findSubBeforeTimemark(story.subtitles_ja, time);
            }

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

                scrollSubtitleIntoView();
            }

            target.innerHTML = html;
        }

        updateSubGuideLine(time, subs);
    }

    display(time - story.subtitles_en_offset, enSubtitles, captionsEn);
    display(time - story.subtitles_ja_offset, jaSubtitles, captionsJa, true);
}

function splitSubtitle(subs, subtitleIdx, wordIdx) {

    subs.splice(subtitleIdx, 0, structuredClone(subs[subtitleIdx]));

    var original = subs[subtitleIdx];
    var copy = subs[subtitleIdx + 1];

    const duration = 10;

    copy.start_time = original.end_time;
    copy.end_time = copy.start_time + duration;
    
    original.words = original.words.splice(0, wordIdx);
    copy.words = copy.words.slice(wordIdx);

    generateSubtitleHTML(original);
    generateSubtitleHTML(copy);

    generateSubtitleText(original);
    generateSubtitleText(copy);

    // shift all the following subtitles back
    for (let i = subtitleIdx + 2; i < subs.length; i++) {
        let sub = subs[i];
        sub.start_time += duration;
        sub.end_time += duration;
    }

    let idx = 0;
    for (let sub of subs) {
        sub.idx = idx;
        idx++;
    }

    console.log(original, copy);

    displayStoryText();
    displaySubtitles(true);

    snackbarMessage('Split subtitle');

    clearTimeout(subtitleAdjustTimeoutHandle);
    subtitleAdjustTimeoutHandle = setTimeout(
        function () {
            updateSubtitles(story, () => {
                snackbarMessage(`saved updates to subtitle timings`);
            });
        },
        3000
    );
}

function isSubtitleScrolledIntoView() {
    var currentlyHighlighted = document.querySelectorAll(`#story_lines .subtitle.highlight`);
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
    var container = document.querySelector(`#story_lines`);
    var currentlyHighlighted = document.querySelectorAll(`#story_lines .subtitle.highlight`);
    for (ele of currentlyHighlighted) {
        var pos = ele.documentOffsetTop();
        var top = pos - (container.offsetHeight * 0.25);
        container.scrollTo(0, top);
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

    generateAllSubtitleHTML(story);
}

function getWords(id) {
    fetch('words', {
        method: 'POST', // or 'PUT'
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({
            story_id: id
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

function generateSubtitleHTML(sub) {
    let highlightWords = {};
    {
        for (word of sub.words) {
            let w = wordMap[word.base_form];
            if (w && w.archived == 0 && w.count > 0) {
                highlightWords[w.base_form] = w;
            }
        }
    }

    let idx = 0;
    let html = '<div>';
    for (word of sub.words) {
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
        html += `<span word_idx="${idx}" base_form="${word.base_form}" class="${_class}">${word.display}</span>`;
        idx++;
    }
    html += '</div>';
    sub.html = html;
}


function generateSubtitleText(sub) {
    let text = '';
    for (word of sub.words) {
        text += word.display;
    }
    sub.text = text;
}

function generateAllSubtitleHTML(story) {
    for (sub of story.subtitles_ja) {
        generateSubtitleHTML(sub);
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
                    document.getElementById('caption_container').style.transform = 'translate(0px, 0px)';
                    player.style.height = '60px';
                }

                let time = '';
                if (story.end_time > 0) {
                    time = `#t=${Math.trunc(story.start_time)},${Math.trunc(story.end_time)}`;
                }

                player.src = path + time;
            }

            getWords(id);
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
    let lang = 'ja';
    let subs = story.subtitles_ja;
    if (subs == null) {
        storyLines.innerHTML = "";
        return;
    }
    let html = '';
    //let html = '<div class="scroll_subtitle_into_view_msg">Scroll current subtitle into view</div>';
    for (let subIdx = 0; subIdx < subs.length; subIdx++) {
        let sub = subs[subIdx];

        let startTime = formatTrackTime(sub.start_time + story.subtitles_ja_offset, 2);

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


function updateSubGuideLine(time, subs) {
    if (subs == null || subs.length == 0) {
        document.getElementById('captions_meter').style.visibility = 'hidden';
        return;
    }

    // because of overlap, more than one cue can be active
    let longestCue = null;
    let longestDuration = 0;
    for (let i = 0; i < subs.length; i++) {
        let sub = subs[i];
        let duration = sub.end_time - sub.start_time;
        if (duration > longestDuration) {
            longestDuration = duration;
            longestCue = sub;
        }
    }

    if (longestDuration > 0) {
        document.getElementById('captions_meter').style.visibility = 'visible';
        displaySubGuideLine(time, longestCue);
    }
}

function displaySubGuideLine(time, sub) {
    let duration = sub.end_time - sub.start_time;
    if (duration <= 0) {
        return;
    };

    // this can happen some times when seeking because of the event queue order
    if (time < sub.start_time || time > sub.end_time) {
        return;
    }

    const widthPercentagePointsPerSecond = 3;
    const maxWidth = 90;
    const minWidth = 5;
    let width = duration * widthPercentagePointsPerSecond;
    width = Math.min(Math.max(width, minWidth), maxWidth); // clamp 

    subGuideLineElement.style.width = width + '%';

    let subProgress = (time - sub.start_time) / duration;
    subGuideLineIndicator.style.marginLeft = (subProgress * 100).toFixed(5) + '%';
}

player.addEventListener("seeked", (evt) => {
    displaySubtitles(true);
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