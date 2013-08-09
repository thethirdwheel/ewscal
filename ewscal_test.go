package main

import (
	"io/ioutil"
	"testing"
	//    "bytes"
	"encoding/xml"
	"log"
)

func TestParseAvailabilityResponse(t *testing.T) {
	r := readRoomRecords()
	testData, err := ioutil.ReadFile("data/test_response.xml")
	if err != nil {
		log.Fatal("couldn't read test_response.xml: ", err)
	}
	v := FreeBusyResponseEnvelope{}
	log.Printf("%v", v)
	if err := xml.Unmarshal(testData, &v); err != nil {
		log.Fatal("error: %v", err)
	}
	log.Printf("%v", v)
	if len(r) != len(v.Body.Response.ResponseArray.Responses) {
		t.Errorf("Length of r (%v) not equal to length of responses (%v)", len(r), len(v.Body.Response.ResponseArray.Responses))
	}
	//    updateRoomsFromResponse(&r, *bytes.NewBuffer(testData))
}
