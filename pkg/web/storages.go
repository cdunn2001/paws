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
	AcquireStorageObject(sid string, mid string) *StorageObject
}

func CreatePathIfNeeded(path string) {
	err := os.MkdirAll(path, 0777) // Does not guarantee 0777 if already exists.
	if err != nil {
		msg := fmt.Sprintf("Could not create directory %q: %v", path, err)
		panic(msg)
	}
}

var DefaultStorageRootNrt string = "/data/nrta"
var DefaultStorageRootIcc string = "/data/icc"

type MultiDirStore struct {
	NrtDir string
	IccDir string
}

func CreateDefaultStore() *MultiDirStore {
	return CreateMultiDirStore(DefaultStorageRootNrt, DefaultStorageRootIcc)
}

// For testing, use root + "/icc" and root + "/nrt" as output dirs.
func CreateMultiDirStoreFromOne(root string) *MultiDirStore {
	nrtroot := filepath.Join(root, "nrt")
	iccroot := filepath.Join(root, "icc")
	return CreateMultiDirStore(nrtroot, iccroot)
	// TODO: Figure a way to avoid accidentally reversing the args.
}
func CreateMultiDirStore(nrtroot string, iccroot string) *MultiDirStore {
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
	nrtroot = abspath(nrtroot)
	iccroot = abspath(iccroot)
	CreatePathIfNeeded(nrtroot)
	CreatePathIfNeeded(iccroot)
	return &MultiDirStore{
		NrtDir: nrtroot,
		IccDir: iccroot,
	}
}

func (self *MultiDirStore) AcquireStorageObject(socketId string, mid string) *StorageObject {
	//paths := GetStorageObjectPaths(self.NrtDir, self.IccDir, socketId, mid) // TODO: socketId should be decoupled from choice of partition.
	obj := &StorageObject{
		Mid:           mid,
		RootUrl:       filepath.Join("http://storages", mid, "files"),
		RootUrlPath:   filepath.Join("/storages", mid, "files"),
		LinuxIccPath:  filepath.Join(self.IccDir, mid),
		LinuxNrtaPath: filepath.Join(self.NrtDir, socketId, mid),
		//LinuxNrtbPath: filepath.Join(self.NrtbDir, socketId, mid), // TODO
		UrlPath2Item: make(map[string]*StorageItemObject),
	}
	CreatePathIfNeeded(obj.LinuxIccPath)
	CreatePathIfNeeded(obj.LinuxNrtaPath)
	os.MkdirAll(obj.LinuxIccPath, 0777)
	os.MkdirAll(obj.LinuxNrtaPath, 0777)
	return obj
}
func (self *MultiDirStore) Free(obj *StorageObject) {
	for _, sio := range obj.Files {
		url := sio.Url
		fn, err := StorageObjectUrlToLinuxPath(obj, url)
		if err != nil {
			log.Printf("WARNING: Failed to convert URL %q to LinuxPath: %v.\n  Cannot remove from disk.", url, err)
			continue
		}
		log.Printf("Removing %q (%s)", fn, url)
		err = os.Remove(fn)
		if err != nil {
			log.Printf("WARNING: Failed to remove %q: %v", fn, err)
		}
	}
	obj.Files = obj.Files[:0]
}

/*
func GetStorageObjectPaths(nrtbasedir string, iccbasedir string, partition string, mid string) StorageObjectPaths {
	nrtdir := filepath.Join(nrtbasedir, partition, mid)
	iccdir := filepath.Join(iccbasedir, mid)
	os.MkdirAll(nrtdir, 0777)
	os.MkdirAll(iccdir, 0777)
	Ppa_OutputPrefix := filepath.Join(iccdir, mid)
	paths := StorageObjectPaths{
		Darkcal_CalibFile:    filepath.Join(nrtdir, mid+".darkcal.h5"),
		Darkcal_Log:          filepath.Join(nrtdir, mid+".darkcal.log"),
		Loadingcal_CalibFile: filepath.Join(nrtdir, mid+".loadingcal.h5"),
		Loadingcal_Log:       filepath.Join(nrtdir, mid+".loadingcal.log"),
		Basecaller_Baz:       filepath.Join(nrtdir, mid+".baz"),
		Basecaller_Log:       filepath.Join(nrtdir, mid+".basecaller.log"),
		//Basecaller_SimulationFile:          filepath.Join(nrtdir, mid+".baz"),
		Basecaller_TraceFile: filepath.Join(nrtdir, mid+".trc.h5"),
		//ReduceStats_Log:         filepath.Join(iccdir, mid+".reducestats.log"),
		Ppa_OutputPrefix:        Ppa_OutputPrefix,
		Ppa_OutputStatsH5:       Ppa_OutputPrefix + ".sts.h5",
		Ppa_OutputStatsXml:      Ppa_OutputPrefix + ".sts.xml",
		Ppa_OutputReduceStatsH5: Ppa_OutputPrefix + ".rsts.h5",
	}
	return paths
}
*/

// Currently only used for tests.
func GetLocalStorageObject(nrtbasedir string, iccbasedir string, sra string, mid string) *StorageObject {
	//paths := GetStorageObjectPaths(nrtbasedir, iccbasedir, sra, mid)
	baseurl := "http://storages"
	obj := &StorageObject{
		Mid:           mid,
		RootUrl:       filepath.Join(baseurl, mid, "files"),
		RootUrlPath:   filepath.Join("/storages", mid, "files"),
		LinuxNrtaPath: filepath.Join(nrtbasedir, sra, mid),
		LinuxIccPath:  filepath.Join(iccbasedir, sra),
		UrlPath2Item:  make(map[string]*StorageItemObject),
	}
	return obj
}

// TODO: Stop passing store.
func GetStorageObjectForMid(store IStore, mid string, state *State) *StorageObject {
	obj, _ := state.Storages[mid]
	// If not found, return nil. Storage URLs need a StorageObject only if they are
	//   http://host:port/storage/...
	return obj
}

// Creates a storages resource for a movie.
func createStorage(c *gin.Context, state *State) {
	obj := new(StorageObject)
	if err := c.BindJSON(obj); err != nil {
		c.String(http.StatusBadRequest, "Could not parse body into struct.\n")
		return
	}
	socketId := "0" // TODO: This should not be hard-coded.
	mid := obj.Mid
	if socketId == "" || mid == "" {
		c.String(http.StatusBadRequest, "/storages endpoint requires both 'mid' and 'socketId' fields.\n")
		return
	}

	obj = state.Store.AcquireStorageObject(socketId, mid)
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
	if len(obj.Files) != 0 {
		c.String(http.StatusConflict, "For mid '%s', %d files have not been freed.\n", mid, len(obj.Files))
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
func StorageUrlToLinuxPath(url string, state *State) (string, error) {
	log.Println("Converting:", url)
	if strings.HasPrefix(url, "/") {
		//log.Println("Already linuxed: ",url)
		return url, nil
	}
	if strings.HasPrefix(url, "file:") {
		//log.Println("Removing file: prefix from ",url)
		return url[5:], nil
	}
	for _, so := range state.Storages {
		//log.Printf("StorageUrlToLinuxPath so:%v\n", *so)
		// r, _ := regexp.Compile("^" + so.RootUrl)
		if strings.HasPrefix(url, so.RootUrl) {
			//log.Println("url[0:l]:",url[0:l])
			//log.Println("Found match, linux path:", linuxPath)
			return StorageObjectUrlToLinuxPath(so, url)
		}
	}
	return "/dev/null", errors.New("Could not convert " + url)
	// symbolic link to storage directory which points back to this StorageObject
	// Example: http://localhost:23632/storages/m123456_987654
	//RootUrl string `json:"rootUrl"`

	// physical path to storage directory (should only be used for debugging and logging)
	// Example: file:/data/pa/m123456_987654
	//LinuxPath string `json:"linuxPath"`

	// Destination URL for the log file. Logging happens during construction and freeing.
	// Example: http://localhost:23632/storages/m123456_987654/storage.log
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
			linuxpath = filepath.Join(so.LinuxNrtaPath, linuxtail) // TODO: a/b
		default:
			msg := fmt.Sprintf("Unexpected StoragePathEnum %v", loc)
			panic(msg)
		}
		log.Printf("linuxpath: %q from %q or %q", linuxpath, so.LinuxNrtaPath, so.LinuxIccPath)
		item = &StorageItemObject{
			UrlPath:   urlpath,
			LinuxPath: linuxpath,
		}
		so.UrlPath2Item[urlpath] = item
		//so.LinuxPath2Item[item.LinuxPath] = item
	}
	log.Printf("item: %+v", item)
	return Url
}
