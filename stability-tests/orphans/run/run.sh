#!/bin/bash
rm -rf /tmp/harbid-temp

harbid --simnet --appdir=/tmp/harbid-temp --profile=6061 &
harbid_PID=$!

sleep 1

orphans --simnet -alocalhost:24511 -n20 --profile=7000
TEST_EXIT_CODE=$?

kill $harbid_PID

wait $harbid_PID
harbid_EXIT_CODE=$?

echo "Exit code: $TEST_EXIT_CODE"
echo "harbid exit code: $harbid_EXIT_CODE"

if [ $TEST_EXIT_CODE -eq 0 ] && [ $harbid_EXIT_CODE -eq 0 ]; then
  echo "orphans test: PASSED"
  exit 0
fi
echo "orphans test: FAILED"
exit 1
