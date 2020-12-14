package orapass

import (
	"fmt"
	"runtime"
	"testing"
)

/*
localhost:1521:emp:scott:tiger
localhost:1521:*:scott:lion
localhost:1522:emp:scott:bear
localhost:*:emp:scott:cat
otherhost:*:emp:scott:elephant
otherhost:1521:newemp:eve:seal
*:1521:oldemp:alice:wolf

type Parser struct {
	Host        string
	Port        string
	DbName      string
	Username    string
	Password    string
	OrapassFile string
}
*/

func debug() bool {
	return true
}

// Test001 test for missing orapass file
// This may still find and orapass file so checking for an error isn't useful.
func Test001(t *testing.T) {

	var p Parser
	p.Host = "localhost"
	p.Port = "1521"
	p.DbName = "emp"
	p.Username = "scott"
	p.OrapassFile = ""
	p.Debug = debug()

	_, _ = p.GetPasswd()
}

// Test002 test for directory as orapass
// This may still find and orapass file so checking for an error isn't useful.
func Test002(t *testing.T) {

	var p Parser
	p.Host = "localhost"
	p.Port = "1521"
	p.DbName = "emp"
	p.Username = "scott"
	p.OrapassFile = ".git"
	p.Debug = debug()

	_, _ = p.GetPasswd()
}

// Test003 test invalid orapass file permissions
func Test003(t *testing.T) {

	switch runtime.GOOS {
	case "windows":
		// nada

	default:
		var p Parser
		p.Host = "localhost"
		p.Port = "1521"
		p.DbName = "emp"
		p.Username = "scott"
		p.OrapassFile = "run_test.sh"
		p.Debug = debug()

		_, err := p.GetPasswd()
		if err == nil {
			t.Errorf("FAILED: should have complained about file permissions.\n")
		}
	}
}

// Test004 test parsing of valid orapass file
func Test004(t *testing.T) {

	// If the field is specified it should be able to either exact match or match a wild-card in the file
	// But if the field is not specified, should we expect to fill in based on that which does match?
	// It would be nice to be able to just specify the database and have the rest populated from the file...
	// ...seems error prone however and may not get the expected match.
	// Perhaps if the database name is required and the others are optional?

	cases := []struct {
		id         int
		Host       string
		Port       string
		DbName     string
		Username   string
		Password   string
		shouldPass bool
	}{
		{1, "localhost", "1521", "emp", "scott", "tiger", true},       // no error, password should match
		{2, "localhost", "1521", "emp", "walter", "", false},          // no such user
		{3, "localhost", "1521", "nosuchdb", "scott", "lion", true},   // match on wild-card db name
		{4, "localhost", "1522", "emp", "scott", "bear", true},        //
		{5, "localhost", "", "emp", "scott", "tiger", false},          // no port specified
		{6, "localhost", "1234", "emp", "scott", "cat", true},         // match on wild-card port
		{7, "otherhost", "1523", "", "scott", "elephant", false},      // no database specified
		{8, "otherhost", "1521", "anydb", "scott", "elephant", false}, // invalid database specified
		{9, "", "", "newemp", "", "", false},                          // only database specified
		{10, "newhost", "", "oldemp", "", "", false},                  // only database and host specified,
		{11, "", "", "oldemp", "", "", false},                         // only database specified
	}
	for _, c := range cases {

		if debug() {
			if c.shouldPass {
				fmt.Printf("\nStarting test case %d:\n", c.id)
			} else {
				fmt.Printf("\nStarting test case %d (should fail):\n", c.id)
			}
		}

		var sent Parser
		sent.Host = c.Host
		sent.Port = c.Port
		sent.DbName = c.DbName
		sent.Username = c.Username
		sent.OrapassFile = "orapass" // we want the local testing item, not one that the user is actually using...
		sent.Debug = debug()

		got, err := sent.GetPasswd()

		// don't actually send the password but set it here for comparing
		sent.Password = c.Password

		if c.shouldPass {
			switch {
			case err != nil:
				t.Errorf("FAILED on %d: should not have received error, sent(%s) got %q.\n", c.id, dump_fields(sent), err)
			case !compare_fields(c.id, sent, got):
				t.Errorf("FAILED on %d: unexpected result. Sent(%s), got(%s).\n", c.id, dump_fields(sent), dump_fields(got))
			}
		} else {
			switch {
			case err == nil:
				t.Errorf("FAILED on %d: should have received error. Sent(%s), got(%s).\n", c.id, dump_fields(sent), dump_fields(got))
			}
		}
	}
}

func compare_fields(caseID int, sent, got Parser) bool {

	if !okayMatch(sent.Host, got.Host) {
		return false
	}
	if !okayMatch(sent.Port, got.Port) {
		return false
	}
	if !okayMatch(sent.DbName, got.DbName) {
		return false
	}
	if !okayMatch(sent.Username, got.Username) {
		return false
	}
	if !okayMatch(sent.Password, got.Password) {
		return false
	}
	return true
}

func dump_fields(p Parser) string {
	return fmt.Sprintf("%q, %q, %q, %q, %q", p.Host, p.Port, p.DbName, p.Username, p.Password)
}

// okayMatch
func okayMatch(sent, got string) bool {
	switch {
	//case sent == "" && got != "":
	// reverse wild-card match?
	//	return true
	case sent == got:
		return true
	}
	return false
}
