# !/bin/bash
curl https://github.com/dtm-labs/dtm/tree/main/bench/setup.sh -o setup.sh
sh setup.sh

docker-compose -f helper/compose.mysql.yml up -d

echo 'all prepared. you shoud run following commands to test in different terminal'
echo
echo 'cd dtm && go run bench/main.go db'
echo 'cd dtm && bench/run-mysql.sh'
