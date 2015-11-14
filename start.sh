#! /bin/bash
source ./config
if [ "$1" != "" ]
then
 export CS4032_LAB_4_PORT=$1
fi
./server
