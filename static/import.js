let sourcesDiv = document.getElementById('sources');
let malformedPathsDiv = document.getElementById('malformed_paths');

document.body.onload = function (evt) {
    getSources(displaySources);
};

function displaySources(data) {
    let tableHeader = `<table class="sources_table">
        <tr style="text-align: left;"><th></th><th>Source</th><th>Count<br>unimported</th><th>Count<br>imported</th></td>`;
    let html = tableHeader;

    for (let source in data.storyFilePathsBySource) {
        let importedCount = 0;
        if  (data.storiesBySource[source]) {
            importedCount = data.storiesBySource[source].length;
        }
        let count = data.storyFilePathsBySource[source].length;
        count = Math.max(0, count - importedCount);
        html += `<tr>
            <td><a source="${source}" class="import_link">import</a></td>
            <td><span class="story_source">${source}</span></td>
            <td><span>${count}</span></td>
            <td><span>${importedCount}</span></td>
            </tr>`;
    }
    html += `</table>`;

    sourcesDiv.innerHTML = html;

    html = ``;
    for (let path of data.malformedPaths) {
        html += `<div>${path}</div>`;
    }
    malformedPathsDiv.innerHTML = html;
};

