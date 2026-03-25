#! /bin/bash

date=`date "+%Y-%m-%d %H:%M:%S"`

if [ $# -ge 1 ]; then commit_msg="$date $1"
else commit_msg="$date"
fi

# echo "Submit info is: $commit_msg

go mod tidy
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o plant_be_exec ./cmd/plant_be/main.go
upx -9 plant_be_exec

git add .
git commit -m "$commit_msg" 
git push

scp plant_be_exec plant_be:
echo "请执行该命令：scp plant_be_exec plant_be:"