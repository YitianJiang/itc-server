#!/bin/bash

psm=toutiao.clientqa.itcserver

if [ -f "output/bin/${psm}" ]; then
	echo "Delete the old ${psm}"
	rm output/bin/${psm}
fi

sh build.sh
# 本地测试必须使用如下命令方式启动程序
SEC_MYSQL_AUTH=1 TCE_PSM=${psm} doas -p ${psm} sh output/bootstrap.sh
