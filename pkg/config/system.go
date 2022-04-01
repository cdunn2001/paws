package config

import (
	"fmt"
	"log"
	"os"
	"os/exec"
)

func check(e error) {
	if e != nil {
		panic(fmt.Sprintf("Stacktrace: %+v", e))
	}
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
