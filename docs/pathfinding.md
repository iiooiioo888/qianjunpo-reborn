# 寻路（pkg/pathfind）

Phase 2 权威寻路与 `pkg/validate` 移动校验共用同一套网格规则。

## 移动模型

- 19×19 整数格点，八方向相邻，每一步 Chebyshev 代价 **1**（与 `board.Chebyshev` 一致）。
- 全整数、无浮点，结果可复现。

## 算法

| 入口 | 用途 |
|------|------|
| `AStar` | 基准最优路径；启发式为 Chebyshev 距离。 |
| `JPS` | 开放地图上加速搜索；内部为跳点折线，返回前经 `ExpandJumpPath` 展开为逐步相邻路径。 |
| `FindPath` | 可选 `BuildConnectivity` 快检 + 选 `AlgoAStar` / `AlgoJPS` + 可选 `SmoothPath`。 |
| `BuildConnectivity` / `QuickUnreachable` | 并查集对可走格八连通分量，`O(α(n))` 判定起终点是否同分量。 |

## A* 与 JPS

在**全开放**且代价均匀时，展开后的 JPS 路径 **Chebyshev 步数**与 A* 最优解一致（见 `TestJPSExpandedMatchesAStar`）。JPS 仅减少搜索扩展量，不改变最优步数。

有障碍或单位占位时，两者仍应给出相同最优步数；若仅比较折线跳点，长度以 `PathLength` 为准，使用前需 `ExpandJumpPath`。

## 路径平滑

`SmoothPath` 对已有相邻路径做贪心视线合并：沿 Chebyshev 直线逐格检测 `HasLineOfSight`（与 `ExpandSegment` 同 trace）。输出仍为相邻步序列，**步数不会变长**。

`pkg/validate` 在 A* 成功后默认 `Smooth: true`，减少冗余拐点。

## 性能（参考）

`go test -bench=. ./pkg/pathfind/...` 在开放地图上约 200 格对角路径可测 A*/JPS/平滑耗时；目标 \<1ms 为 aspirational，以 CI 与本地 benchmark 为准。
