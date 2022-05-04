package web

import (
	"github.com/stretchr/testify/assert"
	"path/filepath"
	"testing"
)

func TestStorageObjectUrlToLinuxPath(t *testing.T) {
	m1234 := &StorageObject{
		Mid:          "m1234",
		RootUrl:      "http://localhost:23632/storages/m1234/files",
		RootUrlPath:  "/storages/m1234/files",
		LinuxNrtPath: "/data/nrta/0/m1234",
		UrlPath2Item: make(map[string]*StorageItemObject),
	}
	{
		url := ChooseUrlThenRegister(m1234, "", StoragePathNrt, "somefile.txt")
		//url := "http://localhost:23632/storages/m1234/files/somefile.txt"
		actual, err := StorageObjectUrlToLinuxPath(m1234, url)
		assert.Nil(t, err)
		expected := "/data/nrta/0/m1234/somefile.txt"
		assert.Equal(t, expected, actual, "From linux path %q", url)
	}
	{
		url := "http://localhost:23632/storages/m5678/files/otherfile.txt"
		_, err := StorageObjectUrlToLinuxPath(m1234, url)
		assert.NotNil(t, err, "Expected the linux path of %s for %v to cause an error!", url, m1234)
	}
	{
		url := "file:/data/nrta/0/justafile.txt"
		_, err := StorageObjectUrlToLinuxPath(m1234, url)
		assert.NotNil(t, err, "Expected the linux path of %s for %v to cause an error!", url, m1234)
	}
	{
		url := "/data/nrta/0/justafile.txt"
		_, err := StorageObjectUrlToLinuxPath(m1234, url)
		assert.NotNil(t, err, "Expected the linux path of %s for %v to cause an error!", url, m1234)
	}
}

func TestStorageUrlToLinuxPath(t *testing.T) {
	var state State
	state.Storages = make(map[string]*StorageObject)
	m1234 := &StorageObject{
		Mid:          "m1234",
		RootUrl:      "http://localhost:23632/storages/m1234/files",
		RootUrlPath:  "/storages/m1234/files",
		LinuxNrtPath: "/data/nrta/0/m1234",
		UrlPath2Item: make(map[string]*StorageItemObject),
	}
	state.Storages["m1234"] = m1234
	{
		url := "http://localhost:23632/storages/m5678/files/otherfile.txt"
		_, err := StorageUrlToLinuxPath(url, &state)
		assert.NotNil(t, err, "Expected err for url %q, as we did not register that StorageObject yet!", url)
	}
	m5678 := &StorageObject{
		Mid:          "m5678",
		RootUrl:      "http://localhost:23632/storages/m5678/files",
		RootUrlPath:  "/storages/m5678/files",
		LinuxNrtPath: "/data/nrta/1/m5678",
		UrlPath2Item: make(map[string]*StorageItemObject),
	}
	state.Storages["m5678"] = m5678
	{
		url := ChooseUrlThenRegister(m1234, "", StoragePathNrt, "somefile.txt")
		expectedUrl := "http://localhost:23632/storages/m1234/files/somefile.txt"
		assert.Equal(t, expectedUrl, url)
		actual, err := StorageUrlToLinuxPath(url, &state)
		assert.Nil(t, err)
		expected := "/data/nrta/0/m1234/somefile.txt"
		assert.Equal(t, expected, actual, "From linux path %q", url)
	}
	{
		url := ChooseUrlThenRegister(m5678, "", StoragePathNrt, "otherfile.txt")
		expectedUrl := "http://localhost:23632/storages/m5678/files/otherfile.txt"
		assert.Equal(t, expectedUrl, url)
		actual, err := StorageUrlToLinuxPath(url, &state)
		assert.Nil(t, err)
		expected := "/data/nrta/1/m5678/otherfile.txt"
		assert.Equal(t, expected, actual, "From linux path %q", url)
	}
	{
		url := "file:/data/nrta/0/justafile.txt"
		actual, err := StorageUrlToLinuxPath(url, &state)
		assert.Nil(t, err)
		expected := "/data/nrta/0/justafile.txt"
		assert.Equal(t, expected, actual, "From linux path %q", url)
	}
	{
		url := "/data/nrta/0/justafile.txt"
		actual, err := StorageUrlToLinuxPath(url, &state)
		assert.Nil(t, err)
		expected := "/data/nrta/0/justafile.txt"
		assert.Equal(t, expected, actual, "From linux path %q", url)
	}
}
func TestNextPartition(t *testing.T) {
	try := func(arg, expected string) {
		got := NextPartition(arg)
		assert.Equal(t, expected, got)
	}
	try("0", "1")
	try("1", "2")
	try("2", "3")
	try("3", "0")
	try("", "0")
}
func TestAcquireStorageObject(t *testing.T) {
	// This call creates directories, so we need to use testtmpdir.

	store := &MultiDirStore{
		NrtaDir:       filepath.Join(testtmpdir, "nrta"),
		NrtbDir:       filepath.Join(testtmpdir, "nrtb"),
		IccDir:        filepath.Join(testtmpdir, "icc"),
		LastPartition: "",
		LastNrt:       "",
	}
	try := func(expectedNrtDir, expectedPartition, mid string) {
		so := store.AcquireStorageObject(mid)

		expectedNrtPath := filepath.Join(expectedNrtDir, expectedPartition, mid)
		actualNrtPath := so.LinuxNrtPath
		assert.Equal(t, expectedNrtPath, actualNrtPath)

		expectedIccPath := filepath.Join(store.IccDir, mid)
		actualIccPath := so.LinuxIccPath
		assert.Equal(t, expectedIccPath, actualIccPath)
	}
	try(store.NrtaDir, "0", "m123")
	try(store.NrtbDir, "0", "m123")
	try(store.NrtaDir, "1", "m123")
	try(store.NrtbDir, "1", "m123")
	try(store.NrtaDir, "2", "m123")
	try(store.NrtbDir, "2", "m123")
	try(store.NrtaDir, "3", "m123")
	try(store.NrtbDir, "3", "m123")
	try(store.NrtaDir, "0", "m123")
}
