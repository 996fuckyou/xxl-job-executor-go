#!/usr/bin/env bash
set -ex
echo deploy executor test 
appDir="/data01/app/executor_go_test"
logdir="/data01/log/executor_go_test.err.log"

sudo mv executor_go_test.tar.gz $appDir &&
cd $appDir && sudo tar -xzf executor_go_test.tar.gz &&
sudo killall -v executor_go_test && sudo supervisorctl status && sudo tail -n 100 "$logdir"