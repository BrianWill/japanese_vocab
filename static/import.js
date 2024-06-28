let sourcesDiv = document.getElementById('sources');
let malformedPathsDiv = document.getElementById('malformed_paths');

document.body.onload = function (evt) {
    getSources(displaySources);
};


sourcesDiv.onclick = function (evt) {
    if (evt.target.classList.contains('import_link')) {
        evt.preventDefault();
        let source = evt.target.getAttribute('source');

        importSource(source, () => {
            snackbarMessage(`imported source: ${source}`);
        });
    }
};

function displaySources(data) {
    let tableHeader = `<table class="sources_table">
        <tr style="text-align: left;"><th></th><th>Source</th><th>Count<br>imported</th><th>Count<br>unimported</th></td>`;
    let html = tableHeader;

    for (let source in data.storyFilePathsBySource) {
        let importedCount = 0;
        if  (data.storiesBySource[source]) {
            importedCount = data.storiesBySource[source].length;
        }
        let count = data.storyFilePathsBySource[source].length;
        count = Math.max(0, count - importedCount);
        html += `<tr>
            <td><a href="#" source="${source}" class="import_link">import</a></td>
            <td><span class="story_source">${source}</span></td>
            <td><span>${importedCount}</span></td>
            <td><span>${count}</span></td>
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

