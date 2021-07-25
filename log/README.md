# log
A simple golang log library.


This project bases on [uber/zap](https://github.com/uber-go/zap), a wonderful log library for golang.


### Trace logs
Logs should be tracable. Logs in one request or one flow should be grouped.

Example:
```go
package main

import (
    "context"

    "github.com/gosolu/solu/log"
)

func doThings(ctx context.Context, num int) {
    log.In(ctx).With(log.Int("num", num)).Info("do things")
    // other logic codes
}

func main() {
    ctx := context.Background()
    ctx = log.Trace(ctx) // Trace this context

    doThings(ctx, 1)
    doThings(ctx, 2)
}
```
It will log somethings like below:
```
{"g_level":"info","g_ts":"2021-01-17T21:18:42+08:00","g_caller":"example/main.go:10","g_msg":"do things","g_tid":"93afab744b0139d6511353d92fce6fc5","num":1}
{"g_level":"info","g_ts":"2021-01-17T21:18:42+08:00","g_caller":"example/main.go:10","g_msg":"do things","g_tid":"93afab744b0139d6511353d92fce6fc5","num":2}
```


### Rotate Log File
Why rotate?

- If you keep write logs into a file, it may grow too large.
- You may want split your log files daily.



### LICENSE
MIT LICENSE

