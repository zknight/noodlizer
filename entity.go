package main

import (
	"html/template"
	"time"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type Track struct {
	Id      int64
	Title   string
	Tempo   int
	Click   bool
	KeyTone string
	Vox     Vox
	Era     Era
	Genre   Genre
	Kit     Kit
	Lyrics  Lyrics
}

func (t Track) ProperTitle() string {
	return toTitle(t.Title)
}

type Child interface {
	ProperName() string
}

type Vox struct {
	Id   int64
	Name string
}

func (v Vox) ProperName() string {
	return toTitle(v.Name)
}

type Era struct {
	Id   int64
	Name string
}

func (e Era) ProperName() string {
	return toTitle(e.Name)
}

type Genre struct {
	Id   int64
	Name string
}

func (g Genre) ProperName() string {
	return toTitle(g.Name)
}

type Kit struct {
	Id   int64
	Name string
}

func (k Kit) ProperName() string {
	return toTitle(k.Name)
}

func toTitle(s string) string {
	c := cases.Title(language.AmericanEnglish)
	return c.String(s)
}

type Lyrics struct {
	Id      int64
	RawText string
}

func (l Lyrics) PrettyText(maxRows int) template.HTML {
	m := NewMarkText(l.RawText)
	return m.PrettyText(maxRows)
}

type Set struct {
	Id        int64
	SetlistId int64
	SetNum    int
	Name      string
	Tracks    []Track
}

func (s Set) ProperName() string {
	return toTitle(s.Name)
}

func (s Set) TrackCount() int {
	return len(s.Tracks)
}

type Setlist struct {
	Id        int64
	Name      string
	Sets      []Set
	Timestamp int64
}

func (s Setlist) ProperName() string {
	return toTitle(s.Name)
}

func (s Setlist) SetCount() int {
	return len(s.Sets)
}

func (s Setlist) CreatedAt() string {
	t := time.Unix(s.Timestamp, 0)
	return t.Local().Format("2 Jan 2006 - 15:04:05")
}

type Gig struct {
	Id       int64
	Name     string // comes from setlist name
	CurSet   int    // index to current set
	CurTrack int    // index to current track
	Sets     []Set
}

func NewGig(setlist Setlist) *Gig {
	g := &Gig{Name: setlist.Name, CurSet: 0, CurTrack: 0}
	g.Sets = append(g.Sets, setlist.Sets...)
	return g
}

func (g Gig) ProperName() string {
	return toTitle(g.Name)
}

// return next Track or false if no more
/*
func (g *Gig) NextTrack() (Track, bool) {
	if len(g.Sets[g.CurSet].Tracks) < 1 {
		// go to next set, if exists
		g.CurSet += 1
		if g.CurSet >= len(g.Sets) {
			// no more tracks
			return Track{}, false
		}
	}
	t := g.Sets[g.CurSet].Tracks[0]
	g.Sets[g.CurSet].Tracks = g.Sets[g.CurSet].Tracks[1:]
	return t, true
}
*/
