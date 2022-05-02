package web

import (
	"testing"
)

func TestStorageObjectUrlToLinuxPath(t *testing.T) {
	m1234 := &StorageObject{
		Mid:           "m1234",
		RootUrl:       "http://localhost:23632/storages/m1234/files",
		RootUrlPath:   "/storages/m1234/files",
		LinuxNrtaPath: "/data/nrta/0/m1234",
		UrlPath2Item:  make(map[string]*StorageItemObject),
	}
	{
		url := ChooseUrlThenRegister(m1234, "", StoragePathNrt, "somefile.txt")
		//url := "http://localhost:23632/storages/m1234/files/somefile.txt"
		actual, err := StorageObjectUrlToLinuxPath(m1234, url)
		check(err)
		expected := "/data/nrta/0/m1234/somefile.txt"
		if actual != expected {
			t.Errorf("Expected the linux path of %q to be %q but got %q!", url, expected, actual)
		}
	}
	{
		url := "http://localhost:23632/storages/m5678/files/otherfile.txt"
		_, err := StorageObjectUrlToLinuxPath(m1234, url)
		if err == nil {
			t.Errorf("Expected the linux path of %s for %v to cause an error!", url, m1234)
		}
	}
	{
		url := "file:/data/nrta/0/justafile.txt"
		_, err := StorageObjectUrlToLinuxPath(m1234, url)
		if err == nil {
			t.Errorf("Expected the linux path of %s for %v to cause an error!", url, m1234)
		}
	}
	{
		url := "/data/nrta/0/justafile.txt"
		_, err := StorageObjectUrlToLinuxPath(m1234, url)
		if err == nil {
			t.Errorf("Expected the linux path of %s for %v to cause an error!", url, m1234)
		}
	}
}

func TestStorageUrlToLinuxPath(t *testing.T) {
	var state State
	state.Storages = make(map[string]*StorageObject)
	m1234 := &StorageObject{
		Mid:           "m1234",
		RootUrl:       "http://localhost:23632/storages/m1234/files",
		RootUrlPath:   "/storages/m1234/files",
		LinuxNrtaPath: "/data/nrta/0/m1234",
		UrlPath2Item:  make(map[string]*StorageItemObject),
	}
	state.Storages["m1234"] = m1234
	{
		url := "http://localhost:23632/storages/m5678/files/otherfile.txt"
		_, err := StorageUrlToLinuxPath(url, &state)
		if err == nil {
			t.Errorf("Expected err for url %q, as we did not register that StorageObject yet!", url)
		}
	}
	m5678 := &StorageObject{
		Mid:           "m5678",
		RootUrl:       "http://localhost:23632/storages/m5678/files",
		RootUrlPath:   "/storages/m5678/files",
		LinuxNrtaPath: "/data/nrta/1/m5678",
		UrlPath2Item:  make(map[string]*StorageItemObject),
	}
	state.Storages["m5678"] = m5678
	{
		url := ChooseUrlThenRegister(m1234, "", StoragePathNrt, "somefile.txt")
		expectedUrl := "http://localhost:23632/storages/m1234/files/somefile.txt"
		if url != expectedUrl {
			t.Errorf("URL:\nGot %q\nNot %q", url, expectedUrl)
		}
		actual, err := StorageUrlToLinuxPath(url, &state)
		check(err)
		expected := "/data/nrta/0/m1234/somefile.txt"
		if actual != expected {
			t.Errorf("Expected the linux path of %q to be %q but got %q!", url, expected, actual)
		}
	}
	{
		url := ChooseUrlThenRegister(m5678, "", StoragePathNrt, "otherfile.txt")
		expectedUrl := "http://localhost:23632/storages/m5678/files/otherfile.txt"
		if url != expectedUrl {
			t.Errorf("URL:\nGot %q\nNot %q", url, expectedUrl)
		}
		actual, err := StorageUrlToLinuxPath(url, &state)
		check(err)
		expected := "/data/nrta/1/m5678/otherfile.txt"
		if actual != expected {
			t.Errorf("Expected the linux path of %q to be %q but got %q!", url, expected, actual)
		}
	}
	{
		url := "file:/data/nrta/0/justafile.txt"
		actual, err := StorageUrlToLinuxPath(url, &state)
		check(err)
		expected := "/data/nrta/0/justafile.txt"
		if actual != expected {
			t.Errorf("Expected the linux path of %s to be %s but got %s!", url, expected, actual)
		}
	}
	{
		url := "/data/nrta/0/justafile.txt"
		actual, err := StorageUrlToLinuxPath(url, &state)
		check(err)
		expected := "/data/nrta/0/justafile.txt"
		if actual != expected {
			t.Errorf("Expected the linux path of %s to be %s but got %s!", url, expected, actual)
		}
	}
}
func TestNextPartition(t *testing.T) {
	try := func(arg, expected string) {
		got := NextPartition(arg)
		if got != expected {
			t.Errorf("\nGot: %q\nNot: %q", got, expected)
		}
	}

	try("0", "1")
	try("1", "2")
	try("2", "3")
	try("3", "0")
	try("", "0")
}
