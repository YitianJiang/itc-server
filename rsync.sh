#!/bin/bash

fswatch_cmd=`which fswatch`
if [[ -z ${fswatch_cmd} ]]; then 
	brew install fswatch
fi

rules=/Users/apple/iWorks/tt_server/rsync_py_rules
src=/Users/apple/iWorks/tt_server/plugin/admin/                  # 需要同步的源路径
des=/home/apple/repos/toutiao/app/admin/                         # 目标服务器上 rsync --daemon 发布的名称，rsync --daemon这里就不做介绍了，网上搜一下，比较简单。
dest_ip=**.**.**.**                 	# 目标服务器ip地址
user=apple                        		# rsync --daemon定义的验证用户名
cd ${src}                              	# 此方法中，由于rsync同步的特性，这里必须要先cd到源目录，fswatch ./ 才能rsync同步后目录结构一致，有兴趣的同学可以进行各种尝试观看其效果

# 脚本启动的时候全量同步一下
rsync -avz --delete ${src} ${user}@${dest_ip}:${des} --exclude-from=${rules}

# rsync -avz --delete ${src} ${user}@${dest_ip}:${des} 


fswatch ${src} | while read file         # 把监控到有发生更改的"文件路径列表"循环
do

    echo $file
	# rsync -avz ${src} ${user}@${dest_ip}:${des}
	rsync -avz --delete ${src} ${user}@${dest_ip}:${des} --exclude-from=${rules}
done