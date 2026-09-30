// Package pathfind implements deterministic grid pathfinding for the 19×19 board.
//
// Movement model: 8-connected steps with Chebyshev cost 1 per step (integer only, no floats).
// AStar uses that metric with a Chebyshev heuristic (admissible on open grids).
//
// JPS (Jump Point Search) accelerates A* on uniform-cost grids by skipping symmetric
// expansions; jump-point polylines are converted to adjacent steps via ExpandJumpPath before
// use in move validation. On fully open maps, JPS expanded path length matches A* optimal length.
//
// BuildConnectivity + QuickUnreachable provide O(α(n)) reachability checks via union-find
// over passable 8-neighbor components, avoiding full search when start and goal differ in component.
//
// SmoothPath applies greedy line-of-sight merging (Chebyshev trace / Bresenham-style walk)
// to reduce redundant waypoints without increasing step count.
package pathfind
