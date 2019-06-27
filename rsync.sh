#!/bin/bash

fswatch_cmd=`which fswatch`
if [[ -z ${fswatch_cmd} ]]; then 
	brew install fswatch
fi

rules=/Users/bytedance/go/src/code.byted.org/clientQA/itc-server/exclude_files   # 排除文件规则
src=/Users/bytedance/go/src/code.byted.org/clientQA/itc-server/                  # 需要同步的源路径
des=/home/kanghuaisong/go/src/code.byted.org/clientQA/itc-server/                # 目标服务器上 rsync --daemon 发布的名称，rsync --daemon这里就不做介绍了，网上搜一下，比较简单。
dest_ip=10.224.10.61                 	# 目标服务器ip地址
user=kanghuaisong                       # rsync --daemon定义的验证用户名
cd ${src}                              	# 此方法中，由于rsync同步的特性，这里必须要先cd到源目录，fswatch ./ 才能rsync同步后目录结构一致，有兴趣的同学可以进行各种尝试观看其效果

# 脚本启动的时候全量同步一下
rsync -avz --delete ${src} ${user}@${dest_ip}:${des} --exclude-from=${rules}
