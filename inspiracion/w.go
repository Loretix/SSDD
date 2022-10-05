package main

//Ejecutar go run ./Worker.go hostIP:puerto
//hostIP = Ip propia del worker con el exterior (Al que se le conecta el proxy)
//puerto = El puerto al que se le conecta el proxy

type respuestaWorker struct {
	found    []int
	conexion net.Conn
}

func calcularResultadoWorker(conn net.Conn, clientConn int) {
	encoder := gob.NewEncoder(conn)

	foundPrimes := FindPrimes(60000)
	fmt.Println("Worker -> Trabajo terminado, a mandar resultados")
	encoder.Encode(foundPrimes)
}

func main() {
	fmt.Println("Worker -> Creado")
	listener, err := net.Listen(CONN_TYPE, os.Args[1])
	checkError(err)
	conn, err := listener.Accept()
	if err != nil {
		os.Exit(1)
	}
	dec := gob.NewDecoder(conn)
	fmt.Println("Worker -> Conexion con proxy establecida")
	for {
		var clientConn int
		fmt.Println("Worker ->Creado cliente")
		dec.Decode(&clientConn)
		fmt.Println("Worker -> Obtengo trabajo e id. A trabajar !")
		calcularResultadoWorker(conn, clientConn)
	}
}