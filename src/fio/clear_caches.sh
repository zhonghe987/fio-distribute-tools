#!/bin/sh

set -x

sync
sleep 1
echo 3 > /proc/sys/vm/drop_caches
