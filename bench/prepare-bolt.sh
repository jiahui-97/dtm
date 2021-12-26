# !/bin/bash
curl https://raw.githubusercontent.com/dtm-labs/dtm/alpha/bench/setup.sh -o setup.sh
sh setup.sh

echo 'all prepared. you shoud run following commands to test in different terminal'
echo
echo 'cd dtm && go run bench/main.go boltdb'
echo 'cd dtm && bench/run-boltdb.sh'

