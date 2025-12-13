#! /bin/bash

date=`date "+%Y-%m-%d %H:%M:%S"`

if [ $# -ge 1 ]; then commit_msg="$date $1"
else commit_msg="$date"
fi

# echo "Submit info is: $commit_msg"

git add .
git commit -m "$commit_msg" 
git push
