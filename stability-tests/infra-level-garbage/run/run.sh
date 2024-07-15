#!/bin/bash
rm -rf /tmp/harbid-temp

harbid --devnet --appdir=/tmp/harbid-temp --profile=6061 &
harbid_PID=$!

sleep 1

infra-level-garbage --devnet -alocalhost:24611 -m messages.dat --profile=7000
TEST_EXIT_CODE=$?

kill $harbid_PID

wait $harbid_PID
harbid_EXIT_CODE=$?

echo "Exit code: $TEST_EXIT_CODE"
echo "harbid exit code: $harbid_EXIT_CODE"

if [ $TEST_EXIT_CODE -eq 0 ] && [ $harbid_EXIT_CODE -eq 0 ]; then
  echo "infra-level-garbage test: PASSED"
  exit 0
fi
echo "infra-level-garbage test: FAILED"
exit 1
