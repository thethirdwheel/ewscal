package main

import (
	"bufio"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
    "os/exec"
    "io/ioutil"
    "bytes"
)

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
    log.Print("About to make me an encoder")
	enc := xml.NewEncoder(output)
	enc.Indent("  ", "    ")
	_, err := fmt.Fprint(output, xml.Header)
    if err != nil {
        log.Fatal("Couldn't write to pipe", err)
    }
	if err := enc.Encode(envelope); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
	}
    log.Print("just closing")
    output.Close()
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

//Mocked for now, get the data from Exchange eventually
func getRooms(all bool) (r Rooms) {
	r = readRoomRecords()
    exchangeHost, err := ioutil.ReadFile("data/host")
    if err != nil {
        log.Fatal("couldn't read host", err)
    }
    authFile, err := ioutil.ReadFile("data/authfile")
    if err != nil {
        log.Fatal("couldn't read auth", err)
    }

    log.Print("Got host and auth")
    log.Printf("host: %v", strings.TrimSpace(string(exchangeHost)))
    log.Printf("auth: %v", strings.TrimSpace(string(authFile)))
    cmd := exec.Command("curl", "--ntlm", strings.TrimSpace(string(exchangeHost)), "-u", strings.TrimSpace(string(authFile)), "--data", "@-", "--header", "content-type: text/xml; charset=utf-8")
    log.Printf("cmd: %v", cmd)
//    cmd := exec.Command("cat")

    read, write := io.Pipe()
    cmd.Stdin = read
    var buffer bytes.Buffer
    cmd.Stdout = &buffer
    log.Print("doing fine so far")

    if err := cmd.Start(); err != nil {
        log.Fatal("Couldn't start command", err)
    }
    log.Print("About to write request")
	go writeAvailabilityRequest(r, time.Now().Format(time.RFC3339), time.Now().Add(time.Hour).Format(time.RFC3339), write)
    if err := cmd.Wait(); err != nil {
        log.Fatal("Failed on wait", err)
    }
    log.Print("Made it through the wait")
    io.Copy(os.Stdout, &buffer)
//	r[2].Start = time.Now().Add(time.Minute * 15)
//	r[2].Open = false
//	r[4].Start = time.Now().Add(time.Hour)
//	r[4].Open = false
	//	r = append(r, Room{"Room1", "2nd", 4, "CR-PL2-Room1@place.com", time.Now(), time.Duration(10), true})
//	if all {
		//		r = append(r, Room{"Room2", "8th", 6, "CR-PL8-Room2@place.com", time.Now().Add(time.Minute * 5), time.Duration(60), false})
//	}
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

func readRoomRecords() Rooms {
	rooms := make([]Room, 0)
	roomfile, err := os.Open("data/roomRecords")
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
