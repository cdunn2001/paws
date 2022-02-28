package main

import (
	"net/http"
	//"fmt"
	"log" // log.Fatal()
	// "pacb.com/seq/paws/pkg/stuff"
	// "pacb.com/seq/paws/pkg/stiff"
	//"github.com/gofiber/fiber/v2"
	//_ "github.com/gofiber/fiber/v2/middleware/recover" // to trap panics
	//"github.com/gofiber/fiber/v2/utils"
	//"github.com/gofiber/template/html"
	"github.com/gin-gonic/gin"
	"pacb.com/seq/paws/pkg/web"
	"runtime" // only for GOOS
)

// Returns top level status of the pa-ws process.
func getStatus(c *gin.Context) {
	var status web.PawsStatusObject
	c.IndentedJSON(http.StatusOK, status)
}

// Returns a list of socket ids.
func getSockets(c *gin.Context) {
	var socketIds []string = []string{"1", "2", "3", "4"}
	c.IndentedJSON(http.StatusOK, socketIds)
}

// Returns the socket object indexed by the sock_id.
func getSocketById(c *gin.Context) {
	//panic("I AM LOST!")
	//id := c.Param("id")

	//var sockets []SocketObject = []SocketObject{}
	//socket := sockets[id]
	var obj web.SocketObject
	c.IndentedJSON(http.StatusOK, obj)
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
	var obj web.SocketBasecallerObject
	c.IndentedJSON(http.StatusOK, obj)
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
	var obj web.SocketDarkcalObject
	c.IndentedJSON(http.StatusOK, obj)
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
	var obj web.SocketLoadingcalObject
	c.IndentedJSON(http.StatusOK, obj)
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
	//router := gin.Default()
	// Or explicitly:
	router := gin.New()
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

	log.Fatal(router.Run(":5000")) // maybe not needed, but does not seem to hurt
}
