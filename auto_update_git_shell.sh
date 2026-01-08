#! /bin/bash

date=`date "+%Y-%m-%d %H:%M:%S"`

if [ $# -ge 1 ]; then commit_msg="$date $1"
else commit_msg="$date"
fi

# echo "Submit info is: $commit_msg

go mod tidy
#CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o plant_be ./cmd/plant_be/main.go
#upx -9 plant_be

git add .
git commit -m "$commit_msg" 
git push
