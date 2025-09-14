package db

import (
	"bytes"
	"crypto/rand"
	"encoding/gob"
	"fmt"
	"math/big"
)

func (d *DB) NewGig(sl Setlist) (*Gig, error) {
	// generate key (rando)
	id, err := rand.Int(rand.Reader, big.NewInt(0x7FFFFFFF))
	if err != nil {
		return nil, err
	}
	// create a new gig
	g := NewGig(sl)
	// GOBinate the gig
	g.Id = id.Int64()
	var buf bytes.Buffer
	gob_g := gob.NewEncoder(&buf)
	err = gob_g.Encode(g)
	if err != nil {
		return nil, err
	}
	// store in db
	q := "insert into gig (id, obj) values ($1, $2);"
	_, err = d.db.Exec(q, id.Int64(), buf.Bytes())

	//return it
	return g, err
}

func (d *DB) GetGig(id int64) (*Gig, error) {
	fmt.Println("getGig: ", id)
	q := "select obj from gig where id=$1;"
	var (
		objd []byte
	)
	err := d.db.QueryRow(q, id).Scan(&objd)
	if err != nil {
		return nil, err
	}
	buf := bytes.NewBuffer(objd)
	g := Gig{}
	gob_g := gob.NewDecoder(buf)
	err = gob_g.Decode(&g)
	if err == nil {
		fmt.Printf("getGig: %s id %d set %d track %d\n", g.Name, g.Id, g.CurSet, g.CurTrack)
	}

	return &g, err
}

func (d *DB) UpdateGig(g *Gig) error {
	var buf bytes.Buffer
	fmt.Printf("updateGig:%s id %d set %d track %d\n", g.Name, g.Id, g.CurSet, g.CurTrack)
	gob_g := gob.NewEncoder(&buf)
	err := gob_g.Encode(*g)
	if err != nil {
		return err
	}
	q := "update gig set obj=$2 where id=$1;"
	objd := buf.Bytes()
	fmt.Println("length of objd (bytes)=", len(objd))
	res, err := d.db.Exec(q, g.Id, objd)
	nr, rerr := res.RowsAffected()
	if rerr != nil {
		fmt.Println("err RowsAffected: ", rerr.Error())
	}
	fmt.Println("rows affected: ", nr)
	return err
}

func (d *DB) RemoveGig(id int64) error {
	fmt.Println("removing gig ", id)
	q := "delete from gig where id=$1;"
	res, err := d.db.Exec(q, id)
	nr, rerr := res.RowsAffected()
	if rerr != nil {
		fmt.Println("err RowsAffected", rerr.Error())
	}
	fmt.Println("rows affected: ", nr)
	return err
}

func (d *DB) CleanupGigs() error {
	q := "delete from gig where 1"
	res, err := d.db.Exec(q)
	nr, rerr := res.RowsAffected()
	if rerr != nil {
		fmt.Println("err RowsAffected", rerr.Error())
	}
	fmt.Println("rows affected: ", nr)
	return err
}
