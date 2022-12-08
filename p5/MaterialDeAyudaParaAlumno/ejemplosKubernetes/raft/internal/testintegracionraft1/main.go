package main

import (
	"fmt"
	"raft/internal/comun/rpctimeout"
	"raft/internal/raft"
	"time"
)

/*
	REPLICA1 = "ss-0.ss-service.default.svc.cluster.local:6000"
	REPLICA2 = "ss-1.ss-service.default.svc.cluster.local:6000"
	REPLICA3 = "ss-2.ss-service.default.svc.cluster.local:6000"
*/
func main() {

	var endPoint []rpctimeout.HostPort
	// Resto de argumento son los end points como strings
	// De todas la replicas-> pasarlos a HostPort´
	fmt.Println("Comenzamos el cliente")
	var direcciones [3]string 
	direcciones[0]= "ss-0.ss-service.default.svc.cluster.local:6000" 
	direcciones[1]= "ss-1.ss-service.default.svc.cluster.local:6000"
	direcciones[2]= "ss-2.ss-service.default.svc.cluster.local:6000"

	for j := 0; j < 5; j++ {
		respuestaLider := false

		// preparamos los argumentos
		var args raft.TipoOperacion
		args.Operacion = "escritura"
		args.Valor = "hola"
		var reply raft.ResultadoRemoto
		fmt.Println("Direcciones", direcciones[0:])
		endPoint = rpctimeout.StringArrayToHostPortArray(direcciones[0:])
		fmt.Println("Enpoint ya con las direccion", endPoint)
		// enviamos a los nodos hasta encontrar al lider
		for i := 0; i < len(endPoint) && !respuestaLider; i++ {
			fmt.Println("Comenzamos el envio de solicitud : ", endPoint[i])
			err := endPoint[i].CallTimeout("NodoLE.LecEsc", &args, &reply, time.Duration(25)*time.Millisecond)
			if err == nil {
				respuestaLider = true
				fmt.Println("El servidor nos ha respondido")

			} else {
				fmt.Println("No se ha recibido respuesta por parte del servidor: ", err)
			}
		}
		time.Sleep(time.Duration(300) * time.Millisecond)
	}

}
