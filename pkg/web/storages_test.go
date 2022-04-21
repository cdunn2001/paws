package web

import (
	"testing"
)

func TestStorageUrlToLinuxPath(t *testing.T) {
	var state State
	state.Storages = make(map[string]*StorageObject)
	m1234 := &StorageObject{
		Mid:       "m1234",
		RootUrl:   "http://localhost:23632/storages/m1234/files",
		LinuxPath: "/data/nrta/0/m1234",
	}
	state.Storages["m1234"] = m1234
	m5678 := &StorageObject{
		Mid:       "m5678",
		RootUrl:   "http://localhost:23632/storages/m5678/files",
		LinuxPath: "/data/nrta/1/m5678",
	}
	state.Storages["m5678"] = m5678
	{
		url := "http://localhost:23632/storages/m1234/files/somefile.txt"
		actual, _ := StorageUrlToLinuxPath(url, &state)
		expected := "/data/nrta/0/m1234/somefile.txt"
		if actual != expected {
			t.Errorf("Expected the linux path of %s to be %s but got %s!", url, expected, actual)
		}
	}
	{
		url := "http://localhost:23632/storages/m5678/files/otherfile.txt"
		actual, _ := StorageUrlToLinuxPath(url, &state)
		expected := "/data/nrta/1/m5678/otherfile.txt"
		if actual != expected {
			t.Errorf("Expected the linux path of %s to be %s but got %s!", url, expected, actual)
		}
	}
	{
		url := "file:/data/nrta/0/justafile.txt"
		actual, _ := StorageUrlToLinuxPath(url, &state)
		expected := "/data/nrta/0/justafile.txt"
		if actual != expected {
			t.Errorf("Expected the linux path of %s to be %s but got %s!", url, expected, actual)
		}
	}
	{
		url := "/data/nrta/0/justafile.txt"
		actual, _ := StorageUrlToLinuxPath(url, &state)
		expected := "/data/nrta/0/justafile.txt"
		if actual != expected {
			t.Errorf("Expected the linux path of %s to be %s but got %s!", url, expected, actual)
		}
	}
}
