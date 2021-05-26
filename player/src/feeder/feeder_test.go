package feeder

import (
	"bytes"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/majeinfo/chaingun/config"
	log "github.com/sirupsen/logrus"
)

func TestBadCSVFile(t *testing.T) {
	feeder := config.Feeder{Type: "csv", Separator: ";", Filename: "badfile"}
	var buf bytes.Buffer
	log.SetOutput(&buf)

	if Csv(feeder, "baddir") {
		t.Errorf("Giving a bad filename and bad directory should fail")
	}

	log.SetOutput(os.Stderr)
}

func TestEmptyCSVFile(t *testing.T) {
	file, err := ioutil.TempFile("/tmp", "feeder")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(file.Name())

	var buf bytes.Buffer
	log.SetOutput(&buf)

	feeder := config.Feeder{Type: "csv", Separator: ";", Filename: file.Name()}
	if !Csv(feeder, "") {
		t.Errorf("Cannot read file %s, message: %s", file.Name(), buf.String())
	}
	if !strings.Contains(buf.String(), "fed with 0 line"){
		t.Errorf("Wrong message: Expected: %s, got: %s", "CSV feeder fed with 0 lines of data", buf.String())
	}

	log.SetOutput(os.Stderr)
}

func TestWrongHeadersCSVFile(t *testing.T) {
	file, err := ioutil.TempFile("/tmp", "feeder")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(file.Name())

	csv_data := []byte("Name,Surname\nDoe,John,3")
	if err := ioutil.WriteFile(file.Name(), csv_data, 0644); err != nil {
		t.Errorf("Cannot write data in file %s, err=%s", file.Name(), err)
		return
	}

	var buf bytes.Buffer
	log.SetOutput(&buf)

	feeder := config.Feeder{Type: "csv", Separator: ",", Filename: file.Name()}
	if !Csv(feeder, "") {
		t.Errorf("Cannot read file %s, message: %s", file.Name(), buf.String())
		log.SetOutput(os.Stderr)
		return
	}
	if !strings.Contains(buf.String(), "columns mismatch"){
		t.Errorf("Wrong message: Expected: %s, got: %s", "Number of columns mismatch with headers", buf.String())
	}

	log.SetOutput(os.Stderr)
}

func TestGoodCSVFile(t *testing.T) {
	file, err := ioutil.TempFile("/tmp", "feeder")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(file.Name())

	csv_data := []byte(`Name,Surname,Age
Doe,John,35
Smith,Alice,30
Jones,Bob,60
`)
	if err := ioutil.WriteFile(file.Name(), csv_data, 0644); err != nil {
		t.Errorf("Cannot write data in file %s, err=%s", file.Name(), err)
		return
	}

	var buf bytes.Buffer
	log.SetOutput(&buf)

	my_feeder := config.Feeder{Type: "csv", Separator: ",", Filename: file.Name()}
	if !Csv(my_feeder, "") {
		t.Errorf("Cannot read file %s, message: %s", file.Name(), buf.String())
		log.SetOutput(os.Stderr)
		return
	}

	go NextFromFeeder()
	feedData := <-FeedChannel
	if feedData["Name"] != "Doe" || feedData["Surname"] != "John" || feedData["Age"] != "35" {
		t.Errorf("Bad values returned: %v", feedData)
	}
	go NextFromFeeder()
	feedData = <-FeedChannel
	if feedData["Name"] != "Smith" || feedData["Surname"] != "Alice" || feedData["Age"] != "30" {
		t.Errorf("Bad values returned: %v", feedData)
	}
	go NextFromFeeder()
	feedData = <-FeedChannel
	if feedData["Name"] != "Jones" || feedData["Surname"] != "Bob" || feedData["Age"] != "60" {
		t.Errorf("Bad values returned: %v", feedData)
	}
	// Loop on data
	go NextFromFeeder()
	feedData = <-FeedChannel
	if feedData["Name"] != "Doe" || feedData["Surname"] != "John" || feedData["Age"] != "35" {
		t.Errorf("Bad values returned: %v", feedData)
	}
	/*
	for item := range feedData {
		t.Errorf("Item: %s, value: %s", item, feedData[item])
	}
	*/

	log.SetOutput(os.Stderr)
}