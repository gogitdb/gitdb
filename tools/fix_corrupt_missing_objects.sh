#!/usr/bin/env bash

#https://stackoverflow.com/questions/33012869/broken-branch-in-git-fatal-your-current-branch-appears-to-be-broken

git fsck --full >& fsck_report.txt
cat fsck_report.txt | grep "missing" | awk '{print $7}' > fsck_report.txt


Did it report a corrupted file?
If so delete the file, go back to step #1.

for item in `cat fsck_report.txt`
do
  echo "deleting $item"
  rm -f $item
done

Do del .git/index
Do git reset

rm -f fsck_report


#remove invalid reflog references
git reflog expire --stale-fix --all

clone down repo to a separate dir
and run
cat ../fresh/.git/objects/pack/pack-*.pack | git unpack-objects
within the broken repo