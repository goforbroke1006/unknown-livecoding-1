#!/bin/bash

test_args=(
  "-test.v"                  # verbose logs
#  "-test.paniconexit0"       # exit with code 0 on panic
#  "-test.benchtime" "1000x" # b.N = 1000
  "-test.benchtime" "10000x" # b.N = 10000
  "-test.cpu" "1,2,3,4,8"    # repeat each benchmark for 1 3 2 4 8 CPUs
  "-test.run" '^$'           # disable any unit tests
)

bench_list=(
  "BenchmarkPubSubPrimitive_Subscribe"
  "BenchmarkConnectionStorageOnMutex_GetConnection"
  "BenchmarkConnectionStorageOnChan_GetConnection"
  "BenchmarkConnectionStorageOnChan_Shutdown"
)

for bench_name in "${bench_list[@]}"; do
  go test "${test_args[@]}" -test.bench "^\Q${bench_name}\E$" ./... | tee "${bench_name}.bench.log"
done
