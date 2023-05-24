#!/bin/sh
set -xe
mkdir -p /run/nginx
if [ -n "$VUE_APP_API_URL" ];then
  VUE_APP_API_URL=$VUE_APP_API_URL npm run build
fi
exec nginx -g 'daemon off;'
