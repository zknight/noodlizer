package db

import "database/sql"

type PendingUser struct {
	FirstName string
	LastName  string
	ReqLogin  string
	Email     string
	Notes     string
}

type User struct {
	FirstName string
	LastName  string
	Email     string
}

func (d *DB) createPenduserTable() error {
	_, err := d.db.Exec(`
		drop table if exists penduser;
		create table user (
			first STRING NOT NULL,
			last STRING NOT NULL,
			email STRING NOT NULL,
			notes STRING NOT NULL
		);
	`)
	return err
}

func (d *DB) createUserTable() error {
	_, err := d.db.Exec(`
		drop table if exists user;
		create table user (
			id INTEGER PRIMARY KEY,
			first STRING NOT NULL,
			last STRING NOT NULL,
			email STRING NUT NULL,
			band_id INTEGER,
			pref_id INTEGER
		)
	`)
	return err
}

func (d *DB) AddPending(first, last, email, loginid, notes string) error {
	q := "insert into penduser (first, last, email, loginid, notes) values ($1, $2, $3, $4, $5);"
	_, err := d.db.Exec(q, first, last, email, loginid, notes)
	return err
}

func (d *DB) GetPending() ([]PendingUser, error) {
	q := "SELECT first, last, email, loginid, notes FROM penduser;"
	pus := []PendingUser{}
	rows, err := d.db.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var first string
		var last string
		var email string
		var loginid string
		var notes_null sql.NullString
		var notes string

		err = rows.Scan(&first, &last, &email, &loginid, &notes_null)
		if err != nil {
			return nil, err
		}
		if notes_null.Valid {
			notes = notes_null.String
		}
		pu := PendingUser{
			FirstName: first,
			LastName:  last,
			Email:     email,
			ReqLogin:  loginid,
			Notes:     notes,
		}
		pus = append(pus, pu)
	}
	return pus, nil
}
