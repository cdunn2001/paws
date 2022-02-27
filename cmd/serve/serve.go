package main

import (
	"net/http"
	//"fmt"
	//"log" // log.Fatal()
	// "pacb.com/seq/paws/pkg/stuff"
	// "pacb.com/seq/paws/pkg/stiff"
	//"github.com/gofiber/fiber/v2"
	//_ "github.com/gofiber/fiber/v2/middleware/recover" // to trap panics
	//"github.com/gofiber/fiber/v2/utils"
	//"github.com/gofiber/template/html"
	"github.com/gin-gonic/gin"
	"runtime" // only for GOOS
)

// Top level status of the pa-ws process
type PawsStatusObject struct {
	// Real time seconds that pa-ws has been running
	Uptime float64 `json:"uptime"`

	// Time that pa-ws has been running, formatted to be human readable as hours, minutes, seconds, etc
	UptimeMessage string `json:"uptimeMessage"`

	// Current epoch time in seconds as seen by pa-ws (UTC)
	Time float64 `json:"time"`

	// ISO8601 timestamp (with milliseconds) of time field
	Timestamp string `json:"timestamp"`

	// Version of software, including git hash of last commit
	version string `json:"version"`
}

type SocketObject struct {
	// The socket identifier, typically "1" thru "4".
	SocketId string
	//darkcal
	//loadingcal
	//basecal
}

// Returns top level status of the pa-ws process.
func getStatus(c *gin.Context) {
	var status PawsStatusObject
	c.IndentedJSON(http.StatusOK, status)
}

// Returns a list of socket ids.
func getSockets(c *gin.Context) {
	var socketIds []string = []string{"1", "2", "3", "4"}
	c.IndentedJSON(http.StatusOK, socketIds)
}

// Returns the socket object indexed by the sock_id.
func getSocketById(c *gin.Context) {
	//id := c.Param("id")

	//var sockets []SocketObject = []SocketObject{}
	//socket := sockets[id]
	var socket SocketObject
	c.IndentedJSON(http.StatusOK, socket)
}

// Resets all "one shot" app resources for each of the sockets.
func resetSockets(c *gin.Context) {
}

// Resets all "one shot" app resources for the socket.
func resetSocketById(c *gin.Context) {
}

// Returns a single image from the socket.
func getImageBySocketId(c *gin.Context) {
}

// Returns the basecaller object indexed by the socket {id}.
func getBasecallerBySocketId(c *gin.Context) {
}

// Start the basecaller process on socket {id}.
func startBasecallerBySocketId(c *gin.Context) {
}

// Gracefully aborts the basecalling process on socket {id}. This must be called before a POST to "reset". Note The the process will not stop immediately. The client must poll the endpoint until the "process_status.execution_status" is "COMPLETE".
func stopBasecallerBySocketId(c *gin.Context) {
}

// Resets the basecaller resource on socket {id}.
func resetBasecallerBySocketId(c *gin.Context) {
}

// Returns the darkcal object indexed by socket {id}.
func getDarkcalBySocketId(c *gin.Context) {
}

// Starts a darkcal process on socket {id}.
func startDarkcalBySocketId(c *gin.Context) {
}

// Gracefully aborts the darkcal process on socket {id}.
func stopDarkcalBySocketId(c *gin.Context) {
}

// Resets the darkcal resource on socket {id}.
func resetDarkcalBySocketId(c *gin.Context) {
}

// Returns the loadingcal object indexed by socket {id}.
func getLoadingcalBySocketId(c *gin.Context) {
}

// Starts a loadingcal process on socket {id}.
func startLoadingcalBySocketId(c *gin.Context) {
}

// Gracefully aborts the loadingcal process on socket {id}.
func stopLoadingcalBySocketId(c *gin.Context) {
}

// Resets the loadingcal resource on socket {id}.
func resetLoadingcalBySocketId(c *gin.Context) {
}

// Returns a list of MIDs for each storage object.
func listStorageMids(c *gin.Context) {
}

// Creates a storages resource for a movie.
func createStorage(c *gin.Context) {
}

// Returns the storage object by MID.
func getStorageByMid(c *gin.Context) {
}

// Deletes the storages resource for the provided movie context name (MID).
func deleteStorageByMid(c *gin.Context) {
}

// Frees all directories and files associated with the storages resources and reclaims disk space.
func freeStorageByMid(c *gin.Context) {
}

// Returns a list of MIDs for each postprimary object.
func listPostprimaryMids(c *gin.Context) {
}

// Starts a postprimary process on the provided urls to basecalling artifacts files.
func startPostprimary(c *gin.Context) {
}

// Deletes all existing postprimaries resources.
func deletePostprimaries(c *gin.Context) {
}

// Returns the postprimary object by MID.
func getPostprimaryByMid(c *gin.Context) {
}

// Deletes the postprimary resource.
func deletePostprimaryByMid(c *gin.Context) {
}

// Gracefully aborts the postprimary proces associated with MID.
func stopPostprimaryByMid(c *gin.Context) {
}
func main() {
	router := gin.Default()
	router.GET("/hello", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Hello World!",
		})
	})

	router.GET("/os", func(c *gin.Context) {
		c.String(200, runtime.GOOS)
	})

	router.GET("/status", getStatus)
	router.GET("/sockets", getSockets)
	router.GET("/sockets/:id", getSocketById)
	router.POST("/sockets/reset", resetSockets)
	router.POST("/sockets/:id/reset", resetSocketById)
	router.GET("/sockets/:id/image", getImageBySocketId)
	router.GET("/sockets/:id/basecaller", getBasecallerBySocketId)
	router.POST("/sockets/:id/basecaller/start", startBasecallerBySocketId)
	router.POST("/sockets/:id/basecaller/stop", stopBasecallerBySocketId)
	router.POST("/sockets/:id/basecaller/reset", resetBasecallerBySocketId)
	router.GET("/sockets/:id/darkcal", getDarkcalBySocketId)
	router.POST("/sockets/:id/darkcal/start", startDarkcalBySocketId)
	router.POST("/sockets/:id/darkcal/stop", stopDarkcalBySocketId)
	router.POST("/sockets/:id/darkcal/reset", resetDarkcalBySocketId)
	router.GET("/sockets/:id/loadingcal", getLoadingcalBySocketId)
	router.POST("/sockets/:id/loadingcal/start", startLoadingcalBySocketId)
	router.POST("/sockets/:id/loadingcal/stop", stopLoadingcalBySocketId)
	router.POST("/sockets/:id/loadingcal/reset", resetLoadingcalBySocketId)
	router.GET("/storages", listStorageMids)
	router.POST("/storages", createStorage)
	router.GET("/storages/:mid", getStorageByMid)
	router.DELETE("/storages/:mid", deleteStorageByMid)
	router.POST("/storages/:mid/free", freeStorageByMid)
	router.GET("/postprimaries", listPostprimaryMids)
	router.POST("/postprimaries", startPostprimary)
	router.DELETE("/postprimaries", deletePostprimaries)
	router.GET("/postprimaries/:mid", getPostprimaryByMid)
	router.DELETE("/postprimaries/:mid", deletePostprimaryByMid)
	router.POST("/postprimaries/:mid/stop", stopPostprimaryByMid)

	router.Run(":5000")
}
