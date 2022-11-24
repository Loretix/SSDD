go build internal/raft/raft.go 
go build cmd/srvraft/main.go 
go build pkg/cltraft/main.go

#go run cmd/srvraft/main.go 0 127.0.0.1:29001
go run cmd/srvraft/main.go 0 127.0.0.1:29001 127.0.0.1:29002&
go run cmd/srvraft/main.go 1 127.0.0.1:29001 127.0.0.1:29002&
#go run cmd/srvraft/main.go 0 127.0.0.1:29001 127.0.0.1:29002 127.0.0.1:29003
#go run cmd/srvraft/main.go 1 127.0.0.1:29001 127.0.0.1:29002 127.0.0.1:29003
#go run cmd/srvraft/main.go 2 127.0.0.1:29001 127.0.0.1:29002 127.0.0.1:29003

go run pkg/cltraft/main.go 127.0.0.1:3000 127.0.0.1:3001 127.0.0.1:3002

# --------------------------------------------------------------------------------- # 

go run cmd/srvraft/main.go 0 155.210.154.201:29001 155.210.154.191:29002 155.210.154.208:29003
go run cmd/srvraft/main.go 1 155.210.154.201:29001 155.210.154.191:29002 155.210.154.208:29003
go run cmd/srvraft/main.go 2 155.210.154.201:29001 155.210.154.191:29002 155.210.154.208:29003
