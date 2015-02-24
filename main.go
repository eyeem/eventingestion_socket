package main

import (
	"bufio"
	"crypto/md5"
	"fmt"
	backoff "github.com/cenkalti/backoff"
	kinesis "github.com/sendgridlabs/go-kinesis"
	"net"
	"os"
	"os/signal"
	"syscall"
)

func SendToKinesis(ksis *kinesis.Kinesis, record *kinesis.RequestArgs) {
	var resp *kinesis.PutRecordResp
	var retErr error
	err := backoff.Retry(func() error {
		resp, retErr = ksis.PutRecord(record)
		return retErr
	}, backoff.NewExponentialBackOff())
	if err != nil {
		fmt.Printf("PutRecord err: %v %v\n", resp, err)
	} else {
		//fmt.Printf("PutRecord: %v\n", resp)
	}
}

func readEventLoop(events chan string, eventstream string, aws_access_key_id string, aws_secret_access_key string, aws_region kinesis.Region) {
	ksis := kinesis.New(aws_access_key_id, aws_secret_access_key, aws_region)
	for {
		event := <-events
		args := kinesis.NewArgs()
		partitionkey := fmt.Sprintf("%x", md5.Sum([]byte(event)))
		args.Add("StreamName", eventstream)
		args.AddData([]byte(event))
		args.Add("PartitionKey", partitionkey)
		go SendToKinesis(ksis, args)
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
	configuration := getConfigurationFromEnv(func() { os.Exit(1) })
	l, err := net.Listen("unix", configuration.SocketPath)
	if err != nil {
		println("listen error", err.Error())
		return
	} else {
		println("Listening on " + configuration.SocketPath)
	}
	if configuration.SocketMode != 0 {
		fmt.Printf("Changing socket permissions to " + string(configuration.SocketMode) + "\n")
		os.Chmod(configuration.SocketPath, os.FileMode(configuration.SocketMode))
	}
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)
	go signalHandler(c, l)
	events := make(chan string)
	go readEventLoop(events, configuration.EventStream, configuration.AwsAccessKeyId, configuration.AwsSecretAccessKey, configuration.AwsRegion)

	for {
		fd, err := l.Accept()
		if err != nil {
			fmt.Printf("Socket listen error " + err.Error())
			c <- os.Signal(syscall.SIGTERM)
		}
		go readFromFD(fd, events)
	}
}
