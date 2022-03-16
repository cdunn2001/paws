package web

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"sort"
)

// fixtures (TEMPORARY)
var (
	Sockets = map[string]SocketObject{
		"1": SocketObject{
			SocketId: "1",
		},
		"2": SocketObject{
			SocketId: "2",
		},
		"3": SocketObject{
			SocketId: "3",
		},
		"4": SocketObject{
			SocketId: "4",
		},
	}
	Storages      = make(map[string]StorageObject)
	Basecallers   = make(map[string]SocketBasecallerObject)
	Darkcals      = make(map[string]SocketDarkcalObject)
	Loadingcals   = make(map[string]SocketLoadingcalObject)
	Postprimaries = make(map[string]PostprimaryObject)
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
	var status PawsStatusObject // TODO
	c.IndentedJSON(http.StatusOK, status)
}

// Returns a list of socket ids.
func getSockets(c *gin.Context) {
	var socketIds = []string{}
	for k := range Sockets {
		socketIds = append(socketIds, k)
	}
	sort.Strings(socketIds)
	c.IndentedJSON(http.StatusOK, socketIds)
}

// Returns the socket object indexed by the sock_id.
func getSocketById(c *gin.Context) {
	id := c.Param("id")

	obj, found := Sockets[id]
	if !found {
		c.String(http.StatusNotFound, "The socket '%s' was not found in the list of attached sensor FPGA boards.\n", id)
		return
	}
	c.IndentedJSON(http.StatusOK, obj)
}

// Resets all "one shot" app resources for each of the sockets.
func resetSockets(c *gin.Context) {
	c.Status(http.StatusOK) // TODO: Should reset all 3.
}

// Resets all "one shot" app resources for the socket.
func resetSocketById(c *gin.Context) {
	id := c.Param("id")

	_, found := Sockets[id]
	if !found {
		c.String(http.StatusNotFound, "The socket '%s' was not found in the list of attached sensor FPGA boards.\n", id)
		return
	}
	c.Status(http.StatusOK)
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
	obj, found := Basecallers[id]
	if !found {
		c.String(http.StatusNotFound, "The basecaller process for socket '%s' was not found.\n", id)
		return
	}
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
	obj.ProcessStatus.ExecutionStatus = Running
	Basecallers[id] = obj
	c.IndentedJSON(http.StatusOK, obj)
}

// Gracefully aborts the basecalling process on socket {id}. This must be called before a POST to "reset". Note The the process will not stop immediately. The client must poll the endpoint until the "process_status.execution_status" is "COMPLETE".
func stopBasecallerBySocketId(c *gin.Context) {
	id := c.Param("id")
	obj, found := Basecallers[id]
	if !found {
		//c.String(http.StatusNotFound, "The socket '%s' was not found in the list of attached sensor FPGA boards.\n", id)
		c.String(http.StatusNotFound, "The basecaller process for socket '%s' was not found.\n", id)
		return
	}
	if obj.ProcessStatus.ExecutionStatus != Running &&
		obj.ProcessStatus.ExecutionStatus != Complete {
		c.String(http.StatusConflict, "Fails if basecaller is not still in progress (was %s). Do not call after /reset. Call after /start.\n", obj.ProcessStatus.ExecutionStatus)
	}
	obj.ProcessStatus.ExecutionStatus = Complete
	Basecallers[id] = obj // TODO: not thread-safe!!!
	c.Status(http.StatusOK)
}

// Resets the basecaller resource on socket {id}.
func resetBasecallerBySocketId(c *gin.Context) {
	id := c.Param("id")
	obj, found := Basecallers[id]
	if !found {
		c.String(http.StatusNotFound, "The basecaller process for socket '%s' was not found.\n", id)
		return
	}
	if obj.ProcessStatus.ExecutionStatus != Complete &&
		obj.ProcessStatus.ExecutionStatus != Ready {
		c.String(http.StatusConflict, "Fails if basecaller is still in progress (was %s). POST to stop first.\n", obj.ProcessStatus.ExecutionStatus)
		return
	}
	obj.ProcessStatus.ExecutionStatus = Ready
	Basecallers[id] = obj // TODO: not thread-safe!!!
	c.Status(http.StatusOK)
}

// Returns the darkcal object indexed by socket {id}.
func getDarkcalBySocketId(c *gin.Context) {
	id := c.Param("id")
	var obj SocketDarkcalObject
	obj, found := Darkcals[id]
	if !found {
		c.String(http.StatusNotFound, "The darkcal process for socket '%s' was not found.\n", id)
		return
	}
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
	obj.ProcessStatus.ExecutionStatus = Running
	Darkcals[id] = obj
	c.IndentedJSON(http.StatusOK, obj)
}

// Gracefully aborts the darkcal process on socket {id}.
func stopDarkcalBySocketId(c *gin.Context) {
	id := c.Param("id")
	obj, found := Darkcals[id]
	if !found {
		//c.String(http.StatusNotFound, "The socket '%s' was not found in the list of attached sensor FPGA boards.\n", id)
		c.String(http.StatusNotFound, "The darkcal process for socket '%s' was not found.\n", id)
		return
	}
	if obj.ProcessStatus.ExecutionStatus != Running &&
		obj.ProcessStatus.ExecutionStatus != Complete {
		c.String(http.StatusConflict, "Fails if darkcal is not still in progress (was %s). Do not call after /reset. Call after /start.\n", obj.ProcessStatus.ExecutionStatus)
	}
	obj.ProcessStatus.ExecutionStatus = Complete
	Darkcals[id] = obj // TODO: not thread-safe!!!
	c.Status(http.StatusOK)
}

// Resets the darkcal resource on socket {id}.
func resetDarkcalBySocketId(c *gin.Context) {
	id := c.Param("id")
	obj, found := Darkcals[id]
	if !found {
		c.String(http.StatusNotFound, "The darkcal process for socket '%s' was not found.\n", id)
		return
	}
	if obj.ProcessStatus.ExecutionStatus != Complete &&
		obj.ProcessStatus.ExecutionStatus != Ready {
		c.String(http.StatusConflict, "Fails if darkcal is still in progress. POST to stop first. State:'%s'\n", obj.ProcessStatus.ExecutionStatus)
		return
	}
	c.Status(http.StatusOK)
}

// Returns the loadingcal object indexed by socket {id}.
func getLoadingcalBySocketId(c *gin.Context) {
	id := c.Param("id")
	var obj SocketLoadingcalObject
	obj, found := Loadingcals[id]
	if !found {
		c.String(http.StatusNotFound, "The loadingcal process for socket '%s' was not found.\n", id)
		return
	}
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
	obj.ProcessStatus.ExecutionStatus = Running
	Loadingcals[id] = obj
	c.IndentedJSON(http.StatusOK, obj)
}

// Gracefully aborts the loadingcal process on socket {id}.
func stopLoadingcalBySocketId(c *gin.Context) {
	id := c.Param("id")
	obj, found := Loadingcals[id]
	if !found {
		//c.String(http.StatusNotFound, "The socket '%s' was not found in the list of attached sensor FPGA boards.\n", id)
		c.String(http.StatusNotFound, "The loadingcal process for socket '%s' was not found.\n", id)
		return
	}
	if obj.ProcessStatus.ExecutionStatus != Running &&
		obj.ProcessStatus.ExecutionStatus != Complete {
		c.String(http.StatusConflict, "Fails if loadingcal is not still in progress. Do not call after /reset. Call after /start.\n")
	}
	obj.ProcessStatus.ExecutionStatus = Complete
	Loadingcals[id] = obj // TODO: not thread-safe!!!
	c.Status(http.StatusOK)
}

// Resets the loadingcal resource on socket {id}.
func resetLoadingcalBySocketId(c *gin.Context) {
	id := c.Param("id")
	obj, found := Loadingcals[id]
	if !found {
		c.String(http.StatusNotFound, "The loadingcal process for socket '%s' was not found.\n", id)
		return
	}
	if obj.ProcessStatus.ExecutionStatus != Complete &&
		obj.ProcessStatus.ExecutionStatus != Ready {
		c.String(http.StatusConflict, "Fails if loadingcal is still in progress. POST to stop first.\n")
		return
	}
	c.Status(http.StatusOK)
}

// Returns a list of MIDs for each storage object.
func listStorageMids(c *gin.Context) {
	mids := []string{}
	for mid := range Storages {
		mids = append(mids, mid)
	}
	sort.Strings(mids)
	c.JSON(http.StatusOK, mids)
}

// Creates a storages resource for a movie.
func createStorage(c *gin.Context) {
	var obj StorageObject
	if err := c.BindJSON(&obj); err != nil {
		c.Writer.WriteString("Could not parse body into struct.\n")
		return
	}
	mid := obj.Mid
	Storages[mid] = obj
	c.IndentedJSON(http.StatusOK, obj)
}

// Returns the storage object by MID.
func getStorageByMid(c *gin.Context) {
	mid := c.Param("mid")
	var obj StorageObject
	obj, found := Storages[mid]
	if !found {
		c.String(http.StatusNotFound, "The storage for mid '%s' was not found.\n", mid)
		return
	}
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

// Returns a list of MIDs, one for each postprimary object.
func listPostprimaryMids(c *gin.Context) {
	mids := []string{}
	for mid := range Postprimaries { // TODO: thread-safety
		mids = append(mids, mid)
	}
	sort.Strings(mids)
	c.JSON(http.StatusOK, mids)
}

// Starts a postprimary process on the provided urls to basecalling artifacts files.
func startPostprimary(c *gin.Context) {
	var obj PostprimaryObject
	if err := c.BindJSON(&obj); err != nil {
		c.Writer.WriteString("Could not parse body into struct.\n")
		return
	}
	mid := obj.Mid
	if mid == "" {
		c.String(http.StatusBadRequest, "Must provide mid to start a postprimary process.\n")
		return
	}
	_, found := Storages[mid]
	if found {
		c.String(http.StatusConflict, "The postprimary process for mid '%s' already exists. (But maybe we should allow a duplicate call?)\n", mid)
		return
	}
	obj.ProcessStatus.ExecutionStatus = Running
	Postprimaries[mid] = obj
	c.IndentedJSON(http.StatusOK, obj)
}

// Deletes all existing postprimaries resources.
func deletePostprimaries(c *gin.Context) {
	mids := []string{}
	for mid := range Postprimaries { // TODO: thread-safety
		mids = append(mids, mid)
	}
	sort.Strings(mids)
	for _, mid := range mids {
		obj, found := Postprimaries[mid]
		if !found {
			panic("This is not possible unless we have a race condition.")
		}
		if obj.ProcessStatus.ExecutionStatus != Ready && // Should we ever allow Ready?
			obj.ProcessStatus.ExecutionStatus != Complete {
			c.String(http.StatusConflict, "Failed to delete postprimary process for mid '%s', still in progress (%s). Either call /stop first, or wait for Complete.\n", mid, obj.ProcessStatus.ExecutionStatus)
			return
		}
		delete(Postprimaries, mid) // TODO: thread-safety (as elsewhere)
	}
	c.String(http.StatusOK, "All postprimary resources were successfully deleted.\n")
}

// Returns the postprimary object by MID.
func getPostprimaryByMid(c *gin.Context) {
	mid := c.Param("mid")
	var obj PostprimaryObject
	obj, found := Postprimaries[mid]
	if !found {
		c.String(http.StatusNotFound, "The postprimary process for mid '%s' was not found.\n", mid)
		return
	}
	c.IndentedJSON(http.StatusOK, obj)
}

// Deletes the postprimary resource.
func deletePostprimaryByMid(c *gin.Context) {
	mid := c.Param("mid")
	var obj PostprimaryObject
	obj, found := Postprimaries[mid]
	if !found {
		c.String(http.StatusOK, "The postprimary process for mid '%s' was not found, which is fine.\n", mid)
		return
	}
	if obj.ProcessStatus.ExecutionStatus != Ready && // Should we ever allow Ready?
		obj.ProcessStatus.ExecutionStatus != Complete {
		c.String(http.StatusConflict, "Fails if postprimary for mid '%s' is not still in progress. Either call /stop first, or wait for Complete.\n", mid)
		return
	}
	delete(Postprimaries, mid) // TODO: thread-safety (as elsewhere)
	c.String(http.StatusOK, "The postprimary resource for mid '%s' was successfully deleted.\n", mid)
}

// Gracefully aborts the postprimary proces associated with MID.
func stopPostprimaryByMid(c *gin.Context) {
	mid := c.Param("mid")
	obj, found := Postprimaries[mid]
	if !found {
		//c.String(http.StatusNotFound, "The socket '%s' was not found in the list of attached sensor FPGA boards.\n", id)
		c.String(http.StatusNotFound, "The postprimary process for mid '%s' was not found.\n", mid)
		return
	}
	if obj.ProcessStatus.ExecutionStatus != Running &&
		obj.ProcessStatus.ExecutionStatus != Complete {
		c.String(http.StatusConflict, "Fails if postprimary is not still in progress. Do not call after /reset. Call after /start.")
	}
	obj.ProcessStatus.ExecutionStatus = Complete
	Postprimaries[mid] = obj // TODO: not thread-safe!!!
	c.String(http.StatusOK, "The process for mid '%s' was stopped, and now the resource can be DELETEd.\n", mid)
}

// TODO: Is StatusConflict appropriate?
