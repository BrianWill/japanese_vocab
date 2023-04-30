var newStoryText = document.getElementById('new_story_text');
var newStoryButton = document.getElementById('new_story_button');
var storyList = document.getElementById('story_list');
var newStoryTitle = document.getElementById('new_story_title');
var newStoryLink = document.getElementById('new_story_link');
var storyTitle = document.getElementById('story_title');
var wordList = document.getElementById('word_list');
var definitionsDiv = document.getElementById('definitions');
var kanjiResultsDiv = document.getElementById('kanji_results');

newStoryButton.onclick = function (evt) {
    let data = {
        content: newStoryText.value,
        title: newStoryTitle.value,
        link: newStoryLink.value
    };

    fetch('/create_story', {
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
    getStoryList();

    if (window.location.pathname.startsWith('/read/')) {
        let storyId = window.location.pathname.substring(6);
        console.log("opening story with id " + storyId);
        openStory(storyId);
    }
};

function updateStoryList(stories) {
    let html = '';

    html += `<h3>Active stories</h3>`
    for (let s of stories) {
        if (s.state !== 'active') {
            continue;
        }
        html += `<li>
            <a story_id="${s.id}" action="drill" href="#">drill</a>&nbsp;&nbsp;
            <a story_id="${s.id}" action="mark_inactive" href="#">mark inactive</a>&nbsp;&nbsp;
            <a story_id="${s.id}" action="mark_unread" href="#">mark unread</a>&nbsp;&nbsp;
            <a story_id="${s.id}" href="#">${s.title}</a>
            </li>`;
    }

    html += `<h3>Inactive stories</h3>`
    for (let s of stories) {
        if (s.state !== 'inactive') {
            continue;
        }
        html += `<li>
            <a story_id="${s.id}" action="drill" href="#">drill</a>&nbsp;&nbsp;
            <a story_id="${s.id}" action="mark_active" href="#">make active</a>&nbsp;&nbsp;
            <a story_id="${s.id}" action="mark_unread" href="#">mark unread</a>&nbsp;&nbsp;
            <a story_id="${s.id}" href="#">${s.title}</a>
            </li>`;
    }

    html += `<h3>Unread stories</h3>`
    for (let s of stories) {
        if (s.state !== 'unread') {
            continue;
        }
        html += `<li>
            <a story_id="${s.id}" action="drill" href="#">drill</a>&nbsp;&nbsp;
            <a story_id="${s.id}" action="mark_active" href="#">mark active</a>&nbsp;&nbsp;
            <a story_id="${s.id}" href="#">${s.title}</a>
            </li>`;
    }
    storyList.innerHTML = html;
};

storyList.onclick = function (evt) {
    if (evt.target.tagName == 'A') {
        evt.preventDefault();
        var storyId = evt.target.getAttribute('story_id');

        var action = evt.target.getAttribute('action');
        switch (action) {
            case 'drill':
                window.location.href = `/drill.html?storyId=${storyId}`;
                break;
            case 'mark_inactive':
                markStory(storyId, 'inactive');
                break;
            case 'mark_unread':
                markStory(storyId, 'unread');
                break;
            case 'mark_active':
                markStory(storyId, 'active');
                break;
            default:
                window.location.href = `/story.html?storyId=${storyId}`;
                break;
        }
    }
};


function getStoryList() {
    fetch('/stories_list', {
        method: 'GET', // or 'PUT'
        headers: {
            'Content-Type': 'application/json',
        }
    }).then((response) => response.json())
        .then((data) => {
            console.log('Stories list success:', data);
            updateStoryList(data);
        })
        .catch((error) => {
            console.error('Error:', error);
        });
}

function markStory(id, action) {
    fetch(`/mark/${action}/${id}`, {
        method: 'GET',
        headers: {
            'Content-Type': 'application/json',
        }
    }).then((response) => response.json())
        .then((data) => {
            console.log(`Success ${action}:`, data);
            getStoryList();
        })
        .catch((error) => {
            console.error('Error marking story:', error);
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
            storyTitle.innerText = data.title;
            story.tokens = JSON.parse(story.tokens);
            console.log(`/story/${id} success:`, story);
            displayStory(data);
        })
        .catch((error) => {
            console.error('Error:', error);
        });
}

// function displayDefinition(index) {
//     var token = story.tokens[index];
//     getKanji(token.baseForm + token.surface); // might as well get all possibly relevant kanji
//     html = '';
//     for (let entry of token.entries) {
//         html += displayEntry(entry);
//     }
//     definitionsDiv.innerHTML = html;
// }

// storyText.onmouseup = function (evt) {
//     console.log(document.getSelection().toString());
// };

// storyText.onmouseleave = function (evt) {
//     console.log(document.getSelection().toString());
// };

// document.body.addEventListener('selectionchange', (event) => {
//     console.log('changed');
// });


// window.setInterval(function () {
//     console.log(document.getSelection().toString());
// }, 300);

