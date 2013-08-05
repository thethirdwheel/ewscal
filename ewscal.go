package main

import (
	"encoding/json"
	"net/http"
	"time"
)

/*
type TimeWindow struct {
	Start    time.Time
	Duration time.Duration
}
*/

type Room struct {
	Name  string
	Floor string
	Size  int
	Email string
	//	Vacancies []TimeWindow
	Start    time.Time
	Duration time.Duration
	Open     bool
}

type Rooms []Room

func makeHandler(fn func(http.ResponseWriter, *http.Request, bool), all bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fn(w, r, all)
	}
}

//Mocked for now, get the data from Exchange eventually
func getRooms(all bool) (r Rooms) {
	r = append(r, Room{"Room1", "2nd", 4, "CR-PL2-Room1@place.com", time.Now(), time.Duration(10), true})
	if all {
		r = append(r, Room{"Room2", "8th", 6, "CR-PL8-Room2@place.com", time.Now().Add(time.Minute * 5), time.Duration(60), false})
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
