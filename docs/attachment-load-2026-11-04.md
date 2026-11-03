# 附件并发压测（2026-11-04）

## 结果

- 5 连接主测的修复前、修复后各 2,176 次正常上传均成功；逐文件回读校验长度和 SHA-256，未发现损坏。
- 每次主测包含 50 轮、每轮 16 个事务竞争同一附件：每轮恰好一个事务成功，未发生跨帖重复归属。
- 每次主测包含 50 轮绑定与取消竞争：未出现已绑定文件被删、附件行缺失或文件内容丢失。
- 修复后另以 Go race detector 执行 128 次正常上传、10 轮附件抢占和 10 轮绑定／取消竞争，未报告 Go 数据竞争；额度测试仍为 10/10 轮超发，最多接受 13 个文件。
- **每日额度存在业务竞态，尚未修复。** 每日上限设为 1，32 个请求同时上传，主测修复前后均为 20/20 轮超发；单轮最多分别接受 24、20 个文件。额度查询和上传记录创建是两个步骤，不能保证并发下只接受一次。此问题独立于 Go 数据竞争检测。

## 环境与范围

- 本机 macOS arm64，8 个逻辑 CPU、16 GiB 内存，Go 1.27.3。
- 两个独立临时 SQLite 文件，WAL、NORMAL 同步级别、5 秒 busy timeout；论坛库、文件库均为 5 个连接，与本地 `config.toml` 连接数一致。运行完成关闭连接并删除临时数据库。
- 本机随机端口 HTTP，真实 Gin 上传处理器、权限策略、文件存储、附件模型；测试夹具固定注入普通成员身份，**不包含真实登录认证、中间件和公网网络开销**。
- 正常吞吐阶段每日额度提高到 100000，保留真实额度计数查询；注册冷静期已满足。额度边界使用独立账号测试，账号没有角色豁免。
- 上传内容为有效 ASCII `.txt`；64 KiB 每档 512 个请求，5 MiB 每档 32 个请求。请求数固定、并发 worker 固定，属于短时闭环压测，不代表长期稳定容量或生产 SLA。
- 堆峰值按 5 ms 采样，累计分配量来自 Go `runtime.MemStats`，均包含同进程 HTTP 客户端与服务端；不是服务器独立 RSS。回读校验与清理不计入上传阶段耗时和分配量。
- GORM SQL 日志关闭，避免日志 I/O 干扰；没有验证 MySQL、PostgreSQL 或多实例部署。

## 通过 Profile 定位并修复的瓶颈

`filedata.UpdateFileContent` 原先只指定 `Table`，随后用 `UpdateColumn` 传入包含文件字节的 map。GORM 没有完整模型时进入 `schema.ParseWithSpecialTableName` 的不支持类型分支，通过 `fmt.Errorf("%w: %+v", ..., dest)` 格式化整个文件字节。该错误在指定表名的更新路径中被忽略，但已产生大量 CPU 和内存开销。

分配 Profile 的一轮 8 个 5 MiB 上传及校验流程中，`gorm.io/gorm/schema.ParseWithSpecialTableName → fmt.Errorf` 累计分配约 796 MiB，占总分配约 80%。补充 `.Model(&Entity{})` 后，以相同主测参数复测。

下表的“前／后”只比较这一处模型声明修复，均已包含此前的流式 multipart、缓冲区复用和批量绑定优化。

| 文件 | 并发 | 成功请求／秒（前 → 后） | P95 毫秒（前 → 后） | 采样堆峰值 MiB（前 → 后） |
| --- | ---: | ---: | ---: | ---: |
| 64 KiB | 1 | 225.6 → 777.6 | 5.7 → 2.6 | 23.3 → 21.7 |
| 64 KiB | 4 | 483.8 → 962.2 | 22.4 → 13.1 | 24.4 → 22.3 |
| 64 KiB | 16 | 521.0 → 917.1 | 52.5 → 25.2 | 27.9 → 24.5 |
| 64 KiB | 32 | 529.4 → 968.7 | 95.2 → 60.4 | 32.0 → 27.0 |
| 5 MiB | 1 | 3.9 → 21.5 | 286.5 → 51.1 | 105.2 → 43.8 |
| 5 MiB | 4 | 12.1 → 25.2 | 376.3 → 502.9 | 215.4 → 72.8 |
| 5 MiB | 16 | 12.5 → 25.0 | 1616.2 → 884.0 | 569.7 → 163.6 |
| 5 MiB | 32 | 12.3 → 22.1 | 2460.6 → 1402.6 | 900.7 → 251.6 |

5 MiB 请求累计分配约 **108.9 → 10.7 MiB／请求**，64 KiB 请求约 **1.38 → 0.24 MiB／请求**。高并发仍增加排队和内存；4 并发大文件这一轮 P95 没有改善。每档大文件仅 32 个样本，尾延迟需要更多重复轮次和长时间负载才能评估。

修复前另以 2 个连接、64 KiB、每档 256 次请求进行对照，4/16/32 并发分别约 534/539/495 次／秒，均无上传错误；绑定和取消各 10 轮未发现一致性问题。该对照样本较少，不能据此确定最佳连接数。

## 复现

仓库根目录执行；脚本强制使用自建临时数据库，启动时验证实际 SQLite 文件路径，不接受生产目标 URL。

```sh
go build -o /tmp/gooseforum-attachment-load ./scripts/attachment-load
/tmp/gooseforum-attachment-load -requests 512 -out /tmp/gooseforum-attachment-load-after.json

go build -race -o /tmp/gooseforum-attachment-load-race ./scripts/attachment-load
/tmp/gooseforum-attachment-load-race -sizes-kib 64 -workers 4,16 -requests 64 -rounds 10 -out /tmp/gooseforum-attachment-load-race.json

/tmp/gooseforum-attachment-load -sizes-kib 5120 -workers 4 -requests 8 -rounds 1 -alloc-profile /tmp/gooseforum-attachment-load-current.allocs.pprof
go tool pprof -alloc_space -top /tmp/gooseforum-attachment-load-current.allocs.pprof
```

JSON 中的 `claimViolations`、`cancelViolations`、`quotaExceededRounds` 和 `raceErrors` 必须分别检查；命令正常退出只表示采样流程执行完毕，不代表所有业务约束通过。带 `-race` 的吞吐数不能与普通构建直接比较。

原始主测数据保存在本次运行环境的 `/tmp/gooseforum-attachment-load-5conn.json`（模型声明修复前）、`/tmp/gooseforum-attachment-load-after.json`（修复后）；修复前分配 Profile 位于 `/tmp/gooseforum-attachment-load.allocs.pprof`。这些临时结果不属于仓库持久文件。
