
#!/bin/sh
set -x
ssh root@$1 pkill -9 fio
