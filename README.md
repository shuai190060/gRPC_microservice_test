## gRPC vs REST

- use both REST and gRPC to push data to PostgreSQL database
- monitor metrics (cpu, latency)


## Preparation

Command to generate the protofiles

```jsx
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative account_proto/account.proto
```

