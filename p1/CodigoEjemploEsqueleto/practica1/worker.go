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

	//"io"
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
func procesarResultado(conn net.Conn) {

	// decode sera el receptor del mensaje
	decoder := gob.NewDecoder(conn)
	// decode sera el emisor del mensaje
	encoder := gob.NewEncoder(conn)
	// creamos una variable para almacenar el mensaje
	var request com.Request
	//recibir id e intervalo de elementos del cliente con el que se establece la conexión
	err := decoder.Decode(&request)
	checkError(err)
	fmt.Println(request)
	fmt.Println("Recibir peticion del cliente mediante el master")
	//una vez recibido encontrar primos en el intervalor con FindPrimes
	var reply com.Reply
	reply.Primes = FindPrimes(request.Interval)
	reply.Id = request.Id
	//enviar al cliente los numeros primos
	err = encoder.Encode(reply)
	checkError(err)
	fmt.Println("Enviar respuesta al master")
	conn.Close()
}

func main() {
	CONN_HOST, CONN_PORT := os.Args[1], os.Args[2]
	listener, err := net.Listen("tcp", CONN_HOST+":"+CONN_PORT)
	checkError(err)

	for {
		conn, err := listener.Accept()
		checkError(err)
		procesarResultado(conn)
	}

}
