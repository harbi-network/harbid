#!/bin/bash

APPDIR=/tmp/harbid-temp
harbid_RPC_PORT=29587

rm -rf "${APPDIR}"

harbid --simnet --appdir="${APPDIR}" --rpclisten=0.0.0.0:"${harbid_RPC_PORT}" --profile=6061 &
harbid_PID=$!

sleep 1

RUN_STABILITY_TESTS=true go test ../ -v -timeout 86400s -- --rpc-address=127.0.0.1:"${harbid_RPC_PORT}" --profile=7000
TEST_EXIT_CODE=$?

kill $harbid_PID

wait $harbid_PID
harbid_EXIT_CODE=$?

echo "Exit code: $TEST_EXIT_CODE"
echo "harbid exit code: $harbid_EXIT_CODE"

if [ $TEST_EXIT_CODE -eq 0 ] && [ $harbid_EXIT_CODE -eq 0 ]; then
  echo "mempool-limits test: PASSED"
  exit 0
fi
echo "mempool-limits test: FAILED"
exit 1
