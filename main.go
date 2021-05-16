package main

import (
	"bufio"
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strings"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh/terminal"

	"miwifi-termui/app"
)

const logFile = "miwifi.out.log"

var version = "n/a"

func main() {
	var (
		versionFlag  = flag.Bool("version", false, "application version")
		debugFlag    = flag.Bool("debug", false, "run application in debug mode")
		hostFlag     = flag.String("host", "", "MiWiFi host address")
		usernameFlag = flag.String("username", "admin", "username for login")
		passwordFlag = flag.String("password", "", "password for login")
		intervalFlag = flag.Duration("interval", time.Second*10, "fetch data interval")
		uiFlag       = flag.String("ui", "dash", `ui controller {"dash", "cpu", "dev", "info", "mem", "net"}`)
	)

	flag.Parse()

	if *versionFlag {
		fmt.Println(version)
		os.Exit(0)
	}

	var prompt bool

	if *hostFlag == "" {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Enter host: ")
		host, err := reader.ReadString('\n')
		if err != nil {
			panic(err)
		}
		*hostFlag = strings.TrimSpace(host)
		prompt = true
	}

	if !strings.HasPrefix(*hostFlag, "http://") {
		*hostFlag = "http://" + *hostFlag
	}

	if *passwordFlag == "" {
		fmt.Print("Enter admin password: ")
		bytePassword, err := terminal.ReadPassword(syscall.Stdin)
		if err != nil {
			panic(err)
		}
		*passwordFlag = strings.TrimSpace(string(bytePassword))
		prompt = true
	}

	if prompt {
		fmt.Println()
	}

	logger := log.New()
	logger.Out = ioutil.Discard

	if *debugFlag {
		logger.SetLevel(log.DebugLevel)

		file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
		if err != nil {
			panic(errors.New("failed to log to file"))
		}
		logger.Out = file
	}

	a := app.New(getMacAddr(), *hostFlag, *usernameFlag, *passwordFlag, *intervalFlag, logger)
	os.Exit(a.Run(*uiFlag))
}

func getMacAddr() (addr string) {
	interfaces, err := net.Interfaces()
	if err == nil {
		for _, i := range interfaces {
			if i.Flags&net.FlagUp != 0 && !bytes.Equal(i.HardwareAddr, nil) {
				addr = i.HardwareAddr.String()
				break
			}
		}
	}
	return
}
