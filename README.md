# go-distributed-kv-db
Golang实现的分布式key-value数据库  
数据读写基于[bbolt](https://github.com/etcd-io/bbolt)数据库实现

## 部署与运行：
配置launch.sh脚本，并在项目根目录下运行 ` $ ./launch.sh `

launch.sh 参数说明:  
  
db-location: bolt数据库名  
http-addr: 监听地址与端口  
config-file: .toml配置文件名  
