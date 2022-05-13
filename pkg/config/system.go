package config

import (
	"fmt"
	"github.com/pkg/errors"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
)

func check(e error) {
	if e != nil {
		panic(fmt.Sprintf("Stacktrace: %+v", e))
	}
}

func FindBinary(basename string) string {
	out, err := exec.Command("which", basename).Output() // .CombinedOutput()?
	if err != nil {
		log.Printf("Failed to find %q in PATH:%s\n", basename, os.Getenv("PATH"))
		return ""
	} else {
		return strings.TrimSpace(string(out))
	}
}

func GetStdout(prog string, args ...string) string {
	out, err := exec.Command(prog, args...).Output()
	if err != nil {
		log.Printf("Command failed %s %s: %s\n", prog, strings.Join(args, " "), err)
		return ""
	} else {
		return strings.TrimSpace(string(out))
	}
}

type BinaryDescription struct {
	Path    string
	Version string
}

type BinaryDescriptions map[string]BinaryDescription

func DescribeBinaries(bps BinaryPaths) BinaryDescriptions {
	result := make(BinaryDescriptions)

	{
		name := "baz2bam"
		path := FindBinary(name)
		version := GetStdout(path, "--version")
		result[name] = BinaryDescription{
			Path:    path,
			Version: version,
		}
	}
	{
		name := "basecaller"
		path := FindBinary("smrt-basecaller-launch.sh") // this script is necessary to configure NUMA. don't call smrt-basecaller binary directly.
		version := GetStdout(path, "--version")
		result[name] = BinaryDescription{
			Path:    path,
			Version: version,
		}
	}
	{
		name := "pa-cal"
		path := FindBinary(name)
		version := GetStdout(path, "--version")
		result[name] = BinaryDescription{
			Path:    path,
			Version: version,
		}
	}
	{
		name := "reducestats"
		path := FindBinary("ppa-reducestats")
		version := ""
		if path != "" {
			version = GetStdout(path, "--version")
		}
		result[name] = BinaryDescription{
			Path:    path,
			Version: version,
		}
	}
	{
		name := "pa-wsgo"
		path, err := os.Executable()
		if err != nil {
			log.Printf("Error from os.Executable(): %+v\n", err)
			path = "<err>"
		}
		version := Version
		result[name] = BinaryDescription{
			Path:    path,
			Version: version,
		}
	}

	return result
}

func VerifyBinary(label string, path string) {
	log.Printf("Verifying %s: '%s'\n", label, path)
	{
		log.Printf("which %s: \n", path)
		out, err := exec.Command("which", path).Output() // .CombinedOutput()?
		if err != nil {
			log.Printf("%s\n", err)
			log.Println("PATH:", os.Getenv("PATH"))
		} else {
			log.Println(string(out))
		}
	}
	{
		log.Printf("%s version:", path)
		out, err := exec.Command(path, "--version").Output()
		if err != nil {
			log.Printf("%s", err)
			check(err)
		} else {
			log.Println(string(out))
		}
	}
}
func VerifyBinaries(tc BinaryPaths) {
	log.Printf("Verifying binaries.\n")
	VerifyBinary("Binary_baz2bam", tc.Binary_baz2bam)
	VerifyBinary("Binary_pa_cal", tc.Binary_pa_cal)
	VerifyBinary("Binary_smrt_basecaller", tc.Binary_smrt_basecaller)
	//VerifyBinary("Binary_reducestats", tc.Binary_reducestats)
}

func GetModificationTime(fn string) (time.Time, error) {
	file, err := os.Stat(fn)
	if err != nil {
		return time.Time{}, errors.Wrapf(err, "Failed to stat %q", fn)
	}
	return file.ModTime(), nil
}
