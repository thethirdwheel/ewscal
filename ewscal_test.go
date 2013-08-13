package main

import (
	"log"
	"os"
	"testing"
	"time"
)

func TestParseAvailabilityResponse(t *testing.T) {
	r := readRoomRecords("data/testRoomRecords")
	testFile, err := os.Open("data/test_response.xml")
	if err != nil {
		log.Fatal("couldn't read test_response.xml: ", err)
	}
	startTime, err := time.Parse(time.RFC3339, "2013-08-05T08:00:00Z")
	if err != nil {
		log.Fatal("couldn't parse date: ", err)
	}
	initRoomTimes(&r, startTime)
	updateRoomsFromResponse(&r, testFile, startTime)
}
