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

---

### Dial
Dial the address, create a TCP Conn, and hook the conn to the corresponding coder in the argument.

### Go
Build a Call object from the method parameters and so on, assign a sequence number, and put it in its own parameter Pending Map.
Then use the request method and sequence number to build a header.
Benefit from your own coder, write the header and parameters to the conn.

### Receive
Send and receive are actually asynchronous, with the possibility of sending A B C and receiving C B A.
To receive a message, read the header first, use seq to verify if the request is still waiting, if A's reply arrives but A is no longer waiting, just discard it.
Based on Call's ch to notify the caller that the reply has arrived.