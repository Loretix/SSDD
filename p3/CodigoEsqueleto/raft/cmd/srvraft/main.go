package main

import (
	"fmt"
	"net"
	"net/rpc"
	"os"
	"raft/internal/comun/check"
	"raft/internal/comun/rpctimeout"
	"raft/internal/raft"
	"strconv"
 "strings"
  "time"
)

type NodoLE struct {
	Nr *raft.NodoRaft
}


func (nodo *NodoLE) LecEsc(args *raft.TipoOperacion, reply *raft.ResultadoRemoto) error {
	fmt.Println("Estoy en LecEsc")
	if nodo.Nr.Roll == raft.LIDER {
		reply.ValorADevolver = args.Valor
		nodo.Nr.SometerOperacionRaft(*args, reply)

	}
	return nil
}

func conexionCliente(nodo *NodoLE, me int, ip string) {

	// Parte Servidor
	rpc.Register(nodo)
	cliente, err := net.Listen("tcp", ip[:strings.Index(string(ip), ":")]+":300"+strconv.Itoa(me))
	check.CheckError(err, "Error escuchando al cliente")
	for {
		conn, err := cliente.Accept()
		check.CheckError(err, "Error con la conexión al cliente")
		go rpc.ServeConn(conn)
	}

	time.Sleep(100 * time.Millisecond)

}

func main() {
	nodo := new(NodoLE)
	// obtener entero de indice de este nodo
	me, err := strconv.Atoi(os.Args[1])
	check.CheckError(err, "Main, mal numero entero de indice de nodo:")
	var nodos []rpctimeout.HostPort
	// Resto de argumento son los end points como strings
	// De todas la replicas-> pasarlos a HostPort
	for _, endPoint := range os.Args[2:] {
		nodos = append(nodos, rpctimeout.HostPort(endPoint))
	}
	// Parte Servidor
	nodo.Nr = raft.NuevoNodo(nodos, me, make(chan raft.AplicaOperacion, 1000))
	rpc.Register(nodo.Nr)

	fmt.Println("Replica escucha en :", me, " de ", os.Args[2:])

	l, err := net.Listen("tcp", os.Args[2:][me])
	check.CheckError(err, "Main listen error:")
	go conexionCliente(nodo, me, os.Args[2+me])
	rpc.Accept(l)
}
