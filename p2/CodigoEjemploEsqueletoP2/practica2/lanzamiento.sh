go build -o lector cmd/lector/main.go
go build -o escritor cmd/escritor/main.go

./lector 1 3 &
./lector 2 3 &
./escritor 3 3 &