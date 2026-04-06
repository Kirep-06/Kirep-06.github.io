# Harness 用于 autoresearch

2026-04-01

这篇文章记录我如何把 Harness 引入 autoresearch 流程，用来减少重复检索步骤并稳定产出格式。

## 核心做法

我的策略是先定义最小输入，再把检索、筛选、汇总拆成固定阶段，每一阶段只输出下一步需要的数据。

```text
input -> search -> filter -> summarize -> final note
```

## 结果

在相同时间窗口内，我能更快得到可复用结论，也更容易追踪每一步的错误来源。
