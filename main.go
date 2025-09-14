package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/zknight/noodlizer/db"
)

func main() {
	argc := len(os.Args)

	if argc < 2 {
		os.Args = append(os.Args, "serve")
	}

	switch os.Args[1] {
	case "import":
		if argc < 4 {
			fmt.Printf("Dude. If you want to import, you need <infile> and <outfile>.")
			os.Exit(2)
		}
		inpath := os.Args[2]
		outpath := os.Args[3]
		flimport(inpath, outpath)
	case "serve":
		// no params for now
		serve()
	default:
		fmt.Printf("Sorry. Dunno what you mean, \"%s\"....?\n", os.Args[1])
		os.Exit(2)
	}

}

func serve() {
	//http.HandleFunc("/{$}", Index)
	// open DB
	tdb, err := db.OpenDB("tracks.db")
	if err != nil {
		fmt.Println("Error opening track database: ", err.Error())
		os.Exit(1)
	}
	_ = NewView(tdb)
	svr := http.Server{Addr: ":80"}
	done := make(chan struct{})
	go func() {
		sigc := make(chan os.Signal, 100)
		signal.Notify(sigc, os.Interrupt, syscall.SIGTERM)
		<-sigc

		if err := svr.Shutdown(context.Background()); err != nil {
			fmt.Println("Error shutting shit down: ", err.Error())
		}
		close(done)
	}()

	if err := svr.ListenAndServe(); err != http.ErrServerClosed {
		fmt.Println("Error ListenAndServe: ", err.Error())
	}
	// clean up the gigs...
	fmt.Printf("Cleaning up...")
	err = db.cleanupGigs()
	if err != nil {
		fmt.Println("Error cleaning up: ", err.Error())
	}
}

func flimport(infile, dbfile string) {
	fmt.Printf("hello. playing around with: %s\n", infile)

	csvf, err := os.Open(infile)
	if err != nil {
		fmt.Println("Can't. Just can't. ", err.Error())
		os.Exit(1)
	}
	reader := csv.NewReader(csvf)
	contents, err := reader.ReadAll()
	if err != nil {
		fmt.Println("I Tried...", err.Error())
		os.Exit(1)
	}
	/* don't need these if we just use indecies (sing it)
	keys := []string {
		"title",
		"kit",
		"tempo",
		"vox",
		"click",
		"era",
		"genre",
	}
	*/
	era_txt := map[string]string{
		"40": "forties",
		"50": "fifties",
		"60": "sixties",
		"70": "seventies",
		"80": "eighties",
		"90": "nineties",
		"0":  "oughts",
		"10": "twenty-tens",
		"20": "modern",
	}

	type set map[string]struct{}
	vox_set := set(make(set))
	era_set := set(make(set))
	genre_set := set(make(set))
	kit_set := set(make(set))

	songs := []db.Track{}

	for _, row := range contents {
		track := db.Track{}
		for i, v := range row {
			//fmt.Printf("%s:%s ", keys[i], v)
			//switch keys[i] {
			switch i {
			//case "title":
			case 0:
				track.Title = strings.ToLower(v)
			//case "kit":
			case 1:
				v = strings.ToLower(v)
				kit_set[v] = struct{}{}
				track.Kit.Name = v
			//case "tempo":
			case 2:
				track.Tempo, _ = strconv.Atoi(v)
			//case "vox":
			case 3:
				v = strings.ToLower(v)
				vox_set[v] = struct{}{}
				track.Vox.Name = v
			//case "click":
			case 4:
				track.Click = v == "True"
			//case "era":
			case 5:
				v = era_txt[v]
				era_set[v] = struct{}{}
				track.Era.Name = v
			//case "genre":
			case 6:
				v = strings.ToLower(v)
				genre_set[v] = struct{}{}
				track.Genre.Name = v
			}
		}
		songs = append(songs, track)
		fmt.Println()
	}

	var tdb *db.DB

	// check for exist
	_, err = os.Stat(dbfile)
	if os.IsNotExist(err) {
		fmt.Println("Creating new...")
		tdb, err = db.NewDB(dbfile)
	} else {
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
		fmt.Println("Opening...")
		tdb, err = db.OpenDB(dbfile)
	}
	if err != nil {
		fmt.Println("failed to open database.", err.Error())
		os.Exit(1)
	}

	// iterate setssss
	for k := range vox_set {
		tdb.AddVox(k)
	}

	for k := range era_set {
		tdb.AddEra(k)
	}

	for k := range genre_set {
		tdb.AddGenre(k)
	}

	for k := range kit_set {
		tdb.AddKit(k)
	}

	// loopty thru the songs
	for _, s := range songs {
		songid, err := db.AddSong(s.Title, s.Vox.Name, s.Tempo, s.Kit.Name, s.Click, s.Era.Name, s.Genre.Name)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
		fmt.Println("Song ID :", songid)
	}

}
