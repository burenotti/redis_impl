# My Redis Server

![coverage badge](https://img.shields.io/endpoint?url=https://gist.githubusercontent.com/burenotti/cdabc4087e0fb9c2ec9c827cef65974e/raw/redis_impl__refs_heads_master.json)

This is a simple redis implementation written in Go.
It is developed for only for educational purposes and not supposed to be used in real world projects.

The main goal of this project is to implement complicated parts of redis mainly persistence and replication.

## Features

- [x] PING
- [x] Pipelining
- [x] GET/SET
- [x] Transactions
- [ ] Keys expiration
- [ ] Key eviction
- [ ] Key eviction policies
- [ ] Other data structures ()
- [ ] Persistence:
    - [ ] Append only file
        - [ ] AOF compression
    - [ ] Snapshot
- [ ] Replication

## Building & Running

**Make**

```shell
# Clone the repository
git clone https://github.com/burenotti/redis_impl.git
cd redis_impl

# Build server using make
make build

# Run server
./build/redis_liunx_amd64 -config <path_to_config>
```

**Docker**

```shell
# Clone the repository
git clone https://github.com/burenotti/redis_impl.git
cd redis_impl

# Build docker image
docker build -t burenotti/redis .

# Run built image 
docker run -v ./config.yaml:/etc/redis/config.yaml -it -p 6379:6379 burenotti/redis_impl
```

## Configuring

This implementation is configured using yaml file, that is passed to server using `-config` parameter.

```yaml
server:
  host: 0.0.0.0
  port: 7379
  shutdown_timeout: 10s
```

## Implementation details

This section explains design of the project.

**Coming soon...**