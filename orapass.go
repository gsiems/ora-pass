// Package orapass retrieves Oracle passwords from a PostgreSQL
// inspired .pgpass-like file so that they do not need to be hard-coded
// in scripts, applications, or config files.
//
// The orapass file is a colon separated file consisting of one line
// per entry where each entry has five fields:
//
//      host:port:database(SID):username:password
//
// Each of the first four fields can be a case-insensitive literal
// value or "*" which acts as a match-anything-wildcard.
// Blank and commented out lines are ignored.
package orapass

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/phayes/permbits" // MIT license
)

// PwRec contains the fields for the orapass file
type PwRec struct {
	Host        string
	Port        string
	DbName      string
	Username    string
	Password    string
	OrapassFile string
}

// GetPasswd retrieves the password for the specified (host, port,
// database, username).
func (p *PwRec) GetPasswd(verbose bool) (PwRec, error) {

	var p2 PwRec
	err := p.findPasswordFile(verbose)
	if err != nil {
		return p2, err
	}

	switch {
	case runtime.GOOS != "windows":
		_, err = p.checkFilePerms()
		if err != nil {
			return p2, err
		}
	}
	return p.searchFile(verbose)
}

// findPasswordFile searches for an orapass file and returns the first one found
func (p *PwRec) findPasswordFile(verbose bool) error {

	var files []string

	if p.OrapassFile != "" {
		files = append(files, p.OrapassFile)
	}

	f := os.Getenv("ORAPASSFILE")
	if f != "" {
		carp(verbose, "ORAPASSFILE environment variable is defined")
		files = append(files, f)
	}

	switch runtime.GOOS {
	case "windows":
		//os.Getenv("APPDATA") or maybe os.Getenv("LOCALAPPDATA")
		dir := os.Getenv("APPDATA")
		files = append(files, filepath.Join(dir, "oracle", ".orapass"))
		files = append(files, filepath.Join(dir, "oracle", "orapass"))

	default:
		dir := os.Getenv("HOME")
		files = append(files, filepath.Join(dir, ".orapass"))
		files = append(files, filepath.Join(dir, "orapass"))
	}

	for _, f := range files {
		ok, err := fileExists(f, verbose)
		switch {
		case err != nil:
			return err
		case ok:
			p.OrapassFile = f
			return nil
		}
	}

	carp(verbose, "No orapass file found")
	return nil
}

// checkFilePerms verifies the permissions on the orapass file
func (p *PwRec) checkFilePerms() (bool, error) {

	permissions, err := permbits.Stat(p.OrapassFile)
	if err != nil {
		return false, err
	}
	if permissions != 0600 {
		errstr := fmt.Sprintf(
			"Permissions on file are incorrect. Should be 600, got %o",
			permissions)
		err = errors.New(errstr)
		return true, err
	}
	return true, nil
}

// searchFile searches the orapass file for a matching entry. If more
// than one entry could match then the first matching entry is returned.
func (p *PwRec) searchFile(verbose bool) (PwRec, error) {

	var p2 PwRec
	carp(verbose, fmt.Sprintf("Reading %q", p.OrapassFile))

	dat, err := ioutil.ReadFile(p.OrapassFile)
	if err != nil {
		return p2, err
	}

	re := regexp.MustCompile("^ *#")
	s := string(dat[:])
	lines := strings.Split(s, "\n")
	for i, line := range lines {

		// Ignore commented outlines
		if re.MatchString(line) {
			continue
		}

		carp(verbose, fmt.Sprintf("    Parsing line %d", i))
		tokens := strings.SplitN(line, ":", 5)
		if len(tokens) < 5 {
			continue
		}

		hostMatch := chkForMatch(p.Host, tokens[0])
		portMatch := chkForMatch(p.Port, tokens[1])
		dbNameMatch := chkForMatch(p.DbName, tokens[2])
		userMatch := chkForMatch(p.Username, tokens[3])

		carp(verbose, fmt.Sprintf("        hostMatch is %t", hostMatch))
		carp(verbose, fmt.Sprintf("        portMatch is %t", portMatch))
		carp(verbose, fmt.Sprintf("        dbNameMatch is %t", dbNameMatch))
		carp(verbose, fmt.Sprintf("        userMatch is %t", userMatch))


		if hostMatch && portMatch && dbNameMatch && userMatch {

			p2.Host = pickParm(p.Host, tokens[0])
			p2.Port = pickParm(p.Port, tokens[1])
			p2.DbName = pickParm(p.DbName, tokens[2])
			p2.Username = tokens[3]
			p2.Password = tokens[4]

			return p2, nil
		}
	}

	err = errors.New("Could not find a suitable password entry")
	return p2, err
}

// fileExists checks to ensure that the specified file exists and is a regular file
func fileExists(pathname string, verbose bool) (bool, error) {

	carp(verbose, fmt.Sprintf("Looking for file %q", pathname))
	fi, err := os.Stat(pathname)
	if err != nil {
		// For our purposes, a non-existent file is not considered an error
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	carp(verbose, fmt.Sprintf("Found %q", pathname))
	switch mode := fi.Mode(); {
	case mode.IsRegular():
		return true, nil
	}

	carp(verbose, fmt.Sprintf("%q is not a regular file", pathname))
	return false, nil
}

// chkForMatch checks the calling parameter against the same file
// parameter, taking into account wild-card characters and returns true
// on a match
func chkForMatch(callingParm, fileParm string) bool {
	if strings.ToUpper(callingParm) == strings.ToUpper(fileParm) {
		return true
	} else if fileParm == "*" {
		return true
	}
	return false
}

// pickParm chooses between the calling parameter and file parameter
// and returns the appropriate value
func pickParm(callingParm, fileParm string) string {
	if fileParm != "*" && fileParm != "" {
		return fileParm
	}
	return callingParm
}

func carp(verbose bool, s string) {
	if verbose {
		fmt.Fprintln(os.Stderr, s)
	}
}
