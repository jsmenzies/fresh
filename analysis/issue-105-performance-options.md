# Issue #105 Performance Notes

Date: 2026-03-13
Branch: `perf-timing-info`
Worktree: `/Users/james/git/fresh/.worktrees/perf-timing-info`

## Captured Timing Samples (Info Column)

Observed values from your screenshot:

```text
refresh 3391ms f:3330ms(ok) b:60ms
refresh 2984ms f:2926ms(ok) b:57ms
refresh 3284ms f:3209ms(ok) b:74ms
refresh 3220ms f:3153ms(ok) b:67ms
refresh 3457ms f:3407ms(ok) b:50ms
refresh 3284ms f:3212ms(ok) b:71ms
refresh 3538ms f:3478ms(ok) b:60ms
refresh 3118ms f:3055ms(ok) b:62ms
refresh 3845ms f:3792ms(ok) b:52ms
refresh 3824ms f:3761ms(ok) b:63ms
refresh 3383ms f:3320ms(ok) b:63ms
refresh 3201ms f:3135ms(ok) b:66ms
refresh 167ms  f:87ms(ok)  b:79ms
refresh 3674ms f:3622ms(ok) b:51ms
refresh 157ms  f:68ms(ok)  b:89ms
```

## Initial Read

- `fetch` dominates runtime for most repos (typically ~2.9s to ~3.8s).
- `build` is usually much smaller (~50ms to ~90ms).
- Two repos are much faster end-to-end (~157ms to ~167ms), likely due to local/nearby remotes, no-op transport, or cached/auth/session behavior.

Implication: startup optimization should prioritize reducing or deferring `fetch`, then reduce process fan-out.

## Possible Implementations

### 1) Do not auto-refresh immediately on first listing render
- What: Keep scan-built repo state and skip initial `performRefresh` fan-out.
- Impact: Very high startup win.
- Risk: Low.
- Effort: Low.

### 2) Add `refresh.fetchOnStartup` config (default `false`)
- What: Keep refresh command behavior, but disable startup fetch by default.
- Impact: Very high where network fetch is slow.
- Risk: Low.
- Effort: Low.

### 3) Keep startup refresh but split fetch from build
- What: Show scan/build data immediately, run fetch updates asynchronously afterward.
- Impact: High first-paint improvement.
- Risk: Low-medium (state update timing).
- Effort: Medium.

### 4) Concurrency cap for startup refresh workers
- What: Replace “all repos at once” with bounded worker pool (e.g. 4–8).
- Impact: High stability under many repos; smoother CPU/network load.
- Risk: Low.
- Effort: Medium.

### 5) Prioritize visible rows first
- What: Refresh current viewport repos first; defer offscreen.
- Impact: High perceived speed.
- Risk: Medium (queue complexity).
- Effort: Medium.

### 6) Add per-repo fetch cooldown / TTL cache
- What: Skip fetch if recently fetched within N minutes.
- Impact: High for repeated launches.
- Risk: Medium (staleness tradeoff).
- Effort: Medium.

### 7) Add a “local-only startup mode”
- What: Startup computes only local status + branch, no remote fetch.
- Impact: Very high startup speed.
- Risk: Low.
- Effort: Low-medium.

### 8) Batch/merge branch commands
- What: Reduce `GetCurrentBranch` + `ListLocalBranches` + `FilterMergedBranches` command count.
- Impact: Medium per repo.
- Risk: Medium (parser correctness).
- Effort: Medium.

### 9) Defer merged-branch pruning candidates
- What: Compute `Branches.Merged` only when user triggers prune flow or row focus.
- Impact: Medium.
- Risk: Low.
- Effort: Medium.

### 10) Defer `LastCommitTime` and link metadata
- What: Render table without these fields first, fill in background.
- Impact: Medium first paint.
- Risk: Low.
- Effort: Low-medium.

### 11) Remove redundant `IsRepository` git call after `.git` discovery
- What: Trust `.git` directory detection by scanner unless edge case fallback needed.
- Impact: Medium startup command reduction.
- Risk: Low-medium (worktree edge cases).
- Effort: Low.

### 12) Smarter fetch strategy
- What: Use optional fetch flags/strategy (e.g. per remote, timeout tuning, prune toggle).
- Impact: Medium-high depending on network topology.
- Risk: Medium.
- Effort: Medium.

### 13) Stagger refresh with jitter
- What: Add small randomized delay per repo before fetch.
- Impact: Medium (reduces burst contention).
- Risk: Low.
- Effort: Low.

### 14) Persist previous run snapshot
- What: Load last-known repo states from disk and paint instantly; reconcile in background.
- Impact: Very high perceived speed.
- Risk: Medium (cache invalidation).
- Effort: Medium-high.

### 15) Add startup mode CLI flags
- What:
  - `--refresh-on-start` (bool)
  - `--fetch-on-start` (bool)
  - `--refresh-workers` (int)
  - `--local-only` (bool)
- Impact: Medium-high (operator control).
- Risk: Low.
- Effort: Low-medium.

## Recommended Rollout Order

1. Skip immediate full startup refresh (Option 1).
2. Add `fetchOnStartup=false` default + CLI/config override (Option 2 + 15 partial).
3. Add bounded worker pool for refresh (Option 4).
4. Defer merged-branch computation and other non-critical fields (Option 9 + 10).
5. Reduce command count in branch pipeline (Option 8).

## Success Metrics to Track

- Startup total duration.
- Time to first full table render.
- Git command count per repo at startup.
- `fetch` p50/p95/p99 duration.
- CPU/process spike during startup (optional).
- Manual refresh correctness parity (`r` still deterministic).
