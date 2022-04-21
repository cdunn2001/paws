package web

import (
	"testing"
)

func TestStorageObjectUrlToLinuxPath(t *testing.T) {
	m1234 := &StorageObject{
		Mid:       "m1234",
		RootUrl:   "http://localhost:23632/storages/m1234/files",
		LinuxPath: "/data/nrta/0/m1234",
	}
	{
		url := "http://localhost:23632/storages/m1234/files/somefile.txt"
		actual, err := StorageObjectUrlToLinuxPath(m1234, url)
		check(err)
		expected := "/data/nrta/0/m1234/somefile.txt"
		if actual != expected {
			t.Errorf("Expected the linux path of %s to be %s but got %s!", url, expected, actual)
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
		Mid:       "m1234",
		RootUrl:   "http://localhost:23632/storages/m1234/files",
		LinuxPath: "/data/nrta/0/m1234",
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
		Mid:       "m5678",
		RootUrl:   "http://localhost:23632/storages/m5678/files",
		LinuxPath: "/data/nrta/1/m5678",
	}
	state.Storages["m5678"] = m5678
	{
		url := "http://localhost:23632/storages/m1234/files/somefile.txt"
		actual, err := StorageUrlToLinuxPath(url, &state)
		check(err)
		expected := "/data/nrta/0/m1234/somefile.txt"
		if actual != expected {
			t.Errorf("Expected the linux path of %s to be %s but got %s!", url, expected, actual)
		}
	}
	{
		url := "http://localhost:23632/storages/m5678/files/otherfile.txt"
		actual, err := StorageUrlToLinuxPath(url, &state)
		check(err)
		expected := "/data/nrta/1/m5678/otherfile.txt"
		if actual != expected {
			t.Errorf("Expected the linux path of %s to be %s but got %s!", url, expected, actual)
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
