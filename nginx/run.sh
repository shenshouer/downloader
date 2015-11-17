#!/usr/bin/env bash

echo "IP:`hostname -I` hostname: `hostname`" >> /usr/share/nginx/html/index.html

nginx -g "daemon off;"