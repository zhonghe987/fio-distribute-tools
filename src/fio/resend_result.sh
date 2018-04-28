
#!/bin/sh
local_ip=`/sbin/ifconfig -a | grep inet | grep -v 127.0.0.1 | grep -v inet6 | awk '{print $2}' | tr -d addrs`
summary_ip=192.168.32.19
target=root@${summary_ip}
cd ~/
#result_new_dir=`hostname`_fio
result_vm_path=fio_test
#result_new_name=${result_new_dir}.tar.gz
#mv ${result_vm_path} ${result_new_dir}
#tar cvf  ${result_new_name} ${result_new_dir}
result_target_path=/root/result_fio_report/${local_ip}_fio
result_target=${target}:${result_target_path}

scp ${result_vm_path}/read_1024k_100G_1_libaio_1000_192.168.32.11.json ${result_target}
rm -rf ${result_vm_path}/read_1024k_100G_1_libaio_1000_192.168.32.11.json

