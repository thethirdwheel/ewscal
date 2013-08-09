package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

var RFC3339NoTZ = strings.TrimSuffix(time.RFC3339, "Z07:00")

type CalendarEvent struct {
	XMLName   xml.Name `xml:"CalendarEvent"`
	StartTime string
	EndTime   string
	BusyType  string
}

type CalendarEventArray struct {
	XMLName xml.Name        `xml:"CalendarEventArray"`
	Events  []CalendarEvent `xml:"CalendarEvent"`
}

type ResponseFreeBusyView struct {
	XMLName       xml.Name `xml:"FreeBusyView"`
	CalendarArray CalendarEventArray
}

type FreeBusyResponse struct {
	XMLName xml.Name `xml:"FreeBusyResponse"`
	View    ResponseFreeBusyView
}

type FreeBusyResponseArray struct {
	XMLName   xml.Name           `xml:"FreeBusyResponseArray"`
	Responses []FreeBusyResponse `xml:"FreeBusyResponse"`
}

type UserAvailabilityResponse struct {
	XMLName       xml.Name `xml:"GetUserAvailabilityResponse"`
	ResponseArray FreeBusyResponseArray
}

type SoapBody struct {
	XMLName  xml.Name `xml:"Body"`
	Response UserAvailabilityResponse
}

type FreeBusyResponseEnvelope struct {
	XMLName xml.Name `xml:"Envelope"`
	Body    SoapBody
}

type Timeblock struct {
	Bias      int
	Time      string
	DayOrder  int
	Month     int
	DayOfWeek string
}

type TimeZone struct {
	XMLName      xml.Name `xml:"t:TimeZone"`
	Xmlns        string   `xml:"xmlns,attr"`
	Bias         int
	StandardTime Timeblock
	DaylightTime Timeblock
}

type Mailboxes struct {
	XMLName xml.Name `xml:"MailboxDataArray"`
	Boxes   []Mailbox
}

type Mailbox struct {
	XMLName          xml.Name `xml:"t:MailboxData"`
	EmailAddress     string   `xml:"t:Email>t:Address"`
	AttendeeType     string   `xml:"t:AttendeeType"`
	ExcludeConflicts bool     `xml:"t:ExcludeConflicts"`
}

type TimeWindow struct {
	StartTime string `xml:"t:StartTime"`
	EndTime   string `xml:"t:EndTime"`
}

type FreeBusyView struct {
	XMLName        xml.Name   `xml:"t:FreeBusyViewOptions"`
	TimeWindow     TimeWindow `xml:"t:TimeWindow"`
	MergedInterval int        `xml:"t:MergedFreeBusyIntervalInMinutes"`
	RequestedView  string     `xml:"t:RequestedView"`
}

type AvailabilityRequest struct {
	XMLName          xml.Name `xml:"GetUserAvailabilityRequest"`
	Xmlns            string   `xml:"xmlns,attr"`
	XmlnsT           string   `xml:"xmlns:t,attr"`
	Tz               TimeZone
	MailboxDataArray Mailboxes
	Fbv              FreeBusyView
}

type AvailabilityEnvelopeBody struct {
	XMLName xml.Name `xml:"soap:Body"`
	Request AvailabilityRequest
}

type AvailabilityEnvelope struct {
	XMLName   xml.Name `xml:"soap:Envelope"`
	XmlnsXsi  string   `xml:"xmlns:xsi,attr"`
	XmlnsXsd  string   `xml:"xmlns:xsd,attr"`
	XmlnsSoap string   `xml:"xmlns:soap,attr"`
	XmlnsT    string   `xml:"xmlns:t,attr"`
	Body      AvailabilityEnvelopeBody
}

func generateMailboxes(roomlist Rooms) (m Mailboxes) {
	for _, r := range roomlist {
		m.Boxes = append(m.Boxes, Mailbox{EmailAddress: r.Email, AttendeeType: "Required"})
	}
	return
}

func writeAvailabilityRequest(roomlist Rooms, startdate string, enddate string, output io.WriteCloser) {
	defer output.Close()
	boxen := generateMailboxes(roomlist)

	//Set timezone information for return values
	tz := TimeZone{Xmlns: "http://schemas.microsoft.com/exchange/services/2006/types", Bias: 480}
	tz.StandardTime = Timeblock{Bias: 0, Time: "02:00:00", DayOrder: 5, Month: 10, DayOfWeek: "Sunday"}
	tz.DaylightTime = Timeblock{Bias: -60, Time: "02:00:00", DayOrder: 1, Month: 4, DayOfWeek: "Sunday"}

	//Set time window of interest
	tw := TimeWindow{StartTime: startdate, EndTime: enddate}
	requestWindow := FreeBusyView{MergedInterval: 60, RequestedView: "FreeBusy"}
	requestWindow.TimeWindow = tw

	request := AvailabilityRequest{Xmlns: "http://schemas.microsoft.com/exchange/services/2006/messages", XmlnsT: "http://schemas.microsoft.com/exchange/services/2006/types", Tz: tz, MailboxDataArray: boxen, Fbv: requestWindow}

	body := AvailabilityEnvelopeBody{Request: request}
	envelope := AvailabilityEnvelope{XmlnsXsi: "http://www.w3.org/2001/XMLSchema-instance", XmlnsXsd: "http://www.w3.org/2001/XMLSchema", XmlnsSoap: "http://schemas.xmlsoap.org/soap/envelope/", XmlnsT: "http://schemas.microsoft.com/exchange/services/2006/types", Body: body}
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

func makeHandler(fn func(http.ResponseWriter, *http.Request, bool), all bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fn(w, r, all)
	}
}

func updateRoomsFromResponse(r *Rooms, b bytes.Buffer, startTime time.Time) {
    loc, err := time.LoadLocation("America/Los_Angeles")
    if err != nil {
        log.Fatal("error: %v", err)
    }
	v := FreeBusyResponseEnvelope{}
	if err := xml.Unmarshal(b.Bytes(), &v); err != nil {
		log.Fatal("error: %v", err)
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
			if (*r)[i].Start.Before(eventStart) {
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

	read, write := io.Pipe()
	cmd.Stdin = read
	var buffer bytes.Buffer
	cmd.Stdout = &buffer

	if err := cmd.Start(); err != nil {
		log.Fatal("Couldn't start command", err)
	}
	go writeAvailabilityRequest(r, startTime.Format(time.RFC3339), endTime.Format(time.RFC3339), write)
	if err := cmd.Wait(); err != nil {
		log.Fatal("Failed on wait", err)
	}
	updateRoomsFromResponse(&r, buffer, startTime)
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
	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/api/v1/room/all", makeHandler(apiHandler, true))
	http.HandleFunc("/api/v1/room/available", makeHandler(apiHandler, false))
	http.ListenAndServe(":8080", nil)
}
