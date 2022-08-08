# Design

---
*eth_block_indexer* contain two services. API to get block and transaction related information and Ethereum block indexer

### YAML config example
eth_block_indexer use config.yml to configurate
example:
```
core:
  start_block_num: 21709284 # the latest I know block number
  worker_num: 0 # default worker number is runtime.NumCPU()
  queue_num: 0 # default queue number is 2
  address: ""
  http_port: "8080"
  https_port: "8081"
  mode: "release"
api:
  blocks_uri: "/blocks"
  block_by_id_uri: "/blocks/:id"
  transaction_uri: "/transaction/:txHash"
log:
  format: "string" # string or json
  access_log: "/var/eth_block_indexer_log" # stdout: output to console,or define log path like "log/access_log"
  access_level: "trace"
  error_log: "/var/eth_block_error_log" # stderr: output to console,or define log path like "log/error_log"
  error_level: "trace"
```

# Indexer db schema

---
### *blocks*

| Name | DataType |
| ------ | ------ |
| ID   | uint (primary key)   |
| create_at   | Date   |
| updated_at   | Date   |
| deleted_at   | Date   |
| block_number   | uint64   |
| block_hash   | bytea   |
| block_time   | uint64   |
| parent_hash   | bytea   |

### *transactions*

| Name | DataType |
| ------ | ------ |
| ID   | uint (primary key)   |
| create_at   | Date   |
| updated_at   | Date   |
| deleted_at   | Date   |
| block_number   | uint64   |
| from   | bytea   |
| to   | bytea   |
| nonce   |  uint64  |
| data   |  bytea  |
| value   | uint64   |

### *transaction_logs*

| Name | DataType |
| ------ | ------ |
| ID   | uint (primary key)   |
| create_at   | Date   |
| updated_at   | Date   |
| deleted_at   | Date   |
| tx_hash   | bytea   |
| index   | bytea   |
| data   | bytea   |

