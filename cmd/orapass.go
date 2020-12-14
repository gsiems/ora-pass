// Retrieves a database password for an Oracle user.
//
// Parameters:
//  Database - The database may be specified by the ORACLE_SID
//      environment variable or by the -d flag.
//
//  Host - The host for the database may be specified by the
//      ORACLE_HOST environment variable or by the -h flag. If neither the
//      ORACLE_HOST environment variable or -h flag are specified then
//      this defaults to localhost.
//
//  Port - The port (if other than the default 1521) may be
//      specified by the ORACLE_PORT environment variable or the -p flag.
//
//  Username - The username may be specified by the ORACLE_USER
//      environment variable or by the -u flag. If neither are specified
//      then this defaults to the logged in user.
package main

import (
	"flag"
	"fmt"
	"os"

	orap "github.com/gsiems/orapass"
)

func main() {

	var (
		p      orap.Parser
		fQuiet bool
	)
	flag.StringVar(&p.Username, "u", "", "The username to obtain a password for. Overrides the ORACLE_USER environment variable. Defaults to the OS user.")
	flag.StringVar(&p.Host, "h", "", "The hostname that the database is on. Overrides the ORACLE_HOST environment variable. Defaults to localhost.")
	flag.StringVar(&p.Port, "p", "", "The port that the database is listening on. Overrides the ORACLE_PORT environment variable. Defaults to 1521.")
	flag.StringVar(&p.DbName, "d", "", "The database to connect to. Overrides the ORACLE_SID environment variable.")
	flag.StringVar(&p.OrapassFile, "f", "", "The orapass file to search for first.")
	flag.BoolVar(&p.Debug, "debug", false, "Debug mode.")
	flag.BoolVar(&fQuiet, "q", false, "Quiet mode. Do not print any error messages.")

	flag.Parse()

	p2, err := p.GetPasswd()
	if err != nil {
		if !fQuiet {
			os.Stderr.WriteString(fmt.Sprintf("%s\n", err))
		}
		os.Exit(1)
	}
	fmt.Println(p2.Password)
}
