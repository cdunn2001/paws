package web

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Returns a list of MIDs for each storage object.
func listStorageMids(c *gin.Context, state *State) {
	mids := []string{}
	for mid, _ := range state.Storages {
		mids = append(mids, mid)
	}
	sort.Strings(mids)
	c.IndentedJSON(http.StatusOK, mids)
}

type IStore interface {
	Free(obj *StorageObject)
	AcquireStorageObject(mid string) *StorageObject
}

func Exists(path string) bool {
	_, err := os.Stat("temp")
	return !os.IsNotExist(err)
}

var BadPath = "/data/icc"

func CheckIllegalPathToCreate(path string) {
	if strings.HasPrefix(filepath.Clean(path), BadPath) && !Exists(BadPath) {
		msg := fmt.Sprintf("Trying to create %q, which must already exist. (for %q)", BadPath, path)
		panic(msg)
	}
}

func CreatePathIfNeeded(path string) {
	if !Exists(path) {
		log.Printf("CreatePathIfNeeded(%q)\n", path)
	}
	CheckIllegalPathToCreate(path)
	err := os.MkdirAll(path, 0777) // Does not guarantee 0777 if already exists.
	if err != nil {
		msg := fmt.Sprintf("Could not create directory %q: %v", path, err)
		panic(msg)
	}
}
func DeletePathIfExists(path string) {
	log.Printf("DeletePathIfNeeded(%q)\n", path)
	base := filepath.Base(path)
	if !strings.HasPrefix(base, "m") {
		log.Printf("For safety, we do not RemoveAll for a directory that does not start with 'm'. (%q)\n", path)
		return
	}
	err := os.RemoveAll(path)
	if err != nil {
		log.Printf("WARNING: Failed to remove directory %q: %v", path, err)
	}
}

var DefaultStorageRootNrta string = "/data/nrta"
var DefaultStorageRootNrtb string = "/data/nrtb"
var DefaultStorageRootIcc string = "/data/icc"

type MultiDirStore struct {
	NrtaDir       string
	NrtbDir       string
	IccDir        string
	LastPartition string
	LastNrt       string // a|b
}

func CreateDefaultStore() *MultiDirStore {
	return CreateMultiDirStore(DefaultStorageRootNrta, DefaultStorageRootNrtb, DefaultStorageRootIcc)
}

// For testing, use root + "/icc" and root + "/nrta|b" as output dirs.
func CreateMultiDirStoreFromOne(root string) *MultiDirStore {
	nrtaroot := filepath.Join(root, "nrta")
	nrtbroot := filepath.Join(root, "nrtb")
	iccroot := filepath.Join(root, "icc")
	return CreateMultiDirStore(nrtaroot, nrtbroot, iccroot)
	// TODO: Figure a way to avoid accidentally reversing the args.
}
func CreateMultiDirStore(nrtaroot, nrtbroot, iccroot string) *MultiDirStore {
	abspath := func(root string) string {
		if filepath.IsAbs(root) {
			return root
		}
		absroot, err := filepath.Abs(root)
		if err != nil {
			msg := fmt.Sprintf("Unable to run filepath.Abs(%q). Someone must have deleted the current working directory.", root)
			panic(msg)
		}
		return absroot
	}
	nrtaroot = abspath(nrtaroot)
	nrtbroot = abspath(nrtbroot)
	iccroot = abspath(iccroot)
	CreatePathIfNeeded(nrtaroot)
	CreatePathIfNeeded(nrtbroot)
	CreatePathIfNeeded(iccroot)
	return &MultiDirStore{
		NrtaDir:       nrtaroot,
		NrtbDir:       nrtbroot,
		IccDir:        iccroot,
		LastPartition: "",
		LastNrt:       "",
	}
}

// Someday, this will have smart logic, to load-balance the partitions.
// For now, just cycle.
func NextPartition(last string) string {
	switch last {
	case "0":
		return "1"
	case "1":
		return "2"
	case "2":
		return "3"
	case "3":
		return "0"
	default:
		return "0"
	}
}
func (self *MultiDirStore) AcquireStorageObject(mid string) *StorageObject {
	var (
		partition string
		nrt       string
		nrtDir    string
	)
	// Temporary, until we split BAZ files to both NRTs at the same time.
	// TODO: Reimplement
	if self.LastNrt == "a" {
		nrt = "b"
		nrtDir = self.NrtbDir
		partition = self.LastPartition
	} else {
		nrt = "a"
		nrtDir = self.NrtaDir
		partition = NextPartition(self.LastPartition)
	}
	obj := &StorageObject{
		Mid:          mid,
		RootUrl:      filepath.Join("http://storages", mid, "files"),
		RootUrlPath:  filepath.Join("/storages", mid, "files"),
		LinuxIccPath: filepath.Join(self.IccDir, mid),
		LinuxNrtPath: filepath.Join(nrtDir, partition, mid),
		UrlPath2Item: make(map[string]*StorageItemObject),
	}
	// To start fresh. Also, we can allow debug logs to linger. But we can also drop this.
	DeletePathIfExists(obj.LinuxIccPath)
	DeletePathIfExists(obj.LinuxNrtPath)

	CreatePathIfNeeded(obj.LinuxIccPath)
	CreatePathIfNeeded(obj.LinuxNrtPath)
	self.LastPartition = partition
	self.LastNrt = nrt
	return obj
}
func (self *MultiDirStore) Free(obj *StorageObject) {
	for _, sio := range obj.UrlPath2Item {
		url := sio.Url
		fn := sio.LinuxPath
		log.Printf("Removing %q (%s)", fn, url)
		err := os.Remove(fn)
		if err != nil {
			log.Printf("WARNING: Failed to remove %q: %v", fn, err)
		}
	}
	obj.UrlPath2Item = nil // len() is still valid.
	DeletePathIfExists(obj.LinuxIccPath)
	DeletePathIfExists(obj.LinuxNrtPath)
}

// Currently only used for tests.
func GetLocalStorageObject(nrtbasedir string, iccbasedir string, partition string, mid string) *StorageObject {
	baseurl := "http://storages"
	obj := &StorageObject{
		Mid:          mid,
		RootUrl:      filepath.Join(baseurl, mid, "files"),
		RootUrlPath:  filepath.Join("/storages", mid, "files"),
		LinuxNrtPath: filepath.Join(nrtbasedir, partition, mid),
		LinuxIccPath: filepath.Join(iccbasedir, partition),
		UrlPath2Item: make(map[string]*StorageItemObject),
	}
	return obj
}

func RequireStorageObjectForMid(mid string, state *State) *StorageObject {
	obj, _ := state.Storages[mid]
	if obj == nil {
		msg := fmt.Sprintf("Must first call /storages endpoint for mid %q.", mid)
		panic(msg)
	}
	return obj
}

// Creates a storages resource for a movie.
func createStorage(c *gin.Context, state *State) {
	obj := new(StorageObject)
	if err := c.BindJSON(obj); err != nil {
		c.String(http.StatusBadRequest, "Could not parse body into struct.\n")
		return
	}
	mid := obj.Mid
	if mid == "" {
		c.String(http.StatusBadRequest, "/storages endpoint requires 'mid' (movie id).\n")
		return
	}

	obj = state.Store.AcquireStorageObject(mid)
	state.Storages[mid] = obj
	//log.Printf("New StorageObject: %+v", obj)
	log.Printf("New StorageObject: %q\n  -> %q", obj.RootUrl, obj.LinuxIccPath)
	c.IndentedJSON(http.StatusOK, obj)
}

// Returns the storage object by MID.
func getStorageByMid(c *gin.Context, state *State) {
	mid := c.Param("mid")
	obj, found := state.Storages[mid]
	if !found {
		c.String(http.StatusNotFound, "The storage for mid '%s' was not found.\n", mid)
		return
	}
	c.IndentedJSON(http.StatusOK, obj)
}

// Deletes the storages resource for the provided movie context name (MID).
func deleteStorageByMid(c *gin.Context, state *State) {
	mid := c.Param("mid")
	obj, found := state.Storages[mid]
	if !found {
		c.String(http.StatusOK, "The storage for mid '%s' was not found. Must have been deleted already, which is fine.\n", mid)
		return
	}
	if len(obj.UrlPath2Item) != 0 {
		c.String(http.StatusConflict, "For mid '%s', %d files have not been freed.\n", mid, len(obj.UrlPath2Item))
		return
	}
	delete(state.Storages, mid)
	c.Status(http.StatusOK)
}

// Frees all directories and files associated with the storages resources and reclaims disk space.
func freeStorageByMid(c *gin.Context, state *State) {
	mid := c.Param("mid")
	obj, found := state.Storages[mid]
	if !found {
		c.String(http.StatusConflict, "The storage for mid '%s' was not found. Must have been deleted already.\n", mid)
		return
	}
	// TODO: Do this in the background. PTSD-1282
	state.Store.Free(obj)
	c.Status(http.StatusOK)
}

func StorageObjectUrlToLinuxPath(so *StorageObject, Url string) (string, error) {
	//log.Printf("url: %q; so: %+v", Url, so)
	urlpath, err := url.Parse(Url)
	if err != nil {
		return "", err
	}
	if !strings.HasPrefix(urlpath.Path, so.RootUrlPath) {
		msg := fmt.Sprintf("Precondition violated. RootUrlPath %q is not a prefix of URL %q.",
			so.RootUrlPath, urlpath.Path)
		return "/dev/null", errors.New(msg)
	}
	sio, found := so.UrlPath2Item[urlpath.Path]
	if !found {
		msg := fmt.Sprintf("Failed to find urlpath %q (from URL %q) among registered paths. Someone forget to call ChooseUrlThenRegister()", urlpath.Path, Url)
		return "/dev/null", errors.New(msg)
	}
	return sio.LinuxPath, nil
}

// A URL can be:
//
// 1) symbolic, via this StorageObject
// Example: http://localhost:23632/storages/m123456_987654
//
// 2) physical path to storage directory (should only be used for debugging and logging)
// Example: file:/data/pa/m123456_987654
// Example: /data/pa/m123456_987654
//
// Supports:
//  file:/path   <- strips off file: and returns /path
//  /path        <- returns /path
//  localfile    <- I would like to drop support for this, but I don't want to break anything (MTL) I want all paths to be absolute.
//  discard:     <- returns ""
// eventually will support
//  file://host/path  <- returns /path assuming the path is NFS mounted, otherwise panics
//  http://host:port/storages/mid  <- will convert to a Linux path after being processed by the storages framework
func TranslateUrl(so *StorageObject, Url string) string {
	if strings.HasPrefix(Url, "/") {
		return Url
	}
	parsed, err := url.Parse(Url)
	if err != nil {
		msg := fmt.Sprintf("URL parsing error: %v", err)
		panic(msg)
	}
	if parsed.Scheme == "file" {
		return parsed.Path
	} else if parsed.Scheme == "discard" {
		return "" // TODO: or "/dev/null" ? not sure
	} else if parsed.Scheme == "" {
		return parsed.Path
	} else if parsed.Scheme != "http" {
		msg := fmt.Sprintf("Unsupported scheme %q in URL %q", parsed.Scheme, Url)
		panic(msg)
	}

	// Must be storages endpoint.
	if !strings.HasPrefix(parsed.Path, "/storages/") {
		msg := fmt.Sprintf("Unable to translate URL %q w/ path %q into linux path. Expected 'http://host:port/storages/path...'", Url, parsed.Path)
		log.Printf(msg)
		panic(msg)
	}
	if so == nil {
		msg := fmt.Sprintf("Nil StorageObject for URL %q", Url)
		panic(msg)
	}
	result, err := StorageObjectUrlToLinuxPath(so, Url)
	if err != nil {
		msg := fmt.Sprintf("Unable to translate URL %q into linux path via StorageObject %+v: %v", Url, so, err)
		log.Printf(msg)
		panic(msg)
	}
	return result
}

func ChooseUrl(so *StorageObject, linuxtail string) string {
	return so.RootUrl + "/" + linuxtail
}

// If Url is "", then create a new StorageItem.
// Register the Url if not already registered.
// Then, return the Url
func ChooseUrlThenRegister(so *StorageObject, Url string, loc StoragePathEnum, linuxtail string) string {
	if Url == "" {
		Url = ChooseUrl(so, linuxtail)
	}
	parsedUrl, err := url.Parse(Url)
	if err != nil {
		msg := fmt.Sprintf("Error parsing URL %q: %v", Url, err)
		panic(msg)
	}
	urlpath := parsedUrl.Path
	item, found := so.UrlPath2Item[urlpath]
	if !found {
		var linuxpath string
		switch loc {
		case StoragePathIcc:
			linuxpath = filepath.Join(so.LinuxIccPath, linuxtail)
		case StoragePathNrt:
			linuxpath = filepath.Join(so.LinuxNrtPath, linuxtail)
		default:
			msg := fmt.Sprintf("Unexpected StoragePathEnum %v", loc)
			panic(msg)
		}
		log.Printf("linuxpath: %q from %q or %q", linuxpath, so.LinuxNrtPath, so.LinuxIccPath)
		item = &StorageItemObject{
			UrlPath:   urlpath,
			LinuxPath: linuxpath,
			Loc:       loc,
		}
		so.UrlPath2Item[urlpath] = item
		//so.LinuxPath2Item[item.LinuxPath] = item
	}
	log.Printf("item: %+v", item)
	return Url
}
