#!/bin/sh

if [ -z "$1" ]; then
    echo "Usage: $0 <VERSION>"
    exit 1
fi

docker build -t harbor.zjuqsc.com/ippool/ipmanager:$1 . --build-arg CONFIG=/config.json --platform linux/amd64
exit 0

docker push harbor.zjuqsc.com/ippool/ipmanager:$1

# run with
sudo docker run -d --name ipmanager -p 19000:19000 -p 9091:9091 -v /var/log/nginx/:/var/log/nginx harbor.zjuqsc.com/ippool/ipmanager:latest