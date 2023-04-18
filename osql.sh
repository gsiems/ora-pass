#!/bin/bash
########################################################################
#
# osql -d [ -h ] [ -p ] [ -u ]
#
# Use pgpass to simplify connecting to Oracle using sqlcl
#
########################################################################

sqlclBin="$HOME/bin/sqlcl/bin/sql"

function usage() {
    echo "Usage: $(basename $0) -d [-u] [-p]" 2>&1
    echo '   -d   the database to connect to'
    echo '   -h   the hostname of the database to connect to (default to DNS lookup then localhost)'
    echo '   -p   the port to connect on (default 1521)'
    echo '   -u   the username to connect as (default OS user)'
    exit 1
}

if [[ ${#} -eq 0 ]]; then
    usage
fi

while getopts ':d:h:p:u' opt; do
    case "$opt" in
        d)
            dbname="${OPTARG}"
            ;;
        u)
            username="${OPTARG}"
            ;;
        p)
            port="${OPTARG}"
            ;;
        h)
            host="${OPTARG}"
            ;;
        *)
            echo "Invalid option $opt"
            echo
            usage
            ;;
    esac
done

if [[ -z $dbname ]]; then
    dbname=$1
fi

if [[ -z $dbname ]]; then
    echo "No database specified. Did you forget the -d switch?"
    echo
    usage
fi
dbname=$(echo ${dbname} | tr "[a-z]" "[A-Z]")

if [[ -z $username ]]; then
    username=$(whoami)
fi

if [[ -z $port ]]; then
    port=1521
fi

if [[ -z $host ]]; then

    # If no host was specified then look for a DNS cname for the database
    dbHost=$(host ${dbname})
    ret=$?

    if ((ret == 0)); then
        hostName=$(host ${dbname} | tail -n 1 | cut -d ' ' -f 1)
        hostName=$(echo ${hostName} | tr "[A-Z]" "[a-z]")
    else
        # no DNS entry found
        hostName='localhost'
    fi

fi

cmd="orapass -d ${dbname} -u ${username} -h ${hostName} -p ${port} -q"
passwd=$($cmd)

if [[ -z $passwd ]]; then
    echo "Unable to determine a password for ${username}@${hostName}:${port}:${dbname}"
    exit 1
fi

echo "Connecting to ${username}@${hostName}:${port}:${dbname}"

cmdFile=$(mktemp -t)
touch ${cmdFile}
chmod 600 ${cmdFile}
echo "connect ${username}/${passwd}@${hostName}:${port}:${dbname}" >$cmdFile
echo "! rm $cmdFile" >>$cmdFile

${sqlclBin} /NOLOG @${cmdFile}
