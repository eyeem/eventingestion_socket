package main

import (
	"bufio"
	"crypto/md5"
	"errors"
	"fmt"
	kinesis "github.com/sendgridlabs/go-kinesis"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"
)

func getConfiguration(variable string) (conf string, err error) {
	if len(os.Getenv(variable)) > 0 {
		conf = os.Getenv(variable)
	} else {
		return "", errors.New("Environment variable " + variable + " not found\n")
	}
	return conf, nil
}

func readEventsAndSendToKinesis(events chan string, eventstream string, aws_access_key_id string, aws_secret_access_key string, aws_region kinesis.Region) {
    for {
        event := <-events
        ksis := kinesis.New(aws_access_key_id, aws_secret_access_key, aws_region)
        args := kinesis.NewArgs()
        partitionkey := fmt.Sprintf("%x", md5.Sum([]byte(event)))
        args.Add("StreamName", eventstream)
        args.AddData([]byte(event))
        args.Add("PartitionKey", partitionkey)
        go func() {
            resp4, err := ksis.PutRecord(args)
            if err != nil {
                fmt.Printf("PutRecord err: %v %v\n", resp4, err)
                if (err.Error() == "500 Internal Server Error") {
                    resp4, err := ksis.PutRecord(args)
                    if err != nil {
                        fmt.Printf("PutRecord for %v '%v' failed after retry, skipping.\n", resp4, event)
                    }
                }
            } else {
                //fmt.Printf("PutRecord: %v\n", resp4)
            }
        }()
    }
}

func readFromFD(fd net.Conn, events chan string) {
	for {
        data, err := bufio.NewReader(fd).ReadString('\n')
        if err != nil {
            if err.Error() == "EOF" {
                break
            } else {
                fmt.Printf(" BufReader error " + err.Error())
            }
        }
        events <- data
    }
}

func signalHandler(c chan os.Signal, l net.Listener) {
    <-c
    fmt.Printf("Got SIGTERM, exiting.\n")
    l.Close()
    os.Exit(0)
}

func main() {
	eventstream, err := getConfiguration("EVENT_STREAM")
	if err != nil {
		fmt.Print(err.Error() + "Configuration for EVENT_STREAM not set, will default to 'PrimaryEventStream'\n")
		eventstream = "PrimaryEventStream"
	}
	println("Sending data to " + eventstream)
	aws_access_key_id, err := getConfiguration("AWS_ACCESS_KEY_ID")
	if err != nil {
		fmt.Printf(err.Error() + "Configuration for AWS_ACCESS_KEY_ID not set\n")
		os.Exit(1)
	}
	aws_secret_access_key, err := getConfiguration("AWS_SECRET_ACCESS_KEY")
	if err != nil {
		fmt.Printf(err.Error() + "Configuration for AWS_SECRET_ACCESS_KEY not set\n")
		os.Exit(1)
	}
	aws_region_string, err := getConfiguration("AWS_REGION")
	aws_region := kinesis.Region{"us-east-1"}
	if err != nil {
		fmt.Printf(err.Error() + "No configuration set for AWS_REGION, will default to us-east-1\n")
	} else {
		aws_region = kinesis.Region{aws_region_string}
	}
	socket, err := getConfiguration("SOCKET_PATH")
	if err != nil {
		fmt.Printf(err.Error() + "\nSocket not configured, will set to /tmp/eventingestion.sock\n")
		socket = "/tmp/eventingestion.sock"
	}
	l, err := net.Listen("unix", socket)
	if err != nil {
		println("listen error", err.Error())
		return
	} else {
		println("Listening on " + socket)
		permission, err := getConfiguration("SOCKET_MODE")
		if err != nil {
			fmt.Printf(err.Error() + "SOCKET_MODE is not set in your environment variables, will leave it readable for $USER\n")
		} else {
			perm, err := strconv.ParseUint(permission, 0, 32)
			if err != nil {
				fmt.Printf(err.Error() + "SOCKET_MODE is not readable, will leave it readable for $USER\n")
			} else {
				fmt.Printf("Changing socket permissions to " + permission + "\n")
				os.Chmod(socket, os.FileMode(perm))
			}
		}
	}
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)
	go signalHandler(c, l)
    events := make(chan string)
    go readEventsAndSendToKinesis(events, eventstream, aws_access_key_id, aws_secret_access_key, aws_region)

	for {
		fd, err := l.Accept()
		if err != nil {
			fmt.Printf("Socket listen error " + err.Error())
			return
		}
		go readFromFD(fd, events)

	}

}
