package go_fritzbox_api

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// PyAdapter is a struct that controls the Python Adapter to Selenium.
// The Python Implementation of Selenium Wire is used to fetch the Arguments that are required for specific Requests.
//
// Starting a Python-Adapter will result in a Chromedriver being started in the Background and a Session for the Fritz!Box being created.
// This Session is automatically refreshed periodically, unless RefreshSession is set to false.
// For more Information, see the Readme (todo)
// todo: add the option to give arguments to the chromedriver options on initialize
type PyAdapter struct {
	Client  *Client
	Debug   bool
	adapter *exec.Cmd
	running bool
	writer  *bufio.Writer
	reader  *bufio.Reader
}

const (
	Timeout             = 2 * time.Second
	RefreshSession      = true
	RefreshSessionDelay = 17 * time.Minute
)

// StartAdapter starts the Python Script, logs in and starts the sessionRefresher if RefreshSession is set to true.
func (pya *PyAdapter) StartAdapter() error {
	err := pya.openAdapter()
	if err != nil {
		return err
	}

	if pya.Debug {
		err = pya.setDebug()
		if err != nil {
			return err
		}
	}

	// Login
	err = pya.login()
	if err != nil {
		return err
	}

	if RefreshSession {
		go pya.sessionRefresher()
	}

	return nil
}

func (pya *PyAdapter) GetArgsHKR(device Hkr) (params map[string]string, err error) {
	if err = pya.Client.checkExpiry(); err != nil {
		return
	}

	// Send Request
	err = write(pya.writer, fmt.Sprintf("HKR %s %s", pya.Client.BaseUrl, device.device.ID))
	if err != nil {
		return
	}

	result, err := readSuccess(pya.reader, time.Second*10)
	params = make(map[string]string)
	err = json.Unmarshal([]byte(result), &params)
	return
}

func (pya *PyAdapter) sessionRefresher() {
	for pya.running {
		time.Sleep(RefreshSessionDelay)
		err := pya.login()
		if err != nil {
			fmt.Println(err)
		}
	}
}

func (pya *PyAdapter) StopPyAdapter() error {
	pya.running = false
	return pya.adapter.Process.Kill()
}

func (pya *PyAdapter) openAdapter() (err error) {
	pya.adapter = exec.Command("python", "./pyAdapter/main.py")

	// Get the Pipes
	out, err := pya.adapter.StdoutPipe()
	if err != nil {
		return
	}
	in, err := pya.adapter.StdinPipe()
	if err != nil {
		return
	}

	// Start the Python Script
	err = pya.adapter.Start()
	if err != nil {
		return
	}

	// Create the Reader and Writer
	pya.writer = bufio.NewWriter(in)
	pya.reader = bufio.NewReader(out)

	// Wait for the Script to start
	err = expect(pya.reader, "HELO")
	if err != nil {
		return
	}

	// Send ok
	err = write(pya.writer, "OK")
	if err != nil {
		return
	}

	return nil
}

func (pya *PyAdapter) login() (err error) {
	if err = pya.Client.checkExpiry(); err != nil {
		return
	}

	// Send Login
	err = write(pya.writer, fmt.Sprintf("LOGIN %s %s", pya.Client.BaseUrl, pya.Client.SID()))
	if err != nil {
		return
	}

	// Wait for Response
	err = expectWithTimeout(pya.reader, "OK", 30*time.Second)
	return
}

func (pya *PyAdapter) setDebug() (err error) {
	err = write(pya.writer, "DEBUG")
	if err != nil {
		return
	}

	err = expect(pya.reader, "OK")
	return
}

func readSuccess(reader *bufio.Reader, timeout time.Duration) (out string, err error) {
	out, err = readWithTimeout(reader, timeout)
	if err != nil {
		return
	}

	if !strings.HasPrefix(out, "SUCCESS") {
		err = fmt.Errorf("expected \"SUCCESS\", got \"%s\"", out)
		return
	}

	return strings.TrimPrefix(out, "SUCCESS"), nil
}

func expect(reader *bufio.Reader, expected string) (err error) {
	return expectWithTimeout(reader, expected, Timeout)
}

func expectWithTimeout(reader *bufio.Reader, expected string, timeout time.Duration) (err error) {
	_, err = expectHelper(reader, expected, timeout)
	return
}

func expectHelper(reader *bufio.Reader, expected string, timeout time.Duration) (out string, err error) {
	output, err := readWithTimeout(reader, timeout)
	if err != nil {
		err = fmt.Errorf("%w %w", err, fmt.Errorf("expected %s", expected))
		return
	}

	if output != expected {
		err = fmt.Errorf("expected \"%s\", got \"%s\"", expected, output)
		return
	}

	return output, nil
}

func write(writer *bufio.Writer, msg string) error {
	_, err := writer.WriteString(msg + "\n")
	if err != nil {
		return err
	}
	err = writer.Flush()
	if err != nil {
		return err
	}
	return nil
}

func read(reader *bufio.Reader) (string, error) {
	return readWithTimeout(reader, Timeout)
}

func readWithTimeout(reader *bufio.Reader, timeout time.Duration) (string, error) {
	c1 := make(chan string, 1)
	go func() {
		output, _ := reader.ReadString('\n')
		c1 <- strings.Trim(output, "\r\n")
	}()

	select {
	case res := <-c1:
		return res, nil
	case <-time.After(timeout):
		return "", fmt.Errorf("timed out waiting for response")
	}
}
