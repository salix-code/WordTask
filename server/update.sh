#!/bin/bash
set -e # 遇到任何错误立即停止执行

cd ~/web/WordTask/server/
git pull origin web
sudo systemctl stop wordtask
go build -o ./dist/wordtask-server main.go
sudo systemctl start wordtask 