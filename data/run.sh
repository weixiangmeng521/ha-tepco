#!/usr/bin/with-contenv bashio
set -e
#  ______  ____    ____    ____     _____      
# /\__  _\/\  _`\ /\  _`\ /\  _`\  /\  __`\    
# \/_/\ \/\ \ \L\_\ \ \L\ \ \ \/\_\\ \ \/\ \   
#    \ \ \ \ \  _\L\ \ ,__/\ \ \/_/_\ \ \ \ \  
#     \ \ \ \ \ \L\ \ \ \/  \ \ \L\ \\ \ \_\ \ 
#      \ \_\ \ \____/\ \_\   \ \____/ \ \_____\
#       \/_/  \/___/  \/_/    \/___/   \/_____/
                                             
                                             
echo " ______  ____    ____    ____     _____      "
echo "/\\__  _\\/\\  _\`\\ /\\  _\`\\ /\\  _\`\\  /\\  __\`\\    "
echo "\\/_/\\ \\/\\ \\ \\L\\_\\ \\ \\L\\ \\ \\ \\/\\_\\\\ \\ \\/\\ \\   "
echo "   \\ \\ \\ \\ \\  _\L\\ \\ ,__/\\ \\ \\/_/_\\ \\ \\ \\ \\  "
echo "    \\ \\ \\ \\ \\ \\L\\ \\ \\ \\/  \\ \\ \\L\\ \\\\ \\ \\_\\ \\ "
echo "     \\ \\_\\ \\ \\____/\\ \\_\\   \\ \\____/ \\ \\_____/ "
echo "      \\/_/  \\/___/  \\/_/    \\/___/   \\/_____/ "

while true; do
    CONFIG_USERNAME=$(bashio::config 'username')
    CONFIG_PASSWORD=$(bashio::config 'password')

    if (bashio::config.is_empty 'mqtt' || ! (bashio::config.has_value 'mqtt.server' || bashio::config.has_value 'mqtt.username' || bashio::config.has_value 'mqtt.password')) && bashio::var.has_value "$(bashio::services 'mqtt')"; then
        if bashio::var.true "$(bashio::services 'mqtt' 'ssl')"; then
            export TEPCO2MQTT_CONFIG_MQTT_SERVER="mqtts://$(bashio::services 'mqtt' 'host'):$(bashio::services 'mqtt' 'port')"
        else
            export TEPCO2MQTT_CONFIG_MQTT_SERVER="mqtt://$(bashio::services 'mqtt' 'host'):$(bashio::services 'mqtt' 'port')"
        fi
        export TEPCO2MQTT_CONFIG_MQTT_USERNAME="$(bashio::services 'mqtt' 'username')"
        export TEPCO2MQTT_CONFIG_MQTT_PASSWORD="$(bashio::services 'mqtt' 'password')"
    fi

    bashio::log.info "Starting Server..."

    export CONFIG_USERNAME
    export CONFIG_PASSWORD
    export TEPCO2MQTT_CONFIG_MQTT_USERNAME
    export TEPCO2MQTT_CONFIG_MQTT_PASSWORD


    echo "======================================================="
    echo "Username: ${CONFIG_USERNAME}"
    echo "Password: ${#CONFIG_PASSWORD}"
    echo "MQTT username: ${TEPCO2MQTT_CONFIG_MQTT_USERNAME}"
    echo "MQTT password length: ${#TEPCO2MQTT_CONFIG_MQTT_PASSWORD}"
    echo "======================================================="

    # excute
    /usr/bin/tepco \
        -u "$CONFIG_USERNAME" \
        -p "$CONFIG_PASSWORD" \
        -mqtt_username "$TEPCO2MQTT_CONFIG_MQTT_USERNAME" \
        -mqtt_password "$TEPCO2MQTT_CONFIG_MQTT_PASSWORD"
        

    sleep 3600
done
 