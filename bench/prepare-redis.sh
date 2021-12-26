# !/bin/bash
curl https://github.com/dtm-labs/dtm/tree/main/bench/setup.sh -o setup.sh
sh setup.sh

echo 'all prepared. you shoud run following commands to test in different terminal'
echo
echo 'cd dtm && go run bench/main.go redis'
echo 'cd dtm && bench/run-redis.sh'
