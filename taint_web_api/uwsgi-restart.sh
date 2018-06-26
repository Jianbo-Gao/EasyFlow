#!/bin/bash

process_num=`ps aux | grep uwsgi-evm-taint | grep -v grep | wc -l`
if [[ $process_num -ne "0" ]]
then
    ps aux | grep uwsgi-evm-taint | grep -v grep | cut -c 9-15 | xargs kill
	sleep 3
fi

uwsgi uwsgi-evm-taint.ini