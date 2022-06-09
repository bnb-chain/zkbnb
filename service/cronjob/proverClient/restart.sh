# restart.sh 

# get port from .yaml
PORT=`cat ./etc/globalrpc.yaml | grep "ListenOn" |awk -F ':' '{print $3}'` 
echo "service Port":$PORT 

# get PID according to port
PID=` lsof -ti:${PORT} ` 
echo "service PID":$PID 

# kill this process
kill -9 $PID

# start service
nohup ./main -f ./etc/*.yaml > ./log/log.file 2>&1 &