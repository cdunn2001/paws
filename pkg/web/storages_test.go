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
func TestChooseNextNrtPartition(t *testing.T) {
	try := func(index int, nrt string, expectedIndex int, expectedNrt string) {
		got := ChooseNextNrtPartition(NrtPartition{index, nrt})
		assert.Equal(t, expectedIndex, got.PartitionIndex)
		assert.Equal(t, expectedNrt, got.Nrt) // constant for now
	}
	try(0, "a", 0, "b")
	try(0, "b", 1, "a")
	try(1, "a", 1, "b")
	try(1, "b", 2, "a")
	try(2, "a", 2, "b")
	try(2, "b", 3, "a")
	try(3, "a", 3, "b")
	try(3, "b", 0, "a")
	// Note that the function has no side-effects, so these
	// are stateless tests.
}
func TestAcquireStorageObject(t *testing.T) {
	// This test creates directories, so we need to use testtmpdir.

	nrta := filepath.Join(testtmpdir, "nrta")
	nrtb := filepath.Join(testtmpdir, "nrtb")
	icc := filepath.Join(testtmpdir, "icc")
	CreatePathIfNeeded(nrta)
	CreatePathIfNeeded(nrtb)
	CreatePathIfNeeded(icc)
	store := NewMultiDirStore(nrta, nrtb, icc)

	try := func(expectedNrtDir, expectedPartition, mid string) {
		so := store.AcquireStorageObject(mid)

		expectedNrtPath := filepath.Join(expectedNrtDir, expectedPartition, mid)
		actualNrtPath := so.LinuxNrtPath
		assert.Equal(t, expectedNrtPath, actualNrtPath)

		expectedIccPath := filepath.Join(store.IccDir, mid)
		actualIccPath := so.LinuxIccPath
		assert.Equal(t, expectedIccPath, actualIccPath)
	}
	mid := "m123"
	so := store.AcquireStorageObject(mid) // Do not panic.
	// Normally, we would use different mids, but that actually makes no difference.
	try(store.NrtbDir, "0", mid)
	try(store.NrtaDir, "1", mid)
	try(store.NrtbDir, "1", mid)
	try(store.NrtaDir, "2", mid)
	try(store.NrtbDir, "2", mid)
	try(store.NrtaDir, "3", mid)
	try(store.NrtbDir, "3", mid)
	// We do not actually care about the order above, but for now we know it.

	assert.Panics(t, func() { store.AcquireStorageObject(mid) })
	store.Free(so)                      // Free one, so we can aquire another.
	_ = store.AcquireStorageObject(mid) // Do not panic.
}
func TestCheckIllegalPathToCreate(t *testing.T) {
	assert.Equal(t, []string{"/data/icc", "/data/nrta", "/data/nrtb"}, BadPaths)
	CheckIllegalPathToCreate("/tmp")
	was := BadPaths
	defer func() {
		BadPaths = was
	}()
	BadPaths = append(BadPaths, "/tmp/foo")
	assert.Panics(t, func() { CheckIllegalPathToCreate("/tmp/foo") })
}
func TestExists(t *testing.T) {
	assert.False(t, Exists("/fubar"), "/fubar")
	assert.True(t, Exists("/tmp"), "/tmp")
}
func TestFindFirstFalseIndex(t *testing.T) {
	assert.Equal(t, -1, FindFirstFalseIndex(0, []bool{}))
	assert.Equal(t, 0, FindFirstFalseIndex(0, []bool{false}))
	assert.Equal(t, 1, FindFirstFalseIndex(1, []bool{false, false, false}))
	assert.Equal(t, 2, FindFirstFalseIndex(1, []bool{false, true, false}))
	assert.Equal(t, 0, FindFirstFalseIndex(1, []bool{false, true, true})) // wrap
	assert.Equal(t, -1, FindFirstFalseIndex(2, []bool{true, true, true, true}))
	// Usually we have 4 indices, but this function does not care.
}
