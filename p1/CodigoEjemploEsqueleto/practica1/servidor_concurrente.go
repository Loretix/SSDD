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

	defer conn.Close() //cerramos canal pq defer se ejecuta al final del uso
	// conn guarda la información de la conexión con el cliente
	// decode sera el receptor del mensaje
	decoder := gob.NewDecoder(conn)
	// decode sera el emisor del mensaje
	encoder := gob.NewEncoder(conn)
	// creamos una variable para almacenar el mensaje
	var request com.Request
	//recibir id e intervalo de elementos del cliente con el que se establece la conexión
	err := decoder.Decode(&request)
	checkError(err)
	//una vez recibido encontrar primos en el intervalor con FindPrimes
	var reply com.Reply
	reply.Primes = FindPrimes(request.Interval)
	reply.Id = request.Id
	//enviar al cliente los numeros primos
	err = encoder.Encode(reply)
	checkError(err)
}

func main() {
	//CONN_HOST, CONN_PORT := os.Args[1], os.Args[2] //pasar argumentos al hacer el go run, por lo demás todo guay
	//listener, err := net.Listen("tcp", CONN_HOST+":"+CONN_PORT)
	listener, err := net.Listen("tcp", "127.0.0.1:30003")
	checkError(err)

	for {
		conn, err := listener.Accept() //acepta a un cliente
		checkError(err)
		go procesarResultado(conn)
	}

}
