package web

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
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

type Store struct {
}

var DefaultStorageRoot string // Must be asbolute.

func init() {
	root := "./tmp/storage"
	if !filepath.IsAbs(root) {
		absroot, err := filepath.Abs(root)
		if err != nil {
			msg := fmt.Sprintf("Unable to run filepath.Abs(%q). Someone must have deleted the current working directory.", root)
			panic(msg)
		}
		root = absroot
	}
	DefaultStorageRoot = root
}

func GetLocalStorageObject(mid string) *StorageObject {
	obj := &StorageObject{
		Mid:     mid,
		RootUrl: filepath.Join(DefaultStorageRoot, mid),
	}
	return obj
}

// If not found, generate a local file.
func GetStorageObjectForMid(mid string, state *State) *StorageObject {
	obj, found := state.Storages[mid]
	if !found {
		obj = GetLocalStorageObject(mid)
	}
	return obj
}

func (self *Store) AcquireStorageObjectFromMid(mid string, state *State) *StorageObject {
	obj := GetLocalStorageObject(mid)
	dir, err := StorageUrlToLinuxPath(obj.RootUrl, state)
	check(err)
	os.MkdirAll(dir, 0777)
	return obj
}
func (self *Store) Free(obj *StorageObject) {
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

// Creates a storages resource for a movie.
func createStorage(c *gin.Context, state *State) {
	obj := new(StorageObject)
	if err := c.BindJSON(obj); err != nil {
		c.Writer.WriteString("Could not parse body into struct.\n")
		return
	}
	mid := obj.Mid // This is the only thing we get from the input obj, right?

	obj = state.store.AcquireStorageObjectFromMid(mid, state)
	state.Storages[mid] = obj
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
	state.store.Free(obj)
	c.Status(http.StatusOK)
}

func StorageObjectUrlToLinuxPath(so *StorageObject, url string) (string, error) {
	if !strings.HasPrefix(url, so.RootUrl) {
		msg := fmt.Sprintf("Precondition violated. RootURL %q is not a prefix of URL %q.",
			so.RootUrl, url)
		return "/dev/null", errors.New(msg)
	}
	l := len(so.RootUrl)
	filepath := url[l:]
	linuxPath := so.LinuxPath + filepath
	return linuxPath, nil
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
