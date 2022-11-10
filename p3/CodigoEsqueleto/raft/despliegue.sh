go build internal/raft/raft.go 
go build cmd/srvraft/main.go 
go build pkg/cltraft/main.go

#go run cmd/srvraft/main.go 0 127.0.0.1:29001
go run cmd/srvraft/main.go 0 127.0.0.1:29001 127.0.0.1:29002&
go run cmd/srvraft/main.go 1 127.0.0.1:29001 127.0.0.1:29002&
#go run cmd/srvraft/main.go 0 127.0.0.1:29001 127.0.0.1:29002 127.0.0.1:29003&
#go run cmd/srvraft/main.go 1 127.0.0.1:29001 127.0.0.1:29002 127.0.0.1:29003&
#go run cmd/srvraft/main.go 2 127.0.0.1:29001 127.0.0.1:29002 127.0.0.1:29003&