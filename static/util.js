function displayEntry(entry) {
    html = `<div class="entry"><div class="word">`;

    html += '<div class="readings">'
    for (var reading of entry.readings || []) {
        html += `<div class="reading">${reading.reading}</div>`;
    }

    html += '</div><div class="kanji_spellings">'
    for (var kanji of entry.kanji_spellings || []) {
        html += `<div class="kanji_spelling">${kanji.kanji_spelling}</div>`;
    }

    html += `</div></div><div class="senses">`;

    for (var sense of entry.senses || []) {
        html += `<div class="sense">
            <span class="pos">${sense.parts_of_speech.join(', ')}</span>
            <span class="glosses">${sense.glosses.map(x => x.value).join('; &nbsp;&nbsp;')}</span>
        </div>`;
    }

    return html + `</div></div><hr>`;
}

function displayKanji(kanji) {
    html = '';

    if (!kanji || kanji.length === 0) {
        kanjiResultsDiv.innerHTML = '';
        return;
    }

    for (let k of kanji) {       
        for (let group of k.readingmeaning.group) {
            onyomi = group.reading.filter(x => x.type === 'ja_on').map(x => `<span class="kanji_reading">${x.value}</span>`);
            kunyomi = group.reading.filter(x => x.type === 'ja_kun').map(x => `<span class="kanji_reading">${x.value}</span>`);

            var meanings = group.meaning.filter(x => !x.language).map(x => x.value);

            var misc = '';
            if (k.misc.stroke_count) {
                misc += `<span class="strokes">strokes: ${k.misc.stroke_count}</span>`;
            }
            if (k.misc.frequency) {
                misc += `<span class="frequency">frequency: ${k.misc.frequency}</span>`;
            }

            html += `<div class="kanji">
                    <div>
                    <span class="literal">${k.literal}</span>
                    <span class="onyomi_readings">${onyomi.join('')}</span>
                    <span class="kunyomi_readings">${kunyomi.join('')}</span>
                    </div>
                    <div class="kanji_meanings">${meanings.join(';  &nbsp;&nbsp;')}</div>
                    <div class="kanji_misc">${misc}</div>
                    </div>`;
        }
    }

    kanjiResultsDiv.innerHTML = html;
}
