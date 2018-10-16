package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"time"
)

// Host contains the settings needed to make a health check.
type Host struct {
	// The Name or IP Address of the host that is to be scanned.
	Name string
	// The tcp Port to be scanned.
	Port string
	// The max number of failures that considers a node down.
	MaxFailureCount int
	// How long to continue to check if a host is down or up.
	TimeLimit time.Duration
	// The interval to run the actual tcp check.
	CheckInterval time.Duration
}

// Result constains the result of the health check.
type Result struct {
	// The Host struct.
	Host *Host
	// Number of times a check failed.
	FailedCount int
	// Number of times a check passed.
	PassedCount int
	// Is the node to be considered flapping.
	IsFlapping bool
	// Is the node to be considered up.
	IsUp bool
	// Is the node to be considered down.
	IsDown bool
	// The reporting message.
	Msg string
}

// EnvStringToReader will grab an ENV variable and convert it to  an io.Reader interface
func EnvStringToReader(env string) (reader io.Reader, err error) {
	envData := os.Getenv(env)
	if envData == "" {
		return reader, fmt.Errorf("ENV variable %s does not exist", env)
	}
	reader = strings.NewReader(envData)
	return reader, nil
}

// CheckStdin will check if anything was passed to STDIN.
func CheckStdin(file *os.File) bool {
	stat, err := file.Stat()
	if err != nil {
		return false
	}
	if stat.Size() == 0 {
		return false
	}
	return true
}

// ReadInput takes in the settings for check and returns a list of *Host structs from STDIN.
func ReadInput(host Host, reader io.Reader) (hosts []*Host, err error) {
	scanner := bufio.NewScanner(reader)
	if err := scanner.Err(); err != nil {
		return hosts, err
	}
	for scanner.Scan() {
		hostname := scanner.Text()
		hosts = append(hosts, &Host{
			Name:            hostname,
			Port:            host.Port,
			MaxFailureCount: host.MaxFailureCount,
			TimeLimit:       host.TimeLimit,
			CheckInterval:   host.CheckInterval,
		})
	}

	return hosts, nil
}

// Run the health check for a time limit and output the results into the Result channel.
// This function is meant to be run with go func.
func (h *Host) Run(r chan Result, c Connector) {
	timeout := time.After(h.TimeLimit)
	tick := time.Tick(h.CheckInterval)
	result := Result{
		Host:        h,
		FailedCount: 0,
		PassedCount: 0,
		IsFlapping:  false,
		IsUp:        false,
		IsDown:      false,
		Msg:         "",
	}

DONE:
	for {
		select {
		case <-timeout:
			if result.FailedCount == h.MaxFailureCount {
				result.IsDown = true
				result.Msg = fmt.Sprintf("%s is down", h.Name)
				r <- result
			} else if result.FailedCount > 0 && result.PassedCount == 0 {
				result.IsDown = true
				result.Msg = fmt.Sprintf("%s is down", h.Name)
				r <- result
			} else if result.FailedCount > 0 && result.PassedCount > 0 {
				result.IsFlapping = true
				result.Msg = fmt.Sprintf("%s is flapping", h.Name)
				r <- result
			} else if result.FailedCount == 0 && result.PassedCount > 0 {
				result.IsUp = true
				result.Msg = fmt.Sprintf("%s is reporting ok", h.Name)
				r <- result
			}
		case <-tick:
			err := h.CheckPort(c)
			if err != nil {
				result.FailedCount++
				if result.FailedCount == result.Host.MaxFailureCount {
					if result.FailedCount == h.MaxFailureCount {
						result.IsDown = true
						result.Msg = fmt.Sprintf("%s is down", h.Name)
						r <- result
						break DONE
					}
				}
			} else {
				result.PassedCount++
			}
		}
	}
}

func ReadResults(rq chan Result, hostCount int) {
	resultCount := 0
DONE:
	for {
		select {
		case result := <-rq:
			resultCount++
			fmt.Println(result.Msg)
			if hostCount == resultCount {
				fmt.Println("Finished processing results")
				break DONE
			}
		}
	}
}

// CheckPort will check the TCP port if it is up or down and report an error if one happens.
func (h *Host) CheckPort(c Connector) (err error) {
	timeOut := time.Duration(1) * time.Second
	address := net.JoinHostPort(h.Name, h.Port)
	err = c.ConnectTCP(address, timeOut)
	if err != nil {
		return err
	}
	return nil
}

// Client is used for the Connector interface.
type Client struct{}

// Connector interface is used for mocking ConnectTcp.
type Connector interface {
	ConnectTCP(string, time.Duration) error
}

// ConnectTcp will perform a tcp connection on the address specfied.
func (c Client) ConnectTCP(address string, timeOut time.Duration) (err error) {
	conn, err := net.DialTimeout("tcp", address, timeOut)
	if err != nil {
		return err
	}
	defer conn.Close()
	return nil

}
