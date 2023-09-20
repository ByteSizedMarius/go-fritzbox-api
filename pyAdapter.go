package go_fritzbox_api

import (
	"bufio"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"
)

//go:embed pyAdapter/main.py
var adapterEmbed embed.FS

// PyAdapter is a struct that controls the Python Adapter to Selenium.
// The Python Implementation of Selenium Wire is used to fetch the Arguments that are required for specific Requests.
//
// Starting a Python-Adapter will result in a Chromedriver being started in the Background and a Session for the Fritz!Box being created.
// This Session is automatically refreshed periodically, unless RefreshSession is set to false.
// For more Information, see the Readme (todo)
type PyAdapter struct {
	Client *Client
	Debug  bool
	// DriverArgs are the arguments that are passed to the chromedriver using options.add_argument().
	DriverArgs []string
	pyaClient  *Client
	adapter    *exec.Cmd
	running    bool
	writer     *bufio.Writer
	reader     *bufio.Reader
}

const (
	RefreshSession      = true
	Timeout             = 2 * time.Second
	RefreshSessionDelay = 17 * time.Minute
)

// StartAdapter starts the Python Script, logs in and starts the sessionRefresher if RefreshSession is set to true.
func (pya *PyAdapter) StartAdapter() error {

	// SID is invalidated after is us used with the webdriver
	// Thus, the Webdriver needs its own client
	pya.pyaClient = pya.Client.Copy()
	err := pya.pyaClient.Initialize()
	if err != nil {
		return err
	}

	err = pya.openAdapter()
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

	if err = pya.pyaClient.checkExpiry(); err != nil {
		return
	}

	// Send Request
	err = write(pya.writer, fmt.Sprintf("HKR %s %s", pya.pyaClient.BaseUrl, device.device.ID))
	if err != nil {
		return
	}

	result, err := readSuccess(pya.reader, time.Second*10)
	if err != nil {
		return
	}

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
	// if we cannot find the adapter locally, try to take it from embed
	var fileContent []byte
	fileContent, err = adapterEmbed.ReadFile("pyAdapter/main.py")
	if err != nil {
		return
	}

	adapter := path.Join(os.TempDir(), "pyAdapter.py")
	if err = os.WriteFile(adapter, fileContent, 0666); err != nil {
		return
	}

	pya.adapter = exec.Command("python", adapter)

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

	// Handle Chromedriver Arguments
	var input string
	if pya.DriverArgs != nil && len(pya.DriverArgs) > 0 {
		input = strings.Join(pya.DriverArgs, "|")
		err = write(pya.writer, fmt.Sprintf("ARGS %s", input))
		if err != nil {
			return
		}

		err = expect(pya.reader, "OK")
		if err != nil {
			return
		}

	}

	return nil
}

func (pya *PyAdapter) login() (err error) {
	if err = pya.pyaClient.checkExpiry(); err != nil {
		return
	}

	// Send Login
	err = write(pya.writer, fmt.Sprintf("LOGIN %s %s", pya.pyaClient.BaseUrl, pya.pyaClient.SID()))
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
	_, err = expectHelper(reader, expected, timeout, false)
	return
}

func expectHelper(reader *bufio.Reader, expected string, timeout time.Duration, keepReading bool) (out string, err error) {
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
		output, err := reader.ReadString('\n')
		if errors.Is(err, io.EOF) {
			c1 <- ""
		}
		c1 <- strings.Trim(output, "\r\n")
	}()

	select {
	case res := <-c1:
		return res, nil
	case <-time.After(timeout):
		return "", fmt.Errorf("timed out waiting for response")
	}
}
