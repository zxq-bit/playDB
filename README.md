# playDB
my replica db for play

## tips

### split

作为一个命令由 dataNode 来执行，对 raft 透明

### raft

采用 etcd 的 raft

因单节点上会运行大量复制组，需要 multi raft，而 etcd 版的有单独 Tick，且本身使用较广

注意 snapshot 需要自己维护并管理，传输也是自己负责

### store

目前实现为全量内存 cache

采用 google 的 btree 的 golang 实现

- 简单实现，先不用 c/c++ 版本
- 单线程写， copy on write 实现，不用处理锁
- 代码简单易理解、调试、修改

### 调度使用 kubernetes ?

有考虑过各个复制组都单独作为一个 deployment(replica:n) 来进行管理，把调度丢出去，但考虑了下还是有不少问题

- 分裂可能会非常花时间，带来的写操作阻塞可能难以接收
- k8s 的调度实现比较简单，对于偶发的分布问题导致无法调度时会很难搞
  - 同复制组的 pod 必须互相排斥，可能出现总资源有剩余但无法调度的情况
  - k8s 不会杀 pod，需要 manager 自行处理，脱离了省事的初衷
  - 过于依赖 k8s，耦合会很严重
- io 性能问题
  - raft 本身就会导致写操作很慢，容器的 io 性能还会进一步下降
