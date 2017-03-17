#!/bin/bash
# This script automaticaly copies project folder to new computer, and run the program
clear
user="student"
echo "Type in the workstation number to start"
read workstationNumber

if [ $workstationNumber == "1" ]; then
  IP=140
elif [ $workstationNumber == "2" ]; then
  IP=149
elif [ $workstationNumber == "3" ]; then
  IP=150
elif [ $workstationNumber == "4" ]; then
  IP=141
elif [ $workstationNumber == "5" ]; then
  IP=143
elif [ $workstationNumber == "6" ]; then
  IP=146
elif [ $workstationNumber == "7" ]; then
  IP=154
elif [ $workstationNumber == "8" ]; then
  IP=161
elif [ $workstationNumber == "9" ]; then
  IP=156
elif [ $workstationNumber == "10" ]; then
  IP=158
elif [ $workstationNumber == "11" ]; then
  IP=159
elif [ $workstationNumber == "12" ]; then
  IP=144
elif [ $workstationNumber == "13" ]; then
  IP=152
elif [ $workstationNumber == "14" ]; then
  IP=142
elif [ $workstationNumber == "15" ]; then
  IP=148
elif [ $workstationNumber == "16" ]; then
  IP=147
elif [ $workstationNumber == "17" ]; then
  IP=145
elif [ $workstationNumber == "18" ]; then
  IP=151
elif [ $workstationNumber == "19" ]; then
  IP=157
elif [ $workstationNumber == "20" ]; then
  IP=155
elif [ $workstationNumber == "21" ]; then
  IP=153
elif [ $workstationNumber == "22" ]; then
  IP=38
elif [ $workstationNumber == "23" ]; then
  IP=48
elif [ $workstationNumber == "24" ]; then
  IP=46
elif [ $workstationNumber == "25" ]; then
  IP=43
else
  echo "Not a valid workstation"
  exit 1
fi

# Create new rsa key
#ssh-keygen -t rsa # uncomment when starting from a new computer

# Send RSA key to remote conputer
cat ~/.ssh/id_rsa.pub | ssh $user@129.241.187.$IP 'cat >> .ssh/authorized_keys'

##echo "Connecting to Workstation" $workstationNumber "at 129.241.187."$IP
ssh-copy-id -i ~/.ssh/id_rsa.pub  $user@129.241.187.$IP
echo "Delete old files and folder"
#ssh $user@129.241.187.$IP 'rm -rf ~/work/src/github.com/andersliland/ttk4145-project/'
echo "Create new folder path"
#ssh $user@129.241.187.$IP 'mkdir -p ~/work/src/github.com/andersliland/ttk4145-project' # create directory path if it does not exsist
echo "Copy project content"
#scp -rq ~/work/src/github.com/andersliland/ttk4145-project/. $user@129.241.187.$IP:~/work/src/github.com/andersliland/ttk4145-project &>/dev/null
echo 'SSH into remote and execute go run main.go'
#ssh -t $user@129.241.187.$IP "cd /home/student/work/src/github.com/andersliland/ttk4145-project/ && ./setupRemote.sh ; bash"

clear
echo "Copy project content to workstation 4,5,6"
scp -rq ~/work/src/github.com/andersliland/ttk4145-project/. $user@129.241.187.141:~/work/src/github.com/andersliland/ttk4145-project &>/dev/null
scp -rq ~/work/src/github.com/andersliland/ttk4145-project/. $user@129.241.187.143:~/work/src/github.com/andersliland/ttk4145-project &>/dev/null
scp -rq ~/work/src/github.com/andersliland/ttk4145-project/. $user@129.241.187.146:~/work/src/github.com/andersliland/ttk4145-project &>/dev/null
echo "Finished"
