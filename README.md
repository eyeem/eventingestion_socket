# eventingestion_socket
## Overview

Provides a file descriptor at `SOCKET_PATH` with permissions `SOCKET_MODE` that will allow arbitrary input, split by newlines (`\n`) to be pushed to Kinesis `EVENT_STREAM` in AWS region `AWS_REGION` using credentials available in `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY`
```
Application                                       ->  eventingestion_socket

$foo = Socket_Client('unix:///var/run/foobar')        read('unix:///var/run/foobar')
write($foo,'bar\nbaz')                                kinesis.PutRecords('bar')
                                                      kinesis.PutRecords('baz')
```

The goal of this project is, to offload the CPU time needed to HTTP PutRecord into Kinesis away from the application itself and thus not slowing the application down and hiding the concurrent processing behind a simple fast UNIX filesocket.

To build and run the project, git clone into your GOPATH like $GOPATH/src/github.com/LarsFronius/eventingestion_socket.
Dependencies are locked using gom, so go get github.com/mattn/gom and gom install && gom build should build the package.

To run the socket, the following environment variables can be used to configure the socket:
```
export AWS_ACCESS_KEY_ID=
export AWS_SECRET_ACCESS_KEY=
export SOCKET_PATH=
export SOCKET_MODE=
export EVENT_STREAM=
export AWS_REGION=
```
Afterwards a ./eventingestion_socket should start the socket.

## Packaging

The packaging branch contains a debian folder that will always build the latest pushed version on github and create a debian package from it using `debuild`. This is tested on Ubuntu >= 14.04.
