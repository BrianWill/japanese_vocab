var newStoryText = document.getElementById('new_story_text');
var newStoryButton = document.getElementById('new_story_button');
var storyList = document.getElementById('story_list');
var newStoryTitle = document.getElementById('new_story_title');
var storyTitle = document.getElementById('story_title');
var storyText = document.getElementById('story');

newStoryButton.onclick = function (evt) {
    let data = {
        content: newStoryText.value,
        title: newStoryTitle.value
    };

    fetch('story', {
        method: 'POST', // or 'PUT'
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(data),
    }).then((response) => response.json())
        .then((data) => {
            console.log('Success:', data);
        })
        .catch((error) => {
            console.error('Error:', error);
        });
};

document.body.onload = function (evt) {
    console.log('asdf');

    fetch('stories_list', {
        method: 'GET', // or 'PUT'
        headers: {
            'Content-Type': 'application/json',
        }
    }).then((response) => response.json())
        .then((data) => {
            console.log('Success:', data);
            updateStoryList(data);
        })
        .catch((error) => {
            console.error('Error:', error);
        });
};


function updateStoryList(stories) {
    html = '';
    for (let s of stories) {
        html += `<li><a story_id="${s._id}" href="#">${s.title}</a></li>`;
    }
    storyList.innerHTML = html;
};

storyList.onclick = function (evt) {
    if (evt.target.tagName == 'A') {
        var storyId = evt.target.getAttribute('story_id');
        console.log(storyId);
        openStory(storyId)
    }
};

function openStory(id) {
    fetch('story/' + id, {
        method: 'GET', // or 'PUT'
        headers: {
            'Content-Type': 'application/json',
        }
    }).then((response) => response.json())
        .then((data) => {
            console.log('Success:', data);
            storyTitle.innerText = data.title;
            storyText.innerText = data.content;
        })
        .catch((error) => {
            console.error('Error:', error);
        });
}

storyText.onmousedown = function (evt) {
    //console.log(document.getSelection().toString());
};

storyText.onmouseup = function (evt) {
    console.log(document.getSelection().toString());
};

storyText.onmouseleave = function (evt) {
    console.log(document.getSelection().toString());
};

// document.body.addEventListener('selectionchange', (event) => {
//     console.log('changed');
// });


// window.setInterval(function () {
//     console.log(document.getSelection().toString());
// }, 300);

