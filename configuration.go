package main

import (
	"errors"
	"fmt"
	kinesis "github.com/sendgridlabs/go-kinesis"
	"os"
	"strconv"
)

type Configuration struct {
	AwsAccessKeyId     string
	AwsSecretAccessKey string
	AwsRegion          kinesis.Region
	EventStream        string
	SocketPath         string
	SocketMode         uint64
}

var (
	EnvVariableNotFound = errors.New("Environment variable not found")
)

func getConfiguration(variable string) (conf string, err error) {
	if len(os.Getenv(variable)) > 0 {
		conf = os.Getenv(variable)
	} else {
		return "", EnvVariableNotFound
	}
	return conf, nil
}

func getConfigurationFromEnv(exit func()) (conf Configuration) {
	err := errors.New("")
	conf.EventStream, err = getConfiguration("EVENT_STREAM")
	if err == EnvVariableNotFound {
		fmt.Print(err.Error() + "\nConfiguration for EVENT_STREAM not set, will default to 'PrimaryEventStream'\n")
		conf.EventStream = "PrimaryEventStream"
	}
	println("Sending data to " + conf.EventStream)
	conf.AwsAccessKeyId, err = getConfiguration("AWS_ACCESS_KEY_ID")
	if err == EnvVariableNotFound {
		fmt.Printf(err.Error() + "\nConfiguration for AWS_ACCESS_KEY_ID not set\n")
		exit()
	}
	conf.AwsSecretAccessKey, err = getConfiguration("AWS_SECRET_ACCESS_KEY")
	if err == EnvVariableNotFound {
		fmt.Printf(err.Error() + "\nConfiguration for AWS_SECRET_ACCESS_KEY not set\n")
		exit()
	}
	aws_region_string, err := getConfiguration("AWS_REGION")
	conf.AwsRegion = kinesis.Region{"us-east-1"}
	if err == EnvVariableNotFound {
		fmt.Printf(err.Error() + "\nNo configuration set for AWS_REGION, will default to us-east-1\n")
	} else {
		conf.AwsRegion = kinesis.Region{aws_region_string}
	}
	conf.SocketPath, err = getConfiguration("SOCKET_PATH")
	if err == EnvVariableNotFound {
		fmt.Printf(err.Error() + "\nSocket not configured, will set to /tmp/eventingestion.sock\n")
		conf.SocketPath = "/tmp/eventingestion.sock"
	}
	permission, err := getConfiguration("SOCKET_MODE")
	if err == EnvVariableNotFound {
		fmt.Printf(err.Error() + "\nSOCKET_MODE is not readable, will leave it readable for $USER\n")
	} else {
		conf.SocketMode, err = strconv.ParseUint(permission, 0, 32)
		if err != nil {
			fmt.Printf(permission + " is not a valid mode")
		}
	}
	return conf

}
