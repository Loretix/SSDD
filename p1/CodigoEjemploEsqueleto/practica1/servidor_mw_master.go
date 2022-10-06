/*
* AUTOR: Rafael Tolosana Calasanz
* ASIGNATURA: 30221 Sistemas Distribuidos del Grado en Ingeniería Informática
*			Escuela de Ingeniería y Arquitectura - Universidad de Zaragoza
* FECHA: septiembre de 2021
* FICHERO: server.go
* DESCRIPCIÓN: contiene la funcionalidad esencial para realizar los servidores
*				correspondientes a la práctica 1
 */
package main

import (
	"encoding/gob"
	"fmt"
	"net"
	"os"
	"practica1/com"
)

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
}

// PRE: verdad
// POST: IsPrime devuelve verdad si n es primo y falso en caso contrario
func IsPrime(n int) (foundDivisor bool) {
	foundDivisor = false
	for i := 2; (i < n) && !foundDivisor; i++ {
		foundDivisor = (n%i == 0)
	}
	return !foundDivisor
}

// PRE: interval.A < interval.B
// POST: FindPrimes devuelve todos los números primos comprendidos en el
//
//	intervalo [interval.A, interval.B]
func FindPrimes(interval com.TPInterval) (primes []int) {
	for i := interval.A; i <= interval.B; i++ {
		if IsPrime(i) {
			primes = append(primes, i)
		}
	}
	return primes
}

// gorutina
func peticionesClientes(envioPeticiones chan net.Conn, workerIp string) {
	//establecer conexión con el worker

	for {
		connClientes := <-envioPeticiones
		//establecer conexión con el worker
		fmt.Println("conexión con worker")
		tcpAddr, err := net.ResolveTCPAddr("tcp", workerIp)
		checkError(err)
		conn, err := net.DialTCP("tcp", nil, tcpAddr)
		checkError(err)
		fmt.Println("conexión con worker establecida")

		encoderW := gob.NewEncoder(conn)
		// recibimos respuesta workers
		decoderW := gob.NewDecoder(conn)
		var request com.Request
		var reply com.Reply

		encoderC := gob.NewEncoder(connClientes)
		decoderC := gob.NewDecoder(connClientes)
		// recibimimos la peticion de los clientes
		err = decoderC.Decode(&request)
		checkError(err)
		fmt.Println(request)
		fmt.Println("Recibir peticion del cliente")
		// enviamos la peticion a los workers
		err = encoderW.Encode(request)
		checkError(err)
		fmt.Println("Peticion enviada a los Workers")
		// recibimos la respuesta de los workers
		err = decoderW.Decode(&reply)
		checkError(err)
		fmt.Println("Recibir respuesta de los workers")
		// enviamos la respuesta de los workers a los clientes
		err = encoderC.Encode(reply)
		checkError(err)
		//connClientes.Close()
	}
}

func main() {
	//CONN_HOST, CONN_PORT := os.Args[1], os.Args[2] //pasar argumentos al hacer el go run, por lo demás todo guay
	//listener, err := net.Listen("tcp", CONN_HOST+":"+CONN_PORT)
	listener, err := net.Listen("tcp", "127.0.0.1:30050")
	checkError(err)
	//creación de canales
	envioPeticiones := make(chan net.Conn)
	//vector de ip:puerto de los workers
	datosWorker := [2]string{"127.0.0.1:30080", "127.0.0.1:30081"}

	for i := 0; i < 2; i++ {
		go peticionesClientes(envioPeticiones, datosWorker[i])
	}
	//envio peticiones de conexión de clientes
	for {
		conn, err := listener.Accept()
		checkError(err)
		envioPeticiones <- conn
	}
}
