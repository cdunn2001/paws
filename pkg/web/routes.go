package web

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func AddRoutes(router *gin.Engine) {
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
	id := c.Param("id")

	var obj SocketObject
	obj.SocketId = id
	c.IndentedJSON(http.StatusOK, obj)
}

// Resets all "one shot" app resources for each of the sockets.
func resetSockets(c *gin.Context) {
	c.Status(http.StatusOK)
}

// Resets all "one shot" app resources for the socket.
func resetSocketById(c *gin.Context) {
	id := c.Param("id")
	//c.Status(http.StatusOK)
	c.String(http.StatusNotFound, "The socket '%s' was not found in the list of attached sensor FPGA boards.\n", id)
}

// Returns a single image from the socket.
func getImageBySocketId(c *gin.Context) {
	id := c.Param("id")
	c.String(http.StatusNotFound, "The socket '%s' was not found in the list of attached sensor FPGA boards.\n", id)
}

// Returns the basecaller object indexed by the socket {id}.
func getBasecallerBySocketId(c *gin.Context) {
	id := c.Param("id")
	var obj SocketBasecallerObject
	obj.Mid = "Mid-for-" + id
	c.IndentedJSON(http.StatusOK, obj)
}

// Start the basecaller process on socket {id}.
func startBasecallerBySocketId(c *gin.Context) {
	id := c.Param("id")
	var obj SocketBasecallerObject
	if err := c.BindJSON(&obj); err != nil {
		c.Writer.WriteString("Could not parse body into struct.\n")
		return
	}
	obj.Mid = "Mid-for-" + id
	c.IndentedJSON(http.StatusOK, obj)
}

// Gracefully aborts the basecalling process on socket {id}. This must be called before a POST to "reset". Note The the process will not stop immediately. The client must poll the endpoint until the "process_status.execution_status" is "COMPLETE".
func stopBasecallerBySocketId(c *gin.Context) {
	id := c.Param("id")
	//c.Status(http.StatusOK)
	c.String(http.StatusNotFound, "The socket '%s' was not found in the list of attached sensor FPGA boards.\n", id)
}

// Resets the basecaller resource on socket {id}.
func resetBasecallerBySocketId(c *gin.Context) {
	id := c.Param("id")
	//c.Status(http.StatusOK)
	c.String(http.StatusNotFound, "The socket '%s' was not found in the list of attached sensor FPGA boards.\n", id)
	//c.String(http.StatusConflict, "basecaller was not in the READY or COMPLETE state.\n")
}

// Returns the darkcal object indexed by socket {id}.
func getDarkcalBySocketId(c *gin.Context) {
	id := c.Param("id")
	var obj SocketDarkcalObject
	obj.Mid = "Mid-for-" + id
	c.IndentedJSON(http.StatusOK, obj)
}

// Starts a darkcal process on socket {id}.
func startDarkcalBySocketId(c *gin.Context) {
	id := c.Param("id")
	var obj SocketDarkcalObject
	if err := c.BindJSON(&obj); err != nil {
		c.Writer.WriteString("Could not parse body into struct.\n")
		return
	}
	obj.Mid = "Mid-for-" + id
	c.IndentedJSON(http.StatusOK, obj)
}

// Gracefully aborts the darkcal process on socket {id}.
func stopDarkcalBySocketId(c *gin.Context) {
	id := c.Param("id")
	//c.Status(http.StatusOK)
	c.String(http.StatusNotFound, "The socket '%s' was not found in the list of attached sensor FPGA boards.\n", id)
}

// Resets the darkcal resource on socket {id}.
func resetDarkcalBySocketId(c *gin.Context) {
	id := c.Param("id")
	//c.Status(http.StatusOK)
	c.String(http.StatusNotFound, "The socket '%s' was not found in the list of attached sensor FPGA boards.\n", id)
	//c.String(http.StatusConflict, "Fails if darkcal is still in progress. POST to stop first.")
}

// Returns the loadingcal object indexed by socket {id}.
func getLoadingcalBySocketId(c *gin.Context) {
	id := c.Param("id")
	var obj SocketLoadingcalObject
	obj.Mid = "Mid-for-" + id
	c.IndentedJSON(http.StatusOK, obj)
}

// Starts a loadingcal process on socket {id}.
func startLoadingcalBySocketId(c *gin.Context) {
	id := c.Param("id")
	var obj SocketLoadingcalObject
	if err := c.BindJSON(&obj); err != nil {
		c.Writer.WriteString("Could not parse body into struct.\n")
		return
	}
	obj.Mid = "Mid-for-" + id
	c.IndentedJSON(http.StatusOK, obj)
}

// Gracefully aborts the loadingcal process on socket {id}.
func stopLoadingcalBySocketId(c *gin.Context) {
	id := c.Param("id")
	//c.Status(http.StatusOK)
	c.String(http.StatusNotFound, "The socket id was not found in the list of attached sensor FPGA boards.")
	c.String(http.StatusNotFound, "The socket '%s' was not found in the list of attached sensor FPGA boards.\n", id)
}

// Resets the loadingcal resource on socket {id}.
func resetLoadingcalBySocketId(c *gin.Context) {
	id := c.Param("id")
	//c.Status(http.StatusOK)
	c.String(http.StatusNotFound, "The socket '%s' was not found in the list of attached sensor FPGA boards.\n", id)
	//c.String(http.StatusConflict, "Fails if loadingcal is still in progress. POST to stop first.")
}

// Returns a list of MIDs for each storage object.
func listStorageMids(c *gin.Context) {
	found := []string{"NOT", "IMPLEMENTED"}
	c.IndentedJSON(http.StatusOK, found)
}

// Creates a storages resource for a movie.
func createStorage(c *gin.Context) {
	var obj StorageObject
	if err := c.BindJSON(&obj); err != nil {
		c.Writer.WriteString("Could not parse body into struct.\n")
		return
	}
	c.IndentedJSON(http.StatusOK, obj)
}

// Returns the storage object by MID.
func getStorageByMid(c *gin.Context) {
	mid := c.Param("mid")
	var obj StorageObject
	obj.Mid = mid
	c.IndentedJSON(http.StatusOK, obj)
}

// Deletes the storages resource for the provided movie context name (MID).
func deleteStorageByMid(c *gin.Context) {
	mid := c.Param("mid")
	c.String(http.StatusConflict, "For mid '%s', if all files have not been freed, the DELETE will fail.\n", mid)
}

// Frees all directories and files associated with the storages resources and reclaims disk space.
func freeStorageByMid(c *gin.Context) {
	//mid := c.Param("mid")
	c.Status(http.StatusOK)
}

// Returns a list of MIDs for each postprimary object.
func listPostprimaryMids(c *gin.Context) {
	found := []string{"NOT", "IMPLEMENTED"}
	c.IndentedJSON(http.StatusOK, found)
}

// Starts a postprimary process on the provided urls to basecalling artifacts files.
func startPostprimary(c *gin.Context) {
	var obj PostprimaryObject
	if err := c.BindJSON(&obj); err != nil {
		c.Writer.WriteString("Could not parse body into struct.\n")
		return
	}
	c.IndentedJSON(http.StatusOK, obj)
}

// Deletes all existing postprimaries resources.
func deletePostprimaries(c *gin.Context) {
	mid := c.Param("mid")
	//c.String(http.StatusOK, "All postprimary resources were successfully deleted.")
	c.String(http.StatusConflict, "For mid '%s', one or more of the postprimaries processes were not stopped.\n", mid)
}

// Returns the postprimary object by MID.
func getPostprimaryByMid(c *gin.Context) {
	mid := c.Param("mid")
	var obj PostprimaryObject
	obj.Mid = mid
	c.IndentedJSON(http.StatusOK, obj)
}

// Deletes the postprimary resource.
func deletePostprimaryByMid(c *gin.Context) {
	mid := c.Param("mid")
	c.String(http.StatusConflict, "The postprimaries processes for mid '%s' were not stopped. POST to the stop endpoint first.\n", mid)
}

// Gracefully aborts the postprimary proces associated with MID.
func stopPostprimaryByMid(c *gin.Context) {
	mid := c.Param("mid")
	c.String(http.StatusOK, "The process for mid '%s' was stopped, and now the resource can be DELETEd.\n", mid)
}

// TODO: Is StatusConflict appropriate?
