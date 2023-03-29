#!/bin/bash

# Start the first process
# /go/bin/zincsearch &
  
# Start the second process
/go/bin/wzinc start &
  
# Wait for any process to exit
wait -n
  
# Exit with status of process that exited first
exit $?