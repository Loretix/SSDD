package main

import (
	"os"
	"practica2/gestorfichero"
	"practica2/ra"
	"strconv"
)

func main() { // Escritor main

	me, err := strconv.Atoi(os.Args[1])
	gestorfichero.CheckError(err)
	N, err := strconv.Atoi(os.Args[2])
	gestorfichero.CheckError(err)

	raLector := ra.New(me, N, 1, "../ms/users.txt") // Creacion ra (el 1 indica que son escritores)
	go raLector.Recibir()                           // lanzar la gorutina de recibir
	for {

		raLector.PreProtocol()                // ejecutamos el preprotocol pq queremos acceder a SC
		gestorfichero.EscribirFichero("hola") // accedemos a SC (leemos funcion Leer() )
		raLector.PostProtocol()               // ejecutamos el postprotocolo (salimos de SC)

		//time.Sleep(time.Duration(50) * time.Microsecond)

	}
}
