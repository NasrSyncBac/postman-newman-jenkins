package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"time"

	"github.com/Syncbak-Git/controldb"
	"github.com/Syncbak-Git/jsconfig"
	"github.com/Syncbak-Git/log"
	"github.com/julienschmidt/httprouter"
)

type HomeDisplay struct {
	Catchers  map[string][]string
	Redirects []*Redirect
	Title     string
}

var maxCatcher int
var maxAdapter int

func main() {

}

func Home(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	ids, err := getCatchers()
	ids.Title = "Catchers"
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if len(ids.Catchers) == 0 {
		http.Error(w, "No ids in redis db", http.StatusInternalServerError)
		return
	}
	for k := range ids.Catchers {
		v := ids.Catchers[k]
		for i := range v {
			v[i] = strings.Replace(v[i], ".syncbak.corp", "", 1)
		}
	}
	err = templates.Execute(w, ids)

	if err != nil {
		log.Error("error executing template %s\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

var templates = template.Must(template.ParseFiles("views/index.html"))

func db() *controldb.SourceStreamDb {
	time := time.Duration(4 * time.Second)
	redisdb := jsconfig.S.FindString("Redis")
	redispw := jsconfig.S.FindString("RedisPwd")
	return controldb.NewSourceStreamDb(redisdb, redispw, time)
}

func getCatchers() (*HomeDisplay, error) {
	db := db()
	if db == nil {
		return nil, fmt.Errorf("Could not get db it is nil %+v", db)
	}

	assigned, err := db.FetchAllCatchers()
	check(err)

	rd := newRedirectDb(jsconfig.S.FindString("Redis"), jsconfig.S.FindString("RedisPwd"), jsconfig.S.FindString("RedirectPrefix"))
	rds, err := rd.streams()
	check(err)

	return &HomeDisplay{Catchers: assigned, Redirects: rds}, nil
}

func writeAdapterInfo(w http.ResponseWriter, r *http.Request, write func(http.ResponseWriter, map[string][]string)) {
	ids, err := adapterAssignments()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if len(ids) == 0 {
		http.Error(w, "No adapters in rabbitmq", http.StatusInternalServerError)
		return
	}
	write(w, ids)
}

func AdapterCount(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	writeAdapterInfo(w, r, func(newWriter http.ResponseWriter, vals map[string][]string) {
		fmt.Fprintf(w, "%d", len(vals))
	})
}

func AdapterSlots(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	writeAdapterInfo(w, r, func(newWriter http.ResponseWriter, vals map[string][]string) {
		fmt.Fprintf(w, "%d", len(vals)*maxAdapter)
	})
}

func AdapterSlotsUsed(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	writeAdapterInfo(w, r, func(newWriter http.ResponseWriter, vals map[string][]string) {
		count := 0
		for _, streams := range vals {
			count += len(streams)
		}
		fmt.Fprintf(w, "%d", count)
	})
}

func Adapters(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	writeAdapterInfo(w, r, func(newWriter http.ResponseWriter, vals map[string][]string) {
		hd := &HomeDisplay{Catchers: vals, Title: "CDN Adapters"}
		err := templates.Execute(w, hd)
		if err != nil {
			log.Error("error executing templates %s\n", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
}

func Transcoders(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	t, err := connectedTranscoders()
	if err != nil {
		log.Error("error executing template %s\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	hd := &HomeDisplay{Catchers: t, Title: "Transcoders"}
	err = templates.Execute(w, hd)
	if err != nil {
		log.Error("error executing template %s\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func writeCatcherStat(w http.ResponseWriter, r *http.Request, isSlots bool) {
	ids, err := getCatchers()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if len(ids.Catchers) == 0 {
		http.Error(w, "No ids in redis db", http.StatusInternalServerError)
	}
	val := len(ids.Catchers)
	if isSlots {
		val = val * maxCatcher
	}
	fmt.Fprintf(w, "%d", val)
}

func Catchers(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	db := db()
	if db == nil {
		log.Error("could not get database! it is nil")
		http.Error(w, "db is nil", http.StatusInternalServerError)
		return
	}
	slots, err := db.FetchAllCatchers()
	if err != nil {
		http.Error(w, "Could not fetch catchers error:"+err.Error(), http.StatusInternalServerError)
		return
	}
	serveJson(w, slots)
}

func serveJson(w http.ResponseWriter, obj interface{}) {
	b, err := json.Marshal(obj)
	if err != nil {
		http.Error(w, "Could not marshal json from adapter command", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, string(b))

}
