package orapass

import (
	"testing"
)

/*
localhost:1521:emp:scott:tiger
localhost:1521:*:scott:lion
localhost:1522:*:scott:bear
otherhost:*:*:scott:elephant

type PwRec struct {
	Host        string
	Port        string
	DbName      string
	Username    string
	Password    string
	OrapassFile string
}
*/

// Test001 test for an "everything is specified" existing entry
func Test001(t *testing.T) {

	var p PwRec
	p.Host = "localhost"
	p.Port = "1521"
	p.DbName = "emp"
	p.Username = "scott"
	p.OrapassFile = "orapass" // we want the local testing item, not one that the user is actually using...

	_, err := p.GetPasswd(true)
	if err != nil {
		t.Errorf("ERR: %q\n", err)
	}
}
