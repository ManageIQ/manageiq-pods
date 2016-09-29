#!/bin/sh

APP_ROOT=/var/www/miq/vmdb

find "$APP_ROOT" -exec chgrp 0 {} \;
find "$APP_ROOT" -exec chmod g+rw {} \;
find "$APP_ROOT" -type d -exec chmod g+x {} +
chmod 777 $APP_ROOT/log
