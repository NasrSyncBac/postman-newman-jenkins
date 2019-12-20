package main

import (
	"net/http"

	"github.com/Syncbak-Git/controldb"
	"github.com/Syncbak-Git/jsconfig"
	"github.com/julienschmidt/httprouter"
)

type HomeDisplay struct {
	Catchers  map[string][]string
	Redirects []*Redirect
	Title     string
}

func main() {

}

func Home(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	//ids, err := getCatchers()
}

func db() *controldb.SourceStreamDb {
	redisdb := jsconfig.S.FindString("Redis")
	redispw := jsconfig.S.FindString("RedisPwd")
}
