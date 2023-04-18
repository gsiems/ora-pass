# orapass

An Oracle password file parsing library and utility based on the
.pgpass file available to PostgreSQL users, orapass searches a
similarly formatted file for Oracle password information. The intent is
to provide an alternative to hard-coding password information in
scripts, applications, or config files. Keeping the Oracle password
information in a separate file also lessens the chances of having
password information showing up in source-code control systems.

The orapass file is a colon separated file consisting of one line
per entry where each entry has five fields:

     host:port:database(SID):username:password

Each of the first four fields can be a case-insensitive literal value
or "*" which acts as a match-anything-wildcard. Blank lines and lines
that are commented out using the "#" character are ignored.

When parsing, the first matching line is used.

The location of the orapass file is defermined by first checking for
the ORAPASSFILE environment variable. If there is no ORAPASSFILE
environment variable or if the ORAPASSFILE does not point to a valid
file then:

 * on Unix-like OSes, orapass looks for a $HOME/.orapass or
 $HOME/orapass file. Additionally the file must be chmod 600 or orapass
 will refuse to use it.

 * on Windows, orapass looks for a APPDATA/oracle/.orapass or
 APPDATA/oracle/orapass file.

# Example

osql.sh is an example bash script that uses orapass to automate connecting
to Oracle using sqlcl.
