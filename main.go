package main

import (
    "fmt"
    "net"
    "os"
    "os/signal"
    "syscall"
    "bufio"
    "errors"
    "strconv"
    kinesis "github.com/sendgridlabs/go-kinesis"
)

func getConfiguration(variable string) (conf string, err error) {
    if (len(os.Getenv(variable)) > 0) {
        conf = os.Getenv(variable)
    } else {
        return "", errors.New("Environment variable " + variable + " not found\n")
    }
    return conf, nil
}

func sendToKinesis(data string) {
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
    ksis := kinesis.New(aws_access_key_id, aws_secret_access_key)
    args := kinesis.NewArgs()
    args.Add("StreamName", "PrimaryEventStream")
    args.AddData([]byte(data))
    args.Add("PartitionKey", "123")
    resp4, err := ksis.PutRecord(args)
    if err != nil {
        fmt.Printf("PutRecord err: %v\n", err)
    } else {
        fmt.Printf("PutRecord: %v\n", resp4)
    }
}

func main() {
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
            fmt.Printf(err.Error() + "\nSOCKET_MODE is ot set in your environment variables, will leave it readable for $USER\n")
        } else {
            perm, err := strconv.ParseUint(permission, 0, 32)
            if err != nil {
                fmt.Printf(err.Error() + "\nSOCKET_MODE is not readable, will leave it readable for $USER\n")
            } else {
                fmt.Printf("Changing socket permissions to " + permission + "\n")
                os.Chmod(socket, os.FileMode(perm))
            }
        }
    }
    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt)
    signal.Notify(c, syscall.SIGTERM)
    go func() {
        <-c
        l.Close()
        os.Exit(0)
    }()

    for {
        fd, err := l.Accept()
        if err != nil {
            fmt.Printf("Socket listen error " + err.Error())
            return
        }
        for {
            data, err := bufio.NewReader(fd).ReadString('\n')
            if err != nil {
                fmt.Printf("BufReader error " + err.Error())
                return
            }
            go sendToKinesis(string(data))
        }

    }

}
