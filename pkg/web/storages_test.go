package web

import (
	"github.com/stretchr/testify/assert"
	"path/filepath"
	"strings"
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
		urlExpected := "http://localhost:23632/storages/m1234/files/somefile.txt"
		assert.Equal(t, urlExpected, url) // can change, so not a great test

		actual, err := StorageObjectUrlToLinuxPath(m1234, url)
		assert.Nil(t, err)
		expected := "/data/nrta/0/m1234/somefile.txt"
		assert.Equal(t, expected, actual, "From linux path %q", url)
	}
	{
		url := "http://localhost:23632/storages/m5678/files/otherfile.txt"
		_, err := StorageObjectUrlToLinuxPath(m1234, url)
		assert.NotNil(t, err, "Expected the linux path of %s for %v to cause an error, because the url was not registered.", url, m1234)
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

func TestTranslateUrl(t *testing.T) {
	{
		got := TranslateUrl(nil, "file:/bar")
		expected := "/bar"
		assert.Equal(t, expected, got)
	}
	{
		got := TranslateUrl(nil, "file://hostname/bar")
		expected := "/bar"
		assert.Equal(t, expected, got)
	}
	{
		got := TranslateUrl(nil, "local")
		expected := "local"
		assert.Equal(t, expected, got)
	}
	{
		so := &StorageObject{
			RootUrl:      "http://hostname:9999/storages/MID/files",
			RootUrlPath:  "/storages/MID/files",
			LinuxIccPath: "/var",
			UrlPath2Item: make(map[string]*StorageItemObject),
		}
		url := ChooseUrlThenRegister(so, "", StoragePathIcc, "foo/bar")
		expectedUrl := "http://hostname:9999/storages/MID/files/foo/bar"
		assert.Equal(t, expectedUrl, url)

		got := TranslateUrl(so, url)
		expected := "/var/foo/bar"
		assert.Equal(t, expected, got)

		// Change port. Should not matter.
		url1 := strings.Replace(url, "9999", "8888", 1)
		assert.NotEqual(t, url, url1)
		got1 := TranslateUrl(so, url)
		expected1 := "/var/foo/bar"
		assert.Equal(t, expected1, got1)

		// Change scheme.
		url2 := strings.Replace(url, "http:", "https:", 1)
		assert.NotEqual(t, url, url2)
		assert.Panics(t, func() { TranslateUrl(so, url2) })
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
func TestCheckIllegalPathToCreate(t *testing.T) {
	assert.Panics(t, func() { CheckIllegalPathToCreate("/data/icc") })
	assert.Panics(t, func() { CheckIllegalPathToCreate("/data/icc/foo") })
	CheckIllegalPathToCreate("/tmp")
	was := BadPath
	defer func() {
		BadPath = was
	}()
	BadPath = "/tmp/foo"
	assert.Panics(t, func() { CheckIllegalPathToCreate("/tmp/foo") })
}
