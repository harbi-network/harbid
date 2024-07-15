#!/bin/bash
rm -rf /tmp/harbid-temp

NUM_CLIENTS=128
harbid --devnet --appdir=/tmp/harbid-temp --profile=6061 --rpcmaxwebsockets=$NUM_CLIENTS &
harbid_PID=$!
harbid_KILLED=0
function killharbidIfNotKilled() {
  if [ $harbid_KILLED -eq 0 ]; then
    kill $harbid_PID
  fi
}
trap "killharbidIfNotKilled" EXIT

sleep 1

rpc-idle-clients --devnet --profile=7000 -n=$NUM_CLIENTS
TEST_EXIT_CODE=$?

kill $harbid_PID

wait $harbid_PID
harbid_EXIT_CODE=$?
harbid_KILLED=1

echo "Exit code: $TEST_EXIT_CODE"
echo "harbid exit code: $harbid_EXIT_CODE"

if [ $TEST_EXIT_CODE -eq 0 ] && [ $harbid_EXIT_CODE -eq 0 ]; then
  echo "rpc-idle-clients test: PASSED"
  exit 0
fi
echo "rpc-idle-clients test: FAILED"
exit 1
