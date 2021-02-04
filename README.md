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

goworker : 即便把先后2条消息投递到同一个 queue，也无法保证消费顺序。 因为一个 queue 里的 job，可能会给多个并行工作的 worker 消费。

如何保证消费顺序： concurrency = 1，也就是只有一个 worker    
 

一般来说，mq 的消费顺序保证：  
1. 需要保证顺序的消息投递到同一个 queue/partition/... 
2. 同一个queue/partition/... 只允许有一个 worker 处理消息。 

goworker 中 queue 并没有与 worker 绑定，所以只能让全局(所有queue) 共用一个 worker 才能解决消费顺序问题.. 

投递顺序保证了，如果因为网络原因 mq 收到消息的先后顺序不一致怎么办？        
args 里带一个 seq id ，消费时开发者根据 seq id 自己去判断 

## # 程序退出 

收到 quit, terminal, interrupt 信号后。 消费端会先停止 poller，然后 workers 会先把手上的活给干完再停止。 

## # failure mode 

消息丢失?   

消息不会持久化到磁盘(?? 利用了 redis 的持久化?) ?? poller lpop 出来后，还没到 worker 手上就丢了怎么办？  

