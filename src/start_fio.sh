#!/bin/sh

scp fio/crontab root@$1:/etc/
scp fio/fio-signal.sh root@$1:~/
echo """
#!/bin/sh
local_ip=\`/sbin/ifconfig -a | grep inet | grep -v 127.0.0.1 | grep -v inet6 | awk '{print \$2}' | tr -d "addrs"\`
summary_ip=$2
target=root@\${summary_ip}
cd ~/
#result_new_dir=\`hostname\`_fio
result_vm_path=fio_test
#result_new_name=\${result_new_dir}.tar.gz
#mv \${result_vm_path} \${result_new_dir}
#tar cvf  \${result_new_name} \${result_new_dir}
result_target_path=/root/result_fio_report/\${local_ip}_fio
result_target=\${target}:\${result_target_path}

scp \${result_vm_path}/$3 \${result_target}
rm -rf \${result_vm_path}/$3
""" > fio/resend_result.sh
chmod +x fio/resend_result.sh
scp fio/resend_result.sh root@$1:~/
scp fio/clear_caches.sh root@$1:~/
