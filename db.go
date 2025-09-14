package main

import (
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

type DB struct {
	db *sql.DB
}

func new_db(path string) (*DB, error) {
	d, err := open_db(path)
	if err != nil {
		return nil, err
	}
	err = d.init()
	return d, err
}

func open_db(path string) (*DB, error) {
	d, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	return &DB{db: d}, nil
}

func (d *DB) init() error {
	_, err := d.db.Exec(`
	drop table if exists track;
	create table track (
		id INTEGER primary key,
		title TEXT NOT NULL UNIQUE,
		tempo INTEGER NOT NULL,
		click INTEGER NOT NULL,
		kit_id INTEGER,
		vox_id INTEGER,
		era_id INTEGER,
		genre_id INTEGER,
		lyrics_id INTEGER,
		key_tone TEXT
	);
	drop table if exists vox;
	create table vox (
		id INTEGER primary key,
		name TEXT NOT NULL UNIQUE
	);
	drop table if exists era;
	create table era (
		id INTEGER primary key,
		name TEXT NOT NULL UNIQUE
	);
	drop table if exists genre;
	create table genre (
		id INTEGER primary key,
		name TEXT NOT NULL UNIQUE
	);
	drop table if exists kit;
	create table kit (
		id INTEGER primary key,
		name TEXT NOT NULL UNIQUE
	);
	drop table if exists lyrics;
	create table lyrics (
		id INTEGER primary key,
		text TEXT
	);
	drop table if exists setlist;
	create table setlist (
		id INTEGER primary key,
		name TEXT NOT NULL,
		timestamp INTEGER NOT NULL
	);
	drop table if exists a_set;
	create table a_set {
		id INTEGER primary key,
		setlist_id INTEGER NOT NULL
		name TEXT NOT NULL,
		setnum INTEGER NOT NULL
	};
	drop table if exists sets_tracks;
	create table sets_tracks {
		set_id INTEGER NOT NULL,
		track_id INTEGER NOT NULL,
		seq INTEGER NOT NULL
	};
	drop table if exists gig {
		id INTEGER primary key,
		obj BLOB NOT NULL,
	}
	`)

	return err
}

func (d *DB) getVoxByName(name string) (int64, error) {
	q := fmt.Sprintf("select id from vox WHERE name = '%s';", name)
	var id int64
	err := d.db.QueryRow(q).Scan(&id)
	return id, err
}

func (d *DB) getEraByName(name string) (int64, error) {
	q := fmt.Sprintf("select id from era WHERE name = '%s';", name)
	var id int64
	err := d.db.QueryRow(q).Scan(&id)
	return id, err
}

func (d *DB) getGenreByName(name string) (int64, error) {
	q := fmt.Sprintf("select id from genre WHERE name = '%s';", name)
	var id int64
	err := d.db.QueryRow(q).Scan(&id)
	return id, err
}

func (d *DB) getKitByName(name string) (int64, error) {
	q := fmt.Sprintf("select id from kit WHERE name = '%s';", name)
	var id int64
	err := d.db.QueryRow(q).Scan(&id)
	return id, err
}

var trackSelect string = `
select 
	track.id, track.title, track.tempo, track.click, track.key_tone, vox.id, vox.name, era.id, era.name, genre.id, genre.name, kit.id, kit.name 
from track 
	join vox on track.vox_id = vox.id
	join era on track.era_id = era.id
	join genre on track.genre_id = genre.id
	join kit on track.kit_id = kit.id
`

func (d *DB) getTracksByVox(vox_id int64) ([]Track, error) {
	q := trackSelect + "where track.vox_id = $1;"
	rows, err := d.db.Query(q, vox_id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return d._extractTracks(rows)
}

func (d *DB) getTracksByEra(era_id int64) ([]Track, error) {
	q := trackSelect + "where track.era_id = $1;"
	rows, err := d.db.Query(q, era_id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return d._extractTracks(rows)
}

func (d *DB) getTracksByGenre(genre_id int64) ([]Track, error) {
	q := trackSelect + "where track.genre_id = $1;"
	rows, err := d.db.Query(q, genre_id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return d._extractTracks(rows)
}

func (d *DB) getTracksByKit(kit_id int64) ([]Track, error) {
	q := trackSelect + "where track.kit_id = $1;"
	rows, err := d.db.Query(q, kit_id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return d._extractTracks(rows)
}

func (d *DB) getAllTracks() ([]Track, error) {
	q := trackSelect + ";"

	rows, err := d.db.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return d._extractTracks(rows)
}

func (d *DB) _extractTracks(rows *sql.Rows) ([]Track, error) {
	tracks := []Track{}
	for rows.Next() {
		var (
			id       int64
			title    string
			tempo    int64
			click    int64
			key_tone string
			vox_id   int64
			vox      string
			era_id   int64
			era      string
			genre_id int64
			genre    string
			kit_id   int64
			kit      string
		)
		err := rows.Scan(&id, &title, &tempo, &click, &key_tone, &vox_id, &vox, &era_id, &era, &genre_id, &genre, &kit_id, &kit)
		if err != nil {
			return nil, err
		}
		voxObj := Vox{Id: vox_id, Name: vox}
		eraObj := Era{Id: era_id, Name: era}
		genreObj := Genre{Id: genre_id, Name: genre}
		kitObj := Kit{Id: kit_id, Name: kit}
		track := Track{
			Id:      id,
			Title:   title,
			Tempo:   int(tempo),
			Click:   click == 1,
			KeyTone: key_tone,
			Vox:     voxObj, //fmt.Sprintf("vox %d", vox_id),
			Era:     eraObj,
			Genre:   genreObj, //fmt.Sprintf("genre %d", genre_id),
			Kit:     kitObj,   //fmt.Sprintf("kit %d", kit_id),
		}
		tracks = append(tracks, track)
	}
	return tracks, nil
}

func (d *DB) getAllVoxes() ([]Vox, error) {
	q := `
select
	id, name
from vox;`

	voxes := []Vox{}
	rows, err := d.db.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var id int64
		var name string

		err = rows.Scan(&id, &name)
		if err != nil {
			return nil, err
		}
		vox := Vox{Id: id, Name: name}
		voxes = append(voxes, vox)
	}
	return voxes, nil
}

func (d *DB) getAllEras() ([]Era, error) {
	q := `
select
	id, name
from era;`

	eras := []Era{}
	rows, err := d.db.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var id int64
		var name string

		err = rows.Scan(&id, &name)
		if err != nil {
			return nil, err
		}
		era := Era{Id: id, Name: name}
		eras = append(eras, era)
	}
	return eras, nil
}

func (d *DB) getAllGenres() ([]Genre, error) {
	q := `
select
	id, name
from genre;`

	genres := []Genre{}
	rows, err := d.db.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var id int64
		var name string

		err = rows.Scan(&id, &name)
		if err != nil {
			return nil, err
		}
		genre := Genre{Id: id, Name: name}
		genres = append(genres, genre)
	}
	return genres, nil
}

func (d *DB) getAllKits() ([]Kit, error) {
	q := `
select
	id, name
from kit;`

	kits := []Kit{}
	rows, err := d.db.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var id int64
		var name string

		err = rows.Scan(&id, &name)
		if err != nil {
			return nil, err
		}
		kit := Kit{Id: id, Name: name}
		kits = append(kits, kit)
	}
	return kits, nil
}

func (d *DB) getTrack(id int64) (Track, error) {
	/*
			q := `
		select
			track.id, track.title, track.tempo,
			track.click, track.vox.id, vox.name,
			era.id, era.name, genre.id, genre.name,
			kit.id, kit.name, lyrics.id, lyrics.text
		from track
			join vox on track.vox_id = vox.id
			join era on track.era_id = era.id
			join genre on track.genre_id = genre.id
			join kit on track.kit_id = kit.id
			join lyrics on track.lyrics_id = lyrics.id
		where track.id = $1;
			`
	*/
	q := `
select 
	track.id, track.title, track.tempo, 
	track.click, track.key_tone, vox.id, vox.name, 
	era.id, era.name, genre.id, genre.name,
	kit.id, kit.name, track.lyrics_id
from track
	join vox on track.vox_id = vox.id
	join era on track.era_id = era.id
	join genre on track.genre_id = genre.id
	join kit on track.kit_id = kit.id
where track.id = $1;
	`
	row := d.db.QueryRow(q, id)
	var (
		title          string
		tempo          int64
		click          int64
		key_tone       string
		vox_id         int64
		vox            string
		era_id         int64
		era            string
		genre_id       int64
		genre          string
		kit_id         int64
		kit            string
		lyrics_id_null sql.NullInt64
		lyrics_id      int64
		lyrics         string
	)

	err := row.Scan(&id, &title, &tempo, &click, &key_tone, &vox_id, &vox, &era_id, &era, &genre_id, &genre, &kit_id, &kit, &lyrics_id_null)
	if err != nil {
		return Track{}, err
	}
	if lyrics_id_null.Valid {
		lyrics_id = lyrics_id_null.Int64
		q = "select text from lyrics where id=$1"
		row := d.db.QueryRow(q, lyrics_id)
		err = row.Scan(&lyrics)
		if err != nil {
			return Track{}, err
		}
	}
	voxObj := Vox{Id: vox_id, Name: vox}
	eraObj := Era{Id: era_id, Name: era}
	genreObj := Genre{Id: genre_id, Name: genre}
	kitObj := Kit{Id: kit_id, Name: kit}
	lyricsObj := Lyrics{Id: lyrics_id, RawText: lyrics}
	t := Track{
		Id:      id,
		Title:   title,
		Tempo:   int(tempo),
		Click:   click == 1,
		KeyTone: key_tone,
		Vox:     voxObj,
		Era:     eraObj,
		Genre:   genreObj,
		Kit:     kitObj,
		Lyrics:  lyricsObj,
	}
	return t, nil
}

func (d *DB) updateTrack(track Track) error {
	q := `
update track
set
	title=$1, tempo=$2, click=$3, key_tone=$4, vox_id=$5, era_id=$6, genre_id=$7, kit_id=$8
where id=$9`
	_, err := d.db.Exec(q, track.Title, track.Tempo, track.Click, track.KeyTone, track.Vox.Id,
		track.Era.Id, track.Genre.Id, track.Kit.Id, track.Id)
	return err
}

func (d *DB) updateLyrics(lyrics Lyrics) error {
	q := `
update lyrics
set
	text=$1
where id=$2;`
	_, err := d.db.Exec(q, lyrics.RawText, lyrics.Id)
	return err
}

func (d *DB) addLyrics(track_id int64, lyrics Lyrics) (int64, error) {
	q := `insert into lyrics (text) values ($1)`
	res, err := d.db.Exec(q, lyrics.RawText)
	if err != nil {
		return -1, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return -1, err
	}
	q = `update track set lyrics_id=$1 where id=$2`
	_, err = d.db.Exec(q, id, track_id)
	if err != nil {
		return -1, err
	}

	return id, nil
}

func (d *DB) getEra(id int64) (Era, error) {
	q := `select name from era where id=$1`
	row := d.db.QueryRow(q, id)
	var name string
	err := row.Scan(&name)
	if err != nil {
		return Era{}, err
	}
	e := Era{Id: id, Name: name}
	return e, err
}

func (d *DB) addEra(name string) (int64, error) {
	return d._insertName("era", name)
}

func (d *DB) getGenre(id int64) (Genre, error) {
	q := `select name from genre where id=$1`
	row := d.db.QueryRow(q, id)
	var name string
	err := row.Scan(&name)
	if err != nil {
		return Genre{}, err
	}
	g := Genre{Id: id, Name: name}
	return g, err
}

func (d *DB) addGenre(name string) (int64, error) {
	return d._insertName("genre", name)
}

func (d *DB) getVox(id int64) (Vox, error) {
	q := `select name from vox where id=$1`
	row := d.db.QueryRow(q, id)
	var name string
	err := row.Scan(&name)
	if err != nil {
		return Vox{}, err
	}
	v := Vox{Id: id, Name: name}
	return v, err
}

func (d *DB) addVox(name string) (int64, error) {
	return d._insertName("vox", name)
}

func (d *DB) getKit(id int64) (Kit, error) {
	q := `select name from kit where id=$1`
	row := d.db.QueryRow(q, id)
	var name string
	err := row.Scan(&name)
	if err != nil {
		return Kit{}, err
	}
	k := Kit{Id: id, Name: name}
	return k, err
}
func (d *DB) addKit(name string) (int64, error) {
	return d._insertName("kit", name)
}

func (d *DB) getAllSetlists() ([]Setlist, error) {
	q := `
select 
	setlist.id, setlist.name, setlist.timestamp, 
	a_set.id, a_set.name, a_set.setnum 
from setlist
left join
	a_set on a_set.setlist_id = setlist.id;`

	rows, err := d.db.Query(q)
	if err != nil {
		return nil, err
	}

	return d._extractSetlists(rows)
}

func (d *DB) getSetlist(id int64) (Setlist, error) {
	q := `
SELECT 
	setlist.name, setlist.timestamp, a_set.id 
FROM setlist
JOIN
	a_set on a_set.setlist_id = setlist.id
WHERE
	setlist.id = $1;`
	rows, err := d.db.Query(q, id)
	if err != nil {
		return Setlist{}, err
	}
	var (
		name      string
		timestamp int64
		set_id    int64
	)
	sets := []Set{}
	for rows.Next() {
		err = rows.Scan(&name, &timestamp, &set_id)
		if err != nil {
			return Setlist{}, err
		}
		s, err := d.getSet(set_id)
		if err != nil {
			return Setlist{}, err
		}
		sets = append(sets, s)
	}
	sl := Setlist{Id: id, Name: name, Timestamp: timestamp, Sets: sets}

	return sl, nil
}

func (d *DB) _extractSetlists(rows *sql.Rows) ([]Setlist, error) {
	setlists := []Setlist{}
	last_id := int64(0)
	var cur_setlist *Setlist

	for rows.Next() {
		var (
			id            int64
			name          string
			tstamp        int64
			set_name_null sql.NullString
			set_id_null   sql.NullInt64
			set_num_null  sql.NullInt64
		)
		err := rows.Scan(&id, &name, &tstamp, &set_id_null, &set_name_null, &set_num_null)

		if err != nil {
			return nil, err
		}
		if id != last_id {
			if cur_setlist != nil {
				setlists = append(setlists, *cur_setlist)
			}
			cur_setlist = &Setlist{Id: id, Name: name, Timestamp: tstamp}
			last_id = id
		}
		if set_id_null.Valid && set_num_null.Valid && set_name_null.Valid {
			set_id := set_id_null.Int64
			set_num := set_num_null.Int64
			set_name := set_name_null.String

			as := Set{Id: set_id, SetlistId: id, SetNum: int(set_num), Name: set_name}
			// DEB: fmt.Printf("adding set %d to setlist %d\n", set_id, id)
			cur_setlist.Sets = append(cur_setlist.Sets, as)
		}
	}
	if cur_setlist != nil {
		setlists = append(setlists, *cur_setlist)
	}

	return setlists, nil

}

func (d *DB) getSet(id int64) (Set, error) {
	// DEB: fmt.Println("getSet ", id)
	q := `
SELECT a_set.name, a_set.setlist_id, a_set.setnum,
	track.id, track.title, track.tempo, track.key_tone,
	vox.id, vox.name, 
	era.id, era.name, 
	genre.id, genre.name,
	kit.id, kit.name
FROM a_set
LEFT JOIN sets_tracks on a_set.id = sets_tracks.set_id
LEFT JOIN track on track.id = sets_tracks.track_id
LEFT JOIN vox on vox.id = track.vox_id
LEFT JOIN era on era.id = track.era_id
LEFT JOIN genre on genre.id = track.genre_id
LEFT JOIN kit on kit.id = track.kit_id
WHERE a_set.id = $1
ORDER BY sets_tracks.seq ASC;
`
	rows, err := d.db.Query(q, id)
	if err != nil {
		return Set{}, err
	}
	defer rows.Close()
	var (
		name            string
		setlist_id      int64
		setnum          int64
		track_id_null   sql.NullInt64
		title_null      sql.NullString
		tempo_null      sql.NullInt64
		key_tone_null   sql.NullString
		vox_id_null     sql.NullInt64
		vox_name_null   sql.NullString
		era_id_null     sql.NullInt64
		era_name_null   sql.NullString
		genre_id_null   sql.NullInt64
		genre_name_null sql.NullString
		kit_id_null     sql.NullInt64
		kit_name_null   sql.NullString
	)
	s := Set{}
	for rows.Next() {
		err = rows.Scan(&name, &setlist_id, &setnum, &track_id_null, &title_null, &tempo_null, &key_tone_null,
			&vox_id_null, &vox_name_null, &era_id_null, &era_name_null, &genre_id_null, &genre_name_null,
			&kit_id_null, &kit_name_null)
		if err != nil {
			return Set{}, err
		}
		if s.Id == 0 {
			s.Id = id
			s.Name = name
			s.SetNum = int(setnum)
		}
		track_id := int64(0)
		// if track id is valid, check everything else. Otherwise, no need
		if track_id_null.Valid {
			track_id = track_id_null.Int64
			// DEB: fmt.Println("track id valid")

			title := ""
			if title_null.Valid {
				title = title_null.String
			}
			tempo := int64(0)
			if tempo_null.Valid {
				tempo = tempo_null.Int64
			}
			key_tone := ""
			if key_tone_null.Valid {
				key_tone = key_tone_null.String
			}
			vox_id := int64(0)
			if vox_id_null.Valid {
				vox_id = vox_id_null.Int64
			}
			vox_name := ""
			if vox_name_null.Valid {
				vox_name = vox_name_null.String
			}
			era_id := int64(0)
			if era_id_null.Valid {
				era_id = era_id_null.Int64
			}
			era_name := ""
			if era_name_null.Valid {
				era_name = era_name_null.String
			}
			genre_id := int64(0)
			if genre_id_null.Valid {
				genre_id = genre_id_null.Int64
			}
			genre_name := ""
			if genre_name_null.Valid {
				genre_name = genre_name_null.String
			}
			kit_id := int64(0)
			if kit_id_null.Valid {
				kit_id = kit_id_null.Int64
			}
			kit_name := ""
			if kit_name_null.Valid {
				kit_name = kit_name_null.String
			}

			v := Vox{Id: vox_id, Name: vox_name}
			e := Era{Id: era_id, Name: era_name}
			g := Genre{Id: genre_id, Name: genre_name}
			k := Kit{Id: kit_id, Name: kit_name}
			t := Track{Id: track_id, Title: title,
				Vox: v, Era: e, Genre: g, Tempo: int(tempo),
				KeyTone: key_tone, Kit: k}
			s.Tracks = append(s.Tracks, t)
		} else {
			// DEB: fmt.Println("no track id")
			s.Tracks = []Track{}
		}
	}
	return s, nil
}

func (d *DB) addSet(s Set) (int64, error) {
	// DEB: fmt.Println("Adding set:", s.Name)
	q := "insert into a_set (setlist_id, name, setnum) values ($1, $2, $3);"
	res, err := d.db.Exec(q, s.SetlistId, s.Name, s.SetNum)
	if err != nil {
		return -1, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return -1, err
	}
	return id, nil
}

func (d *DB) updateSet(s Set) error {
	// DEB: fmt.Println("Updating set:", s.Name)
	q := "update a_set set name=$1 where id=$2"
	_, err := d.db.Exec(q, s.Name, s.Id)
	return err
}

func (d *DB) addTrackToSet(sid int64, tid int64) error {
	// DEB: fmt.Printf("Adding track %d to set %d\n", tid, sid)
	q := "select max(seq) from sets_tracks where set_id=$1;"
	var seq_null sql.NullInt64
	err := d.db.QueryRow(q, sid).Scan(&seq_null)
	if err != nil {
		return err
	}
	seq := 0
	if seq_null.Valid {
		// DEB: fmt.Println(" seq:", seq_null.Int64)
		seq = int(seq_null.Int64)
	}

	q = "insert into sets_tracks (set_id, track_id, seq) values ($1, $2, $3)"
	_, err = d.db.Exec(q, sid, tid, seq)
	if err != nil {
		return err
	}
	return nil
}

func (d *DB) remTrackFromSet(sid int64, tid int64) error {
	// DEB: fmt.Printf("Removing track %d from set %d\n", sid, tid)
	q := "delete from sets_tracks where set_id = $1 and track_id =$2"
	_, err := d.db.Exec(q, sid, tid)
	if err != nil {
		return err
	}
	return nil
}

func (d *DB) updateSetlist(s Setlist) error {
	// DEB: fmt.Println("updating setlist:", s.Name)
	q := "update setlist set name=$1 where id=$2;"
	_, err := d.db.Exec(q, s.Name, s.Id)
	return err
}

func (d *DB) addSetlist(s Setlist) (int64, error) {
	// DEB: fmt.Println("Adding set:", s.Name)
	tstamp := time.Now().Unix()
	q := "insert into setlist (name, timestamp) values ($1, $2);"
	res, err := d.db.Exec(q, s.Name, tstamp)
	if err != nil {
		return -1, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return -1, err
	}
	return id, nil
}

func (d *DB) _insertName(table string, name string) (int64, error) {
	q := fmt.Sprintf("insert into %s (name) values ('%s');", table, name)
	res, err := d.db.Exec(q)
	if err != nil {
		return -1, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return -1, err
	}

	return id, nil
}

func (d *DB) addSong(title string, vox string, tempo int, kit string, click bool,
	era string, genre string) (int64, error) {

	// DEB: fmt.Println("Adding song:", title)

	vox_id, err := d.getVoxByName(vox)
	if err != nil {
		return -1, err
	}
	// DEB: fmt.Println("vox_id ", vox_id)

	era_id, err := d.getEraByName(era)
	if err != nil {
		return -1, err
	}
	// DEB: fmt.Println("era_id ", era_id)

	genre_id, err := d.getGenreByName(genre)
	if err != nil {
		return -1, err
	}
	// DEB: fmt.Println("genre_id ", genre_id)

	kit_id, err := d.getKitByName(kit)
	if err != nil {
		return -1, err
	}
	// DEB: fmt.Println("kit_id ", kit_id)

	q := "insert into track (title, tempo, click, vox_id, era_id, genre_id, kit_id) values ($1, $2, $3, $4, $5, $6, $7)"

	res, err := d.db.Exec(q, title, tempo, click, vox_id, era_id, genre_id, kit_id)
	if err != nil {
		return -1, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return -1, err
	}

	return id, nil
}

/*
func (d *DB) checkVoxByName(name string) (bool, error) {
	_, err := d.getVoxByName(name)
	missing := err == sql.ErrNoRows
	if !missing && err != nil {
		return true, err
	}
	return !missing, nil
}
*/
