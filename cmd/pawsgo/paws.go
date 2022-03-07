package main

import (
	"fmt"
	"log" // log.Fatal()
	"os"
	"strconv"
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
func main() {
	//router := gin.Default()
	// Or explicitly:
	router := gin.New()
	router.SetTrustedProxies(nil) // https://pkg.go.dev/github.com/gin-gonic/gin#readme-don-t-trust-all-proxies
	router.Use(
		//gin.Logger(),
		gin.LoggerWithWriter(gin.DefaultWriter, "/pathsNotToLog/"), // useful!
		//gin.Recovery(),
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
	//lfn := "/var/log/pacbio/pa-wsgo/pa-wsgo.log"
	lfn := "/tmp/pa-wsgo.log"
	f, err := os.Create(lfn)
	check(err)
	f.WriteString("CDUNN WAS HERE\n")
	ns := os.Getenv("NOTIFY_SOCKET")
	wusec := os.Getenv("WATCHDOG_USEC")
	wpid := os.Getenv("WATCHDOG_PID")
	fmt.Fprintf(f, "NOTIFY_SOCKET='%s', WATCHDOG_USEC='%s'\n", ns, wusec)
	fmt.Printf("stdout wrote to '%s'\n", lfn)
	fmt.Fprintf(os.Stderr, "stderr wrote to '%s'\n", lfn)
	if wpid != "" {
		pid, err := strconv.Atoi(wpid)
		check(err)
		usec, err := strconv.Atoi(wusec)
		check(err)
		usec = usec / 2
		fmt.Fprintf(f, "usec='%d', pid='%d'\n", usec, pid)
		if os.Getpid() != pid {
			fmt.Fprintf(os.Stderr, "Wrong pid! '%s'\n", wpid)
			os.Exit(1)
		}
	}

	log.Fatal(router.Run(":5000")) // logger maybe not needed, but does not seem to hurt
}
