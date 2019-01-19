#!/bin/bash

cd /service/images/nginx
bash ./wait-for-it.sh day_trader_web_1:41000 --timeout=300
bash ./wait-for-it.sh day_trader_web_2:41000 --timeout=300
bash ./wait-for-it.sh day_trader_web_3:41000 --timeout=300

exec nginx -g 'daemon off;'