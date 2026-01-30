#!/bin/bash

for r in {32..0}
do
        ((b = 32 - r ))
        for g in {0..32}
        do
                printf "\e[38;2;%d;%d;%dm@" $((r<<3)) $((g<<3)) $((b<<3))
        done
        printf "\n"
done