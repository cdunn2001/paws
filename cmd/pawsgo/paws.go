package main

import (
	"flag"
	"fmt"
	"io"
	"log" // log.Fatal()
	"net/http"
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
	"pacb.com/seq/paws/pkg/web"
	"runtime" // only for GOOS
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

// Someday, keep panic message in response, maybe.
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
	log.Printf("NOTIFY_SOCKET='%s', WATCHDOG_USEC='%s'\n", ns, wusec)
	//fmt.Printf("stdout wrote to '%s'\n", lfn)
	//fmt.Fprintf(os.Stderr, "stderr wrote to '%s'\n", lfn)
	if wpid != "" {
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
		delay, err := daemon.SdWatchdogEnabled(false)
		check(err)
		delay = delay / 2
		log.Printf("For timer, using delay='%s'\n", delay.Round(time.Microsecond))
		timer2 := time.NewTicker(delay)
		defer timer2.Stop()
		log.Printf("Created Ticker w/ arg='%s'\n", delay)
		done := make(chan bool)
		go func() {
			log.Print("gofunc started. Watiing on ticker/done channels...\n")
			for {
				select {
				case <-done:
					log.Print("Done!\n")
					return
				case current := <-timer2.C:
					log.Printf("...Timer 2 fired! current='%s'\n", current)
					supported_and_sent, err := daemon.SdNotify(false, daemon.SdNotifyWatchdog)
					check(err)
					log.Printf("delay='%s', sent='%s'\n", delay.Round(time.Microsecond), supported_and_sent)
				}
			}
			log.Print("End of watchdog gofunc.\n")
		}()
		msg := ""
		msg = fmt.Sprintf("Wait for %s delay.\n", delay)
		log.Print(msg)
		time.Sleep(16 * time.Second)
		done <- true
		msg = "Send done <- true\n"
		log.Print(msg)
	}

	portStr := fmt.Sprintf(":%d", port)
	log.Fatal(router.Run(portStr)) // logger maybe not needed, but does not seem to hurt
}
func main() {
	portPtr := flag.Int("port", 23632, "Listen on this port.")
	cfgPtr := flag.String("config", "", "Read PpaConfig (JSON) from this file, to update default config.")
	lfnPtr := flag.String("logoutput","/var/log/pacbio/pa-wsgo/pa-wsgo.log","Logfile output")
	flag.Parse()
	//flag.PrintDefaults()

	//lfn := "/var/log/pacbio/pa-wsgo/pa-wsgo.log"
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
	ppaConfig := web.PpaConfig{}
	ppaConfig.SetDefaults()
	if *cfgPtr != "" {
		log.Printf("config='%v'\n", *cfgPtr)
		web.UpdatePpaConfigFromFile(*cfgPtr, &ppaConfig)
	}
	listen(*portPtr, lw)
}
