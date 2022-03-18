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
	"strconv"
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
	State            string  // e.g. "progress" or "exception"
	Ready            bool    // turns to true when ICS can resume execution for providing sensor data
	StageNumber      int32   // number starting at 0 representing the stage
	StageName        string  // human readable description of the stage
	Counter          uint64  // a counter that monotonically increments as progress is made. example might be frames or ZMWs
	CounterMax       uint64  // the maximum number that the counter is expected to attain when done. If this is not known, then it should be omitted.
	TimeToNextStatus float64 // maximum time in seconds until the next status update. If there is no status message within the alloted time, pa-ws should kill the process
	StageWeights     []int32 // weights for each stage. Does not need to be normalized to any number
	Timestamp        string  // ISO8601 time stamp with millisecond precision (see PacBio::Utilities::ISO8601)

	// Only for exceptions:
	Message string
}

func Json2StatusReport(raw []byte) (result StatusReport) {
	err := json.Unmarshal(raw, &result)
	if err != nil {
		// TODO: Ignore errors. Assume a heartbeat.
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

// This can be used to set the env for our dummy bash scripts, causing
// them to waste "stall" seconds (float).
// The result is suitable for 'env' arg of WatchBash().
func DummyEnv(stall string) (result []string) {
	secs, err := strconv.ParseFloat(stall, 32)
	if err != nil {
		return result
	}
	if secs == 0.0 {
		return result
	}
	delay := 0.1
	count := int(secs / delay)
	result = []string{
		fmt.Sprintf("STATUS_COUNT=%d", count),
		fmt.Sprintf("STATUS_DELAY_SECONDS=%f", delay),
	}
	fmt.Printf("DummyEnv:'%s'\n", result)
	return result
}

func WatchBash(bash string, ps *ProcessStatusObject, env []string) (*ControlledProcess, error) {
	rpipe, wpipe, err := os.Pipe()
	if err != nil {
		return nil, err
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
	chanStatusReportText := make(chan string)
	chanDone := make(chan bool)
	chanComplete := make(chan bool)
	chanKill := make(chan bool)

	cbp := &ControlledProcess{
		cmd:          cmd,
		status:       ps,
		chanKill:     chanKill,
		chanComplete: chanComplete,
	}
	cbp.status.ExecutionStatus = Running // TODO: Make this thread-safe!!!

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
				fmt.Printf("Killing pid %d\n", cmd.Process.Pid)
				cmd.Process.Kill() // TODO: What happens if not running? Also, check sub-children, or maybe ssid.
			case <-chanKill:
				fmt.Printf("Got chanKill.\n")
				fmt.Printf("Killing pid %d\n", cmd.Process.Pid)
				cmd.Process.Kill() // TODO: What happens if not running? Also, check sub-children, or maybe ssid.
				done = true
				fmt.Println("Called Kill()")
			case <-chanDone:
				done = true
				fmt.Println("Got chanDone!")
			case srText := <-chanStatusReportText:
				sr, err := String2StatusReport(srText)
				if err != nil {
					// TODO: Log unexpected line?
				} else if sr.State == "exception" {
					// TODO: Got an exception. Log and ignore?
				} else {
					timeout = sr.TimeToNextStatus
					fmt.Println("timeout is now", timeout)

					// TODO: Make this thread-safe!!!
					cbp.status.Timestamp = sr.Timestamp
				}
			}
		}
		if timedout {
			fmt.Println("Timedout?")
		} else if done {
			fmt.Println("Done?")
		} else {
			fmt.Println("Not sure!!!")
		}
		fmt.Printf("Now waiting for pid %d\n", cmd.Process.Pid)
		err := cmd.Wait()
		fmt.Printf("Done waiting for pid %d. Exit:%v\n", cmd.Process.Pid, err)

		// TODO: Make these thread-safe!!!
		cbp.status.ExecutionStatus = Complete
		cbp.status.ExitCode = int32(cmd.ProcessState.ExitCode())

		if !cmd.ProcessState.Exited() {
			// Must have been signaled.
			cbp.status.CompletionStatus = Aborted
		} else if !cmd.ProcessState.Success() {
			cbp.status.CompletionStatus = Failed
		} else {
			cbp.status.CompletionStatus = Success
		}
		chanComplete <- true
	}()
	go func() {
		fmt.Println("Started scanner go-func")
		breader := bufio.NewReader(rpipe)
		defer rpipe.Close()
		var err error = nil
		for err == nil {
			text := ""
			line := []byte{}
			isPrefix := true
			for isPrefix {
				line, isPrefix, err = breader.ReadLine()
				if err != nil {
					if err != io.EOF {
						// Unexpected error. // TODO: What else is ok here?
						fmt.Printf("Unexpected error from Readline():'%+v'\n", err)
						//check(err)
					}
					fmt.Println("End of file. Done reading status-reports.")
					break
				}
				text = text + string(line)
			}
			fmt.Println("Got:", text)
			if err == io.EOF {
				break
			}
			chanStatusReportText <- text
		}
		chanDone <- true
		//close(chanStatusReportText) // TODO? Somehow?
	}()

	return cbp, nil
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
