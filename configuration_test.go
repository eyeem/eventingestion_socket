package main

import (
	kinesis "github.com/sendgridlabs/go-kinesis"
	"os"
	"strconv"
	"testing"
)

func TestGetConfiguration(t *testing.T) {
	os.Setenv("FOOBAR", "asd")
	defer os.Clearenv()
	conf, err := getConfiguration("FOOBAR")
	if conf != "asd" {
		t.Fail()
	}
	if err != nil {
		t.Fail()
	}

	conf, err = getConfiguration("BARBAZ")
	if conf != "" {
		t.Fail()
	}
	if err != EnvVariableNotFound {
		t.Fail()
	}
}

func TestGetConfigurationFromEnv(t *testing.T) {
	conf := getConfigurationFromEnv(func() {})
	if conf.AwsAccessKeyId != "" {
		t.Fail()
	}
	if conf.SocketPath != "/tmp/eventingestion.sock" {
		t.Fail()
	}
	region := kinesis.Region{"us-east-1"}
	if conf.AwsRegion != region {
		t.Fail()
	}
	if conf.SocketMode != 0 {
		t.Fail()
	}
	if conf.EventStream != "PrimaryEventStream" {
		t.Fail()
	}

	os.Setenv("AWS_ACCESS_KEY_ID", "AKIAFOOBAR")
	os.Setenv("AWS_SECRET_ACCESS_KEY", "foobarbaz")
	os.Setenv("AWS_REGION", "eu-west-1")
	os.Setenv("EVENT_STREAM", "StagingStream")
	os.Setenv("SOCKET_PATH", "/tmp/foobar")
	os.Setenv("SOCKET_MODE", "0777")
	defer os.Clearenv()

	conf = getConfigurationFromEnv(func() {})
	if conf.AwsAccessKeyId != "AKIAFOOBAR" {
		t.Fail()
	}
	if conf.AwsSecretAccessKey != "foobarbaz" {
		t.Fail()
	}
	region = kinesis.Region{"eu-west-1"}
	if conf.AwsRegion != region {
		t.Fail()
	}
	if conf.EventStream != "StagingStream" {
		t.Fail()
	}
	if conf.SocketPath != "/tmp/foobar" {
		t.Fail()
	}
	socket, _ := strconv.ParseUint("0777", 0, 32)
	if socket != conf.SocketMode {
		t.Fail()
	}

}
