package utils

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestGoodMD5(t *testing.T) {
	file, err := ioutil.TempFile("/tmp", "md5")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(file.Name())

	md5_data := []byte("OK\n")
	if err := ioutil.WriteFile(file.Name(), md5_data, 0644); err != nil {
		t.Errorf("Cannot write data in file %s, err=%s", file.Name(), err)
		return
	}
	if result, err := Hash_file_md5(file.Name()); err != nil {
		t.Errorf("Could not compute MD5 sum: %s", err)
	} else if result != "d36f8f9425c4a8000ad9c4a97185aca5" {
		t.Errorf("Bad MD5")
	}
}

func TestBadMD5(t *testing.T) {
	file, err := ioutil.TempFile("/tmp", "md5")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(file.Name())

	md5_data := []byte("NOK\n")
	if err := ioutil.WriteFile(file.Name(), md5_data, 0644); err != nil {
		t.Errorf("Cannot write data in file %s, err=%s", file.Name(), err)
		return
	}
	if result, err := Hash_file_md5(file.Name()); err != nil {
		t.Errorf("Could not compute MD5 sum: %s", err)
	} else if result == "d36f8f9425c4a8000ad9c4a97185aca5" {
		t.Errorf("Bad MD5")
	}
}
