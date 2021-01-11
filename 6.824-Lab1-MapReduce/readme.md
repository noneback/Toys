# MIT6.824 Lab1 MapReduce

## Description 

Lab1 of mit6.824,the course of distributed system.  

It requires students to read the paper *MapReduce: Simplified Data Processing on Large Clusters*,and to design  **master** and **worker**,trying to implement a simple version of MapReduce System in golang using rpc and go plugin.  

more information in official website

## Main Work
Our job is to design and implement master and worker,accomplish the simple MapReduce.

main code see `./src/mr`

## Test
use bash script auto test all mrapp in ./src/mrapp
```bash
cd ./src/main
sh ./test-mr.sh
```

## Reference

[MapReduce: Simplified Data Processing on Large Clusters](http://static.googleusercontent.com/media/research.google.com/zh-CN//archive/mapreduce-osdi04.pdf)

[Official Website](https://pdos.csail.mit.edu/6.824/labs/lab-mr.html)