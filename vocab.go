package main

import (
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"errors"
	"io"
	"os"
	"strings"
)

type VocabDB struct {
	DB *sql.DB
}

func InitVocabDB() (VocabDB, error) {
	query := `
CREATE TABLE IF NOT EXISTS vocab (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    word TEXT NOT NULL UNIQUE,
    kana TEXT,
    part_of_speech TEXT,
    definition TEXT,
    kanji_meanings TEXT,
	drill_todo INTEGER NOT NULL DEFAULT 0,
    drill_count INTEGER NOT NULL DEFAULT 0,
    archived BOOLEAN NOT NULL DEFAULT 0,
	not_a_word BOOLEAN NOT NULL DEFAULT 0
);`

	db, err := sql.Open("sqlite", "vocab.db")
	if err != nil {
		return VocabDB{}, err
	}

	_, err = db.Exec(query)
	return VocabDB{DB: db}, nil
}

func (v *VocabDB) Close() error {
	return v.DB.Close()
}

func (v *VocabDB) Insert(vocab *Vocab) (int, error) {
	kmJSON, err := json.Marshal(vocab.KanjiMeanings)
	if err != nil {
		return 0, err
	}

	res, err := v.DB.Exec(`
        INSERT INTO vocab (word, kana, part_of_speech, definition, kanji_meanings, drill_count, archived)
        VALUES (?, ?, ?, ?, ?, ?, ?)
    `,
		vocab.Word,
		vocab.Kana,
		vocab.PartOfSpeech,
		vocab.Definition,
		string(kmJSON),
		vocab.DrillCount,
		vocab.Archived,
	)

	if err != nil {
		return 0, err
	}

	id64, err := res.LastInsertId()
	return int(id64), err
}

func (v *VocabDB) UpsertAll(list []Vocab) error {
	tx, err := v.DB.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(`
        INSERT INTO vocab (
            id, word, kana, part_of_speech, definition,
            kanji_meanings, drill_count, archived
        ) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
        ON CONFLICT(word) DO UPDATE SET
            kana = excluded.kana,
            part_of_speech = excluded.part_of_speech,
            definition = excluded.definition,
            kanji_meanings = excluded.kanji_meanings            
    `)
	if err != nil {
		tx.Rollback()
		return err
	}
	defer stmt.Close()

	for _, vocab := range list {
		kmJSON, err := json.Marshal(vocab.KanjiMeanings)
		if err != nil {
			tx.Rollback()
			return err
		}

		// If ID = 0, insert a NULL so SQLite will auto-assign
		var id interface{}
		if vocab.ID == 0 {
			id = nil
		} else {
			id = vocab.ID
		}

		_, err = stmt.Exec(
			id,
			vocab.Word,
			vocab.Kana,
			vocab.PartOfSpeech,
			vocab.Definition,
			string(kmJSON),
			vocab.DrillCount,
			vocab.Archived,
		)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

func (v *VocabDB) Get(id int) (*Vocab, error) {
	row := v.DB.QueryRow(`
        SELECT id, word, kana, part_of_speech, definition, kanji_meanings, drill_count, archived
        FROM vocab
        WHERE id = ?
    `, id)

	var kmJSON string
	var vc Vocab

	err := row.Scan(
		&vc.ID,
		&vc.Word,
		&vc.Kana,
		&vc.PartOfSpeech,
		&vc.Definition,
		&kmJSON,
		&vc.DrillCount,
		&vc.Archived,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if kmJSON != "" {
		err = json.Unmarshal([]byte(kmJSON), &vc.KanjiMeanings)
		if err != nil {
			return nil, err
		}
	}

	return &vc, nil
}

func (v *VocabDB) Update(vocab *Vocab) error {
	kmJSON, err := json.Marshal(vocab.KanjiMeanings)
	if err != nil {
		return err
	}

	_, err = v.DB.Exec(`
        UPDATE vocab
        SET word = ?, kana = ?, part_of_speech = ?, definition = ?,
            kanji_meanings = ?, drill_count = ?, archived = ?
        WHERE id = ?
    `,
		vocab.Word,
		vocab.Kana,
		vocab.PartOfSpeech,
		vocab.Definition,
		string(kmJSON),
		vocab.DrillCount,
		vocab.Archived,
		vocab.ID,
	)

	return err
}

func (v *VocabDB) Delete(id int) error {
	_, err := v.DB.Exec("DELETE FROM vocab WHERE id = ?", id)
	return err
}

func (v *VocabDB) ListAll(excludeArchived bool) ([]Vocab, error) {
	queryStr := `
        SELECT id, word, kana, part_of_speech, definition, kanji_meanings, drill_count, archived
        FROM vocab
        ORDER BY id
    `
	if excludeArchived {
		queryStr = `
			SELECT id, word, kana, part_of_speech, definition, kanji_meanings, drill_count, archived
			FROM vocab
			WHERE archived = 0
			ORDER BY id
		`
	}
	rows, err := v.DB.Query(queryStr)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []Vocab

	for rows.Next() {
		var vc Vocab
		var kmJSON string

		err := rows.Scan(
			&vc.ID,
			&vc.Word,
			&vc.Kana,
			&vc.PartOfSpeech,
			&vc.Definition,
			&kmJSON,
			&vc.DrillCount,
			&vc.Archived,
		)
		if err != nil {
			return nil, err
		}

		if kmJSON != "" {
			if err := json.Unmarshal([]byte(kmJSON), &vc.KanjiMeanings); err != nil {
				return nil, err
			}
		}

		list = append(list, vc)
	}

	return list, nil
}

func (v *VocabDB) FindByWord(term string) ([]Vocab, error) {
	pattern := "%" + term + "%"

	rows, err := v.DB.Query(`
        SELECT id, word, kana, part_of_speech, definition,
               kanji_meanings, drill_count, archived
        FROM vocab
        WHERE word LIKE ?
        ORDER BY word ASC
    `, pattern)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []Vocab

	for rows.Next() {
		var vc Vocab
		var kmJSON string

		err := rows.Scan(
			&vc.ID,
			&vc.Word,
			&vc.Kana,
			&vc.PartOfSpeech,
			&vc.Definition,
			&kmJSON,
			&vc.DrillCount,
			&vc.Archived,
		)
		if err != nil {
			return nil, err
		}

		if kmJSON != "" {
			if err := json.Unmarshal([]byte(kmJSON), &vc.KanjiMeanings); err != nil {
				return nil, err
			}
		}

		results = append(results, vc)
	}

	return results, nil
}

func ParseVocabCSVLine(line string) (Vocab, error) {
	r := csv.NewReader(strings.NewReader(line))
	r.FieldsPerRecord = -1

	fields, err := r.Read()
	if err != nil {
		return Vocab{}, err
	}
	if len(fields) < 5 {
		return Vocab{}, errors.New("invalid line: expected 5 fields")
	}

	kanjiInfos := parseKanjiMeanings(fields[4])

	return Vocab{
		Word:          fields[0],
		Kana:          fields[1],
		PartOfSpeech:  fields[2],
		Definition:    fields[3],
		KanjiMeanings: kanjiInfos,
	}, nil
}

func parseKanjiMeanings(s string) []KanjiInfo {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}

	parts := strings.Split(s, "/")

	var result []KanjiInfo
	for _, part := range parts {
		p := strings.TrimSpace(part)
		if p == "" {
			continue
		}

		// Expect something like: "躊: meaning here"
		kv := strings.SplitN(p, ":", 2)
		if len(kv) != 2 {
			continue // skip malformed segments
		}

		meaning := strings.TrimSpace(kv[1])
		kv = strings.SplitN(kv[0], "(", 2)
		kanji := strings.TrimSpace(kv[0])
		pronunciation := strings.TrimSpace(kv[1])
		pronunciation = strings.TrimSuffix(pronunciation, ")")

		if kanji != "" && meaning != "" {
			result = append(result, KanjiInfo{
				Kanji:         kanji,
				Pronunciation: pronunciation,
				Meaning:       meaning,
			})
		}
	}

	return result
}

func LoadVocabCSVFile(path string) ([]Vocab, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	reader := csv.NewReader(f)
	reader.FieldsPerRecord = -1

	var vocabList []Vocab
	for {
		fields, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if len(fields) < 5 {
			continue // skip invalid lines
		}

		v := Vocab{
			Word:          fields[0],
			Kana:          fields[1],
			PartOfSpeech:  fields[2],
			Definition:    fields[3],
			KanjiMeanings: parseKanjiMeanings(fields[4]),
		}
		vocabList = append(vocabList, v)
	}

	return vocabList, nil
}

func VocabToCSVLine(v Vocab) string {
	// Serialize kanji meanings
	var kmParts []string
	for _, k := range v.KanjiMeanings {
		kmParts = append(kmParts, k.Kanji+"("+k.Pronunciation+"): "+k.Meaning)
	}
	kmField := strings.Join(kmParts, " / ")

	// Escape fields with quotes if necessary
	escape := func(s string) string {
		if strings.ContainsAny(s, `",`) {
			s = strings.ReplaceAll(s, `"`, `""`)
			return `"` + s + `"`
		}
		return s
	}

	return strings.Join([]string{
		escape(v.Word),
		escape(v.Kana),
		escape(v.PartOfSpeech),
		escape(v.Definition),
		escape(kmField),
	}, ",")
}

func SaveVocabCSVFile(path string, vocabList []Vocab) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	writer := csv.NewWriter(f)
	defer writer.Flush()

	// Write header
	writer.Write([]string{"word", "romaji", "part_of_speech", "definition", "kanji_meanings"})

	for _, v := range vocabList {
		line := VocabToCSVLine(v)
		// csv.Writer handles proper escaping again
		if err := writer.Write(strings.Split(line, ",")); err != nil {
			return err
		}
	}
	return nil
}
