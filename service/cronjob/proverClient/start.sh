#start.sh

# start service
nohup ./main -f ./etc/*.yaml > ./log/log.file 2>&1 &