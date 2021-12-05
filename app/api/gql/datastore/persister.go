package datastore

import (
	"database/sql"
	"encoding/json"
	"log"
	"os"
)

/////////////////
// Persistance //
/////////////////
type JSONPersister struct {
	File string
}

var _ Persister = (*JSONPersister)(nil)

func NewJSONPersister() JSONPersister {
	return JSONPersister{
		File: "./data.json",
	}
}

func (p JSONPersister) Write(d *Data) error {
	if p.File == "" {
		log.Println("this data object has no persisted layer, skip file writes")
		return nil
	}

	bytes, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(p.File, bytes, os.ModePerm)
}

func (p JSONPersister) Read() *Data {
	data := &Data{}
	bytes, err := os.ReadFile(p.File)
	if err != nil {
		log.Println(err, ", fail to load data from json, skip persistence")
		return data
	}
	err = json.Unmarshal(bytes, data)
	if err != nil {
		log.Println(err, "fail to load data from json, skip persistence")
		return data
	}
	return data
}

type PostgresPersister struct {
	db *sql.DB
}

var _ Persister = (*PostgresPersister)(nil)

func NewPostgresPersister(db *sql.DB) PostgresPersister {
	return PostgresPersister{
		db: db,
	}
}

func (p PostgresPersister) Write(d *Data) error {
	if p.db == nil {
		log.Println("this data object has no persisted layer, skip file writes")
		return nil
	}
	statement := `
		UPDATE json_persister
		SET data = $1
		WHERE id = 1;
	`
	bytes, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		return err
	}
	_, err = p.db.Exec(statement, bytes)
	return err
}

func (p PostgresPersister) Read() *Data {
	row := p.db.QueryRow(`
	SELECT data FROM json_persister WHERE id = 1;
	`)
	var bytes []byte
	err := row.Scan(&bytes)
	if err != nil {
		log.Println(err, ", fail to load data from postgres, skip persistence")
	}

	data := &Data{}
	err = json.Unmarshal(bytes, data)
	if err != nil {
		log.Println(err, ", fail to load data from json, skip persistence")
		return data
	}
	return data
}
