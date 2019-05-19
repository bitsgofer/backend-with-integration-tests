#/bin/bash

HOSTNAME=$1
PORT=$2

set -e

echo "waiting for $HOSTNAME:$PORT to open"
while true
do
	nc -zv $HOSTNAME $PORT
	if [ "$?" == "0" ]
	then
		exit 0
	fi

	sleep 1
done

exit 0
