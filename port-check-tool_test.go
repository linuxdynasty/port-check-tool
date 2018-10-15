package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time"
)

var TesthostLists = `127.0.0.1
192.168.1.6
192.168.1.32
foo.bar.net`

func TestEnvStringToReaderPass(t *testing.T) {
	err := os.Setenv("HOSTS", TesthostLists)
	if err != nil {
		t.Errorf("Failed to set hostslist %s with err %s", TesthostLists, err.Error())
	}
	_, err = EnvStringToReader("HOSTS")
	if err != nil {
		t.Errorf("Failed to get hostslist %s with err %s", TesthostLists, err.Error())
	}
}

func TestEnvStringToReaderFail(t *testing.T) {
	_ = os.Unsetenv("HOSTS")
	_, err := EnvStringToReader("HOSTS")
	if err == nil {
		t.Error("Failed to produce error on non existent env var HOSTS")
	}
	if err != nil {
		if err.Error() != "ENV variable HOSTS does not exist" {
			t.Error("Errors does not match")
		}
	}

	err = os.Setenv("HOSTS", "")
	if err != nil {
		t.Errorf("Failed to set hostslist \"\" with err %s", err.Error())
	}
	if err != nil {
		if err.Error() != "ENV variable HOSTS does not exist" {
			t.Error("Errors does not match")
		}
	}
}

func RunCheckStdinTests(hostLists string, t *testing.T) bool {
	contents := []byte(hostLists)
	tmpfile, err := ioutil.TempFile("", "hostLists")
	if err != nil {
		t.Error(err.Error())
	}

	defer os.Remove(tmpfile.Name()) // clean up

	if _, err := tmpfile.Write(contents); err != nil {
		t.Error(err.Error())
	}

	if _, err := tmpfile.Seek(0, 0); err != nil {
		t.Error(err.Error())
	}

	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }() // Restore original Stdin

	os.Stdin = tmpfile
	dataExist := CheckStdin(os.Stdin)

	if err := tmpfile.Close(); err != nil {
		t.Error(err.Error())
	}

	return dataExist
}

func TestCheckStdinPass(t *testing.T) {
	dataExist := RunCheckStdinTests(TesthostLists, t)
	if !dataExist {
		t.Errorf("Could not read from STDIN")
	}
}

func TestCheckStdinFail(t *testing.T) {
	hostLists := ``
	dataExist := RunCheckStdinTests(hostLists, t)
	if dataExist {
		t.Errorf("Data should not exist")
	}
}

func TestReadInput(t *testing.T) {
	dataToRead := strings.NewReader(TesthostLists)
	hostSettings := Host{
		Port:            "22",
		MaxFailureCount: 5,
		TimeLimit:       time.Duration(1) * time.Minute,
		CheckInterval:   time.Duration(20) * time.Second,
	}

	hosts, err := ReadInput(hostSettings, dataToRead)
	if err != nil {
		t.Errorf("Error occured while reading host lists %s", err.Error())
	}
	if len(hosts) != 4 {
		t.Errorf("Length of host lists does not match the list of []*Host")
	}

	for index, val := range hosts {
		if val.Port != "22" {
			t.Error("Port does not match what was passed: 22")
		}
		if val.MaxFailureCount != 5 {
			t.Error("Maxfailure count does not match what was passed: 5")
		}
		if val.TimeLimit != time.Duration(1)*time.Minute {
			t.Errorf("TimeLimit does not match what was passed: %v", time.Duration(1)*time.Minute)
		}
		if val.CheckInterval != time.Duration(20)*time.Second {
			t.Errorf("CheckInterval does not match what was passed: %v", time.Duration(20)*time.Second)
		}
		if index == 0 {
			if val.Name != "127.0.0.1" {
				t.Errorf("Name does not match what was passed: %s", val.Name)
			}
		}
		if index == 1 {
			if val.Name != "192.168.1.6" {
				t.Errorf("Name does not match what was passed: %s", val.Name)
			}
		}
		if index == 2 {
			if val.Name != "192.168.1.32" {
				t.Errorf("Name does not match what was passed: %s", val.Name)
			}
		}
		if index == 3 {
			if val.Name != "foo.bar.net" {
				t.Errorf("Name does not match what was passed: %s", val.Name)
			}
		}
	}
}

type FakeClient struct {
	Err error
}

func (f FakeClient) ConnectTCP(address string, timeout time.Duration) (err error) {
	return f.Err
}

func TestRun(t *testing.T) {
	hosts := `127.0.0.1`
	resultsQueue := make(chan Result)
	resultCount := 0
	hostSettings := Host{
		Port:            Port,
		MaxFailureCount: 4,
		TimeLimit:       time.Duration(30) * time.Second,
		CheckInterval:   time.Duration(6) * time.Second,
	}

	DataToRead = strings.NewReader(hosts)
	hostLists, err := ReadInput(hostSettings, DataToRead)
	if err != nil {
		t.Errorf("Error occured %s", err.Error())
	}
	fc := FakeClient{
		Err: nil,
	}
	for _, hostToRun := range hostLists {
		go hostToRun.Run(resultsQueue, fc)
	}
	hostCount := len(hostLists)
DONE:
	for {
		select {
		case result := <-resultsQueue:
			resultCount++
			if result.Host.Name != "127.0.0.1" {
				t.Errorf("Name %s does not match 127.0.0.1", result.Host.Name)
			}
			if result.PassedCount != 4 {
				t.Errorf("Passed count %d does not == 4", result.PassedCount)
			}
			if hostCount == resultCount {
				fmt.Println("Finished processing results")
				break DONE
			}
		}
	}

}
