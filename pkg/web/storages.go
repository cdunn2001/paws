package web

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"sort"
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

// Creates a storages resource for a movie.
func createStorage(c *gin.Context, state *State) {
	obj := new(StorageObject)
	if err := c.BindJSON(obj); err != nil {
		c.Writer.WriteString("Could not parse body into struct.\n")
		return
	}
	mid := obj.Mid
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
	c.String(http.StatusConflict, "For mid '%s', if all files have not been freed, the DELETE will fail.\n", mid)
}

// Frees all directories and files associated with the storages resources and reclaims disk space.
func freeStorageByMid(c *gin.Context, state *State) {
	//mid := c.Param("mid")
	c.Status(http.StatusOK)
}

func StorageUrlToLinuxPath(url string, state *State) (string, error) {
	fmt.Println("Converting:", url)
	lenUrl := len(url)
	if lenUrl >= 1 {
		if url[0:1] == "/" {
			//fmt.Println("Already linuxed: ",url)
			return url, nil
		}
	}
	if lenUrl >= 5 {
		if url[0:5] == "file:" {
			//fmt.Println("Removing file: prefix from ",url)
			return url[5:], nil
		}
	}
	for _, so := range state.Storages {
		//fmt.Printf("StorageUrlToLinuxPath so:%v\n", *so)
		// r, _ := regexp.Compile("^" + so.RootUrl)
		l := len(so.RootUrl)
		//fmt.Println("l:",l)
		if lenUrl >= l {
			//fmt.Println("url[0:l]:",url[0:l])
			if url[0:l] == so.RootUrl {
				filepath := url[l:]
				linuxPath := so.LinuxPath + filepath
				//fmt.Println("Found match, linux path:",linuxPath)
				return linuxPath, nil
			}
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
