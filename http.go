package main

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/zknight/noodlizer/db"
)

type View struct {
	db      *db.DB
	index   *template.Template
	ws_mtx  sync.Mutex
	subs    map[*subscriber]struct{}
	waiters map[string]struct{}
}

type trackInfo struct {
	Track  db.Track
	Voxes  []db.Vox
	Eras   []db.Era
	Genres []db.Genre
	Kits   []db.Kit
}

func NewView(db *db.DB) *View {
	v := &View{
		db:      db,
		subs:    make(map[*subscriber]struct{}),
		waiters: make(map[string]struct{}),
	}
	// static path
	fs := http.FileServer(http.Dir("./static/"))

	// handler funcs
	http.HandleFunc("/{$}", v.Index)
	//http.HandleFunc("/{obj}/{id}", v.ShowObj)
	http.HandleFunc("/subscribe", v.Subscribe)
	// alias wait and ready to same handler
	http.HandleFunc("POST /wait", v.UpdateGigPause)
	http.HandleFunc("POST /ready", v.UpdateGigPause)
	http.HandleFunc("/setlists", v.ShowSetlists)
	http.HandleFunc("/tracks", v.ShowAllTracks)
	http.HandleFunc("/voxes", v.ShowVoxes)
	http.HandleFunc("/eras", v.ShowEras)
	http.HandleFunc("/genres", v.ShowGenres)
	http.HandleFunc("/kits", v.ShowKits)
	http.HandleFunc("/set/{id}", v.ShowSet)
	http.HandleFunc("/set/{id}/edit", v.EditSet)
	http.HandleFunc("/set/{sid}/add_track/{tid}", v.AddTrackToSet)
	http.HandleFunc("/set/{sid}/del_track/{tid}", v.DelTrackFromSet)
	http.HandleFunc("/setlist/{id}", v.ShowSetlist)
	http.HandleFunc("/setlist/create", v.CreateSetlist)
	http.HandleFunc("/setlist/{id}/edit", v.EditSetlist)
	http.HandleFunc("POST /setlist/save", v.SaveSetlist)
	http.HandleFunc("POST /setlist/{id}/update", v.UpdateSetlist)
	http.HandleFunc("/setlist/{id}/create_set/{setnum}", v.CreateSet)
	http.HandleFunc("POST /setlist/{id}/save_set", v.SaveSet)
	http.HandleFunc("POST /setlist/{id}/update_set", v.UpdateSet)
	http.HandleFunc("/track/{id}", v.ShowTrack)
	http.HandleFunc("/track/{id}/edit", v.EditTrack)
	http.HandleFunc("POST /track/{id}/update", v.UpdateTrack)
	http.HandleFunc("POST /track/{id}/update_lyrics", v.UpdateLyrics)
	http.HandleFunc("/vox/{id}", v.ShowVox)
	http.HandleFunc("/era/{id}", v.ShowEra)
	http.HandleFunc("/genre/{id}", v.ShowGenre)
	http.HandleFunc("/kit/{id}", v.ShowKit)
	http.HandleFunc("/gig/", v.StartGig)
	http.HandleFunc("/gig/next/{id}", v.ShowGigNext)
	http.HandleFunc("/gig/prev/{id}", v.ShowGigPrev)
	http.HandleFunc("/gig/end/{id}", v.EndGig)
	http.HandleFunc("/gig/setlist/{id}", v.DoGig)

	http.Handle("/static/", http.StripPrefix("/static", fs))
	fmap := template.FuncMap{
		"inc": func(i int) int {
			return i + 1
		}}
	v.index = template.Must(template.New("main").Funcs(fmap).ParseGlob("./template/*.tmpl"))

	go v.servicePause()
	return v
}

// Handler functions for web service

// index
func (v *View) Index(w http.ResponseWriter, r *http.Request) {
	//io.WriteString(w, "index.")
	//http.Redirect(w, r, "/tracks", http.StatusFound)
	err := v.index.ExecuteTemplate(w, "index.tmpl", nil)
	if err != nil {
		io.WriteString(w, err.Error())
	}
}

func (v *View) ShowSet(w http.ResponseWriter, r *http.Request) {
	id, serr := strconv.Atoi(r.PathValue("id"))
	if serr != nil {
		io.WriteString(w, serr.Error())
		return
	}
	fmt.Println("show Set", id)
	set, err := v.db.GetSet(int64(id))
	if err != nil {
		io.WriteString(w, err.Error())
		return
	}
	err = v.index.ExecuteTemplate(w, "set.tmpl", set)
	if err != nil {
		io.WriteString(w, err.Error())
	}
}

func (v *View) EditSet(w http.ResponseWriter, r *http.Request) {
	id, serr := strconv.Atoi(r.PathValue("id"))
	if serr != nil {
		io.WriteString(w, serr.Error())
		return
	}
	fmt.Println("edit Set", id)
	set, err := v.db.GetSet(int64(id))
	if err != nil {
		io.WriteString(w, err.Error())
		return
	}
	tracks, err := v.db.GetAllTracks()
	if err != nil {
		io.WriteString(w, err.Error())
		return
	}

	err = v.index.ExecuteTemplate(w, "edit_set.tmpl", struct {
		Set    db.Set
		Tracks []db.Track
		Action string
	}{Set: set, Tracks: tracks, Action: "update"})
	if err != nil {
		io.WriteString(w, err.Error())
	}
}

func (v *View) CreateSet(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		io.WriteString(w, err.Error())
		return
	}
	setnum, err := strconv.Atoi(r.PathValue("setnum"))
	if err != nil {
		io.WriteString(w, err.Error())
		return
	}
	s := db.Set{SetlistId: int64(id), SetNum: setnum}
	t := []db.Track{}
	err = v.index.ExecuteTemplate(w, "edit_set.tmpl", struct {
		Set    db.Set
		Tracks []db.Track
		Action string
	}{Set: s, Tracks: t, Action: "save"})
	if err != nil {
		io.WriteString(w, err.Error())
	}
}

func (v *View) SaveSet(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		io.WriteString(w, err.Error())
		return
	}
	fmt.Println("Save (new) Set for setlist ", id)
	err = r.ParseForm()
	if err != nil {
		io.WriteString(w, err.Error())
		return
	}
	name := r.PostFormValue("Name")
	setnum, _ := strconv.Atoi(r.PostFormValue("Setnum"))
	s := db.Set{SetlistId: int64(id), SetNum: setnum, Name: name}
	_, err = v.db.AddSet(s)
	if err != nil {
		io.WriteString(w, err.Error())
		return
	}

	url := fmt.Sprintf("/setlist/%d", id)
	http.Redirect(w, r, url, http.StatusFound)

}

func (v *View) UpdateSet(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		io.WriteString(w, err.Error())
		return
	}
	fmt.Println("Save (update) Set for setlist ", id)
	err = r.ParseForm()
	if err != nil {
		io.WriteString(w, err.Error())
		return
	}
	name := r.PostFormValue("Name")
	setnum, _ := strconv.Atoi(r.PostFormValue("Setnum"))
	setid, _ := strconv.Atoi(r.PostFormValue("SetId"))
	s := db.Set{Id: int64(setid), SetlistId: int64(id), SetNum: setnum, Name: name}
	err = v.db.UpdateSet(s)
	if err != nil {
		io.WriteString(w, err.Error())
		return
	}
	url := fmt.Sprintf("/setlist/%d", id)
	http.Redirect(w, r, url, http.StatusFound)

}

func (v *View) AddTrackToSet(w http.ResponseWriter, r *http.Request) {
	sid, err := strconv.Atoi(r.PathValue("sid"))
	if err != nil {
		io.WriteString(w, err.Error())
		return
	}
	tid, err := strconv.Atoi(r.PathValue("tid"))
	if err != nil {
		io.WriteString(w, err.Error())
		return
	}
	err = v.db.AddTrackToSet(int64(sid), int64(tid))
	if err != nil {
		io.WriteString(w, err.Error())
		return
	}
	url := fmt.Sprintf("/set/%d/edit", sid)
	http.Redirect(w, r, url, http.StatusFound)
}

func (v *View) DelTrackFromSet(w http.ResponseWriter, r *http.Request) {
	sid, err := strconv.Atoi(r.PathValue("sid"))
	if err != nil {
		io.WriteString(w, err.Error())
		return
	}
	tid, err := strconv.Atoi(r.PathValue("tid"))
	if err != nil {
		io.WriteString(w, err.Error())
		return
	}
	err = v.db.RemTrackFromSet(int64(sid), int64(tid))
	if err != nil {
		io.WriteString(w, err.Error())
		return
	}
	url := fmt.Sprintf("/set/%d/edit", sid)
	http.Redirect(w, r, url, http.StatusFound)
}

func (v *View) ShowSetlists(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Show Setlists.")
	setlists, err := v.db.GetAllSetlists()
	if err != nil {
		io.WriteString(w, err.Error())
		return
	}
	err = v.index.ExecuteTemplate(w, "setlists.tmpl", setlists)
	if err != nil {
		io.WriteString(w, err.Error())
	}
}

func (v *View) ShowSetlist(w http.ResponseWriter, r *http.Request) {
	id, serr := strconv.Atoi(r.PathValue("id"))
	if serr != nil {
		io.WriteString(w, serr.Error())
		return
	}
	fmt.Println("Show Setlist ", id)
	setlist, err := v.db.GetSetlist(int64(id))
	if err != nil {
		io.WriteString(w, err.Error())
		return
	}
	err = v.index.ExecuteTemplate(w, "setlist.tmpl", setlist)
	if err != nil {
		io.WriteString(w, err.Error())
	}
}

func (v *View) CreateSetlist(w http.ResponseWriter, r *http.Request) {
	setlist := db.Setlist{}
	err := v.index.ExecuteTemplate(w, "new_setlist.tmpl", setlist)
	if err != nil {
		io.WriteString(w, err.Error())
	}
}

func (v *View) EditSetlist(w http.ResponseWriter, r *http.Request) {
	id, serr := strconv.Atoi(r.PathValue("id"))
	if serr != nil {
		io.WriteString(w, serr.Error())
		return
	}
	fmt.Println("Edit Setlist")
	setlist, err := v.db.GetSetlist(int64(id))
	if err != nil {
		io.WriteString(w, err.Error())
		return
	}
	err = v.index.ExecuteTemplate(w, "edit_setlist.tmpl", setlist)
	if err != nil {
		io.WriteString(w, err.Error())
	}
}

func (v *View) SaveSetlist(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Saving (new) Setlist")
	err := r.ParseForm()
	if err != nil {
		io.WriteString(w, err.Error())
		return
	}
	name := r.PostFormValue("Name")
	s := db.Setlist{Name: name}
	id, err := v.db.AddSetlist(s)
	fmt.Printf("new setlist id:%d\n", id)
	if err != nil {
		io.WriteString(w, err.Error())
		return
	}
	url := fmt.Sprintf("/setlist/%d", id)
	http.Redirect(w, r, url, http.StatusFound)
}

func (v *View) UpdateSetlist(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		io.WriteString(w, err.Error())
		return
	}
	fmt.Println("Updating setlist ", id)
	err = r.ParseForm()
	if err != nil {
		io.WriteString(w, err.Error())
		return
	}
	name := r.PostFormValue("Name")
	s := db.Setlist{Id: int64(id), Name: name}
	err = v.db.UpdateSetlist(s)
	if err != nil {
		io.WriteString(w, err.Error())
		return
	}
	url := fmt.Sprintf("/setlist/%d", id)
	http.Redirect(w, r, url, http.StatusFound)
}

func (v *View) ShowAllTracks(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Show Tracks.")
	tracks, err := v.db.GetAllTracks()
	fmt.Println("track cnt: ", len(tracks))
	if err != nil {
		io.WriteString(w, err.Error())
		return
	}
	err = v.index.ExecuteTemplate(w, "tracks.tmpl", struct{ Tracks []db.Track }{Tracks: tracks})
	if err != nil {
		io.WriteString(w, err.Error())
	}
}

func (v *View) ShowTrack(w http.ResponseWriter, r *http.Request) {
	id, serr := strconv.Atoi(r.PathValue("id"))
	if serr != nil {
		io.WriteString(w, serr.Error())
		return
	}
	fmt.Println("Show Track ", id)
	t, err := v.db.GetTrack(int64(id))
	if err != nil {
		io.WriteString(w, err.Error())
		return
	}
	err = v.index.ExecuteTemplate(w, "track.tmpl", t)
	if err != nil {
		io.WriteString(w, err.Error())
	}
}

func (v *View) EditTrack(w http.ResponseWriter, r *http.Request) {
	id, serr := strconv.Atoi(r.PathValue("id"))
	if serr != nil {
		io.WriteString(w, serr.Error())
		return
	}
	fmt.Println("Edit Track ", id)
	t, err := v.db.GetTrack(int64(id))
	if err != nil {
		io.WriteString(w, err.Error())
		return
	}
	voxes, err := v.db.GetAllVoxes()
	if err != nil {
		io.WriteString(w, err.Error())
		return
	}
	eras, err := v.db.GetAllEras()
	if err != nil {
		io.WriteString(w, err.Error())
		return
	}
	genres, err := v.db.GetAllGenres()
	if err != nil {
		io.WriteString(w, err.Error())
		return
	}
	kits, err := v.db.GetAllKits()
	if err != nil {
		io.WriteString(w, err.Error())
		return
	}
	err = v.index.ExecuteTemplate(w, "edit_track.tmpl",
		trackInfo{Track: t, Voxes: voxes, Eras: eras, Genres: genres, Kits: kits})

	if err != nil {
		io.WriteString(w, err.Error())
	}
}

func (v *View) UpdateTrack(w http.ResponseWriter, r *http.Request) {
	id, serr := strconv.Atoi(r.PathValue("id"))
	if serr != nil {
		io.WriteString(w, serr.Error())
		return
	}
	fmt.Println("Update Track ", id)
	err := r.ParseForm()
	if err != nil {
		io.WriteString(w, err.Error())
		return
	}
	//io.WriteString(w, fmt.Sprintf("%v", r.PostForm))

	title := strings.ToLower(r.PostForm["Title"][0])
	click := r.PostForm["Click"][0] == "on"
	tempo, _ := strconv.Atoi(r.PostForm["Tempo"][0])
	vox_id, _ := strconv.Atoi(r.PostForm["Vox"][0])
	era_id, _ := strconv.Atoi(r.PostForm["Era"][0])
	genre_id, _ := strconv.Atoi(r.PostForm["Genre"][0])
	kit_id, _ := strconv.Atoi(r.PostForm["Kit"][0])
	key_tone := r.PostForm["KeyTone"][0]

	t := db.Track{
		Id:      int64(id),
		Title:   title,
		Tempo:   tempo,
		Click:   click,
		KeyTone: key_tone,
		Vox:     db.Vox{Id: int64(vox_id)},
		Era:     db.Era{Id: int64(era_id)},
		Genre:   db.Genre{Id: int64(genre_id)},
		Kit:     db.Kit{Id: int64(kit_id)},
	}
	err = v.db.UpdateTrack(t)
	if err != nil {
		io.WriteString(w, err.Error())
	}
	uri := fmt.Sprintf("/track/%d", id)
	http.Redirect(w, r, uri, http.StatusFound)

}

func (v *View) UpdateLyrics(w http.ResponseWriter, r *http.Request) {
	i, serr := strconv.Atoi(r.PathValue("id"))
	if serr != nil {
		io.WriteString(w, serr.Error())
		return
	}
	id := int64(i)
	fmt.Println("Update Lyrics ", id)
	err := r.ParseForm()
	if err != nil {
		io.WriteString(w, err.Error())
		return
	}
	lyrics_id, err := strconv.Atoi(r.PostFormValue("lyrics_id"))
	if err != nil {
		io.WriteString(w, err.Error())
		return
	}
	fmt.Println("lyrics_id from form:", lyrics_id)
	lyrics := db.bLyrics{Id: int64(lyrics_id), RawText: r.PostFormValue("lyrics")}
	if lyrics.Id == 0 {
		_, err = v.db.AddLyrics(id, lyrics)
		if err != nil {
			io.WriteString(w, err.Error())
			return
		}
	} else {
		err = v.db.UpdateLyrics(lyrics)
		if err != nil {
			io.WriteString(w, err.Error())
		}
	}
	//fmt.Println(r.PostForm["lyrics"][0])
	//m := NewMarkText(r.PostForm["lyrics"][0])

	//	io.WriteString(w, "<html><body>"+m.PrettyText()+"</body></html>")
	uri := fmt.Sprintf("/track/%d", id)
	http.Redirect(w, r, uri, http.StatusFound)
}

func (v *View) ShowVoxes(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Show Voxes.")
	voxes, err := v.db.GetAllVoxes()
	if err != nil {
		io.WriteString(w, err.Error())
		return
	}
	err = v.index.ExecuteTemplate(w, "voxes.tmpl", struct{ Voxes []db.Vox }{Voxes: voxes})
	if err != nil {
		io.WriteString(w, err.Error())
	}
}

func (v *View) ShowVox(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		io.WriteString(w, err.Error())
		return
	}
	fmt.Println("Show Vox ", id)
	vox, err := v.db.GetVox(int64(id))
	if err != nil {
		io.WriteString(w, err.Error())
		return
	}

	tracks, err := v.db.GetTracksByVox(int64(id))
	if err != nil {
		io.WriteString(w, err.Error())
		return
	}
	x := struct {
		Kind   string
		Id     int
		Obj    db.Child
		Tracks []db.Track
	}{Kind: "Vox", Id: id, Obj: vox, Tracks: tracks}
	err = v.index.ExecuteTemplate(w, "child.tmpl", x)
	if err != nil {
		io.WriteString(w, err.Error())
	}
}

func (v *View) ShowEras(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Show Eras.")
	eras, err := v.db.GetAllEras()
	if err != nil {
		io.WriteString(w, err.Error())
		return
	}
	err = v.index.ExecuteTemplate(w, "eras.tmpl", struct{ Eras []db.Era }{Eras: eras})
	if err != nil {
		io.WriteString(w, err.Error())
	}
}

func (v *View) ShowEra(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		io.WriteString(w, err.Error())
		return
	}
	fmt.Println("Show Era ", id)
	era, err := v.db.GetEra(int64(id))
	if err != nil {
		io.WriteString(w, err.Error())
		return
	}

	tracks, err := v.db.GetTracksByEra(int64(id))
	if err != nil {
		io.WriteString(w, err.Error())
		return
	}
	x := struct {
		Kind   string
		Id     int
		Obj    db.Child
		Tracks []db.Track
	}{Kind: "Era", Id: id, Obj: era, Tracks: tracks}
	err = v.index.ExecuteTemplate(w, "child.tmpl", x)
	if err != nil {
		io.WriteString(w, err.Error())
	}
}

func (v *View) ShowGenres(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Show Genres.")
	genres, err := v.db.GetAllGenres()
	if err != nil {
		io.WriteString(w, err.Error())
		return
	}
	err = v.index.ExecuteTemplate(w, "genres.tmpl", struct{ Genres []db.Genre }{Genres: genres})
	if err != nil {
		io.WriteString(w, err.Error())
	}
}

func (v *View) ShowGenre(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		io.WriteString(w, err.Error())
		return
	}
	fmt.Println("Show Genre ", id)
	genre, err := v.db.GetGenre(int64(id))
	if err != nil {
		io.WriteString(w, err.Error())
		return
	}

	tracks, err := v.db.GetTracksByGenre(int64(id))
	if err != nil {
		io.WriteString(w, err.Error())
		return
	}
	x := struct {
		Kind   string
		Id     int
		Obj    db.Child
		Tracks []db.Track
	}{Kind: "Genre", Id: id, Obj: genre, Tracks: tracks}
	err = v.index.ExecuteTemplate(w, "child.tmpl", x)
	if err != nil {
		io.WriteString(w, err.Error())
	}
}

func (v *View) ShowKits(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Show Kits.")
	kits, err := v.db.GetAllKits()
	if err != nil {
		io.WriteString(w, err.Error())
		return
	}
	err = v.index.ExecuteTemplate(w, "kits.tmpl", struct{ Kits []db.Kit }{Kits: kits})
	if err != nil {
		io.WriteString(w, err.Error())
	}
}

func (v *View) ShowKit(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		io.WriteString(w, err.Error())
		return
	}
	fmt.Println("Show Kit ", id)
	kit, err := v.db.GetKit(int64(id))
	if err != nil {
		io.WriteString(w, err.Error())
		return
	}

	tracks, err := v.db.GetTracksByKit(int64(id))
	if err != nil {
		io.WriteString(w, err.Error())
		return
	}
	x := struct {
		Kind   string
		Id     int
		Obj    db.Child
		Tracks []db.Track
	}{Kind: "Kit", Id: id, Obj: kit, Tracks: tracks}
	err = v.index.ExecuteTemplate(w, "child.tmpl", x)
	if err != nil {
		io.WriteString(w, err.Error())
	}
}

func (v *View) StartGig(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Start a gig")
	setlists, err := v.db.GetAllSetlists()
	if err != nil {
		io.WriteString(w, err.Error())
		return
	}
	err = v.index.ExecuteTemplate(w, "launch_gig.tmpl", setlists)
	if err != nil {
		io.WriteString(w, err.Error())
	}
}

func (v *View) DoGig(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		io.WriteString(w, fmt.Sprintf("DoGig.1: %s", err.Error()))
		return
	}
	fmt.Println("Doin' gig", id)
	// Load a gig - contains a list of tracks for each set? or point to a set and keep
	// track of current song... temporary set?
	// no form required. Just pass a parameter (hash) that references the information
	// "pickle" a struct?
	sl, err := v.db.GetSetlist(int64(id))
	if err != nil {
		io.WriteString(w, fmt.Sprintf("DoGig.2: %s", err.Error()))
		return
	}
	gig, err := v.db.newGig(sl)
	if err != nil {
		io.WriteString(w, fmt.Sprintf("DoGig.3: %s", err.Error()))
		return
	}
	//m := fmt.Sprintf("should be new gig: id=%d name=%s nsets=%d curset=%d", gig.Id, gig.ProperName(), len(gig.Sets), gig.CurSet)
	//io.WriteString(w, m)
	data := struct {
		Id      int
		Title   string
		Name    string
		SetName string
		Lyrics  template.HTML
		Tempo   int
		KeyTone string
	}{}
	data.Id = int(gig.Id)
	data.Name = gig.ProperName()
	data.SetName = gig.Sets[gig.CurSet].ProperName()
	tid := gig.Sets[gig.CurSet].Tracks[gig.CurTrack].Id
	track, err := v.db.GetTrack(tid)
	if err != nil {
		io.WriteString(w, fmt.Sprintf("DoGig.4: %s", err.Error()))
		return
	}
	data.Lyrics = track.Lyrics.PrettyText(0)
	data.Title = gig.Sets[gig.CurSet].Tracks[gig.CurTrack].ProperTitle()
	data.Tempo = gig.Sets[gig.CurSet].Tracks[gig.CurTrack].Tempo
	data.KeyTone = gig.Sets[gig.CurSet].Tracks[gig.CurSet].KeyTone
	err = v.index.ExecuteTemplate(w, "gig.tmpl", data)
	if err != nil {
		io.WriteString(w, fmt.Sprintf("DoGig.5: %s", err.Error()))
		return
	}
}

func (v *View) ShowGigNext(w http.ResponseWriter, r *http.Request) {
	//fmt.Println("Show Next Song (ShowGigNext)")
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		io.WriteString(w, fmt.Sprintf("ShowGigNext.1: %s", err.Error()))
		return
	}
	g, err := v.db.getGig(int64(id))
	if err != nil {
		io.WriteString(w, fmt.Sprintf("ShowGigNext.2: %s", err.Error()))
		return
	}
	//fmt.Println("len(g.Sets)=", len(g.Sets))
	//fmt.Println("g.CurTrack=", g.CurTrack)
	//fmt.Println("g.CurSet=", g.CurSet)
	//fmt.Println("TrackCount=", g.Sets[g.CurSet].TrackCount())
	//fmt.Println("len(Tracks)=", len(g.Sets[g.CurSet].Tracks))
	g.CurTrack += 1
	if g.CurTrack >= g.Sets[g.CurSet].TrackCount() {
		//fmt.Println("  that was the last track of set: ", g.CurSet)
		data := struct {
			Id      int64
			Name    string
			SetName string
			Status  string
		}{
			Id:      g.Id,
			Name:    g.ProperName(),
			SetName: g.Sets[g.CurSet].ProperName(),
			Status:  "End",
		}

		g.CurTrack = 0
		g.CurSet += 1
		if g.CurSet >= len(g.Sets) {
			// done. Need to go back to setlist page?
			//fmt.Println("Past the end, should show setlist_end.tmpl")
			data := struct {
				Id     int64
				Name   string
				Status string
			}{
				Id: g.Id,
				//Name: g.ProperName(),
				Name:   g.ProperName(),
				Status: "End",
			}
			//fmt.Println("propername: ", g.ProperName())
			// TODO: instead of forpin with the math, play a trick instead.
			// pass the id for the NEXT set and track. It will look like the
			// natural number index instead of 0-base
			err = v.index.ExecuteTemplate(w, "setlist_end.tmpl", data)
			if err != nil {
				io.WriteString(w, fmt.Sprintf("ShowGigNext.7: %s", err.Error()))
			}
			return
		} else {
			g.CurTrack = -1 // next time through here it will get set incremented to the first track
			err = v.db.updateGig(g)
			if err != nil {
				io.WriteString(w, fmt.Sprintf("ShowGigNext.8: %s", err.Error()))
			}
			err = v.index.ExecuteTemplate(w, "set_end.tmpl", data)
			if err != nil {
				io.WriteString(w, fmt.Sprintf("ShowGigNext.6: %s", err.Error()))
			}
			return
		}
	}
	//fmt.Println("Next Track in set: ", g.CurTrack)
	//fmt.Println("Next Set in Gig: ", g.CurSet)
	err = v.db.updateGig(g)
	if err != nil {
		io.WriteString(w, fmt.Sprintf("ShowGigNext.3: %s", err.Error()))
		return
	}

	// simplify template by only pulling what is needed
	data := struct {
		Id      int
		Title   string
		Name    string
		SetName string
		Lyrics  template.HTML
		Tempo   int
		KeyTone string
	}{}
	data.Id = int(g.Id)
	data.Name = g.ProperName()
	data.SetName = g.Sets[g.CurSet].ProperName()
	tid := g.Sets[g.CurSet].Tracks[g.CurTrack].Id
	//fmt.Println("TRACK ID=", tid)
	track, err := v.db.GetTrack(tid)
	if err != nil {
		io.WriteString(w, fmt.Sprintf("ShowGigNext.4: %s", err.Error()))
		return
	}
	data.Lyrics = track.Lyrics.PrettyText(0)
	data.Title = g.Sets[g.CurSet].Tracks[g.CurTrack].ProperTitle()
	data.Tempo = g.Sets[g.CurSet].Tracks[g.CurTrack].Tempo
	data.KeyTone = g.Sets[g.CurSet].Tracks[g.CurTrack].KeyTone
	err = v.index.ExecuteTemplate(w, "gig.tmpl", data)
	if err != nil {
		io.WriteString(w, fmt.Sprintf("ShowGigNext.5: %s", err.Error()))
		return
	}
}

func (v *View) ShowGigPrev(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Show Prev Song (ShowGigPrev)")
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		io.WriteString(w, fmt.Sprintf("ShowGigPrev.1: %s", err.Error()))
		return
	}
	g, err := v.db.getGig(int64(id))
	if err != nil {
		io.WriteString(w, fmt.Sprintf("ShowGigPrev.2: %s", err.Error()))
		return
	}
	fmt.Println("len(g.Sets)=", len(g.Sets))
	fmt.Println("g.CurTrack=", g.CurTrack)
	fmt.Println("g.CurSet=", g.CurSet)
	fmt.Println("TrackCount=", g.Sets[g.CurSet].TrackCount())
	fmt.Println("len(Tracks)=", len(g.Sets[g.CurSet].Tracks))
	g.CurTrack -= 1
	if g.CurTrack < 0 {
		fmt.Println("  that was the first track of set: ", g.CurSet)
		data := struct {
			Id      int64
			Name    string
			SetName string
			Status  string
		}{
			Id:      g.Id,
			Name:    g.ProperName(),
			SetName: g.Sets[g.CurSet].ProperName(),
			Status:  "Beginning",
		}
		g.CurSet -= 1
		if g.CurSet < 0 {
			g.CurSet = 0
			// done. stay on current?
			data := struct {
				Id     int64
				Name   string
				Status string
			}{
				Id:     g.Id,
				Name:   g.ProperName(),
				Status: "Beginning",
			}
			// TODO change to setlist_beg.tmpl or write the string via templ
			err = v.index.ExecuteTemplate(w, "setlist_end.tmpl", data)
			if err != nil {
				io.WriteString(w, fmt.Sprintf("ShowGigNext.6: %s", err.Error()))
			}
			return
		} else {
			// should get subtracked next time through
			g.CurTrack = len(g.Sets[g.CurSet].Tracks)
			err = v.db.updateGig(g)
			if err != nil {
				io.WriteString(w, fmt.Sprintf("ShowGigNext.5: %s", err.Error()))
			}
			err = v.index.ExecuteTemplate(w, "set_end.tmpl", data)
			if err != nil {
				io.WriteString(w, fmt.Sprintf("ShowGigNext.7: %s", err.Error()))
			}
			return
		}
	}
	fmt.Println("Next Track in set: ", g.CurTrack)
	fmt.Println("Next Set in Gig: ", g.CurSet)
	err = v.db.updateGig(g)
	if err != nil {
		io.WriteString(w, fmt.Sprintf("ShowGigNext.3: %s", err.Error()))
		return
	}

	// simplify template by only pulling what is needed
	data := struct {
		Id      int
		Title   string
		Name    string
		SetName string
		Lyrics  template.HTML
		Tempo   int
		KeyTone string
	}{}
	data.Id = int(g.Id)
	data.Name = g.ProperName()
	data.SetName = g.Sets[g.CurSet].ProperName()
	tid := g.Sets[g.CurSet].Tracks[g.CurTrack].Id
	fmt.Println("TRACK ID=", tid)
	track, err := v.db.GetTrack(tid)
	if err != nil {
		io.WriteString(w, fmt.Sprintf("ShowGigNext.4: %s", err.Error()))
		return
	}
	data.Lyrics = track.Lyrics.PrettyText(0)
	data.Title = g.Sets[g.CurSet].Tracks[g.CurTrack].ProperTitle()
	data.Tempo = g.Sets[g.CurSet].Tracks[g.CurTrack].Tempo
	data.KeyTone = g.Sets[g.CurSet].Tracks[g.CurTrack].KeyTone
	err = v.index.ExecuteTemplate(w, "gig.tmpl", data)
	if err != nil {
		io.WriteString(w, fmt.Sprintf("ShowGigNext.5: %s", err.Error()))
		return
	}
}

func (v *View) EndGig(w http.ResponseWriter, r *http.Request) {
	// delete the current gig by id and redirect to the main page
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		io.WriteString(w, fmt.Sprintf("EndGig.1: %s", err.Error()))
		return
	}
	v.db.removeGig(int64(id))
	url := "/"
	http.Redirect(w, r, url, http.StatusFound)
}

/*
func (v *View) ShowObj(w http.ResponseWriter, r *http.Request) {
	s := r.PathValue("obj") + " id: " + r.PathValue("id")
	io.WriteString(w, s)
}
*/

// routing:
// /tracks - list all tracks
// /track/[id]
// /track/edit/[id]
// /track/new/[id]
// /track/del/[id]
// /genres - list all genres
// /genre/new/[id]
// /genre/edit/[id]
// /genre/del/[id]
// /eras - list all eras
// /lyrics
