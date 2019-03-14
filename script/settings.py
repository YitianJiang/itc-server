# coding: utf-8

PRODUCT="toutiao"
SUBSYS="clientqa"
MODULE="itc-server"

# ginex可以不使用nginx就可以布署,api层有完整的metrics,
# 但使用nginx可以额外获得nginx层的metrics,比如499状态
# REQUIRE_NGINX = True
# PRENGINX_SCRIPT = "pre_nginx.sh"

APP_TYPE="binary"
