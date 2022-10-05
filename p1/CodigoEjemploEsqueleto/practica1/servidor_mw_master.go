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

 //gorutina
 func peticionesClientes(envioPeticiones chan net.Conn, workerIp string){
	//establecer conexión con el worker
	fmt.Println("conexión con worker")
	tcpAddr, err := net.ResolveTCPAddr("tcp", workerIp)
	checkError(err)
	fmt.Println("conexión con worker a medio camino")
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	checkError(err)
	fmt.Println("conexión con worker establecida")

	encoderW := gob.NewEncoder(conn)
	// recibimos respuesta workers
	decoderW := gob.NewDecoder(conn)
	var request com.Request
	var reply com.Reply

	for connClientes := range envioPeticiones {
		 encoderC := gob.NewEncoder(connClientes)
		 decoderC := gob.NewDecoder(connClientes)
		 // recibimimos la peticion de los clientes
		 decoderC.Decode(&request)
		 // enviamos la peticion a los workers 
		 encoderW.Encode(request)
		 // recibimos la respuesta de los workers
		 decoderW.Decode(&reply)
		 // enviamos la respuesta de los workers a los clientes
		 encoderC.Encode(reply)	 
	}
	close(envioPeticiones) 
 }
 
 func main() {
	 //CONN_HOST, CONN_PORT := os.Args[1], os.Args[2] //pasar argumentos al hacer el go run, por lo demás todo guay
	 //listener, err := net.Listen("tcp", CONN_HOST+":"+CONN_PORT)
	 listener, err := net.Listen("tcp", "127.0.0.1:30003")
	 checkError(err)
	 //creación de canales 
	 envioPeticiones := make (chan net.Conn) 
	 //vector de ip:puerto de los workers
	 //var datosWorker[2] string {"155.210.154.201:30001", "155.210.154.202:30002"}
	 // probatina en local
	 datosWorker := [2]string{"127.0.0.1:30001", "127.0.0.1:30005"}
	 
	 for i := 0; i < 2; i++{
		go peticionesClientes(envioPeticiones, datosWorker[i])
	 }
	 //envio peticiones de conexión de clientes
	 for{
		conn, err := listener.Accept() 
	 	checkError(err)
		envioPeticiones <- conn
	 }

 }
 