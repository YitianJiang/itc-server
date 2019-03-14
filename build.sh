#!/bin/bash

CURDIR=$(cd $(dirname $0); pwd)
if [ ! -f $CURDIR/script/settings.py ]; then
    echo "there is no settings.py exist."
    exit -1
fi

PRODUCT=$(cd $CURDIR/script; python -c "import settings; print(settings.PRODUCT)")
SUBSYS=$(cd $CURDIR/script; python -c "import settings; print(settings.SUBSYS)")
MODULE=$(cd $CURDIR/script; python -c "import settings; print(settings.MODULE)")
RUN_NAME=${PRODUCT}.${SUBSYS}.${MODULE}

mkdir -p output/bin output/conf
cp script/bootstrap.sh script/pre_nginx.sh script/settings.py output 2>/dev/null
chmod +x output/bootstrap.sh output/pre_nginx.sh
find conf/ -type f ! -name "*_local.*" | xargs -I{} cp {} output/conf/

go build -a -o output/bin/${RUN_NAME}
