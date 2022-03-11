package web

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"log"
	"os"
	"os/exec"
	"regexp"
	"time"
)

var (
	json_regex = regexp.MustCompile(`_STATUS\s*(.*)$`)
)

func plus1(x int) (result int) {
	return 1 + x
}
func check(e error) {
	if e != nil {
		panic(fmt.Sprintf("Stacktrace: %+v", e))
	}
}

type StatusReport struct {
	normal      NormalStatusReport
	exceptional ExceptionalStatusReport
}
type NormalStatusReport struct {
	State            string  // e.g. "progress"
	Ready            bool    // turns to true when ICS can resume execution for providing sensor data
	StageNumber      int32   // number starting at 0 representing the stage
	StageName        string  // human readable description of the stage
	Counter          uint64  // a counter that monotonically increments as progress is made. example might be frames or ZMWs
	CounterMax       uint64  // the maximum number that the counter is expected to attain when done. If this is not known, then it should be omitted.
	TimeToNextStatus float64 // maximum time in seconds until the next status update. If there is no status message within the alloted time, pa-ws should kill the process
	StageWeights     []int32 // weights for each stage. Does not need to be normalized to any number
	Timestamp        string  // ISO8601 time stamp with millisecond precision (see PacBio::Utilities::ISO8601)
}
type ExceptionalStatusReport struct {
	State     string // "exception"
	Message   string
	Timestamp string // ISO8601
}

func Json2StatusReport(raw []byte) (result StatusReport) {
	err := json.Unmarshal(raw, &result.normal)
	if err != nil {
		err = json.Unmarshal(raw, &result.exceptional)
		check(err)
	}
	return result
}
func String2StatusReport(text string) (result StatusReport, err error) {
	found := json_regex.FindStringSubmatch(text)
	if found == nil {
		return result, errors.Errorf("Could not parse JSON from '%s'", text)
	}
	json_raw := []byte(found[1])
	result = Json2StatusReport(json_raw)
	return result, nil
}

func WatchBash(bash string, env []string) error {
	rpipe, wpipe, err := os.Pipe()
	defer rpipe.Close()
	if err != nil {
		return err
	}
	extraFiles := []*os.File{wpipe} // becomes fd 3 in child
	fmt.Println("bash:", bash)
	cmd := exec.Command("bash", "-c", bash)
	cmd.Env = env
	cmd.ExtraFiles = extraFiles
	cmd.Start()
	mustClose := func(f *os.File) {
		err = f.Close()
		check(err)
	}
	// immediately close our dup'ed fds (write end of our signal pipe)
	for _, f := range extraFiles {
		mustClose(f)
	}
	timeToNextStatusChan := make(chan float64)
	doneChan := make(chan bool)
	completeChan := make(chan bool)
	go func() {
		fmt.Println("Started timeout go-func")
		var timeout float64 = 2.0
		timedout := false
		done := false
		for !done && !timedout {
			select {
			case <-time.After(time.Duration((timeout + 0.01) * float64(time.Second))):
				timedout = true
				fmt.Println("Timed out!")
			case <-doneChan:
				done = true
				fmt.Println("Go done chan!")
			case timeout = <-timeToNextStatusChan:
				fmt.Println("timeout is now", timeout)
			}
		}
		if timedout {
			fmt.Println("Timedout?")
		} else if done {
			fmt.Println("Done?")
		} else {
			fmt.Println("Not sure!!!")
		}
		completeChan <- true
	}()
	go func() {
		fmt.Println("Started scanner go-func")
		breader := bufio.NewReader(rpipe)
		var err error = nil
		for err == nil {
			text := ""
			line := []byte{}
			isPrefix := true
			for isPrefix && err != io.EOF {
				line, isPrefix, err = breader.ReadLine()
				timeToNextStatusChan <- 0.1
				if err != nil && err != io.EOF {
					check(err)
				}
				text = text + string(line)
			}
			fmt.Println("Got:", text)
			if err == io.EOF {
				break
			}
			sr, err := String2StatusReport(text)
			if err != nil {
				// TODO: Log unexpected line?
				continue
			}
			timeToNextStatusChan <- sr.normal.TimeToNextStatus
		}
		doneChan <- true
	}()
	select {
	case <-completeChan:
	}
	return nil
}
func RunBash(bash string, env []string) error {
	cmd := exec.Command("bash", "-c", bash)
	cmd.Env = env
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatal(err)
		return errors.New("My bad") // unreachable
	}
	log.Printf("Result of exec:\n%s\n", out)
	return nil
}