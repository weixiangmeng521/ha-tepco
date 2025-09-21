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
    username=$(bashio::config 'username')
    password=$(bashio::config 'password')

    # set env
    export USERNAME="$username"
    export PASSWORD="$password"

    # excute
    /usr/bin/tepco -u "$username" -p "$password"
    sleep 3600
done
 