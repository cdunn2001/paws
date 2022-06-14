package web

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"
)

// If a file of this name exists, then move it to something that does not.
// If this is actually a symlink, remove the symlink.
func MoveExistingLogfile(specified string) {
	fi, err := os.Lstat(specified)
	if err == nil {
		if fi.Mode()&os.ModeSymlink == os.ModeSymlink {
			// This is a symlink, so remove just the symlink.
			err := os.Remove(specified)
			if err != nil {
				fmt.Printf("FATAL: Could not remove symlink of logfile %q: %+v\n",
					specified, err)
				check(err)
			}
		} else {
			// Not a symlink. Must have been created by older version of paws.
			fmt.Printf("ERROR: Old version of paws? Renaming logfile from %q\n",
				specified)
			// Choose a new name and move this file to it.
			newname := ChooseLoggerFilenameLegacy(specified)
			err := os.Rename(specified, newname)
			if err != nil {
				fmt.Printf("ERROR: Could not rename logfile from %q to %q: %+v\nLost old logfile.\n",
					specified, newname, err)
			}
		}
	} else if errors.Is(err, fs.ErrNotExist) {
		// No problem.
	} else {
		fmt.Printf("FATAL: Unexpected error testing logfile %q: %+v\n",
			specified, err)
		check(err)
	}
}

// Caller must eventually call 'result.Close()'.
func RotateLogfile(userfn string) (result *os.File) {
	MoveExistingLogfile(userfn)
	fn := ChooseLoggerFilename(userfn)
	{
		err := os.Symlink(filepath.Base(fn), userfn)
		if err != nil {
			fmt.Printf("ERROR: Failed to create convenient symlink from %q to %q: %+v\nContinuing.",
				fn, userfn, err)
		}
	}
	fmt.Printf("Logging to '%s'\n", fn)
	result, err := os.Create(fn)
	check(err)
	return result
}
func ChooseLoggerFilename(existing string) string {
	return chooseLoggerFilenameTestable(existing, time.Now().UTC(), os.Getpid())
}
func chooseLoggerFilenameTestable(existing string, now time.Time, pid int) string {
	dir, oldbasename := filepath.Split(existing)
	oldpre, oldext := SplitExt(oldbasename)
	// canonical time for layout: "Jan 2 15:04:05 2006 MST"
	layout := "06-01-02"
	datetime := now.Format(layout)
	newbasename := fmt.Sprintf("%s.%s.%d%s",
		oldpre, datetime, pid, oldext)
	return filepath.Join(dir, newbasename)
}

// For a logfile generated by older paws, we want to move it to a new name.
// Do not use PID in that new name.
func ChooseLoggerFilenameLegacy(existing string) string {
	return chooseLoggerFilenameLegacyTestable(existing, GetCtime(existing))
}
func chooseLoggerFilenameLegacyTestable(existing string, ctime time.Time) string {
	dir, oldbasename := filepath.Split(existing)
	oldpre, oldext := SplitExt(oldbasename)
	// canonical time for layout: "Jan 2 15:04:05 2006 MST"
	layout := "06-01-02"
	datetime := ctime.Format(layout)
	newbasename := fmt.Sprintf("%s.%s%s",
		oldpre, datetime, oldext)
	return filepath.Join(dir, newbasename)
}