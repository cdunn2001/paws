package web

import (
    "testing"
)

// TestHelloName calls greetings.Hello with a name, checking
// for a valid return value.
func TestStorage1(t *testing.T) {
	Storages["m1234"].Mid = "m1234"
	Storages["m1234"].RootUrl = "http://localhost:23632/storages/m1234/files"
	Storages["m1234"].LinuxPath = "/data/nrta/0/m1234"

	url := "http://localhost:23632/storages/m1234/files/somefile.txt"
    actual,err := StorageUrlToLinuxPath(url)
	expected := "/data/nrta/0/m1234/somefile.txt"
	if actual != expected {
	    t.Errorf("Expected the linux path of %s to be %s but got %s!", url, expected, actual)
	}
}
