# orapass command

Retrieves a database password for an Oracle user.

Parameters:

 * -d: Use the -d flag to specify the database name. If the database
    name is not specified then the ORACLE_SID environment variable is
    used. There is no default value so the database name must be
    specified using either the -d flag or the ORACLE_SID environment
    variable.

 * -h: Use the -h flag to specify the host that the database resides
    on. If the host is not specified then the ORACLE_HOST environment
    variable is used. If neither is specified then this defaults to
    localhost.

 * -p: Use the -u flag to specify the port for the oracle listener. If
     the port is not specified then the ORACLE_PORT environment
     variable is used. If neither is specified then this defaults to
     1521.

 * -u: Use the -u flag to specify the database user. If the user is not
    specified then the ORACLE_USER environment variable is used. If
    neither is specified then this defaults to the logged in user.

 * -f: Use the -f flag to specify the orapass file to search for first.
    If not specified (or found) then the file specified by the
    ORAPASSFILE environment variable is searched for next. If neither
    is specified (or found) then the search continues. Run with the
    debug flag to see the file locations searched for.

 * -q: Use the -q flag to quiet any error messages.

 * -debug: Use the -debug flag to print debug information.
