package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"time"
)

// Port is the tcp port to scan.
var Port string

// TimeLimit is the amount of time it will continue to run the checks for.
var TimeLimit time.Duration

// CheckInterval is how often a check will run within a time limit.
var CheckInterval time.Duration

// PortDefault is the tcp port to scan.
var PortDefault = "22"

// TimeLimitDefault is the amount of time it will continue to run the checks for.
var TimeLimitDefault = time.Duration(5) * time.Minute

// CheckIntervalDefault is how often a check will run within a time limit.
var CheckIntervalDefault = time.Duration(30) * time.Second

// DataToRead will hold an io.Reader either from STDIN or from ENV variable HOSTS
var DataToRead io.Reader

func init() {
	flag.StringVar(
		&Port,
		"p",
		PortDefault,
		"Port to scan. The env variable is PORT",
	)
	flag.DurationVar(
		&TimeLimit,
		"t",
		TimeLimitDefault,
		"The number of minutes to continue checking the port. The env variable is TIMELIMIT",
	)
	flag.DurationVar(
		&CheckInterval,
		"i",
		CheckIntervalDefault,
		"The number of seconds to wait before each check. The env variable is CHECKINTERVAL",
	)
	flag.Parse()

	if Port == PortDefault && os.Getenv("PORT") != "" {
		Port = os.Getenv("PORT")
	}

	var err error

	if TimeLimit == TimeLimitDefault && os.Getenv("TIMELIMIT") != "" {
		TimeLimit, err = time.ParseDuration(os.Getenv("TIMELIMIT"))
		if err != nil {
			fmt.Println("Please pass time as : '1h', '1m', '30s'")
			os.Exit(1)
		}
	}

	if CheckInterval == CheckIntervalDefault && os.Getenv("CHECKINTERVAL") != "" {
		CheckInterval, err = time.ParseDuration(os.Getenv("CHECKINTERVAL"))
		if err != nil {
			fmt.Println("Please pass time as : '1h', '1m', '30s'")
			os.Exit(1)
		}
	}

}

func main() {
	resultsQueue := make(chan Result)
	hostSettings := Host{
		Port:            Port,
		MaxFailureCount: 5,
		TimeLimit:       TimeLimit,
		CheckInterval:   CheckInterval,
	}

	var localErr error
	if CheckStdin(os.Stdin) {
		DataToRead = os.Stdin
	} else {
		DataToRead, localErr = EnvStringToReader("HOSTS")
		if localErr != nil {
			fmt.Printf("Failed to include STDIN or to include ENV variable HOSTS : %s", localErr.Error())
			os.Exit(1)
		}
	}

	hostLists, err := ReadInput(hostSettings, DataToRead)
	if err != nil {
		fmt.Printf("Error occured %s", err.Error())
	}
	hostCount := len(hostLists)
	if CheckInterval > TimeLimit {
		fmt.Println("The Interval must me less than the time limit")
		os.Exit(1)
	}

	client := Client{}
	for _, hostToRun := range hostLists {
		go hostToRun.Run(resultsQueue, client)
	}

	ReadResults(resultsQueue, hostCount)

}
