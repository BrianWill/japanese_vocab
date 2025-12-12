let sourcesDiv = document.getElementById('sources');
let malformedPathsDiv = document.getElementById('malformed_paths');

document.body.onload = function (evt) {
    getSources(displaySources);
};


sourcesDiv.onclick = function (evt) {
    if (evt.target.classList.contains('import_link')) {
        evt.preventDefault();
        let source = evt.target.getAttribute('source');

        // very large duration to make it practically indefinite
        snackbarMessage(`importing source "${source}". This may take a while.`, 10000 * 1000);

        importSource(source, false, () => {
                clearSnackbarMessage();
                snackbarMessage(`COMPLETED IMPORT: source "${source}"`);
                getSources(displaySources);
            },
            () => {
                clearSnackbarMessage();
                snackbarMessage(`FAILED IMPORT: source "${source}"`);
            });
    } else if (evt.target.classList.contains('import_new_link')) {
        evt.preventDefault();
        let source = evt.target.getAttribute('source');

        // very large duration to make it practically indefinite
        snackbarMessage(`importing new stories from source "${source}". This may take a while.`, 10000 * 1000);

        importSource(source, true, () => {
                clearSnackbarMessage();
                snackbarMessage(`COMPLETED IMPORT: source "${source}"`);
                getSources(displaySources);
            },
            () => {
                clearSnackbarMessage();
                snackbarMessage(`FAILED IMPORT: source "${source}"`);
            });
    } else if (evt.target.classList.contains('remove_link')) {
        evt.preventDefault();
        let source = evt.target.getAttribute('source');

        if (!window.confirm(`Remove all previously imported stories of ${source}?`)) {
            return;
        }

        // very large duration to make it practically indefinite
        snackbarMessage(`removing all previously imported stories of "${source}". This may take a while.`, 10000 * 1000);

        removeSource(source, () => {
            clearSnackbarMessage();
            snackbarMessage(`COMPLETED REMOVAL OF STORIES: source "${source}"`);
            getSources(displaySources);
        },
            () => {
                clearSnackbarMessage();
                snackbarMessage(`FAILED REMOVAL OF STORIES: source "${source}"`);
            });
    }
};

function displaySources(data) {
    let tableHeader = `<table class="sources_table">
        <tr style="text-align: left;"><th></th><th></th><th></th><th>Source</th><th>Imported</th><th>Unimported</th></td>`;
    let html = tableHeader;

    for (let source in data.storyFilePathsBySource) {
        let importedCount = 0;
        if (data.storiesBySource[source]) {
            importedCount = data.storiesBySource[source].length;
        }
        let count = data.storyFilePathsBySource[source].length;
        count = Math.max(0, count - importedCount);
        html += `<tr>
            <td><a href="#" title="import all, including re-import of already imported stories" source="${source}" class="import_link">import all</a></td>
            <td><a href="#" title="import only the new stories" source="${source}" class="import_new_link">import new</a></td>
            <td><a href="#" source="${source}" class="remove_link">remove</a></td>
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

