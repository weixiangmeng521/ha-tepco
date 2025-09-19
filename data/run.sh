#!/usr/bin/with-contenv bashio
set -e

while true; do
    username=$(bashio::config 'username')
    password=$(bashio::config 'password')

    # set env
    export USERNAME="$username"
    export PASSWORD="$password"

    # excute
    /usr/bin/tepco -u "$username" -p "$password"
    sleep 3600
done
 