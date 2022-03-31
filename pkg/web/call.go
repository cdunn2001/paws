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
	"pacb.com/seq/paws/pkg/config"
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
	State                string  // e.g. "progress" or "exception"
	Ready                bool    // turns to true when ICS can resume execution for providing sensor data
	StageNumber          int32   // number starting at 0 representing the stage
	StageName            string  // human readable description of the stage
	Counter              uint64  // a counter that monotonically increments as progress is made. example might be frames or ZMWs
	CounterMax           uint64  // the maximum number that the counter is expected to attain when done. If this is not known, then it should be omitted.
	TimeoutForNextStatus float64 // maximum time in seconds until the next status update. If there is no status message within the alloted time, pa-ws should kill the process
	StageWeights         []int32 // weights for each stage. Does not need to be normalized to any number
	Timestamp            string  // ISO8601 time stamp with millisecond precision (see PacBio::Utilities::ISO8601)

	// Only for exceptions:
	Message string
}

func Json2StatusReport(raw []byte) (result StatusReport) {
	err := json.Unmarshal(raw, &result)
	if err != nil {
		log.Printf("ERROR: Could not unmarshal StatusReport '%v'", string(raw))
		check(err)
	}
	return result
}
func String2StatusReport(text string) (result StatusReport, err error) {
	found := json_regex.FindStringSubmatch(text)
	if found == nil {
		return result, errors.Errorf("Could not find start of JSON from '%s'", text)
	}
	json_raw := []byte(found[1])
	result = Json2StatusReport(json_raw)
	return result, nil
}
func ProgressMetricsObjectFromStatusReport(sr StatusReport) ProgressMetricsObject {
	var (
		wsum          float64 = 0.0
		stageProgress float64 = 0.0
		netProgress   float64 = 0.0
	)
	for _, w := range sr.StageWeights {
		wsum += float64(w)
	}
	if sr.CounterMax > 0 {
		stageProgress = float64(sr.Counter) / float64(sr.CounterMax)
	}
	for i, w := range sr.StageWeights {
		num := int32(i)
		if num < sr.StageNumber {
			netProgress += float64(w) * 1.0 / wsum
		} else if num == sr.StageNumber {
			netProgress += float64(w) * stageProgress / wsum
		}
	}
	result := ProgressMetricsObject{
		Counter:       sr.Counter,
		CounterMax:    sr.CounterMax,
		Ready:         sr.Ready,
		StageName:     sr.StageName,
		StageNumber:   sr.StageNumber,
		StageWeights:  sr.StageWeights,
		StageProgress: stageProgress,
		NetProgress:   netProgress,
	}
	return result
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
	log.Printf("DummyEnv:'%s'\n", result)
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
	if err != nil {
		panic(err)
	}
	pid := result.cmd.Process.Pid
	ps.PID = pid
	log.Printf("New pid:%d\n", pid)
	return result
}

func (cp *ControlledProcess) cleanup() {
	if strings.HasSuffix(cp.temp_dn, ".tmpdir") {
		log.Printf("DEBUG removing dir tree '%s'", cp.temp_dn)
		err := os.RemoveAll(cp.temp_dn)
		check(err)
	} else {
		log.Printf("DEBUG not removing unknown dir tree '%s'", cp.temp_dn)
	}
}

func logContent(fn, intro string) {
	content, err := ioutil.ReadFile(fn)
	if err != nil {
		log.Printf("ERROR: Could not find '%s' file '%s'\n", intro, fn)
	} else {
		log.Printf("INFO %s:\n%s\n", intro, content)
	}
}
func WriteStringToFile(content string, fn string) {
	f, err := os.Create(fn)
	check(err)
	defer f.Close()
	_, err = f.WriteString(content)
	check(err)
}

// date --utc +%Y%m%dT%TZ
//TIMESTAMP="20220223T146198.099Z" # arbitrary
//RFC3339     = "2006-01-02T15:04:05Z07:00"
func Timestamp(t time.Time) string {
	return t.UTC().Format(time.RFC3339)
}
func TimestampNow() string {
	return Timestamp(time.Now())
}

// Assume basename has no separators.
func SplitExt(basename string) (pre string, dotext string) {
	lastDot := strings.LastIndex(basename, ".")
	if lastDot == -1 {
		return basename, ""
	} else {
		return basename[:lastDot], basename[lastDot:]
	}
}
func ChooseLoggerFilename(existing string) string {
	return chooseLoggerFilenameTestable(existing, time.Now().UTC(), os.Getpid())
}
func chooseLoggerFilenameTestable(existing string, now time.Time, pid int) string {
	dir, oldbasename := filepath.Split(existing)
	oldpre, oldext := SplitExt(oldbasename)
	// canonical time for layout: "Jan 2 15:04:05 2006 MST"
	layout := "06-01-02"
	datetime := now.Format(layout)
	newbasename := fmt.Sprintf("%s.%s.%d%s",
		oldpre, datetime, pid, oldext)
	return filepath.Join(dir, newbasename)
}

// For a logfile generated by older paws, we want to move it to a new name.
// Do not use PID in that new name.
func ChooseLoggerFilenameLegacy(existing string) string {
	return chooseLoggerFilenameLegacyTestable(existing, GetCtime(existing))
}
func chooseLoggerFilenameLegacyTestable(existing string, ctime time.Time) string {
	dir, oldbasename := filepath.Split(existing)
	oldpre, oldext := SplitExt(oldbasename)
	// canonical time for layout: "Jan 2 15:04:05 2006 MST"
	layout := "06-01-02"
	datetime := ctime.Format(layout)
	newbasename := fmt.Sprintf("%s.%s%s",
		oldpre, datetime, oldext)
	return filepath.Join(dir, newbasename)
}
func GetCtime(fn string) (ctime time.Time) {
	fi, err := os.Stat(fn)
	if err != nil {
		fmt.Printf("FATAL: Cannot get ctime for file '%s': %+v\n", fn, err)
	}
	check(err)
	stat := fi.Sys().(*syscall.Stat_t)
	ctime = time.Unix(int64(stat.Ctim.Sec), int64(stat.Ctim.Nsec))
	return ctime
}
func FirstWord(sentence string) string {
	words := strings.Fields(sentence)
	if len(words) == 0 {
		return ""
	} else {
		return words[0]
	}
}
func WatchBash(bash string, ps *ProcessStatusObject, envExtra []string) (*ControlledProcess, error) {
	rpipe, wpipe, err := os.Pipe()
	if err != nil {
		return nil, err
	}
	extraFiles := []*os.File{wpipe} // becomes fd 3 in child

	{
		prog := FirstWord(bash)
		log.Printf("which %s: ", prog)
		out, err := exec.Command("which", prog).Output()
		if err != nil {
			log.Printf("%s\n", err)
			log.Println("PATH:", os.Getenv("PATH"))
		} else {
			log.Println(string(out))
		}
	}

	temp_dn, err := ioutil.TempDir("", "WatchBash.*.tmpdir")
	if err != nil {
		temp_dn = "very.tempdir"
		log.Printf("ERROR Weirdly unable to create TempDir(): %#v\nTrying '%s' instead.\n",
			err, temp_dn)
		err = os.MkdirAll(temp_dn, 0777)
		check(err)
	}
	stdout_fn := filepath.Join(temp_dn, "stdout.txt")
	stderr_fn := filepath.Join(temp_dn, "stderr.txt")
	bash = bash + " >" + stdout_fn + " 2> " + stderr_fn
	log.Println("bash:", bash)
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
	cbp.status.Armed = false             //     : ditto

	go func() {
		pid := int(cmd.Process.Pid)
		var timeout float64 = 10.0 // seconds
		log.Printf("PID: %d Started timeout go-func with timeout=%f\n", pid, timeout)
		timedout := false
		done := false
		for !done && !timedout {
			select {
			case <-time.After(time.Duration((timeout + 0.01) * float64(time.Second))):
				timedout = true
				log.Printf("PID: %d Timed out! Killing.\n", pid)
				cmd.Process.Kill() // TODO: What happens if not running? Also, check sub-children, or maybe ssid.
			case <-chanKill:
				log.Printf("PID: %d Got chanKill. Killing.\n", pid)
				cmd.Process.Kill() // TODO: What happens if not running? Also, check sub-children, or maybe ssid.
				done = true
			case <-chanDone:
				done = true
				log.Printf("PID: %d Got chanDone!\n", pid)
			case srText := <-chanStatusReportText:
				sr, err := String2StatusReport(srText)
				if err != nil {
					// Count as a heartbeat, but do not update timeout.
					log.Printf("PID: %d Unparseable status:\n%s\n", pid, srText)
				} else if sr.State == "exception" {
					// Count as a heartbeat, but do not update timeout.
					log.Printf("PID: %d Status exception:%s\n", pid, srText)

					// TODO: Make this thread-safe!!!
					//cbp.status.Timestamp = sr.Timestamp
					cbp.status.Timestamp = TimestampNow() // Much better if paws sets this.
					// Do not update ProgressMetricsObject on exceptions?
				} else {
					// Count as a heartbeat and update timeout.
					if sr.TimeoutForNextStatus > 0.0 {
						timeout = sr.TimeoutForNextStatus
						log.Printf("PID: %d timeout is now %f\n", pid, timeout)
					} else {
						log.Printf("PID: %d Ignoring TimeoutForNextStatus %f\n", pid, sr.TimeoutForNextStatus)
					}

					// TODO: Make this thread-safe!!!
					cbp.status.Timestamp = TimestampNow() // Much better if paws sets this.

					cbp.status.Armed = sr.Ready // Yes, "Ready" means something different here.
					cbp.status.Progress = ProgressMetricsObjectFromStatusReport(sr)
				}
			}
		}
		if timedout {
			log.Printf("PID: %d Timedout?\n", pid)
		} else if done {
			log.Printf("PID: %d Done?\n", pid)
		} else {
			log.Printf("PID: %d Not sure why we stopped watching it.\n", pid)
		}
		log.Printf("PID: %d Waiting...\n", pid)
		err := cmd.Wait()
		log.Printf("PID: %d Done waiting. Exit:%v\n", pid, err)

		// TODO: Make these thread-safe!!!
		cbp.status.ExecutionStatus = Complete
		cbp.status.Armed = false // TODO: Does it matter here?
		cbp.status.Timestamp = TimestampNow()
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
		pid := int(cmd.Process.Pid)
		log.Printf("PID: %d Started scanner go-func\n", pid)
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
						log.Printf("PID: %d Unexpected error from Readline():'%+v'\n", pid, err)
						//check(err)
					}
					log.Printf("PID: %d End of file. Done reading status-reports. isPrefix:%t, line:'%s'\n", pid, isPrefix, line)
					break
				}
				text = text + string(line)
			}
			log.Printf("PID: %d Got:%s\n", pid, text)
			if err == io.EOF {
				break
			}
			chanStatusReportText <- text
		}

		process, err := os.FindProcess(pid)
		if err != nil {
			log.Printf("PID: %d Failed FindProcess: %s\n", pid, err)
		} else {
			err := process.Signal(syscall.Signal(0))
			log.Printf("PID: %d process.Signal returned: %v\n", pid, err)
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
func VerifyBinary(label string, path string) {
	log.Printf("Verifying %s: '%s'\n", label, path)
	{
		log.Printf("which %s: \n", path)
		out, err := exec.Command("which", path).Output() // .CombinedOutput()?
		if err != nil {
			log.Printf("%s\n", err)
			log.Println("PATH:", os.Getenv("PATH"))
		} else {
			log.Println(string(out))
		}
	}
	{
		log.Printf("%s version:", path)
		out, err := exec.Command(path, "--version").Output()
		if err != nil {
			log.Printf("%s", err)
			check(err)
		} else {
			log.Println(string(out))
		}
	}
}
func VerifyBinaries(tc config.BinaryPaths) {
	log.Printf("Verifying binaries.\n")
	VerifyBinary("Binary_baz2bam", tc.Binary_baz2bam)
	VerifyBinary("Binary_pa_cal", tc.Binary_pa_cal)
	VerifyBinary("Binary_smrt_basecaller", tc.Binary_smrt_basecaller)
	//VerifyBinary("Binary_reducestats", tc.Binary_reducestats)
}
