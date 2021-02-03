# goworker 阅读版

源码阅读版 (带着问题去读源码) 。 

# Sched   

- [x] 消息顺序保证 

# Q&A

## # usage  

goworker 是消息生产/消费端，消息通过 redis list 传递。 

作为消费端，goworker 使用一个 poller goroutine 不断 lpop jobs(poll) 然后给多个并行的 worker 消费 jobs。  

作为生产端，goworker.Enqueue() 把 job rpush 到 redis 中。  

```
producer(rpush) -----job-----> redis list <----jobs------ poller(lpop)
                                                                |
                                                                | fan-out jobs
                                                               ↙↓↘
                                                      worker  worker  worker 
```

它兼容 resque.....  所以，resque producer 投递的消息，可以直接用 goworker 来作为消费服务。

## # 消息顺序保证  

goworker : 即便把先后2条消息投递到同一个 queue，也无法保证消费顺序。

poller 会把 jobs fanout 给多个并行工作的 worker。 顺序有先后的2条消息即便按顺序投递到同一个 queue，也可能同时被2个 worker 接收，所以 worker消费顺序无法保证。 

改造 goworker 让其支持 queue 与 worker 绑定?  args 里带一个 seq id ，消费时开发者根据 seq id 自己去判断？ 

## # 程序退出 

收到 quit, terminal, interrupt 信号后。 消费端会先停止 poller，然后 workers 会先把手上的活给干完再停止。 

## # failure mode 

消息丢失?   

消息不会持久化到磁盘(?? 利用了 redis 的持久化?) ?? poller lpop 出来后，还没到 worker 手上就丢了怎么办？  

