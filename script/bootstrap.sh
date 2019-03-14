#!/bin/bash
CURDIR=$(cd $(dirname $0); pwd)
if [ "X$1" != "X" ]; then
    RUNTIME_ROOT=$1
else
    RUNTIME_ROOT=${CURDIR}
fi

if [ "X$RUNTIME_ROOT" == "X" ]; then
    echo "There is no RUNTIME_ROOT support."
    echo "Usage: ./bootstrap.sh $RUNTIME_ROOT"
    exit -1
fi

PORT=$2

RUNTIME_CONF_ROOT=$RUNTIME_ROOT/conf
RUNTIME_LOG_ROOT=$RUNTIME_ROOT/log

if [ ! -d $RUNTIME_LOG_ROOT/app ]; then
    mkdir -p $RUNTIME_LOG_ROOT/app
fi

if [ ! -d $RUNTIME_LOG_ROOT/rpc ]; then
    mkdir -p $RUNTIME_LOG_ROOT/rpc
fi

if [ ! -f $CURDIR/settings.py ]; then
    echo "there is no settings.py exist."
    exit -1
fi

PRODUCT=$(cd $CURDIR; python -c "import settings; print(settings.PRODUCT)")
SUBSYS=$(cd $CURDIR; python -c "import settings; print(settings.SUBSYS)")
MODULE=$(cd $CURDIR; python -c "import settings; print(settings.MODULE)")
REQUIRE_NGINX=$(cd $CURDIR; python -c "import settings; print(settings.REQUIRE_NGINX)" 2>/dev/null)

if [ "$REQUIRE_HTTP_MESH" == "True" ]; then
    REQUIRE_NGINX=False
fi

if [ -z "$PRODUCT" ] || [ -z "$SUBSYS" ] || [ -z "$MODULE" ]; then
    echo "Support PRODUCT SUBSYS MODULE PORT in settings.py"
    exit -1
fi

if [ "$IS_HOST_NETWORK" == "1" ]; then
    if [ "$REQUIRE_NGINX" == "True" ]; then
        export RUNTIME_SERVICE_PORT=$PORT1
        export RUNTIME_DEBUG_PORT=$PORT2
    else
        export RUNTIME_SERVICE_PORT=$PORT0
        export RUNTIME_DEBUG_PORT=$PORT1
    fi
fi

SVC_NAME=${PRODUCT}.${SUBSYS}.${MODULE}

BinaryName=${PRODUCT}.${SUBSYS}.${MODULE}

export GIN_LOG_DIR=$RUNTIME_LOG_ROOT
export PSM=$SVC_NAME
CONF_DIR=$CURDIR/conf/

args="-psm=$SVC_NAME -conf-dir=$CONF_DIR -log-dir=$GIN_LOG_DIR"
if [ "X$PORT" != "X" ]; then
    args+=" -port=$PORT"
fi

echo "$CURDIR/bin/${BinaryName} $args"

exec $CURDIR/bin/${BinaryName} $args
