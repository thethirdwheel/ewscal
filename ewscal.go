package main

import (
	"bufio"
	"encoding/json"
	"encoding/xml"
	"flag"
	"fmt"
	"github.com/thethirdwheel/ewscal/ews"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Room struct {
	Name     string
	Floor    string
	Size     int
	Email    string
	Start    time.Time
	Duration time.Duration
	Open     bool
}

type Rooms []Room

func (r Rooms) Len() int {
	return len(r)
}

func (r Rooms) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}

type ByStart struct{ Rooms }

func (r ByStart) Less(i, j int) bool {
	return r.Rooms[i].Start.Before(r.Rooms[j].Start)
}

var RFC3339NoTZ = strings.TrimSuffix(time.RFC3339, "Z07:00")

func generateMailboxes(roomlist Rooms) (m ews.Mailboxes) {
	for _, r := range roomlist {
		m.Boxes = append(m.Boxes, ews.Mailbox{EmailAddress: r.Email, AttendeeType: "Required"})
	}
	return
}

func writeAvailabilityRequest(roomlist Rooms, startdate string, enddate string, output io.WriteCloser) {
	defer output.Close()
	boxen := generateMailboxes(roomlist)

	//Set timezone information for return values
	tz := ews.TimeZone{Xmlns: "http://schemas.microsoft.com/exchange/services/2006/types", Bias: 480}
	tz.StandardTime = ews.Timeblock{Bias: 0, Time: "02:00:00", DayOrder: 5, Month: 10, DayOfWeek: "Sunday"}
	tz.DaylightTime = ews.Timeblock{Bias: -60, Time: "02:00:00", DayOrder: 1, Month: 4, DayOfWeek: "Sunday"}

	//Set time window of interest
	tw := ews.TimeWindow{StartTime: startdate, EndTime: enddate}
	requestWindow := ews.FreeBusyView{MergedInterval: 60, RequestedView: "FreeBusy"}
	requestWindow.TimeWindow = tw

	request := ews.AvailabilityRequest{Xmlns: "http://schemas.microsoft.com/exchange/services/2006/messages", XmlnsT: "http://schemas.microsoft.com/exchange/services/2006/types", Tz: tz, MailboxDataArray: boxen, Fbv: requestWindow}

	body := ews.AvailabilityEnvelopeBody{Request: request}
	envelope := ews.AvailabilityEnvelope{XmlnsXsi: "http://www.w3.org/2001/XMLSchema-instance", XmlnsXsd: "http://www.w3.org/2001/XMLSchema", XmlnsSoap: "http://schemas.xmlsoap.org/soap/envelope/", XmlnsT: "http://schemas.microsoft.com/exchange/services/2006/types", Body: body}
	enc := xml.NewEncoder(output)
	enc.Indent("  ", "    ")
	_, err := fmt.Fprint(output, xml.Header)
	if err != nil {
		log.Fatal("Couldn't write to pipe", err)
	}
	if err := enc.Encode(envelope); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
	}
}

func makeHandler(fn func(http.ResponseWriter, *http.Request, bool), all bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fn(w, r, all)
	}
}

func updateRoomsFromResponse(r *Rooms, responseReader io.ReadCloser, startTime time.Time) {
	loc, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		log.Fatal(err)
	}
	v := ews.FreeBusyResponseEnvelope{}
	if err := xml.NewDecoder(responseReader).Decode(&v); err != nil {
		log.Fatal(err)
	}
	for i, response := range v.Body.Response.ResponseArray.Responses {
		(*r)[i].Start = startTime
		(*r)[i].Duration = time.Hour
		for _, event := range response.View.CalendarArray.Events {
			eventStart, err := time.ParseInLocation(RFC3339NoTZ, event.StartTime, loc)
			if err != nil {
				log.Fatal("couldn't parse start date: ", err)
			}
			eventEnd, err := time.ParseInLocation(RFC3339NoTZ, event.EndTime, loc)
			if err != nil {
				log.Fatal("couldn't parse end date: ", err)
			}
			if (*r)[i].Start.Before(eventStart) && eventStart.Sub((*r)[i].Start) > time.Duration(10)*time.Minute {
				(*r)[i].Duration = eventStart.Sub((*r)[i].Start)
				break
			} else {
				(*r)[i].Open = false
				(*r)[i].Start = eventEnd
			}
		}
	}
}

func initRoomTimes(r *Rooms, startTime time.Time) {
	for i, _ := range *r {
		(*r)[i].Start = startTime
	}
}

func getRooms(all bool, startTime time.Time, endTime time.Time, rConf string) (r Rooms) {
	r = readRoomRecords(rConf)
	initRoomTimes(&r, startTime)
	exchangeHost, err := ioutil.ReadFile("data/host")
	if err != nil {
		log.Fatal("couldn't read host", err)
	}
	authFile, err := ioutil.ReadFile("data/authfile")
	if err != nil {
		log.Fatal("couldn't read auth", err)
	}

	cmd := exec.Command("curl", "--ntlm", strings.TrimSpace(string(exchangeHost)), "-u", strings.TrimSpace(string(authFile)), "--data", "@-", "--header", "content-type: text/xml; charset=utf-8")

	stdin, err := cmd.StdinPipe()
	if err != nil {
		log.Fatal(err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}

	if err := cmd.Start(); err != nil {
		log.Fatal("Couldn't start command", err)
	}

	go writeAvailabilityRequest(r, startTime.Format(time.RFC3339), endTime.Format(time.RFC3339), stdin)
	updateRoomsFromResponse(&r, stdout, startTime)
	if err := cmd.Wait(); err != nil {
		log.Fatal("Failed on wait", err)
	}
	sort.Sort(ByStart{r})
	return
}

func apiHandler(w http.ResponseWriter, r *http.Request, all bool) {
	listing := getRooms(all, time.Now(), time.Now().Add(time.Hour), "data/roomRecords")
	response, err := json.Marshal(listing)
	if nil == err {
		w.Write(response)
	}
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/api/v1/room/all", http.StatusFound)
}

func rowToRoom(row string) (r Room) {
	rowArray := strings.Split(row, ",")
	r.Name = rowArray[0]
	r.Floor = rowArray[1]
	r.Size, _ = strconv.Atoi(rowArray[2])
	r.Email = rowArray[3]
	r.Start = time.Now()
	r.Duration = time.Minute
	r.Open = true
	return
}

func readRoomRecords(filename string) Rooms {
	rooms := make([]Room, 0)
	roomfile, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	scanner := bufio.NewScanner(roomfile)
	for scanner.Scan() {
		rooms = append(rooms, rowToRoom(scanner.Text()))
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	return Rooms(rooms)
}

func main() {
	var port = flag.Int("port", 8080, "Port to run ewscal from")
	flag.Parse()
	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/api/v1/room/all", makeHandler(apiHandler, true))
	http.HandleFunc("/api/v1/room/available", makeHandler(apiHandler, false))
	http.ListenAndServe(fmt.Sprintf(":%v", *port), nil)
}
