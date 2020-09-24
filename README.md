# Talos

## 表结构
```mysql
CREATE TABLE `modinfos` (
  `module` char(64) NOT NULL DEFAULT '',
  `cwd` char(64) NOT NULL DEFAULT '',
  `env` char(7) NOT NULL DEFAULT '',
  `contact` varchar(256) NOT NULL DEFAULT '',
  `cmdline` varchar(128) NOT NULL DEFAULT '',
  `script` char(32) DEFAULT '',
  `procnum` int DEFAULT '1',
  `logpath` char(64) NOT NULL DEFAULT '',
  `lognum` int DEFAULT '0',
  `logsize` int DEFAULT '0',
  `cmd` char(100) NOT NULL DEFAULT '',
  `restartlimit` int NOT NULL DEFAULT '5',
  PRIMARY KEY (`module`,`env`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8

CREATE TABLE `process_monitor` (
  `module_name` char(64) DEFAULT NULL,
  `env` char(10) DEFAULT NULL,
  `stop_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `start_time` timestamp NULL DEFAULT NULL,
  `cost_time` int DEFAULT NULL,
  `host` char(15) DEFAULT NULL,
  `event_type` int DEFAULT NULL COMMENT '1 down',
  `mail_list` char(255) DEFAULT NULL
) ENGINE=MyISAM DEFAULT CHARSET=utf8
```
## proto
```shell
protoc --go_out=. --go-grpc_out=. --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative internal/rpc/proto/task.proto
```
## 运行
```shell
# -cd 配置文件目录
go run ./cmd/server/server.go run -cd .
go run ./cmd/agent/agent.go run -cd .
```