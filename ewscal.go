package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/thethirdwheel/ewscal/ews"
	"net/http"
	"path"
	"time"
)

func apiHandler(w http.ResponseWriter, r *http.Request, all bool, conf ews.RoomConf) {
	listing := ews.GetRooms(all, time.Now(), time.Now().Add(time.Hour), conf)
	response, err := json.Marshal(listing)
	if nil == err {
		w.Write(response)
	}
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/api/v1/room/all", http.StatusFound)
}

func makeHandler(fn func(http.ResponseWriter, *http.Request, bool, ews.RoomConf), all bool, conf ews.RoomConf) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fn(w, r, all, conf)
	}
}

func main() {
	var port = flag.Int("port", 8080, "Port to run ewscal from")
	var conf = flag.String("confDir", "data", "Folder containing ews configuration")
	flag.Parse()
	authFile := path.Join(*conf, "authfile")
	hostFile := path.Join(*conf, "host")
	roomFile := path.Join(*conf, "roomRecords")
	confObj := ews.MakeConf(authFile, hostFile, roomFile)
	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/api/v1/room/all", makeHandler(apiHandler, true, confObj))
	http.HandleFunc("/api/v1/room/available", makeHandler(apiHandler, false, confObj))
	http.ListenAndServe(fmt.Sprintf(":%v", *port), nil)
}
