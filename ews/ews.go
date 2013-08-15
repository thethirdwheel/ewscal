package ews

import (
	"encoding/xml"
)

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
