package http

import (
	"context"
	"html/template"
	"io"
	"net/http"
	"noodlizer/db"
	"noodlizer/logging"
	"path/filepath"
)

func protect(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie("ndlzr")
		if err == nil {
			ctx := context.WithValue(r.Context(), "id", c.Value)
			next.ServeHTTP(w, r.WithContext(ctx))
		} else {
			http.Redirect(w, r, "/login", http.StatusFound)
			//http.Error(w, "Not logged in, er, uh...", http.StatusForbidden)
		}
	})
}

type View struct {
	db    *db.DB
	index *template.Template
	log   *logging.Logger
	mux   *http.ServeMux
}

func NewView(db *db.DB, log *logging.Logger) *View {
	v := &View{
		db:  db,
		log: log,
		mux: http.NewServeMux(),
	}

	return v
}

func (v *View) Handle() http.Handler {
	var handler http.Handler = v.mux
	//handler = protect(handler)
	return handler
}

func (v *View) AddRoutes(tpath string) {
	// static path
	v.log.Info("template path: ", tpath)
	sp := filepath.Join(tpath, "static/")
	v.log.Info("Adding static route: ", sp)
	fs := http.FileServer(http.Dir(sp))
	tp := filepath.Join(tpath, "template/*.tmpl")
	v.mux.Handle("/static/", http.StripPrefix("/static", fs))
	//http.HandleFunc("/{$}", v.Index)
	v.mux.Handle("/{$}", protect(v.HandleIndex()))
	v.mux.HandleFunc("GET /login", v.Login)
	v.mux.HandleFunc("GET /newacct", v.RequestAcct)
	v.mux.HandleFunc("POST /newacct", v.VerifyAcct)
	// TODO: protect this
	v.mux.HandleFunc("/admin", v.Admin)

	// TODO: dunno if need this in long term after SPA
	fmap := template.FuncMap{
		"inc": func(i int) int {
			return i + 1
		},
	}
	v.index = template.Must(template.New("main").Funcs(fmap).ParseGlob(tp))
}

func (v *View) Cleanup() {
	v.log.Info("Cleaning up... done.")
}

func (v *View) HandleIndex() http.Handler {
	return http.HandlerFunc(v.Index)
}

func (v *View) Index(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie("ndlzr")
	if err != nil {
		if err == http.ErrNoCookie {
			v.log.Info("cookie not set")
			http.SetCookie(w, &http.Cookie{Name: "ndlzr", Value: "mycookie"})
		} else {
			v.log.Error(err.Error())
		}
	} else {
		v.log.Info(c.Name)
		v.log.Info(c.Value)
	}

	err = v.index.ExecuteTemplate(w, "index.tmpl", nil)
	if err != nil {
		// TODO: make this a flash?
		io.WriteString(w, err.Error())
		v.log.Error(err.Error())
	}
}

func (v *View) RequestAcct(w http.ResponseWriter, r *http.Request) {
	err := v.index.ExecuteTemplate(w, "newacct.tmpl", nil)
	if err != nil {
		io.WriteString(w, err.Error())
		v.log.Error(err.Error())
	}
}

func (v *View) VerifyAcct(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		io.WriteString(w, err.Error())
		v.log.Error(err.Error())
	}
	fname := r.PostFormValue("firstname")
	lname := r.PostFormValue("lastname")
	email := r.PostFormValue("email")
	loginid := r.PostFormValue("loginid")
	notes := r.PostFormValue("notes")

	err = v.db.AddPending(fname, lname, email, loginid, notes)
	if err != nil {
		io.WriteString(w, err.Error())
		v.log.Error(err.Error())
	}

	io.WriteString(w, "Thanks! You will receive an email from the administrator soon.")
}

func (v *View) Login(w http.ResponseWriter, r *http.Request) {
	err := v.index.ExecuteTemplate(w, "login.tmpl", nil)
	if err != nil {
		io.WriteString(w, err.Error())
		v.log.Error(err.Error())
	}
}

func (v *View) Authorize(w http.ResponseWriter, r *http.Request) {

}
