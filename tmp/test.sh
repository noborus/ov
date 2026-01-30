#!/usr/bin/zsh

for i in {1..100}
do
echo "OK"
done
sleep 1
for i in {1..100}
do
echo "NG" 
done >&2
