#!/bin/bash

# switch to langfang environment
# sudo /opt/tiger/consul_devbox/bin/switch.sh huailai 

for hall in "lf" "hl" "lq"
do
cmd="sd lookup toutiao.clientqa.itcserver.service.${hall}"
echo ${cmd}

for url in $( ${cmd} | sed "1,5d" | awk '{printf "%s:%s ",$1,$2}' )
do
echo ${url}
curl ${url}/settings GET

echo "Refresh online instance(${ur}) settings..."
curl ${url}/settings -X PUT
echo "Refresh online instance(${ur}) settings... DONE"
done

done

# switch to boe environment
# sudo /opt/tiger/consul_devbox/bin/switch.sh boe