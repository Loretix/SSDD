/*
* AUTOR: Rafael Tolosana Calasanz
* ASIGNATURA: 30221 Sistemas Distribuidos del Grado en Ingeniería Informática
*			Escuela de Ingeniería y Arquitectura - Universidad de Zaragoza
* FECHA: septiembre de 2021
* FICHERO: ricart-agrawala.go
* DESCRIPCIÓN: Implementación del algoritmo de Ricart-Agrawala Generalizado en Go
*/
package ra

import (
    "ms"
    "sync"
)

type Request struct{
    Clock   int
    Pid     int
}

type Reply struct{}

type RASharedDB struct {
    OurSeqNum   int             // Nuestro numero de secuencia (reloj)
    HigSeqNum   int             // Mayor numero de secuencia (reloj)
    OutRepCnt   int             // Numero de procesos pendientes por confirmar acceso a SC
    ReqCS       boolean         // cs_state = si etamos en SC
    RepDefd     int[]           // Procesos que esperan nuestra confirmacion
    ms          *MessageSystem  
    done        chan bool       // avisamos al salir de la SC  
    chrep       chan bool       // chan replay (para la respuesta) y saber si tenemos permiso para acceder a la SC 
    Mutex       sync.Mutex      // mutex para proteger concurrencia sobre las variables
    // TODO: completar
    Me          int             // Mi número de nodo 
    N           int             // numero de nodoso en la red
}


func New(me int, usersFile string) (*RASharedDB) {
    messageTypes := []Message{Request, Reply}
    msgs = ms.New(me, usersFile string, messageTypes)
    ra := RASharedDB{0, 0, 0, false, []int{}, &msgs,  make(chan bool),  make(chan bool), &sync.Mutex{}}
    // TODO completar
    return &ra
}

//Pre: Verdad
//Post: Realiza  el  PreProtocol  para el  algoritmo de
//      Ricart-Agrawala Generalizado
//El preprotocol para un Proceso Pi consiste en enviar a los N - 1 
//procesos distribuidos una peticion de acceso a la seccion critica. 
//Cuando se reciba la respuesta en (5) de los N-1 entonces el Proceso 
//Pi accede a la seccion critica.

func (ra *RASharedDB) PreProtocol(){
    // cs_state <- trying 
    ra.Mutex.Lock() 
    ra.ReqCS = true                      // indicamos que queremos acceder a la SC 
    ra.OurSeqNum = ra.HigSeqNum + 1      // aumentamos el reloj interno 
    ra.Mutex.Unlock() 
    
    ra.OutRepCnt = ra.N - 1              // Numero de procesos que confirman la entrada a SC
    for j := 1; j < ra.N; j++ {
        if j != ra.Me {
            // send (enviamos la peticion de acceso a la SC)
        }
	}

    for ra.OutRepCnt != 0 {
        <-ra.chrep                          // esperamos respuesta (leer del canal)
        ra.OutRepCnt =  ra.OutRepCnt - 1    // decrementas contador de numero de procesos que faltan 
    }   
    ra.ReqCS = false            // puede que no vaya aqui :/ (igual al inicio del post)-------------------------------------------------------------------
}

//Pre: Verdad
//Post: Realiza  el  PostProtocol  para el  algoritmo de
//      Ricart-Agrawala Generalizado
// Enviamos permiso a los procesos que teniamos esperando la confirmacion
// una vez salimos de nuestra SC
func (ra *RASharedDB) PostProtocol(){
    for j := 1; j < ra.N; j++ {
        if ra.RepDefd[j] {
            ra.RepDefd[j] = false           
            // send nuestro permiso para que j entre en SC
        }
	}
}

func (ra *RASharedDB) Stop(){
    ra.ms.Stop()
    ra.done <- true
}
