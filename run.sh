#!/bin/bash

psm=toutiao.clientqa.itc-server

if [ -f "output/bin/${psm}" ]; then
	echo "Delete the old ${psm}"
	rm output/bin/${psm}
fi

sh build.sh
sh output/bootstrap.sh
