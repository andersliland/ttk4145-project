#!/bin/bash
# This script automaticaly copies project folder to new computer, and does a remote login afterwords
clear
user="student"
echo "Type in the workstation number to start"
read workstationNumber
if [ $workstationNumber == "10" ]; then
  IP=155
elif [ $workstationNumber == "11" ]; then
  IP="Na"
  echo "Not a valid workstation"
elif [ $workstationNumber == "12" ]; then
  IP=144
elif [ $workstationNumber == "13" ]; then
  IP=152
elif [ $workstationNumber == "14" ]; then
  IP=142
elif [ $workstationNumber == "15" ]; then
  IP=148
else
  echo "Not a valid workstation"
  exit 1
fi

# Send RSA key to remote conputer
#cat ~/.ssh/id_rsa.pub | ssh $user@129.241.187.$IP 'cat >> .ssh/authorized_keys'

echo "Connecting to Workstation" $workstationNumber "at 129.241.187."$IP
ssh-copy-id $user@129.241.187.$IP
echo "Delete old files and folder"
ssh $user@129.241.187.$IP 'rm -rf ~/work/src/github.com/andersliland/ttk4145-project/'
echo "Create new folder path"
ssh $user@129.241.187.$IP 'mkdir -p ~/work/src/github.com/andersliland/ttk4145-project' # create directory path if it does not exsist
echo "Copy project content"
scp -rq $GOPATH/src/github.com/andersliland/ttk4145-project/. $user@129.241.187.$IP:~/work/src/github.com/andersliland/ttk4145-project &>/dev/null
echo 'SSH into remote and execute go run main.go'
ssh -t $user@129.241.187.$IP "cd /home/student/work/src/github.com/andersliland/ttk4145-project/ && ./setupRemote.sh ; bash"
