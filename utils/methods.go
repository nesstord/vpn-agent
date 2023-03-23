package utils

import (
	"bytes"
	"fmt"
	"math/rand"
	"net"
	"os/exec"
	"time"
)

func GetFreePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer l.Close()

	return l.Addr().(*net.TCPAddr).Port, nil
}

func GeneratePassword(length int) string {
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, length)
	rand.Read(b)
	return fmt.Sprintf("%x", b)[:length]
}

func ExecBash(command string) error {
	errors := bytes.Buffer{}
	cmd := exec.Command("/bin/bash", "-c", command)
	cmd.Stderr = &errors

	output, err := cmd.Output()

	fmt.Println(string(output)) //TODO: wrap for debug mode

	if err != nil {
		return fmt.Errorf("exec bash command error: %s, %s", errors.String(), err.Error())
	}

	return nil
}
