package main

import (
	"bufio"
	"encoding/xml"
//	"flag"
//	"fmt"
	"io"
	"log"
//	"os"
//	"time"
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

func generateMailboxes(source io.Reader) Mailboxes {
	boxen := make([]Mailbox, 0)
	scanner := bufio.NewScanner(source)
	for scanner.Scan() {
		boxen = append(boxen, Mailbox{EmailAddress: scanner.Text(), AttendeeType: "Required"})
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	return Mailboxes{Boxes: boxen}
}

/*
func main() {
	roomlist := flag.String("roomlist", "roomlist", "file containing the list of conference room email addresses to query")
	startdate := flag.String("startdate", time.Now().Add(time.Hour*12).Format(time.RFC3339), "start time in RFC3339 format")
	enddate := flag.String("enddate", time.Now().Add(time.Hour*24).Format(time.RFC3339), "end time in RFC3339 format")
	flag.Parse()

	//Build set of rooms we're interested in
	roomfile, err := os.Open(*roomlist)
	if err != nil {
		log.Fatal(err)
	}
	boxen := generateMailboxes(roomfile)

	//Set timezone information for return values
	tz := TimeZone{Xmlns: "http://schemas.microsoft.com/exchange/services/2006/types", Bias: 480}
	tz.StandardTime = Timeblock{Bias: 0, Time: "02:00:00", DayOrder: 5, Month: 10, DayOfWeek: "Sunday"}
	tz.DaylightTime = Timeblock{Bias: -60, Time: "02:00:00", DayOrder: 1, Month: 4, DayOfWeek: "Sunday"}

	//Set time window of interest
	tw := TimeWindow{StartTime: *startdate, EndTime: *enddate}
	requestWindow := FreeBusyView{MergedInterval: 60, RequestedView: "FreeBusy"}
	requestWindow.TimeWindow = tw

	request := AvailabilityRequest{Xmlns: "http://schemas.microsoft.com/exchange/services/2006/messages", XmlnsT: "http://schemas.microsoft.com/exchange/services/2006/types", Tz: tz, MailboxDataArray: boxen, Fbv: requestWindow}

	body := AvailabilityEnvelopeBody{Request: request}
	envelope := AvailabilityEnvelope{XmlnsXsi: "http://www.w3.org/2001/XMLSchema-instance", XmlnsXsd: "http://www.w3.org/2001/XMLSchema", XmlnsSoap: "http://schemas.xmlsoap.org/soap/envelope/", XmlnsT: "http://schemas.microsoft.com/exchange/services/2006/types", Body: body}
	enc := xml.NewEncoder(os.Stdout)
	enc.Indent("  ", "    ")
	fmt.Print(xml.Header)
	if err := enc.Encode(envelope); err != nil {
		fmt.Printf("error: %v\n", err)
	}
}*/
