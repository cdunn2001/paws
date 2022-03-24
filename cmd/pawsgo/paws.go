package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log" // log.Fatal()
	"net/http"
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
	gin.ForceConsoleColor() // needed for colors w/ MultiWriter
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

	portStr := fmt.Sprintf(":%d", port)
	log.Fatal(router.Run(portStr)) // logger maybe not needed, but does not seem to hurt
}
func main() {
	portPtr := flag.Int("port", 23632, "Listen on this port.")
	cfgPtr := flag.String("config", "", "Read paws config (JSON) from this file, to update default config.")
	//cfgCommonPtr := flag.String("common-config", "", "Read PpaConfig (JSON) from this file, to update default config. (Not implemented.)")
	lfnPtr := flag.String("logoutput", "/var/log/pacbio/pa-wsgo/pa-wsgo.log", "Logfile output")
	dataDirPtr := flag.String("data-dir", "", "Directory for some outputs (usually under SRA subdir")
	flag.Parse()
	//flag.PrintDefaults()

	/*
		ppaConfig := web.PpaConfig{}
		ppaConfig.SetDefaults()
		if *cfgCommonPtr != "" {
			panic("--common-config not implemented")
		}
	*/

	lfn := *lfnPtr
	f, err := os.Create(lfn)
	check(err)
	defer f.Close()
	//lw := os.Stdout
	//lw := f
	lw := io.MultiWriter(f, os.Stdout)
	log.SetOutput(lw)
	log.Println(strings.Join(os.Args[:], " "))
	log.Printf("port='%v'\n", *portPtr)

	if *cfgPtr != "" {
		log.Printf("config='%v'\n", *cfgPtr)
		//web.UpdatePpaConfigFromFile(*cfgPtr, &ppaConfig)
	}

	if *dataDirPtr == "" {
		*dataDirPtr, err = ioutil.TempDir("", "pawsgo.*.datadir")
		check(err)
		//defer os.RemoveAll(*dataDirPtr)
	}
	web.DataDir = *dataDirPtr
	log.Printf("DataDir='%s'\n", web.DataDir)

	log.Printf("tc: %v", config.Top())
	//WriteConfig(config.Top(), "foo.paws.json")

	listen(*portPtr, lw)
}
