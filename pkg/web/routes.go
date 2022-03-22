package web

import (
	"bytes"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	//	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"sort"
	"sync"
	"time"
)

type State struct {
	Sockets       map[string]SocketObject
	Storages      map[string]*StorageObject
	Basecallers   map[string]*SocketBasecallerObject
	Darkcals      map[string]*SocketDarkcalObject
	Loadingcals   map[string]*SocketLoadingcalObject
	Postprimaries map[string]*PostprimaryObject
	AllProcesses  map[int]*ControlledProcess
}

// Someday, move this to separate package, for privacy.
type LockableState struct {
	state State
	lock  sync.Mutex
}

// Caller must *unlock* Mutex later (with 'defer').
// Also, caller must avoid deadlocks!
func (s *LockableState) Get() (*State, *sync.Mutex) {
	s.lock.Lock()
	return &s.state, &s.lock
}

var top LockableState

func init() {
	log.Println("Initializing fixtures")
	top.state = State{
		Sockets: map[string]SocketObject{
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
		},
		Storages:      make(map[string]*StorageObject),
		Basecallers:   make(map[string]*SocketBasecallerObject),
		Darkcals:      make(map[string]*SocketDarkcalObject),
		Loadingcals:   make(map[string]*SocketLoadingcalObject),
		Postprimaries: make(map[string]*PostprimaryObject),
		AllProcesses:  make(map[int]*ControlledProcess),
	}
	for k := range top.state.Sockets {
		top.state.Basecallers[k] = CreateSocketBasecallerObject()
		top.state.Darkcals[k] = CreateSocketDarkcalObject()
		top.state.Loadingcals[k] = CreateSocketLoadingcalObject()
	}
	// TODO: These should be configurable.
	topconfig = TopConfig{
		binaries: FindBinaries(),
		values: ValuesConfig{
			defaultFrameRate: 100.0, // fps
		},
	}
	topconfig.flat = make(map[string]string)
	topconfig.flat["Binary_baz2bam"] = topconfig.binaries.Binary_baz2bam
	topconfig.flat["Binary_pa_cal"] = topconfig.binaries.Binary_pa_cal
	topconfig.flat["Binary_reducestats"] = topconfig.binaries.Binary_reducestats
	topconfig.flat["Binary_smrt_basecaller"] = topconfig.binaries.Binary_smrt_basecaller
}

type BinaryPaths struct {
	Binary_baz2bam         string
	Binary_pa_cal          string
	Binary_reducestats     string
	Binary_smrt_basecaller string
}

type ValuesConfig struct {
	defaultFrameRate float64 // fps
}

//type StringMap map[string]string // would hide map as 'reference' type

type TopConfig struct {
	values   ValuesConfig
	binaries BinaryPaths
	flat     map[string]string // someday maybe put all here?
}

func UpdateWithConfig(kv map[string]string, tc *TopConfig) {
	for k, v := range tc.flat {
		kv[k] = v
	}
}

var topconfig TopConfig // Should be considered "const", as changes would not be thread-safe.

func FindBinaries() BinaryPaths {
	// TODO: Replace w/ PpaConfig
	return BinaryPaths{
		Binary_baz2bam:         "baz2bam",
		Binary_smrt_basecaller: "smrt-basecaller-launch.sh", // this script is necessary to configure NUMA. don't call smrt-basecaller binary directly.
		Binary_pa_cal:          "pa-cal",
		Binary_reducestats:     "ppa-reducestats",
	}
}

type StateHandlerFunc func(*gin.Context, *State)

func SafeState(h StateHandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		state, l := top.Get()
		defer l.Unlock()
		h(c, state)
	}
}

var bootTime time.Time

func AddRoutes(router *gin.Engine) {
	bootTime = time.Now()
	router.GET("/status", SafeState(getStatus))
	router.GET("/sockets", SafeState(getSockets))
	router.GET("/sockets/:id", SafeState(getSocketById))
	router.POST("/sockets/reset", SafeState(resetSockets))
	router.POST("/sockets/:id/reset", SafeState(resetSocketById))
	router.GET("/sockets/:id/image", SafeState(getImageBySocketId))
	router.GET("/sockets/:id/basecaller", SafeState(getBasecallerBySocketId))
	router.POST("/sockets/:id/basecaller/start", SafeState(startBasecallerBySocketId))
	router.POST("/sockets/:id/basecaller/stop", SafeState(stopBasecallerBySocketId))
	router.POST("/sockets/:id/basecaller/reset", SafeState(resetBasecallerBySocketId))
	router.GET("/sockets/:id/darkcal", SafeState(getDarkcalBySocketId))
	router.POST("/sockets/:id/darkcal/start", SafeState(startDarkcalBySocketId))
	router.POST("/sockets/:id/darkcal/stop", SafeState(stopDarkcalBySocketId))
	router.POST("/sockets/:id/darkcal/reset", SafeState(resetDarkcalBySocketId))
	router.GET("/sockets/:id/loadingcal", SafeState(getLoadingcalBySocketId))
	router.POST("/sockets/:id/loadingcal/start", SafeState(startLoadingcalBySocketId))
	router.POST("/sockets/:id/loadingcal/stop", SafeState(stopLoadingcalBySocketId))
	router.POST("/sockets/:id/loadingcal/reset", SafeState(resetLoadingcalBySocketId))
	router.GET("/storages", SafeState(listStorageMids))
	router.POST("/storages", SafeState(createStorage))
	router.GET("/storages/:mid", SafeState(getStorageByMid))
	router.DELETE("/storages/:mid", SafeState(deleteStorageByMid))
	router.POST("/storages/:mid/free", SafeState(freeStorageByMid))
	router.GET("/postprimaries", SafeState(listPostprimaryMids))
	router.POST("/postprimaries", SafeState(startPostprimary))
	router.DELETE("/postprimaries", SafeState(deletePostprimaries))
	router.GET("/postprimaries/:mid", SafeState(getPostprimaryByMid))
	router.DELETE("/postprimaries/:mid", SafeState(deletePostprimaryByMid))
	router.POST("/postprimaries/:mid/stop", SafeState(stopPostprimaryByMid))
}

// Returns top level status of the pa-ws process.
func getStatus(c *gin.Context, state *State) {
	var status PawsStatusObject // TODO
	// Real time seconds that pa-ws has been running
	status.Uptime = time.Now().Sub(bootTime).Seconds()

	// Time that pa-ws has been running, formatted to be human readable as hours, minutes, seconds, etc
	status.UptimeMessage = fmt.Sprintf("%g seconds", status.Uptime)

	// Current epoch time in seconds as seen by pa-ws (UTC)
	utc := time.Now().UTC()
	status.Time = float64(utc.UnixMilli()) * 0.001

	// ISO8601 timestamp (with milliseconds) of time field
	status.Timestamp = utc.Format("2006-01-02T15:04:05-0700Z")

	// Version of software, including git hash of last commit
	status.Version = "0.0.0" // TODO

	c.IndentedJSON(http.StatusOK, status)
}

// Returns a list of socket ids.
func getSockets(c *gin.Context, state *State) {
	var socketIds = []string{}
	for k := range state.Sockets {
		socketIds = append(socketIds, k)
	}
	sort.Strings(socketIds)
	c.IndentedJSON(http.StatusOK, socketIds)
}

// Returns the socket object indexed by the sock_id.
func getSocketById(c *gin.Context, state *State) {
	id := c.Param("id")

	obj, found := state.Sockets[id]
	if !found {
		c.String(http.StatusNotFound, "The socket '%s' was not found in the list of attached sensor FPGA boards.\n", id)
		return
	}
	c.IndentedJSON(http.StatusOK, obj)
}

// Resets all "one shot" app resources for each of the sockets.
func resetSockets(c *gin.Context, state *State) {
	for id, _ := range state.Sockets {
		found := false
		for _, x := range c.Params {
			if x.Key == "id" {
				x.Value = id
			}
		}
		if !found {
			c.Params = append(c.Params, gin.Param{
				Key:   "id",
				Value: id})
		}
		resetSocketById(c, state)
	}
	c.Status(http.StatusOK) // TODO: Should reset all 3.
}

// Resets all "one shot" app resources for the socket.
func resetSocketById(c *gin.Context, state *State) {
	id := c.Param("id")

	_, found := state.Sockets[id]
	if !found {
		c.String(http.StatusNotFound, "The socket '%s' was not found in the list of attached sensor FPGA boards.\n", id)
		return
	}
	resetBasecallerBySocketId(c, state)
	resetDarkcalBySocketId(c, state)
	resetLoadingcalBySocketId(c, state)
	c.Status(http.StatusOK)
}

// Returns a single image from the socket.
func getImageBySocketId(c *gin.Context, state *State) {
	id := c.Param("id")
	c.String(http.StatusNotFound, "The socket '%s' was not found in the list of attached sensor FPGA boards.\n", id)
}

// Returns the basecaller object indexed by the socket {id}.
func getBasecallerBySocketId(c *gin.Context, state *State) {
	id := c.Param("id")
	obj, found := state.Basecallers[id]
	if !found {
		c.String(http.StatusNotFound, "The basecaller process for socket '%s' was not found.\n", id)
		return
	}
	c.IndentedJSON(http.StatusOK, obj)
}

// Start the basecaller process on socket {id}.
func startBasecallerBySocketId(c *gin.Context, state *State) {
	payload, err := httputil.DumpRequest(c.Request, true)
	if err != nil {
		log.Printf("http.StatusInternalServerError %v", err.Error())
		return
	}
	log.Println("startBasecallerBySocketId dump request", string(payload))

	id := c.Param("id")
	obj := &SocketBasecallerObject{}
	if err := c.BindJSON(obj); err != nil {
		c.String(http.StatusBadRequest, "Could not parse body into struct.\n%v\n", err)
		return
	}
	obj.ProcessStatus.ExecutionStatus = Running
	state.Basecallers[id] = obj // TODO: Error if already running?
	wr := new(bytes.Buffer)
	err = WriteBasecallerBash(wr, &topconfig, obj, id)
	if err != nil {
		err = errors.Wrapf(err, "Error in WriteBasecallerBash(%v, %v, %v, %v)", wr, topconfig, obj, id)
		check(err)
		//c.String(http.StatusInternalServerError, "Error generating bash.\n%v\n", err)
		//return
	}
	log.Printf("Wrote:'%s'\n", wr.String())
	stall := c.DefaultQuery("stall", "0")
	cp := StartControlledBashProcess(wr.String(), &obj.ProcessStatus, stall)
	pid := cp.cmd.Process.Pid
	state.AllProcesses[pid] = cp
	log.Printf("Started it\n")
	c.IndentedJSON(http.StatusOK, obj)
}

// Gracefully aborts the basecalling process on socket {id}. This must be called before a POST to "reset". Note The the process will not stop immediately. The client must poll the endpoint until the "process_status.execution_status" is "COMPLETE".
func stopBasecallerBySocketId(c *gin.Context, state *State) {
	id := c.Param("id")
	obj, found := state.Basecallers[id]
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
	obj.ProcessStatus.CompletionStatus = Aborted
	state.Basecallers[id] = obj
	c.Status(http.StatusOK)
}

// Resets the basecaller resource on socket {id}.
func resetBasecallerBySocketId(c *gin.Context, state *State) {
	id := c.Param("id")
	obj, found := state.Basecallers[id]
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
	obj.ProcessStatus.CompletionStatus = Incomplete
	state.Basecallers[id] = obj
	c.Status(http.StatusOK)
}

// Returns the darkcal object indexed by socket {id}.
func getDarkcalBySocketId(c *gin.Context, state *State) {
	id := c.Param("id")
	obj, found := state.Darkcals[id]
	if !found {
		c.String(http.StatusNotFound, "The darkcal process for socket '%s' was not found.\n", id)
		return
	}
	c.IndentedJSON(http.StatusOK, obj)
}

// Starts a darkcal process on socket {id}.
func startDarkcalBySocketId(c *gin.Context, state *State) {
	payload, err := httputil.DumpRequest(c.Request, true)
	if err != nil {
		log.Printf("http.StatusInternalServerError %v", err.Error())
		return
	}
	log.Println("dump request", string(payload))

	/* 	//	bodyReader2, err1 := c.Request.GetBody()
		bodyReader2 := c.Request.Body
		err1 := error(nil)
		if err1 == nil {
			bodyText2, err2 := ioutil.ReadAll(bodyReader2)
			if bodyText2 != nil && err2 == nil {
	    		log.Printf("startDarkcalBySocketId, POSTed payload = %s", bodyText2)
			} else {
				log.Printf("startDarkcalBySocketId, ReadAll failed to read payload: %s", err2)
			}
		} else {
			log.Printf("startDarkcalBySocketId, c.Request.GetBody() failed to read payload: %s", err1)
		}
		c.Request.Body.Seek(0, io.SeekStart)
	*/
	id := c.Param("id")
	obj := &SocketDarkcalObject{}
	if err := c.BindJSON(obj); err != nil {
		c.String(http.StatusBadRequest, "Could not parse body into struct.\n%v\n", err)
		return
	}
	obj.ProcessStatus.ExecutionStatus = Running
	state.Darkcals[id] = obj // TODO: Error if already running?
	wr := new(bytes.Buffer)
	err3 := WriteDarkcalBash(wr, &topconfig, obj, id)
	if err3 != nil {
		err3 = errors.Wrapf(err3, "Error in WriteDarkcalBash(%v, %v, %v, %v)", wr, topconfig, obj, id)
		check(err3)
		//c.String(http.StatusInternalServerError, "Error generating bash.\n%v\n", err)
		//return
	}
	log.Printf("Wrote:'%s'\n", wr.String())
	stall := c.DefaultQuery("stall", "0")
	cp := StartControlledBashProcess(wr.String(), &obj.ProcessStatus, stall)
	pid := cp.cmd.Process.Pid
	state.AllProcesses[pid] = cp
	log.Printf("Started it\n")
	c.IndentedJSON(http.StatusOK, obj)
}

// Gracefully aborts the darkcal process on socket {id}.
func stopDarkcalBySocketId(c *gin.Context, state *State) {
	id := c.Param("id")
	obj, found := state.Darkcals[id]
	if !found {
		//c.String(http.StatusNotFound, "The socket '%s' was not found in the list of attached sensor FPGA boards.\n", id)
		c.String(http.StatusNotFound, "The darkcal process for socket '%s' was not found.\n", id)
		return
	}
	if obj.ProcessStatus.ExecutionStatus != Running &&
		obj.ProcessStatus.ExecutionStatus != Complete {
		c.String(http.StatusConflict, "Fails if darkcal is not still in progress (was %s). Do not call after /reset. Call after /start.\n", obj.ProcessStatus.ExecutionStatus)
	}
	// TODO: Do this elsewhere?
	obj.ProcessStatus.ExecutionStatus = Complete
	obj.ProcessStatus.CompletionStatus = Aborted
	state.Darkcals[id] = obj // TODO: Invalidates ProcessStatus pointer?
	c.Status(http.StatusOK)
}

// Resets the darkcal resource on socket {id}.
func resetDarkcalBySocketId(c *gin.Context, state *State) {
	id := c.Param("id")
	obj, found := state.Darkcals[id]
	if !found {
		c.String(http.StatusNotFound, "The darkcal process for socket '%s' was not found.\n", id)
		return
	}
	if obj.ProcessStatus.ExecutionStatus != Complete &&
		obj.ProcessStatus.ExecutionStatus != Ready {
		c.String(http.StatusConflict, "Fails if darkcal is still in progress. POST to stop first. State:'%s'\n", obj.ProcessStatus.ExecutionStatus)
		return
	}
	obj.ProcessStatus.ExecutionStatus = Ready
	obj.ProcessStatus.CompletionStatus = Incomplete
	state.Darkcals[id] = obj
	c.Status(http.StatusOK)
}

// Returns the loadingcal object indexed by socket {id}.
func getLoadingcalBySocketId(c *gin.Context, state *State) {
	id := c.Param("id")
	obj, found := state.Loadingcals[id]
	if !found {
		c.String(http.StatusNotFound, "The loadingcal process for socket '%s' was not found.\n", id)
		return
	}
	c.IndentedJSON(http.StatusOK, obj)
}

// Starts a loadingcal process on socket {id}.
func startLoadingcalBySocketId(c *gin.Context, state *State) {
	payload, err := httputil.DumpRequest(c.Request, true)
	if err != nil {
		log.Printf("http.StatusInternalServerError %v", err.Error())
		return
	}
	log.Println("startLoadingcalBySocketId dump request", string(payload))

	id := c.Param("id")
	obj := &SocketLoadingcalObject{}
	if err := c.BindJSON(obj); err != nil {
		c.String(http.StatusBadRequest, "Could not parse body into struct.\n%v\n", err)
		return
	}
	obj.ProcessStatus.ExecutionStatus = Running
	state.Loadingcals[id] = obj // TODO: Error if already running?
	wr := new(bytes.Buffer)
	err = WriteLoadingcalBash(wr, &topconfig, obj, id)
	if err != nil {
		err = errors.Wrapf(err, "Error in WriteLoadingcalBash(%v, %v, %v, %v)", wr, topconfig, obj, id)
		check(err)
		//c.String(http.StatusInternalServerError, "Error generating bash.\n%v\n", err)
		//return
	}
	log.Printf("Wrote:'%s'\n", wr.String())
	stall := c.DefaultQuery("stall", "0")
	cp := StartControlledBashProcess(wr.String(), &obj.ProcessStatus, stall)
	pid := cp.cmd.Process.Pid
	state.AllProcesses[pid] = cp
	log.Printf("Started it\n")
	c.IndentedJSON(http.StatusOK, obj)
}

// Gracefully aborts the loadingcal process on socket {id}.
func stopLoadingcalBySocketId(c *gin.Context, state *State) {
	id := c.Param("id")
	obj, found := state.Loadingcals[id]
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
	obj.ProcessStatus.CompletionStatus = Aborted
	state.Loadingcals[id] = obj // TODO: not thread-safe!!!
	c.Status(http.StatusOK)
}

// Resets the loadingcal resource on socket {id}.
func resetLoadingcalBySocketId(c *gin.Context, state *State) {
	id := c.Param("id")
	obj, found := state.Loadingcals[id]
	if !found {
		c.String(http.StatusNotFound, "The loadingcal process for socket '%s' was not found.\n", id)
		return
	}
	if obj.ProcessStatus.ExecutionStatus != Complete &&
		obj.ProcessStatus.ExecutionStatus != Ready {
		c.String(http.StatusConflict, "Fails if loadingcal is still in progress. POST to stop first.\n")
		return
	}
	obj.ProcessStatus.ExecutionStatus = Ready
	obj.ProcessStatus.CompletionStatus = Incomplete
	state.Loadingcals[id] = obj
	c.Status(http.StatusOK)
}

// Returns a list of MIDs, one for each postprimary object.
func listPostprimaryMids(c *gin.Context, state *State) {
	mids := []string{}
	for mid := range state.Postprimaries { // TODO: thread-safety
		mids = append(mids, mid)
	}
	sort.Strings(mids)
	c.JSON(http.StatusOK, mids)
}

// Starts a postprimary process on the provided urls to basecalling artifacts files.
func startPostprimary(c *gin.Context, state *State) {
	obj := &PostprimaryObject{}
	if err := c.BindJSON(obj); err != nil {
		c.String(http.StatusBadRequest, "Could not parse body into struct.\n%v\n", err)
		return
	}
	mid := obj.Mid
	if mid == "" {
		c.String(http.StatusBadRequest, "Must provide mid to start a postprimary process.\n")
		return
	}
	_, found := state.Postprimaries[mid]
	if found {
		c.String(http.StatusConflict, "The postprimary process for mid '%s' already exists. (But maybe we should allow a duplicate call?)\n", mid)
		return
	}
	obj.ProcessStatus.ExecutionStatus = Running
	state.Postprimaries[mid] = obj // TODO: Error if already running?
	wr := new(bytes.Buffer)
	err := WriteBaz2bamBash(wr, &topconfig, obj)
	if err != nil {
		err = errors.Wrapf(err, "Error in WriteBaz2BamBash(%v, %v, %v)", wr, topconfig, obj)
		check(err)
		//c.String(http.StatusInternalServerError, "Error generating bash.\n%v\n", err)
		//return
	}
	log.Printf("Wrote:'%s'\n", wr.String())
	stall := c.DefaultQuery("stall", "0")
	cp := StartControlledBashProcess(wr.String(), &obj.ProcessStatus, stall)
	pid := cp.cmd.Process.Pid
	state.AllProcesses[pid] = cp
	log.Printf("Started it\n")
	c.IndentedJSON(http.StatusOK, obj)
}

// Deletes all existing postprimaries resources.
func deletePostprimaries(c *gin.Context, state *State) {
	mids := []string{}
	for mid := range state.Postprimaries {
		mids = append(mids, mid)
	}
	sort.Strings(mids)
	for _, mid := range mids {
		obj, found := state.Postprimaries[mid]
		if !found {
			panic("This is not possible unless we have a race condition.")
		}
		if obj.ProcessStatus.ExecutionStatus != Ready && // Should we ever allow Ready?
			obj.ProcessStatus.ExecutionStatus != Complete {
			c.String(http.StatusConflict, "Failed to delete postprimary process for mid '%s', still in progress (%s). Either call /stop first, or wait for Complete.\n", mid, obj.ProcessStatus.ExecutionStatus)
			return
		}
		delete(state.Postprimaries, mid)
	}
	c.String(http.StatusOK, "All postprimary resources were successfully deleted.\n")
}

// Returns the postprimary object by MID.
func getPostprimaryByMid(c *gin.Context, state *State) {
	mid := c.Param("mid")
	obj, found := state.Postprimaries[mid]
	if !found {
		c.String(http.StatusNotFound, "The postprimary process for mid '%s' was not found.\n", mid)
		return
	}
	c.IndentedJSON(http.StatusOK, obj)
}

// Deletes the postprimary resource.
func deletePostprimaryByMid(c *gin.Context, state *State) {
	mid := c.Param("mid")
	obj, found := state.Postprimaries[mid]
	if !found {
		c.String(http.StatusOK, "The postprimary process for mid '%s' was not found, which is fine.\n", mid)
		return
	}
	if obj.ProcessStatus.ExecutionStatus != Ready && // Should we ever allow Ready?
		obj.ProcessStatus.ExecutionStatus != Complete {
		c.String(http.StatusConflict, "Fails if postprimary for mid '%s' is not still in progress. Either call /stop first, or wait for Complete.\n", mid)
		return
	}
	delete(state.Postprimaries, mid)
	c.String(http.StatusOK, "The postprimary resource for mid '%s' was successfully deleted.\n", mid)
}

// Gracefully aborts the postprimary proces associated with MID.
func stopPostprimaryByMid(c *gin.Context, state *State) {
	mid := c.Param("mid")
	obj, found := state.Postprimaries[mid]
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
	obj.ProcessStatus.CompletionStatus = Aborted
	state.Postprimaries[mid] = obj
	c.String(http.StatusOK, "The process for mid '%s' was stopped, and now the resource can be DELETEd.\n", mid)
}
