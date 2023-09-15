#!/bin/bash

# 初始化脚本，每个要运行mysql pod实例的node都需要执行

set -e

# 创建mysql数据目录
mkdir /mysql-data/ -p
