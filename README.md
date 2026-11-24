# Redis Go

A Redis-compatible server written in Go as part of the
["Build Your Own Redis" challenge](https://codecrafters.io/challenges/redis).

The server communicates using RESP over TCP and includes an in-memory data store, command handlers, replication, persistence, Pub/Sub, transactions, authentication, and ACL support.

## Features

- RESP request parsing and response encoding
- General commands: `PING`, `ECHO`, `TYPE`, `INFO`, `CONFIG`, `KEYS`
- String commands: `SET`, `GET`, `INCR`
- List commands: `RPUSH`, `LPUSH`, `LRANGE`, `LLEN`, `LPOP`, `BLPOP`
- Stream commands: `XADD`, `XRANGE`, `XREAD`
- Sorted set commands: `ZADD`, `ZRANK`, `ZRANGE`, `ZCARD`, `ZSCORE`, `ZREM`
- Bitmap commands: `SETBIT`, `GETBIT`, `STRLEN`, `BITCOUNT`, `BITOP`
- Geospatial commands: `GEOADD`, `GEOPOS`, `GEODIST`, `GEOSEARCH`
- Pub/Sub: `SUBSCRIBE`, `UNSUBSCRIBE`, `PUBLISH`
- Transactions: `WATCH`, `UNWATCH`, `MULTI`, `EXEC`, `DISCARD`
- Authentication and ACL: `AUTH`, `ACL`
- Replication: `INFO`, `REPLCONF`, `PSYNC`, `WAIT`
- Persistence: RDB loading and optional AOF read/write

## Requirements

- Go 1.27 or newer
- `redis-cli` or another Redis-compatible client for testing

## Running the Server

Start the server on the default port:

```sh
go run ./cmd/redis
```

Connect with `redis-cli`:

```sh
redis-cli -p 6379
```

## Command-Line Options

Run on another port:

```sh
go run ./cmd/redis --port 6380
```

Start as a replica:

```sh
go run ./cmd/redis --port 6380 --replicaof "127.0.0.1 6379"
```

| Option | Default | Description |
|---|---:|---|
| `--port`, `-p` | `6379` | TCP listening port |
| `--replicaof` | empty | Host and port of the master server |
| `--dir` | current directory | Directory containing persistence files |
| `--dbfilename` | `dump.rdb` | RDB filename |
| `--appendonly` | `no` | Enable AOF persistence |
| `--appenddirname` | `appendonlydir` | AOF directory name |
| `--appendfilename` | `appendonly.aof` | AOF filename |
| `--appendfsync` | `everysec` | AOF synchronization policy |

## Project Structure

```text
internal/
  cmd_router/       Command dispatch and command handlers
  collections/      Custom linked lists and queues
  errors/           Redis-compatible error definitions
  resp/             RESP types, parsing, and encoding
  server/           Connections, replication, Pub/Sub, auth, and persistence
  store/            In-memory storage, data types, and RDB support
```

## Persistence

The server can:

- Load an RDB file on startup
- Replay an AOF file when `appendonly=yes`
- Append write commands to the configured AOF file

Configure persistence through the `--dir`, `--dbfilename`, `--appendonly`,
`--appenddirname`, `--appendfilename`, and `--appendfsync` options.


# Notes for Myself

1. set with expiry has bugs. New set, invalidates earlier
9. remove the strong coupling of server and client somehow. Is it required ?
11. Replication manager and seperate its functionality
12. Event loop instead of go routines ?
