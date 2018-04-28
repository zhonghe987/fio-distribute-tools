#!/bin/sh

set -x

mode=$1
bs=$2
size=$3
numjobs=$4
ioengine=$5
runtime=$6
filename=$7
mixread=$8

if [ -z "${ioengine}" ] || [ -z "${size}"  ] || [ -z "${numjobs}" ] || [ -z "${bs}" ] || [ -z "${runtime}" ] || [ -z "${filename}" ] || [ -z "${mode}" ]; then
  echo 'arg error'
  exit 1
fi

local_ip=`/sbin/ifconfig -a | grep inet | grep -v 127.0.0.1 | grep -v inet6 | awk '{print $2}' | tr -d addrs`

root_dir=/root/fio_test

if [ ! -d ${root_dir} ];then
    mkdir -p ${root_dir}
else
   rm -rf ${root_dir}
   mkdir -p ${root_dir}
fi

result_vm_path=${root_dir}


iodepth=64
if [ ${mode} == 'randread' ] || [ ${mode} == 'randwrite' ] || [ ${mode} == 'randrw' ]; then
  iodepth=128
fi

rwmixread=''
if [ ${mode} == 'randrw' ] || [ ${mode} == 'rw' ]; then
  rwmixread='-rwmixread='${mixread}
fi

if [ ${mode} == 'read' ] || [ ${mode} == 'write' ]; then
   if [ ${bs} -lt 1024 ]; then
      iodepth=128
   fi
fi

fio -direct=1 -iodepth=${iodepth} -rw=${mode} ${rwmixread} -ioengine=${ioengine} -bs=${bs}k -size=${size} -numjobs=${numjobs} -runtime=${runtime} -time_based -group_reporting -filename=${filename} -name=${bs}k_${mode}_Testing --output-format=json --output ${result_vm_path}/${mode}_${bs}k_${size}_${numjobs}_${ioengine}_${runtime}_${local_ip}.json 
sleep 2

#./resend_result.sh 
