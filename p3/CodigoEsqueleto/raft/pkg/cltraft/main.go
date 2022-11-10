package main

import (
	"fmt"
	"os"
	"raft/internal/comun/rpctimeout"
	"raft/internal/raft"
	"time"
)

func main() {

	var endPoint []rpctimeout.HostPort
	// Resto de argumento son los end points como strings
	// De todas la replicas-> pasarlos a HostPort´
	fmt.Println("Comenzamos el cliente")
	for j := 0; j < 1; j++ {
		respuestaLider := false

		// preparamos los argumentos
		var args raft.TipoOperacion
		args.Operacion = "escritura"
		args.Valor = "hola"
		var reply raft.ResultadoRemoto
		endPoint = rpctimeout.StringArrayToHostPortArray(os.Args[1:])

		// enviamos a los nodos hasta encontrar al lider
		for i := 1; i <= len(endPoint) && !respuestaLider; i++ {
			fmt.Println("Comenzamos el envio de solicitud : ", endPoint[i-1])
			err := endPoint[i-1].CallTimeout("NodoLE.LecEsc", &args, &reply, time.Duration(25)*time.Millisecond)
			if err == nil {
				respuestaLider = true
				fmt.Println("El servidor nos ha respondido")

			} else {
				fmt.Println("No se ha recibido respuesta por parte del servidor: ", err)
			}
		}
		time.Sleep(time.Duration(100) * time.Millisecond)
	}

}
