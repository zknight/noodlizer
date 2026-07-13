package http

import (
	"io"
	"net/http"
)

func (v *View) Admin(w http.ResponseWriter, r *http.Request) {
	pus, err := v.db.GetPending()
	if err != nil {
		// TODO: make this a flash
		io.WriteString(w, err.Error())
		v.log.Error(err.Error())
		return
	}
	err = v.index.ExecuteTemplate(w, "admin.tmpl", pus)
	if err != nil {
		// TODO: make this a flash
		io.WriteString(w, err.Error())
		v.log.Error(err.Error())
	}
}
