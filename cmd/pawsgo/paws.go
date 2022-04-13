package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"log" // log.Fatal()
	"net/http"
	"path/filepath"
	"sync/atomic"
	//"net/http/httputil"
	"os"
	"strconv"
	"strings"
	"time"
	// "pacb.com/seq/paws/pkg/stuff"
	// "pacb.com/seq/paws/pkg/stiff"
	//"github.com/gofiber/fiber/v2"
	//_ "github.com/gofiber/fiber/v2/middleware/recover" // to trap panics
	//"github.com/gofiber/fiber/v2/utils"
	//"github.com/gofiber/template/html"
	"github.com/coreos/go-systemd/v22/daemon"
	"github.com/gin-gonic/gin"
	"pacb.com/seq/paws/pkg/config"
	"pacb.com/seq/paws/pkg/web"
	"runtime" // only for GOOS
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func PanicHandleRecovery(c *gin.Context, err interface{}) {
	//c.AbortWithStatus(http.StatusInternalServerError)
	msg := fmt.Sprintf("Panic:'%+v'\n", err)
	c.String(http.StatusInternalServerError, msg)
}
func listen(port int, lw io.Writer) {
	//router := gin.Default()
	// Or explicitly:
	router := gin.New()
	router.SetTrustedProxies(nil) // https://pkg.go.dev/github.com/gin-gonic/gin#readme-don-t-trust-all-proxies
	gin.DefaultWriter = lw
	//gin.ForceConsoleColor() // needed for colors w/ MultiWriter
	router.Use(
		gin.Logger(),
		//gin.LoggerWithWriter(gin.DefaultWriter, "/pathsNotToLog/"), // useful!
		gin.CustomRecovery(PanicHandleRecovery),
		//gin.Recovery(),
		//gin.RecoveryWithWriter(log.Writer())
	)

	router.GET("/hello", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Hello World!",
		})
	})

	router.GET("/os", func(c *gin.Context) {
		c.String(200, runtime.GOOS)
	})

	web.AddRoutes(router)
	ns := os.Getenv("NOTIFY_SOCKET")
	wusec := os.Getenv("WATCHDOG_USEC")
	wpid := os.Getenv("WATCHDOG_PID")
	log.Printf("NOTIFY_SOCKET='%s', WATCHDOG_USEC='%s', WATCHDOG_PID=%s\n", ns, wusec, wpid)
	//fmt.Printf("stdout wrote to '%s'\n", lfn)
	//fmt.Fprintf(os.Stderr, "stderr wrote to '%s'\n", lfn)
	if wusec != "" {
		if wpid == "" {
			// We must be testing.
			wpid = strconv.Itoa(os.Getpid())
			log.Printf("Fake WATCHDOG_PID='%s'\n", wpid)
			err := os.Setenv("WATCHDOG_PID", wpid)
			check(err)
		}
		pid, err := strconv.Atoi(wpid)
		check(err)
		usec, err := strconv.Atoi(wusec)
		check(err)
		usec = usec
		interval := time.Duration(usec) * time.Microsecond
		log.Printf("usec='%d', pid='%d', interval='%s'\n", usec, pid, interval)
		if os.Getpid() != pid {
			log.Printf("Wrong pid! '%s'\n", wpid)
			os.Exit(1)
		}
		timeout, err := daemon.SdWatchdogEnabled(false)
		check(err)
		if timeout != interval {
			log.Printf("ERROR: timeout(%s) != our calc(%s)", timeout, interval)
		}
		delay := timeout / 2
		log.Printf("systemd timeout=%s, paws heartbeat%s'\n", timeout, delay)
		timer := time.NewTicker(delay)
		defer timer.Stop()
		log.Printf("Created Ticker w/ arg='%s'\n", delay)

		doneWatchdogCh := make(chan bool)
		defer close(doneWatchdogCh) // closing is as good as sending "true"

		var count int32 = 1

		go func() {
			log.Print("Watchdog gofunc started. Waiting on ticker/done channels...\n")
			for {
				select {
				case <-doneWatchdogCh:
					log.Print("Watchdog done!\n")
					return
				case <-timer.C:
					msg := web.NotifyWatchdog()
					if web.IsPowerOf2(count) {
						log.Printf("Watchdog timer fired. count=%06d -- %s\n", count, msg)
					}
					atomic.AddInt32(&count, 1)
				}
			}
			log.Print("Watchdog gofunc ends.\n")
		}()
		/*
			msg := ""
			msg = fmt.Sprintf("Wait for %s delay.\n", delay)
			log.Print(msg)
			time.Sleep(2 * time.Second)
			doneWatchdogCh <- true
			msg = "Send done <- true\n"
			log.Print(msg)
		*/
	}

	{
		hostname := config.Top().Hostname
		status := web.GetPawsStatusObject()
		marsh, err := json.MarshalIndent(status, "", "  ")
		check(err)
		log.Printf("Status (w/ paths on %s, not necessarily on NRT):\n%s\n",
			hostname, marsh)
	}

	portStr := fmt.Sprintf(":%d", port)
	log.Fatal(router.Run(portStr)) // logger maybe not needed, but does not seem to hurt
}

// If a file of this name exists, then move it to something that does not.
// If this is actually a symlink, remove the symlink.
func MoveExistingLogfile(specified string) {
	fi, err := os.Lstat(specified)
	if err == nil {
		if fi.Mode()&os.ModeSymlink == os.ModeSymlink {
			// This is a symlink, so remove just the symlink.
			err := os.Remove(specified)
			if err != nil {
				fmt.Printf("FATAL: Could not remove symlink of logfile %q: %+v\n",
					specified, err)
				check(err)
			}
		} else {
			// Not a symlink. Must have been created by older version of paws.
			// Choose a new name and move this file to it.
			newname := web.ChooseLoggerFilenameLegacy(specified)
			err := os.Rename(specified, newname)
			if err != nil {
				fmt.Printf("ERROR: Could not rename logfile from %q to %q: %+v\nLost old logfile.\n",
					specified, newname, err)
			}
		}
	} else if errors.Is(err, fs.ErrNotExist) {
		// No problem.
	} else {
		fmt.Printf("FATAL: Unexpected error testing logfile %q: %+v\n",
			specified, err)
		check(err)
	}
}
func main() {
	versionPtr := flag.Bool("version", false, "Print version")
	consolePtr := flag.Bool("console", false, "Log to console. (Ignore --logoutput if any.)")
	portPtr := flag.Int("port", 23632, "Listen on this port.")
	cfgPtr := flag.String("config", "", "Read paws config (JSON) from this file, to update default config.")
	lfnPtr := flag.String("logoutput", "/var/log/pacbio/pa-wsgo/pa-wsgo.log", "Logfile output. We actually choose a unique name (maybe based on timestamp and pid), and (sym?)link the named path to it. We avoid over-writing the pre-existing named path.")
	dataDirPtr := flag.String("data-dir", "", "Directory for some outputs (usually under SRA subdir")
	flag.Parse()
	//flag.PrintDefaults()

	if *versionPtr {
		fmt.Println(config.Version)
		os.Exit(0)
	}

	var lw io.Writer
	if !*consolePtr {
		MoveExistingLogfile(*lfnPtr)
		fn := web.ChooseLoggerFilename(*lfnPtr)
		{
			err := os.Symlink(filepath.Base(fn), *lfnPtr)
			if err != nil {
				fmt.Printf("ERROR: Failed to create convenient symlink from %q to %q: %+v\nContinuing.",
					fn, *lfnPtr, err)
			}
		}
		fmt.Printf("Logging to '%s'\n", fn)
		f, err := os.Create(fn)
		check(err)
		defer f.Close()
		lw = f
	} else {
		lw = os.Stdout
		//lw = io.MultiWriter(f, os.Stdout)
	}
	log.SetOutput(lw)
	log.Println(strings.Join(os.Args[:], " "))
	log.Printf("version=%s\n", config.Version)
	log.Printf("port='%v'\n", *portPtr)

	web.InitFixtures()

	if *cfgPtr != "" {
		log.Printf("config='%v'\n", *cfgPtr)
		//config.UpdateConfigFromFile(*cfgPtr, &ppaConfig)
	}

	if *dataDirPtr == "" {
		var err error
		*dataDirPtr, err = ioutil.TempDir("", "pawsgo.*.datadir")
		check(err)
		//defer os.RemoveAll(*dataDirPtr)
	}
	web.DataDir = *dataDirPtr
	log.Printf("DataDir='%s'\n", web.DataDir)

	log.Printf("tc: %+v", config.Top())
	//WriteConfig(config.Top(), "foo.paws.json")

	config.VerifyBinaries(config.Top().Binaries)
	//web.CheckBaz2bam(config.Top())

	listen(*portPtr, lw)
}
