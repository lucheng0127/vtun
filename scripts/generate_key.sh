#!/bin/bash

key=`/usr/bin/hexdump -vn8 -e'4/4 "%08X" 1 "\n"' /dev/urandom | sed 's/ //g'`
sed -i "s/0123456789ABCDEF/$key/g" ./conf/server/config.yaml
