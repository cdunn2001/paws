package web

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"syscall"
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

type ControlledProcess struct {
	cmd          *exec.Cmd
	status       *ProcessStatusObject
	temp_dn      string // someone must delete
	stdout_fn    string // someone must delete (unless under temp_dn)
	stderr_fn    string // someone must delete (unless under temp_dn)
	chanKill     chan bool
	chanComplete chan bool
}

func StartControlledBashProcess(bash string, ps *ProcessStatusObject, stall string) (result *ControlledProcess) {
	env := DummyEnv(stall)
	result, err := WatchBash(bash, ps, env)
	//result, err := WatchBash(bash, ps, nil)
	if err != nil {
		panic(err) // TODO: Check if panics are working.
	}
	pid := result.cmd.Process.Pid
	fmt.Printf("New pid:%d\n", pid)
	/*
		select {
		case <-result.chanComplete:
		}
		fmt.Println("Did we complete the process?")
	*/
	return result
}

func (cp *ControlledProcess) cleanup() {
	if strings.HasSuffix(cp.temp_dn, ".tempdir") {
		fmt.Printf("DEBUG removing dir tree '%s'", cp.temp_dn)
		err := os.RemoveAll(cp.temp_dn)
		check(err)
	} else {
		fmt.Printf("DEBUG not removing unknown dir tree '%s'", cp.temp_dn)
	}
}

// TODO: Identify this process in the log.
func logContent(fn, intro string) {
	content, err := ioutil.ReadFile(fn)
	if err != nil {
		fmt.Printf("ERROR: Could not find '%s' file '%s'\n", intro, fn)
	} else {
		fmt.Printf("INFO %s:\n%s\n", intro, content)
	}
}
func WriteStringToFile(content string, fn string) {
	f, err := os.Create(fn)
	check(err)
	defer f.Close()
	_, err = f.WriteString(content)
	check(err)
}
func WatchBash(bash string, ps *ProcessStatusObject, envExtra []string) (*ControlledProcess, error) {
	rpipe, wpipe, err := os.Pipe()
	if err != nil {
		return nil, err
	}
	extraFiles := []*os.File{wpipe} // becomes fd 3 in child

	{
		fmt.Println("PATH:", os.Getenv("PATH"))
		log.Println("PATH:", os.Getenv("PATH"))
		out, err := exec.Command("which", "dummy-pa-cal.sh").Output()
		if err != nil {
			fmt.Printf("%s\n", err)
		}
		fmt.Println("Command Successfully Executed")
		output := string(out[:])
		fmt.Println(output)
	}

	temp_dn, err := ioutil.TempDir("", "WatchBash.*.tmpdir")
	if err != nil {
		temp_dn = "very.tempdir"
		fmt.Printf("ERROR Weirdly unable to create TempDir(): %#v\nTrying '%s' instead.\n",
			err, temp_dn)
		err = os.MkdirAll(temp_dn, 0777)
		check(err)
	}
	stdout_fn := filepath.Join(temp_dn, "stdout.txt")
	stderr_fn := filepath.Join(temp_dn, "stderr.txt")
	bash = bash + " >" + stdout_fn + " 2> " + stderr_fn
	fmt.Println("bash:", bash)
	//fn := "/home/UNIXHOME/cdunn/repo/bb/paws/tmp/run.sh"
	//WriteStringToFile(bash, fn)
	//cmd := exec.Command("/bin/bash", fn)
	env := os.Environ()
	env = append(env, envExtra...)

	cmd := exec.Command("/bin/bash", "-c", bash)
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
		temp_dn:      temp_dn,
		stdout_fn:    stdout_fn,
		stderr_fn:    stderr_fn,
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
		logContent(cbp.stdout_fn, "stdout")
		logContent(cbp.stderr_fn, "stderr")
		defer cbp.cleanup()
		chanComplete <- true
	}()
	go func() {
		fmt.Println("Started scanner go-func")
		breader := bufio.NewReader(rpipe)
		defer rpipe.Close()
		var err error = nil
		//time.Sleep(1.0 * time.Second)
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
					fmt.Printf("End of file. Done reading status-reports. isPrefix:%t, line:'%s'\n", isPrefix, line)
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

		pid := cbp.cmd.Process.Pid
		process, err := os.FindProcess(int(pid))
		if err != nil {
			fmt.Printf("Failed to find process: %s\n", err)
		} else {
			err := process.Signal(syscall.Signal(0))
			fmt.Printf("process.Signal on pid %d returned: %v\n", pid, err)
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
