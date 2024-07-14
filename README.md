# My Redis Server

![coverage badge](https://img.shields.io/endpoint?url=https://gist.githubusercontent.com/burenotti/cdabc4087e0fb9c2ec9c827cef65974e/raw/redis_impl__refs_heads_master.json)
**👈 I will fix it. I promise...**

This is a simple redis implementation written in Go.
It is developed for only for educational purposes, so it has no runtime dependencies
(for tests go-mock and testify are used) and not supposed to be used in real world projects.

The main goal of this project is to implement complicated parts of redis mainly persistence and replication.

## Features

- [x] PING
- [x] Pipelining
- [x] GET/SET
- [x] Transactions
- [x] Keys expiration
- [ ] Key eviction
- [ ] Key eviction policies
- [ ] Data structures:
    - [ ] List
    - [ ] Sorted set
    - [ ] Hash map
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
docker run -v ./config/redis.conf:/etc/redis/redis.conf -it -p 6379:6379 burenotti/redis_impl
```

## Configuring

This implementation is configured using yaml file, that is passed to server using `-config` parameter.

```redis
bind 127.0.0.1
port 8379
shutdown_timeout 5
```

## Packages

### Redis config file reader `pkg/conf`

This package allows to bind redis .conf files to go structures.

##### Supported tags

- `redis:"name-of-the-field"` – Name of the field in .conf file
- `redis-required:""` – Field must be presented in .conf file
- `redis-default:"default-value"` – Specifies default value for struct field
- `redis-prefix:"memory_"` – Prefix for all fields in nested structures

#### Example

Assume that we have this configuration file:

```redis
bind 0.0.0.0
port 6379

maxmemory_mb 1024
maxmemory_policy allkeys-lru
```

So we can use this code to parse the file.

```go
package main

import (
	"fmt"
	"github.com/burenotti/redis_impl/pkg/conf"
)

type Config struct {
	Bind      string `redis:"bind" redis-default:"localhost"`
	Port      int    `redis:"int" redis-default:"6379"`
	MaxMemory struct {
		MB     int    `redis:"mb" redis-default:"100"`
		Policy string `redis:"policy" redis-default:"no-evict"`
	} `redis-prefix:"maxmemory_"`
}

func main() {
	var cfg Config
	if err := conf.BindFile(&cfg, "file"); err != nil {
		panic(err.Error())
	}

	fmt.Print(cfg)
}

```

### Redis serialization protocol parser `pkg/resp`

This library implements only RESP of version 2.

- `Marshal(w io.Writer, value interface{}) error` – marshals value.
- `Unmarshal(r ReaderPeaker) (interface{}, error)` – Reads next value from reader.

| Go Type     | RESP2 Type    | RESP prefix |
|-------------|---------------|-------------|
| string      | Simple String | +           |
| error       | Error         | -           |
| []byte      | Bulk String   | $           |
| int64       | Integer       | :           |
| []interface | Array         | *           |

### Algorithms & generic data structures `pkg/algo`

- `algo/heap` – Heap
- `algo/queue` – Linked list queue
- `algo/set` – AVL-Tree sorted set

## Implementation details

This section explains design of the project.

**Coming soon...**