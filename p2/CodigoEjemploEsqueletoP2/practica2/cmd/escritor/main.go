package main

import (
	"fmt"
	"os"
	"practica2/gestorfichero"
	"practica2/ra"
	"strconv"
	"time"
)

func main() { // Escritor main

	me, err := strconv.Atoi(os.Args[1])
	gestorfichero.CheckError(err)
	N, err := strconv.Atoi(os.Args[2])
	gestorfichero.CheckError(err)

	raLector := ra.New(me, N, 1, "./ms/users.txt") // Creacion ra (el 1 indica que son escritores)
	go raLector.Recibir()                          // lanzar la gorutina de recibir
	for {
		time.Sleep(time.Duration(50) * time.Millisecond)
		fmt.Println(strconv.Itoa(me) + "- Empezamos el tratamiento de SC")
		raLector.PreProtocol()                                              // ejecutamos el preprotocol pq queremos acceder a SC
		gestorfichero.EscribirFichero("hola soy "+strconv.Itoa(me)+"\n", N) // accedemos a SC (leemos funcion Leer() )
		raLector.PostProtocol()                                             // ejecutamos el postprotocolo (salimos de SC)

	}
}
