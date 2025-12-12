package main

import (
	"database/sql"
	"encoding/json"
	"errors"
)

type VocabDB struct {
	DB *sql.DB
}

// Create table if not exists
func InitVocabDB() (VocabDB, error) {
	query := `
CREATE TABLE IF NOT EXISTS vocab (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    word TEXT NOT NULL UNIQUE,
    kana TEXT,
    part_of_speech TEXT,
    definition TEXT,
    kanji_meanings TEXT,
    drill_count INTEGER NOT NULL DEFAULT 0,
    archived BOOLEAN NOT NULL DEFAULT 0
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

// INSERT
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

// GET BY ID
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

// UPDATE
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

// DELETE
func (v *VocabDB) Delete(id int) error {
	_, err := v.DB.Exec("DELETE FROM vocab WHERE id = ?", id)
	return err
}

// LIST ALL
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
			ORDER BY id
			WHERE archived = 0
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
