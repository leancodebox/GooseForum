# 访问控制性能基线

基准固定生成 100,000 个已发布主题、100 个分类和 100,000 条有效分类索引，使用首页默认排序与 20 条分页，分别模拟 95%、50%、5% 分类可见率。

复现：

```bash
go test ./app/models/forum/topics -run '^$' \
  -bench BenchmarkAudienceTopicList100K -benchtime=50x -benchmem
```

2026-08-11，Apple M4，SQLite 内存库，Go benchmark 结果：

| 可见率 | 平均延迟 | P95 | 分配 |
| --- | ---: | ---: | ---: |
| 95% | 2.07 ms | 2.68 ms | 52.3 KB / 1,394 allocs |
| 50% | 2.93 ms | 3.30 ms | 46.4 KB / 1,302 allocs |
| 5% | 1.92 ms | 2.48 ms | 41.1 KB / 1,208 allocs |

该基准不进入普通测试路径，避免每次 `go test ./...` 重建 10 万行数据。MySQL 的执行计划和 P95 应在实际部署版本、数据分布与连接参数下用同一查询补录；不能从 SQLite 数字外推。
