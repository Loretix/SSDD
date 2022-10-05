func crearNodo(cs chan net.Conn) {
	for {
		conexionActual := <-cs
		calcularResultado(conexionActual)
	}
}

func Proxy(csp chan net.Conn, ip string) {
	fmt.Println(ip)
	tcpAddr, err := net.ResolveTCPAddr("tcp4", ip)
	checkError(err)

	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	//defer conn.Close()
	checkError(err)

	fmt.Println("Conexion con worker establecida")
	encoder := gob.NewEncoder(conn)
	dec := gob.NewDecoder(conn)

	for {
		conexionActual := <-csp
		fmt.Println("Proxy recibe cliente")
		encoder.Encode(0)
		fmt.Println("Proxy envia cliente")

		var recibe []int
		dec.Decode(&recibe)
		fmt.Println("Proxy recibe respuesta")

		encoderCliente := gob.NewEncoder(conexionActual)
		encoderCliente.Encode(recibe)
		fmt.Println("Proxy envia respuesta")
		conexionActual.Close()
	}

}

func servidor4(argumentos []string) {

	csp := make(chan net.Conn, 20) //Create sync channel with proxies

	for i := 2; i < len(argumentos); i++ { //Create proxies to operate.
		fmt.Println("Creado un proxy")
		go Proxy(csp, argumentos[i])
	}
	fmt.Println("Proxys creados")
	listener, err := net.Listen(CONN_TYPE, CONN_HOST+":"+CONN_PORT)
	checkError(err)

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		fmt.Println("Acepto conexion, la añado a csp")
		csp <- conn
	}
}

