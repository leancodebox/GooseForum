# 访问控制性能基线

基准固定生成 100,000 个已发布主题和 100 个主分类，使用首页默认排序与 20 条分页，分别模拟 95%、50%、5% 分类可见率。受众条件直接作用在 `main_category_id`，不连接分类关系表；排序复合索引保持原列顺序。实测把主分类插入排序索引会让高可见率场景产生大范围分段扫描和额外排序，因此没有采用。

复现：

```bash
go test ./app/models/forum/topics -run '^$' \
  -bench BenchmarkAudienceTopicList100K -benchtime=20x -benchmem
```

2026-08-12，Apple M4，SQLite 内存库，主分类标量过滤的 Go benchmark 结果：

| 可见率 | 平均延迟 | P95 | 分配 |
| --- | ---: | ---: | ---: |
| 95% | 0.277 ms | 0.286 ms | 51.7 KB / 1,413 allocs |
| 50% | 0.287 ms | 0.331 ms | 46.5 KB / 1,322 allocs |
| 5% | 0.443 ms | 0.449 ms | 41.0 KB / 1,228 allocs |

该基准不进入普通测试路径，避免每次 `go test ./...` 重建 10 万行数据。MySQL 的执行计划和 P95 应在实际部署版本、数据分布与连接参数下用同一查询补录；不能从 SQLite 数字外推。
