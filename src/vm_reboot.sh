
#!/bin/sh
iparry=(4 12)
for i in ${iparry[@]}
do
  ssh root@192.168.32.$i reboot
done
