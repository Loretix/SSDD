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
	"fmt"
	"practica2/ms"
	"strconv"
	"sync"
	"time"

	"github.com/DistributedClocks/GoVector/govec"
	"github.com/DistributedClocks/GoVector/govec/vclock"
)

type Request struct {
	Clock int
	Pid   int
	op_t  int // 1 -> escritores 0 -> lectores
}

type Reply struct{}

type RelojVector struct {
	VReloj vclock.VClock
}

type RASharedDB struct {
	OurSeqNum int   // Nuestro numero de secuencia (reloj)
	HigSeqNum int   // Mayor numero de secuencia (reloj)
	OutRepCnt int   // Numero de procesos pendientes por confirmar acceso a SC
	ReqCS     bool  // cs_state = si etamos en SC
	RepDefd   []int // Procesos que esperan nuestra confirmacion
	ms        *ms.MessageSystem
	done      chan bool  // avisamos al salir de la SC
	chrep     chan bool  // chan replay (para la respuesta) y saber si tenemos permiso para acceder a la SC
	Mutex     sync.Mutex // mutex para proteger concurrencia sobre las variables
	// TODO: completar
	Me      int // Mi número de nodo
	N       int // numero de nodoso en la red
	op_type int
	logger  *govec.GoLog
}

// ReturnVCString returns a string encoding of a vector clock
func (vc RelojVector) RelojString() string {
	mapa := vc.VReloj.GetMap()
	var cadena string
	cadena = cadena + "{"
	for i := 1; i <= 3; i++ {
		//buffer.WriteString(fmt.Sprintf("\"%s\":%d", ids[i], vc.VReloj[ids[i]]))
		//cadena = cadena + strconv.Itoa(i) + " - " + strconv.Itoa(int(vc.VReloj[strconv.Itoa(i)])) + ","
		cadena = cadena + strconv.Itoa(i) + " - " + strconv.Itoa(int(mapa[strconv.Itoa(i)])) + ","

	}
	cadena = cadena + "}"
	println(cadena)
	return cadena
}

func New(me int, N int, op_type int, usersFile string) *RASharedDB {
	logger := govec.InitGoVector(strconv.Itoa(me), "LogFile", govec.GetDefaultConfig())
	messageTypes := []ms.Message{Request{}, Reply{}, RelojVector{}}
	msgs := ms.New(me, usersFile, messageTypes)
	ra := RASharedDB{0, 0, 0, false, []int{}, &msgs, make(chan bool), make(chan bool), sync.Mutex{}, me, N, op_type, logger}
	for j := 0; j <= N; j++ {
		ra.RepDefd = append(ra.RepDefd, 0)
	}
	return &ra
}

//Pre: Verdad
//Post: Realiza  el  PreProtocol  para el  algoritmo de
//      Ricart-Agrawala Generalizado
//El preprotocol para un Proceso Pi consiste en enviar a los N - 1
//procesos distribuidos una peticion de acceso a la seccion critica.
//Cuando se reciba la respuesta en (5) de los N-1 entonces el Proceso
//Pi accede a la seccion critica.

func (ra *RASharedDB) PreProtocol() {
	ra.Mutex.Lock()
	ra.ReqCS = true                 // indicamos que queremos acceder a la SC
	ra.OurSeqNum = ra.HigSeqNum + 1 // aumentamos el reloj interno
	ra.Mutex.Unlock()

	ra.OutRepCnt = ra.N - 1 // Numero de procesos que confirman la entrada a SC
	for j := 1; j <= ra.N; j++ {
		if j != ra.Me {
			// Enviamos el reloj vectorial
			ra.ms.Send(j, Request{ra.OurSeqNum, ra.Me, ra.op_type}) // send (enviamos la peticion de acceso a la SC)
			messagePayload := []byte("Request")                     // Codificamos el mensaje y actualizamos el govec
			ra.logger.PrepareSend("Envio peticion acceso SC: "+strconv.Itoa(ra.Me), messagePayload, govec.GetDefaultLogOptions())
			time.Sleep(time.Duration(50) * time.Millisecond)
			ra.ms.Send(j, RelojVector{ra.logger.GetCurrentVC()})
			vc := RelojVector{ra.logger.GetCurrentVC()}
			fmt.Println("RelojVector enviado " + vc.RelojString() + " de " + strconv.Itoa(ra.Me) + " para " + strconv.Itoa(j))

		}
	}

	for ra.OutRepCnt != 0 {
		<-ra.chrep                      // esperamos respuesta (leer del canal)
		ra.OutRepCnt = ra.OutRepCnt - 1 // decrementas contador de numero de procesos que faltan
	}
	ra.ReqCS = false // puede que no vaya aqui :/ (igual al inicio del post)-------------------------------------------------------------------
}

// Pre: Verdad
// Post: Realiza  el  PostProtocol  para el  algoritmo de
//
//	Ricart-Agrawala Generalizado
//
// Enviamos permiso a los procesos que teniamos esperando la confirmacion
// una vez salimos de nuestra SC
func (ra *RASharedDB) PostProtocol() {
	for j := 1; j <= ra.N; j++ {
		if ra.RepDefd[j] == 1 {
			messagePayload := []byte("Reply") // Codificamos el mensaje y actualizamos el govec
			ra.logger.PrepareSend("Envio Reply", messagePayload, govec.GetDefaultLogOptions())
			time.Sleep(time.Duration(50) * time.Millisecond)
			ra.ms.Send(j, RelojVector{ra.logger.GetCurrentVC()}) // Enviamos el reloj vectorial
			vc := RelojVector{ra.logger.GetCurrentVC()}
			fmt.Println("RelojVector enviado " + vc.RelojString() + " de " + strconv.Itoa(ra.Me) + " para " + strconv.Itoa(j))

			ra.RepDefd[j] = 0
			ra.ms.Send(j, Reply{}) // send nuestro permiso para que j entre en SC
		}
	}
}

func (ra *RASharedDB) Stop() {
	ra.ms.Stop()
	ra.done <- true
}

func (ra *RASharedDB) Recibir() {
	var defer_it bool
	fmt.Println("Inicio de la goRutinas de: " + strconv.Itoa(ra.Me))

	for {
		mensaje := ra.ms.Receive()      // Obtenemos un mensaje del mailbox
		switch tipo := mensaje.(type) { // variable que define los casos
		case RelojVector:
			fmt.Println(strconv.Itoa(ra.Me) + "- Recibido RelojVector")
			messagePayload := []byte("GoVect")
			tipo.UnpackReceive("Recibir un reloj vectorial", messagePayload, &messagePayload, govec.GetDefaultLogOptions()) // Decodifica el mensaje y actualiza el reloj local con el recibido
			ra.logger.LogLocalEvent("Recepcion completada GoVec", govec.GetDefaultLogOptions())                             // Log a local event
			vc := RelojVector{ra.logger.GetCurrentVC()}
			fmt.Println("RelojVector recibido " + vc.RelojString() + " soy " + strconv.Itoa(ra.Me))

		case Request: // nos llega un tipo request con su id, y su clock
			fmt.Println(strconv.Itoa(ra.Me) + "- Recibido Request")
			fmt.Println("Request: " + strconv.Itoa(ra.Me) + " from " + strconv.Itoa(tipo.Pid))
			if ra.HigSeqNum < tipo.Clock {
				ra.HigSeqNum = tipo.Clock
			}
			ra.Mutex.Lock()

			exclude := false
			if ra.op_type == 1 {
				exclude = true
			} else if tipo.op_t == 1 {
				exclude = true
			}

			defer_it = ra.ReqCS &&
				((tipo.Clock > ra.OurSeqNum) || (tipo.Clock == ra.OurSeqNum && tipo.Pid > ra.Me)) &&
				(exclude) //exclude(op_type,op_t)  1 -> escritores 0 -> lectores

			ra.Mutex.Unlock()
			if defer_it {
				ra.RepDefd[tipo.Pid] = 1 // Entrariamos nosotros en SC por ello j se queda esperando
			} else {
				messagePayload := []byte("Reply") // Codificamos el mensaje y actualizamos el govec
				ra.logger.PrepareSend("Envio Reply", messagePayload, govec.GetDefaultLogOptions())
				time.Sleep(time.Duration(50) * time.Millisecond)
				ra.ms.Send(tipo.Pid, RelojVector{ra.logger.GetCurrentVC()}) // Enviamos el reloj vectorial

				ra.ms.Send(tipo.Pid, Reply{}) // Enviamos nuestro permiso al proceso j
				fmt.Println("Envio de permiso de " + strconv.Itoa(ra.Me) + " a " + strconv.Itoa(tipo.Pid) + " realizado")

			}

		default:
			fmt.Println(strconv.Itoa(ra.Me) + "- Recibido Reply")
			ra.chrep <- true // recibimos el permiso de un proceso j, y por ello enviamos
			// true por el canal chrep, el cual estaremos esperando en el
			// preprotocol

		}
	}
}
