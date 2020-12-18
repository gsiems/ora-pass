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
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/phayes/permbits" // MIT license
)

// Parser contains the fields for the orapass file
type Parser struct {
	Host        string
	Port        string
	DbName      string
	Username    string
	Password    string
	OrapassFile string
	files       []string
	Debug       bool
}

// GetPasswd retrieves the password for the specified (host, port,
// database, username).
func (p *Parser) GetPasswd() (Parser, error) {

	var osUser string
	usr, err := user.Current()
	if err == nil {
		osUser = usr.Username
	}

	p.Host = coalesce([]string{p.Host, os.Getenv("ORACLE_HOST"), "localhost"})
	p.Port = coalesce([]string{p.Port, os.Getenv("ORACLE_PORT"), "1521"})
	p.DbName = coalesce([]string{p.DbName, os.Getenv("ORACLE_SID")})
	p.Username = coalesce([]string{p.Username, os.Getenv("ORACLE_USER"), osUser})

	var p2 Parser
	err = p.findPasswordFile()
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
	return p.searchFile()
}

// findPasswordFile searches for an orapass file and returns the first one found
func (p *Parser) findPasswordFile() error {

	p.appendFileList(p.OrapassFile)
	p.appendFileList(os.Getenv("ORAPASSFILE"))

	switch runtime.GOOS {
	case "windows":
		//os.Getenv("APPDATA") or maybe os.Getenv("LOCALAPPDATA")
		dir := os.Getenv("APPDATA")
		p.appendFileList(filepath.Join(dir, "oracle", ".orapass"))
		p.appendFileList(filepath.Join(dir, "oracle", "orapass"))

	default:
		dir := os.Getenv("HOME")
		p.appendFileList(filepath.Join(dir, ".orapass"))
		p.appendFileList(filepath.Join(dir, "orapass"))
	}

	for _, f := range p.files {
		ok, err := p.fileExists(f)
		switch {
		case err != nil:
			return err
		case ok:
			p.OrapassFile = f
			return nil
		}
	}

	p.carp("No orapass file found")
	return nil
}

func (p *Parser) appendFileList(f string) {
	if f != "" {
		p.carp(fmt.Sprintf("Adding %q to search list", f))
		p.files = append(p.files, f)
	}
}

// checkFilePerms verifies the permissions on the orapass file
func (p *Parser) checkFilePerms() (bool, error) {

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
func (p *Parser) searchFile() (Parser, error) {

	var p2 Parser
	p.carp(fmt.Sprintf("Searching %q for %s/%s", p.OrapassFile, p.Username, p.DbName))

	re := regexp.MustCompile("^ *#")

	file, err := os.Open(p.OrapassFile)
	if err != nil {
		return p2, err
	}
	defer file.Close()

	i := 0
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		i++

		line := scanner.Text()

		// Ignore commented outlines
		if re.MatchString(line) {
			continue
		}

		p.carp(fmt.Sprintf("    Parsing line %d", i))
		tokens := strings.SplitN(line, ":", 5)
		if len(tokens) < 5 {
			continue
		}

		hostMatch := p.chkForMatch(p.Host, tokens[0])
		portMatch := p.chkForMatch(p.Port, tokens[1])
		dbNameMatch := p.chkForMatch(p.DbName, tokens[2])
		userMatch := p.chkForMatch(p.Username, tokens[3])

		if !hostMatch {
			p.carp("        Host does not match")
		}
		if !portMatch {
			p.carp("        Port does not match")
		}
		if !dbNameMatch {
			p.carp("        DB name does not match")
		}
		if !userMatch {
			p.carp("        Username does not match")
		}

		if hostMatch && portMatch && dbNameMatch && userMatch {
			p.carp("        Match detected")

			p2.Host = p.pickParm(p.Host, tokens[0])
			p2.Port = p.pickParm(p.Port, tokens[1])
			p2.DbName = p.pickParm(p.DbName, tokens[2])
			p2.Username = tokens[3]
			p2.Password = tokens[4]

			return p2, nil
		}
	}
	if err = scanner.Err(); err != nil {
		return p2, err
	}

	err = errors.New("Could not find a suitable password entry")
	return p2, err
}

// fileExists checks to ensure that the specified file exists and is a regular file
func (p *Parser) fileExists(pathname string) (bool, error) {

	p.carp(fmt.Sprintf("Looking for file %q", pathname))
	fi, err := os.Stat(pathname)
	if err != nil {
		// For our purposes, a non-existent file is not considered an error
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	p.carp(fmt.Sprintf("Found %q", pathname))
	switch mode := fi.Mode(); {
	case mode.IsRegular():
		return true, nil
	}

	p.carp(fmt.Sprintf("%q is not a regular file", pathname))
	return false, nil
}

// chkForMatch checks the calling parameter against the same file
// parameter, taking into account wild-card characters and returns true
// on a match
func (p *Parser) chkForMatch(callingParm, fileParm string) bool {
	switch {
	case strings.ToUpper(callingParm) == strings.ToUpper(fileParm):
		return true
	case fileParm == "*" && callingParm != "":
		return true
	}
	return false
}

// pickParm chooses between the calling parameter and file parameter
// and returns the appropriate value
func (p *Parser) pickParm(callingParm, fileParm string) string {
	if fileParm != "*" && fileParm != "" {
		return fileParm
	}
	return callingParm
}

func (p *Parser) carp(s string) {
	if s != "" {
		if p.Debug {
			os.Stderr.WriteString(s)
		}
	}
}

// coalesce picks the first non-empty string from a list
func coalesce(s []string) string {
	for _, v := range s {
		if v != "" {
			return v
		}
	}
	return ""
}
