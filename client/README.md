### Dial
对地址拨号，建立一个TCP Conn，把conn挂在参数中对应的coder上。

### Go
把方法参数啥的构建成一个Call对象，赋值一个序列号，放在自有参数pending Map里面。
再用请求方法和序列号构建成一个Header。
利于自有coder，把header和参数写进conn里。

### Receive
发送和接受实际是异步的，有可能发送A B C，接受C B A。
接受信息要先读取header，用seq验证请求是否还在等待，如果A的回复到，但是A已经不等了，就直接丢弃。
依据Call的ch来通知调用方，回复已经到了。
