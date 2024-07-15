#!/bin/bash
rm -rf /tmp/harbid-temp

harbid --devnet --appdir=/tmp/harbid-temp --profile=6061 --loglevel=debug &
harbid_PID=$!

sleep 1

rpc-stability --devnet -p commands.json --profile=7000
TEST_EXIT_CODE=$?

kill $harbid_PID

wait $harbid_PID
harbid_EXIT_CODE=$?

echo "Exit code: $TEST_EXIT_CODE"
echo "harbid exit code: $harbid_EXIT_CODE"

if [ $TEST_EXIT_CODE -eq 0 ] && [ $harbid_EXIT_CODE -eq 0 ]; then
  echo "rpc-stability test: PASSED"
  exit 0
fi
echo "rpc-stability test: FAILED"
exit 1
