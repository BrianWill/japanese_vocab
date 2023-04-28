function splitOnHighPitch(str, pitch) {
    if (pitch === 0) {
        return ['', '', str];
    }
    let mora = [];
    let s = new Set(['ゅ', 'ょ', 'ゃ', 'ャ', 'ュ', 'ョ']);
    let chars = str.split('');
    for (let i = 0; i < chars.length; i++) {
        if (s.has(chars[i + 1])) {
            mora.push(chars[i] + chars[i + 1]);
            i++;
        } else {
            mora.push(chars[i]);
        }
    }
    return [
        mora.slice(0, pitch - 1).join(''),
        mora.slice(pitch - 1, pitch).join(''),
        mora.slice(pitch).join('')
    ];
}

function displayKanji(kanji, word) {   
    html = '';

    if (!kanji || kanji.length === 0) {
        kanjiResultsDiv.innerHTML = '';
        return;
    }

    for (let ch of new Set(word.split(''))) {
        for (let k of kanji) {
            if (k.literal === ch) {
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
                            <div><span class="onyomi_readings">${onyomi.join('')}</span></div>
                            <div><span class="kunyomi_readings">${kunyomi.join('')}</span></div>
                            </div>
                            <div class="kanji_meanings">${meanings.join(';  &nbsp;&nbsp;')}</div>
                            <div class="kanji_misc">${misc}</div>
                            </div>`;
                }
            }
        }
    }

    

    kanjiResultsDiv.innerHTML = html;
}

function getKanji(str) {
    fetch('/kanji', {
        method: 'POST', // or 'PUT'
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(str),
    }).then((response) => response.json()
    ).then((data) => {
        displayKanji(data.kanji, str);
    }).catch((error) => {
        console.error('Error:', error);
    });
}

function updateWord(word) {
    fetch('/update_word', {
        method: 'POST', // or 'PUT'
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(word),
    }).then((response) => response.json()
    ).then((data) => {
        //console.log(data);
    }).catch((error) => {
        console.error('Error:', error);
    });
}


function shuffle(array) {
    if (array.length < 2) {
        return;
    }

    let idx = array.length;
    while (idx != 0) {
        // Pick a remaining element.
        let randIdx = Math.floor(Math.random() * idx);
        idx--;

        // And swap it with the current element.
        [array[idx], array[randIdx]] = [array[randIdx], array[idx]];
    }
}


function displayEntry(entry) {
    html = `<div class="entry"><div class="word">`;

    html += '<div class="readings">';
    for (var reading of entry.readings || []) {
        if (reading.pitch) {
            let parts = splitOnHighPitch(reading.reading, parseInt(reading.pitch));
            html += `<div class="reading">${parts[0]}<span class="high_pitch">${parts[1]}</span>${parts[2]}</div>`;
        } else {
            html += `<div class="reading unknown_pitch">${reading.reading}﹖</div>`;
        }
    }

    html += '</div><div class="kanji_spellings">';
    for (var kanji of entry.kanji_spellings || []) {
        html += `<div class="kanji_spelling">${kanji.kanji_spelling}</div>`;
    }

    html += `</div></div><div class="senses">`;

    for (var sense of entry.senses || []) {
        let pos = sense.parts_of_speech.map(x => `<span class="pos">${x}</span>`); 
        html += `<div class="sense">
            <div>${pos.join(' ')}</div>
            <div class="glosses">${sense.glosses.map(x => x.value).join('; &nbsp;&nbsp;')}</div>
        </div>`;
    }

    return html + `</div></div>`;
}