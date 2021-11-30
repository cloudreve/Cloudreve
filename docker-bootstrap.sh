#!/bin/sh
GREEN='\033[0;32m'
RESET='\033[0m'
if [ ! -f /etc/cloudreve/aria2c.conf ]; then
    echo -e "[${GREEN}aria2c${RESET}] aria2c config not found. Generating..."
    secret=$(tr -dc A-Za-z0-9 </dev/urandom | head -c 13)
    echo -e "[${GREEN}aria2c${RESET}] Generated port: 6800, secret: $secret"
    cat <<EOF > /etc/cloudreve/aria2c.conf
enable-rpc=true
rpc-listen-port=6800
rpc-secret=$secret
EOF
fi
aria2c --conf-path /etc/cloudreve/aria2c.conf -D
cloudreve
