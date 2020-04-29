
# Introduction
this project is imitate docker technical. that include four core linux function. and pass test in CentOS7.

- [x] namespace

- [x] cgroups

- [x] unified filesystem(aufs)

- [x] network

# how to use 

make sure go environment is ready, and already install gcc.
after that all done, build this project to a binary called "mydocker" like project name.

# support feature

## overview

![commands](.\res\commands.jpg)

## image

currently only support two base image named busybox and nginx by default. the commands of image include delete and list.

## container

the core parts of this project. 



## network

supported.


# TBD

* overlay network 

# reference
[how to write a docker](https://github.com/xianlubird/mydocker)
