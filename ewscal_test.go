package main

import (
	"bytes"
	"encoding/xml"
	"io/ioutil"
	"log"
	"testing"
	"time"
)

func TestParseAvailabilityResponse(t *testing.T) {
	r := readRoomRecords("data/testRoomRecords")
	testData, err := ioutil.ReadFile("data/test_response.xml")
	if err != nil {
		log.Fatal("couldn't read test_response.xml: ", err)
	}
	//    log.Print(time.RFC3339)
	startTime, err := time.Parse(time.RFC3339, "2013-08-05T08:00:00Z")
	if err != nil {
		log.Fatal("couldn't parse date: ", err)
	}
	initRoomTimes(&r, startTime)
	log.Printf("%v", r)
	v := FreeBusyResponseEnvelope{}
	//	log.Printf("%v", v)
	if err := xml.Unmarshal(testData, &v); err != nil {
		log.Fatal("error: %v", err)
	}
	//	log.Printf("%v", v)
	if len(r) != len(v.Body.Response.ResponseArray.Responses) {
		t.Errorf("Length of r (%v) not equal to length of responses (%v)", len(r), len(v.Body.Response.ResponseArray.Responses))
	}
	updateRoomsFromResponse(&r, *bytes.NewBuffer(testData), startTime)
	if len(r) != len(v.Body.Response.ResponseArray.Responses) {
		t.Errorf("Length of r (%v) not equal to length of responses (%v)", len(r), len(v.Body.Response.ResponseArray.Responses))
	}
	log.Printf("%v", r)
}
