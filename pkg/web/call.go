package web

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"time"
)

func plus1(x int) (result int) {
	return 1 + x
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
			case <-time.After(time.Duration(timeout * float64(time.Second))):
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
