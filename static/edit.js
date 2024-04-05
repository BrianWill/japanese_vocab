var storyText = document.getElementById('new_story_text');
var updateStoryButton = document.getElementById('new_story_button');
var storyTitle = document.getElementById('new_story_title');
var storyLink = document.getElementById('new_story_link');
var storyAudio = document.getElementById('new_story_audio');

var storyId;

document.body.onload = function (evt) {
    var url = new URL(window.location.href);
    storyId = parseInt(url.searchParams.get("storyId") || undefined);
    if (storyId === undefined) {
        return;
    }
    getStoryData(storyId);
};

function getStoryData(storyId) {
    fetch(`/story/${storyId}`, {
        method: 'GET', // or 'PUT'
        headers: {
            'Content-Type': 'application/json',
        }
    }).then((response) => response.json())
        .then((data) => {
            storyTitle.value = data.title;
            //storyText.value = data.content;
            storyLink.value = data.link;
            storyAudio.value = data.audio;
        })
        .catch((error) => {
            console.error('Error:', error);
        });
}

updateStoryButton.onclick = function (evt) {
    let data = {
        it: storyId,
        //content: storyText.value,
        title: storyTitle.value,
        link: storyLink.value,
        audio: storyAudio.value,
    };

    fetch('/update_story', {
        method: 'POST', // or 'PUT'
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(data),
    }).then((response) => response.json())
        .then((data) => {
            console.log(data);
        })
        .catch((error) => {
            console.error('Error:', error);
        });
};
