package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/thethirdwheel/ewscal/ews"
	"net/http"
	"time"
)

func apiHandler(w http.ResponseWriter, r *http.Request, all bool) {
	listing := ews.GetRooms(all, time.Now(), time.Now().Add(time.Hour), "data/roomRecords")
	response, err := json.Marshal(listing)
	if nil == err {
		w.Write(response)
	}
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/api/v1/room/all", http.StatusFound)
}

func makeHandler(fn func(http.ResponseWriter, *http.Request, bool), all bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fn(w, r, all)
	}
}

func main() {
	var port = flag.Int("port", 8080, "Port to run ewscal from")
	flag.Parse()
	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/api/v1/room/all", makeHandler(apiHandler, true))
	http.HandleFunc("/api/v1/room/available", makeHandler(apiHandler, false))
	http.ListenAndServe(fmt.Sprintf(":%v", *port), nil)
}
