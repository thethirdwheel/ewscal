package main

import (
	"encoding/json"
	"net/http"
	"time"
)

type TimeWindow struct {
	Start    time.Time
	Duration time.Duration
}

type Room struct {
	Name      string
	Vacancies []TimeWindow
}

type Rooms []Room

func makeHandler(fn func(http.ResponseWriter, *http.Request, bool), all bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fn(w, r, all)
	}
}

//Mocked for now, get the data from Exchange eventually
func getRooms(all bool) (r Rooms) {
	r = append(r, Room{"Room1", []TimeWindow{TimeWindow{time.Now(), time.Duration(10)}, TimeWindow{time.Now(), time.Duration(200)}}})
	if all {
		r = append(r, Room{"Room2", []TimeWindow{TimeWindow{time.Now(), time.Duration(60)}, TimeWindow{time.Now(), time.Duration(12)}}})
	}
	return
}

func apiHandler(w http.ResponseWriter, r *http.Request, all bool) {
	listing := getRooms(all)
	response, err := json.Marshal(listing)
	if nil == err {
		w.Write(response)
	}
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/api/v1/room/all", http.StatusFound)
}

func main() {
	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/api/v1/room/all", makeHandler(apiHandler, true))
	http.HandleFunc("/api/v1/room/available", makeHandler(apiHandler, false))
	http.ListenAndServe(":8080", nil)
}
