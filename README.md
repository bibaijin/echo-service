# echo-service

一个 [LAIN](https://github.com/laincloud/lain)
[service](https://laincloud.gitbooks.io/white-paper/usermanual/service.html)
的例子。

## 功能

它接受 tcp 连接，然后会回写客户端发来的信息，信息以 '\n' 为分隔符。

## 说明

它的 portal 使用 [proxyd](https://github.com/bibaijin/proxyd)。
