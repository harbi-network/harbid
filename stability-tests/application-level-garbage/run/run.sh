#!/bin/bash
rm -rf /tmp/harbid-temp

harbid --devnet --appdir=/tmp/harbid-temp --profile=6061 --loglevel=debug &
harbid_PID=$!
harbid_KILLED=0
function killharbidIfNotKilled() {
    if [ $harbid_KILLED -eq 0 ]; then
      kill $harbid_PID
    fi
}
trap "killharbidIfNotKilled" EXIT

sleep 1

application-level-garbage --devnet -alocalhost:24611 -b blocks.dat --profile=7000
TEST_EXIT_CODE=$?

kill $harbid_PID

wait $harbid_PID
harbid_KILLED=1
harbid_EXIT_CODE=$?

echo "Exit code: $TEST_EXIT_CODE"
echo "harbid exit code: $harbid_EXIT_CODE"

if [ $TEST_EXIT_CODE -eq 0 ] && [ $harbid_EXIT_CODE -eq 0 ]; then
  echo "application-level-garbage test: PASSED"
  exit 0
fi
echo "application-level-garbage test: FAILED"
exit 1
