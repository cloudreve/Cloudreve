#!/bin/bash

file=$1
if [ "$file" == "" ]
then
    file=compose.yaml
fi

docker-compose version
docker-compose -f "$file" up --build -d --pull=never

while true
do
    docker-compose -f "$file" ps | grep testing
    if [ $? -eq 1 ]
    then
        code=-1
        docker-compose -f "$file" logs | grep e2e-cloudreve
        docker-compose -f "$file" logs | grep e2e-testing
        docker-compose ps -a | grep e2e-testing | grep "Exited (0)"
        if [ $? -eq 0 ]
        then
            code=0
            echo "successed"
        fi

        docker-compose -f "$file" down
        set -e
        exit $code
    fi
    sleep 1
done
