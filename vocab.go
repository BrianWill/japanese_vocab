package main

import (
	"context"
	"database/sql"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

type DB struct {
	DB *sql.DB
	context.Context
}

func InitVocabDB() (DB, error) {
	query := `
CREATE TABLE IF NOT EXISTS vocab (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    word TEXT NOT NULL UNIQUE,
    kana TEXT NOT NULL DEFAULT '',
    part_of_speech TEXT NOT NULL DEFAULT '',
    definition TEXT NOT NULL DEFAULT '',
	drill_todo INTEGER NOT NULL DEFAULT 0,
    drill_count INTEGER NOT NULL DEFAULT 0,
    status INTEGER NOT NULL DEFAULT 0
);`

	db, err := sql.Open("sqlite", "vocab.db")
	if err != nil {
		return DB{}, err
	}

	_, err = db.Exec(query)
	return DB{DB: db}, err
}

func (v *DB) Close() error {
	return v.DB.Close()
}

func (v *DB) Insert(vocab *Vocab) (int, error) {
	res, err := v.DB.Exec(`
        INSERT INTO vocab (word, kana, part_of_speech, definition, drill_count, status)
        VALUES (?, ?, ?, ?, ?, ?)
    `,
		vocab.Word,
		vocab.Kana,
		vocab.PartOfSpeech,
		vocab.Definition,
		vocab.DrillCount,
		vocab.Status,
	)

	if err != nil {
		return 0, err
	}

	id64, err := res.LastInsertId()
	return int(id64), err
}

func (db *DB) InsertOrGetVocabBatch(ctx context.Context, tokens []*JpToken) ([]*Vocab, error) {

	if len(tokens) == 0 {
		return nil, nil
	}

	tx, err := db.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT OR IGNORE INTO vocab (
			word,
			status
		) VALUES (?, ?)
	`)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	words := make([]string, 0, len(tokens))

	for _, t := range tokens {
		_, err := stmt.ExecContext(
			ctx,
			t.BaseForm,
			VOCAB_STATUS_DISABLED,
		)
		if err != nil {
			return nil, err
		}
		words = append(words, t.BaseForm)
	}

	// Build: WHERE word IN (?, ?, ...)
	placeholders := strings.Repeat("?,", len(words))
	placeholders = placeholders[:len(placeholders)-1] // get rid of trailing comma

	query := fmt.Sprintf(`
		SELECT
			id,
			word,
			kana,
			part_of_speech,
			definition,
			drill_todo,
			drill_count,
			status
		FROM vocab
		WHERE word IN (%s)
	`, placeholders)

	args := make([]any, len(words))
	for i, w := range words {
		args[i] = w
	}

	rows, err := tx.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*Vocab
	for rows.Next() {
		var v Vocab
		if err := rows.Scan(
			&v.ID,
			&v.Word,
			&v.Kana,
			&v.PartOfSpeech,
			&v.Definition,
			&v.DrillTodo,
			&v.DrillCount,
			&v.Status,
		); err != nil {
			return nil, err
		}

		results = append(results, &v)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return results, nil
}

func (v *DB) UpsertAll(list []Vocab) error {
	tx, err := v.DB.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(`
        INSERT INTO vocab (
            id, word, kana, part_of_speech, definition,
            drill_count, drill_todo, status
        ) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
        ON CONFLICT(word) DO UPDATE SET
            kana = excluded.kana,
            part_of_speech = excluded.part_of_speech,
            definition = excluded.definition
    `)
	if err != nil {
		tx.Rollback()
		return err
	}
	defer stmt.Close()

	for _, vocab := range list {
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
			vocab.DrillCount,
			vocab.DrillTodo,
			vocab.Status,
		)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

func (v *DB) Get(id int) (*Vocab, error) {
	row := v.DB.QueryRow(`
        SELECT id, word, kana, part_of_speech, definition, drill_count, status
        FROM vocab
        WHERE id = ?
    `, id)

	var vc Vocab

	err := row.Scan(
		&vc.ID,
		&vc.Word,
		&vc.Kana,
		&vc.PartOfSpeech,
		&vc.Definition,
		&vc.DrillCount,
		&vc.Status,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &vc, nil
}

func (v *DB) Update(vocab *Vocab) error {
	_, err := v.DB.Exec(`
        UPDATE vocab
        SET word = ?, kana = ?, part_of_speech = ?, definition = ?,
            drill_count = ?, status = ?
        WHERE id = ?
    `,
		vocab.Word,
		vocab.Kana,
		vocab.PartOfSpeech,
		vocab.Definition,
		vocab.DrillCount,
		vocab.Status,
		vocab.ID,
	)

	return err
}

func (v *DB) Delete(id int) error {
	_, err := v.DB.Exec("DELETE FROM vocab WHERE id = ?", id)
	return err
}

func (v *DB) ListAll(activeOnly bool) ([]Vocab, error) {
	queryStr := `
        SELECT id, word, kana, part_of_speech, definition, drill_count, drill_todo, status
        FROM vocab
        ORDER BY id
    `
	if activeOnly {
		queryStr = `
			SELECT id, word, kana, part_of_speech, definition, drill_count, drill_todo, status
			FROM vocab
			WHERE status = ` + strconv.Itoa(int(VOCAB_STATUS_ENABLED)) + `
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

		err := rows.Scan(
			&vc.ID,
			&vc.Word,
			&vc.Kana,
			&vc.PartOfSpeech,
			&vc.Definition,
			&vc.DrillCount,
			&vc.DrillTodo,
			&vc.Status,
		)
		if err != nil {
			return nil, err
		}

		list = append(list, vc)
	}

	return list, nil
}

func (v *DB) FindByWord(term string) ([]Vocab, error) {
	pattern := "%" + term + "%"

	rows, err := v.DB.Query(`
        SELECT id, word, kana, part_of_speech, definition,
               drill_count, drill_todo, status
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

		err := rows.Scan(
			&vc.ID,
			&vc.Word,
			&vc.Kana,
			&vc.PartOfSpeech,
			&vc.Definition,
			&vc.DrillCount,
			&vc.DrillTodo,
			&vc.Status,
		)
		if err != nil {
			return nil, err
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

	return Vocab{
		Word:         fields[0],
		Kana:         fields[1],
		PartOfSpeech: fields[2],
		Definition:   fields[3],
	}, nil
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
			Word:         fields[0],
			Kana:         fields[1],
			PartOfSpeech: fields[2],
			Definition:   fields[3],
		}
		vocabList = append(vocabList, v)
	}

	return vocabList, nil
}
