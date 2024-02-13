### Order

#### Заупскаем приложение

```shell
$ cat ~.env | xargs
$ go run ./cmd/order/main.go
```

#### Запускаем Nats-Streaming

в проекте лежит экзешник

```shell
$ ./nats-server // or ./nats-server.exe if use Windows

> [14172] 2023/08/30 19:43:55.604859 [INF] Starting nats-server
> [14172] 2023/08/30 19:43:55.647513 [INF]   Version:  2.9.21
> [14172] 2023/08/30 19:43:55.647513 [INF]   Git:      [b2e7725]
> [14172] 2023/08/30 19:43:55.647513 [INF]   Name:     NBE6HFAF2SAVPRIAP7PGZUWSLNM5YARVRO5WP3AIWYGOL5M4KVIGQJVU
> [14172] 2023/08/30 19:43:55.647513 [INF]   ID:       NBE6HFAF2SAVPRIAP7PGZUWSLNM5YARVRO5WP3AIWYGOL5M4KVIGQJVU
> [14172] 2023/08/30 19:43:55.652155 [INF] Listening for client connections on 0.0.0.0:4222
> [14172] 2023/08/30 19:43:55.677002 [INF] Server is ready
```