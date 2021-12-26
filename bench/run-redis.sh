# !/bin/bash

set -x

redis-benchmark -n 300000 EVAL "redis.call('SET', 'abcdedf', 'ddddddd')" 0

redis-benchmark -n 300000 EVAL "redis.call('SET', KEYS[1], ARGS[1])" 1 'aaaaaaaaa' 'bbbbbbbbbb'

redis-benchmark -n 3000000 -P 50 SET 'abcdefg' 'ddddddd'

redis-benchmark -n 300000 EVAL "for k=1, 10 do; redis.call('SET', KEYS[1], ARGS[1]); end" 1 'aaaaaaaaa' 'bbbbbbbbbb'

redis-benchmark -n 300000 -P 50 EVAL "redis.call('SET', KEYS[1], ARGS[1])" 1 'aaaaaaaaa' 'bbbbbbbbbb'

redis-benchmark -n 300000 EVAL "local js=cjson.decode(ARGV[1])" 1 'aaaaaaaaa' '{"aaaaa":"bbbbb"}'

ab -n 1000000 -c 10 "http://127.0.0.1:8083/api/busi_bench/benchEmptyUrl"
