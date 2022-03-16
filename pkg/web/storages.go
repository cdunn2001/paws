package web

import (
	"github.com/gin-gonic/gin"
	"sort"
	"net/http"
	"errors"
)


// Returns a list of MIDs for each storage object.
func listStorageMids(c *gin.Context) {
	mids := []string{}
	for mid, _ := range Storages {
		mids = append(mids, mid)
	}
	sort.Strings(mids)
	c.IndentedJSON(http.StatusOK, mids)
}

// Creates a storages resource for a movie.
func createStorage(c *gin.Context) {
	obj :=new(StorageObject)
	if err := c.BindJSON(obj); err != nil {
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
	var obj *StorageObject
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

func StorageUrlToLinuxPath(url string) (string, error) {

	rootUrl := "http://whatever:23632/m1234"
	filepath := "whatever"
	for _, so := range Storages {
		if so.RootUrl == rootUrl {
			linuxPath := so.LinuxPath + filepath
			return linuxPath, nil
		}
	}
	return "/dev/null",errors.New("Could not convert " + url)
			// symbolic link to storage directory which points back to this StorageObject
	// Example: http://localhost:23632/storages/m123456_987654
	//RootUrl string `json:"rootUrl"`

	// physical path to storage directory (should only be used for debugging and logging)
	// Example: file:/data/pa/m123456_987654
	//LinuxPath string `json:"linuxPath"`

	// Destination URL for the log file. Logging happens during construction and freeing.
	// Example: http://localhost:23632/storages/m123456_987654/storage.log
}