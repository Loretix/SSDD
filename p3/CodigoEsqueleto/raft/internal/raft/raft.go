// Escribir vuestro código de funcionalidad Raft en este fichero
//

package raft

//
// API
// ===
// Este es el API que vuestra implementación debe exportar
//
// nodoRaft = NuevoNodo(...)
//   Crear un nuevo servidor del grupo de elección.
//
// nodoRaft.Para()
//   Solicitar la parado de un servidor
//
// nodo.ObtenerEstado() (yo, mandato, esLider)
//   Solicitar a un nodo de elección por "yo", su mandato en curso,
//   y si piensa que es el msmo el lider
//
// nodoRaft.SometerOperacion(operacion interface()) (indice, mandato, esLider)

// type AplicaOperacion

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"

	//"crypto/rand"
	"sync"
	"time"

	//"net/rpc"

	"raft/internal/comun/rpctimeout"
)

const (
	// Constante para fijar valor entero no inicializado
	IntNOINICIALIZADO = -1

	//  false deshabilita por completo los logs de depuracion
	// Aseguraros de poner kEnableDebugLogs a false antes de la entrega
	kEnableDebugLogs = true

	// Poner a true para logear a stdout en lugar de a fichero
	kLogToStdout = false

	// Cambiar esto para salida de logs en un directorio diferente
	kLogOutputDir = "./logs_raft/"

	LIDER     = "lider"
	CANDIDATO = "candidato"
	SEGUIDOR  = "seguidor"
)

type TipoOperacion struct {
	Operacion string // La operaciones posibles son "leer" y "escribir"
	Clave     int
	Valor     string // en el caso de la lectura Valor = ""
}

// A medida que el nodo Raft conoce las operaciones de las  entradas de registro
// comprometidas, envía un AplicaOperacion, con cada una de ellas, al canal
// "canalAplicar" (funcion NuevoNodo) de la maquina de estados
type AplicaOperacion struct {
	Indice    int // en la entrada de registro
	Operacion TipoOperacion
}

// struct que representa la información de cada entrada del
// registro de operaciones (logs)
type RegistroOp struct {
	Comando string // en la entrada de registro
	Mandato int    // mandato
}

type Estado struct {
	CurrentTerm int // ultimo mandato que ha visto el servidor
	VotedFor    int // candidato al que se ha votado
	Log         []RegistroOp
	CommitIndex int // indice de la ultima entrada comprometida
	LastApplied int
	// respecto a las peticiones de los clientes
	NextIndex  []int //indice de la sig entrada de reg para enviar
	MatchIndex []int //indice de la entrada de reg mas alta
}

// Tipo de dato Go que representa un solo nodo (réplica) de raft
type NodoRaft struct {
	Mux sync.Mutex // Mutex para proteger acceso a estado compartido

	// Host:Port de todos los nodos (réplicas) Raft, en mismo orden
	Nodos   []rpctimeout.HostPort
	Yo      int // indice de este nodos en campo array "nodos"
	IdLider int
	// Utilización opcional de este logger para depuración
	// Cada nodo Raft tiene su propio registro de trazas (logs)
	Logger *log.Logger

	// Estado.
	E Estado

	// Roll (SEGUIDOR, LIDER, CANDIDATO)
	Roll string

	// contador de votos en caso de ser candidato
	VotosRecibidos int

	// cuenta el numero de nodos que guardan la entrada
	NodosLogCorrecto int

	TimerEleccion *time.Timer
	TimerLatido   *time.Timer
}

func inicializarEstado(nr *NodoRaft) {
	nr.E.CurrentTerm = 0
	nr.E.VotedFor = -1
	nr.E.CommitIndex = 0
	nr.E.LastApplied = 0
	nr.E.MatchIndex = append(nr.E.MatchIndex, 0)
	nr.Roll = SEGUIDOR
}

// Creacion de un nuevo nodo de eleccion
//
// Tabla de <Direccion IP:puerto> de cada nodo incluido a si mismo.
//
// <Direccion IP:puerto> de este nodo esta en nodos[yo]
//
// Todos los arrays nodos[] de los nodos tienen el mismo orden

// canalAplicar es un canal donde, en la practica 5, se recogerán las
// operaciones a aplicar a la máquina de estados. Se puede asumir que
// este canal se consumira de forma continúa.
//
// NuevoNodo() debe devolver resultado rápido, por lo que se deberían
// poner en marcha Gorutinas para trabajos de larga duracion
func NuevoNodo(nodos []rpctimeout.HostPort, yo int,
	canalAplicarOperacion chan AplicaOperacion) *NodoRaft {
	nr := &NodoRaft{}
	nr.Nodos = nodos
	nr.Yo = yo
	nr.IdLider = -1

	if kEnableDebugLogs {
		nombreNodo := nodos[yo].Host() + "_" + nodos[yo].Port()
		logPrefix := fmt.Sprintf("%s", nombreNodo)

		fmt.Println("LogPrefix: ", logPrefix)

		if kLogToStdout {
			nr.Logger = log.New(os.Stdout, nombreNodo+" -->> ",
				log.Lmicroseconds|log.Lshortfile)
		} else {
			err := os.MkdirAll(kLogOutputDir, os.ModePerm)
			if err != nil {
				panic(err.Error())
			}
			logOutputFile, err := os.OpenFile(fmt.Sprintf("%s/%s.txt",
				kLogOutputDir, logPrefix), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
			if err != nil {
				panic(err.Error())
			}
			nr.Logger = log.New(logOutputFile,
				logPrefix+" -> ", log.Lmicroseconds|log.Lshortfile)
		}
		nr.Logger.Println("logger initialized")
	} else {
		nr.Logger = log.New(ioutil.Discard, "", 0)
	}

	// Realizada
	inicializarEstado(nr)
	nr.VotosRecibidos = 0
	nr.NodosLogCorrecto = 0
	// enviara el primer latido en un nº aleatorio entre 50 y 200 ms
	nr.TimerEleccion = time.NewTimer(time.Duration(rand.Intn(50)+150) * time.Millisecond)
	nr.TimerLatido = time.NewTimer(50 * time.Millisecond)
	return nr
}

// Metodo Para() utilizado cuando no se necesita mas al nodo
//
// Quizas interesante desactivar la salida de depuracion
// de este nodo
func (nr *NodoRaft) para() {
	go func() { time.Sleep(5 * time.Millisecond); os.Exit(0) }()
}

// Devuelve "yo", mandato en curso y si este nodo cree ser lider
//
// Primer valor devuelto es el indice de este  nodo Raft el el conjunto de nodos
// la operacion si consigue comprometerse.
// El segundo valor es el mandato en curso
// El tercer valor es true si el nodo cree ser el lider
// Cuarto valor es el lider, es el indice del líder si no es él
func (nr *NodoRaft) obtenerEstado() (int, int, bool, int) {
	var yo int = nr.Yo
	var mandato int = nr.E.CurrentTerm
	var esLider bool = (nr.Yo == nr.IdLider)
	var idLider int = nr.IdLider
	// Realizada
	return yo, mandato, esLider, idLider
}

// El servicio que utilice Raft (base de datos clave/valor, por ejemplo)
// Quiere buscar un acuerdo de posicion en registro para siguiente operacion
// solicitada por cliente.

// Si el nodo no es el lider, devolver falso
// Sino, comenzar la operacion de consenso sobre la operacion y devolver en
// cuanto se consiga
//
// No hay garantia que esta operacion consiga comprometerse en una entrada de
// de registro, dado que el lider puede fallar y la entrada ser reemplazada
// en el futuro.
// Primer valor devuelto es el indice del registro donde se va a colocar
// la operacion si consigue comprometerse.
// El segundo valor es el mandato en curso
// El tercer valor es true si el nodo cree ser el lider
// Cuarto valor es el lider, es el indice del líder si no es él
func (nr *NodoRaft) someterOperacion(operacion TipoOperacion) (int, int,
	bool, int, string) {
	indice := nr.E.LastApplied
	mandato := nr.E.CurrentTerm
	EsLider := (nr.Yo == nr.IdLider)
	idLider := nr.IdLider
	valorADevolver := ""

	// Realizada
	if EsLider {
		nr.E.Log = append(nr.E.Log, RegistroOp{Comando: operacion.Operacion, Mandato: operacion.Clave})
		valorADevolver = operacion.Valor // Devolvemos valor leido en caso de q exista
		indice = indice + 1
		nr.E.LastApplied = indice // actualizamos el índice, pq hemos añadido una entrada
	}

	return indice, mandato, EsLider, idLider, valorADevolver
}

// -----------------------------------------------------------------------
// LLAMADAS RPC al API
//
// Si no tenemos argumentos o respuesta estructura vacia (tamaño cero)
type Vacio struct{}

func (nr *NodoRaft) ParaNodo(args Vacio, reply *Vacio) error {
	defer nr.para()
	return nil
}

type EstadoParcial struct {
	Mandato int
	EsLider bool
	IdLider int
}

type EstadoRemoto struct {
	IdNodo int
	EstadoParcial
}

func (nr *NodoRaft) ObtenerEstadoNodo(args Vacio, reply *EstadoRemoto) error {
	reply.IdNodo, reply.Mandato, reply.EsLider, reply.IdLider = nr.obtenerEstado()
	return nil
}

type ResultadoRemoto struct {
	ValorADevolver string
	IndiceRegistro int
	EstadoParcial
}

func (nr *NodoRaft) SometerOperacionRaft(operacion TipoOperacion,
	reply *ResultadoRemoto) error {
	reply.IndiceRegistro, reply.Mandato, reply.EsLider,
		reply.IdLider, reply.ValorADevolver = nr.someterOperacion(operacion)
	return nil
}

// ------------------------------- funciones pedirVoto -------------------------------------------------//

// -----------------------------------------------------------------------
// LLAMADAS RPC protocolo RAFT
//
// Structura de ejemplo de argumentos de RPC PedirVoto.
//
// Recordar
// -----------
// Nombres de campos deben comenzar con letra mayuscula !
// REQUEST VOTE, mensaje que envian los candidatos para solicitar voto y convertirse en lideres
type ArgsPeticionVoto struct {
	// Realizada
	Term         int // mandato del candidato
	CandidateId  int // id del cantidado que solicita la votacion
	LastLogIndex int // ultimo valor del indice del log del candidato
	LastLogTerm  int // mandato en el que se registro el ultimo indice del log
}

// Structura de ejemplo de respuesta de RPC PedirVoto,
//
// Recordar
// -----------
// Nombres de campos deben comenzar con letra mayuscula !
type RespuestaPeticionVoto struct {
	// Realizada
	Term        int  // currentTerm, for candidate to update itself
	VoteGranted bool // TRUE si recivimos el voto, false en caso contrario
}

// Metodo para RPC PedirVoto - metodo mediante en cual un receptor
// de ArgsPeticionVoto envia su respuesta mediante el tipo RespuestaPeticionVoto
func (nr *NodoRaft) PedirVoto(peticion *ArgsPeticionVoto,
	reply *RespuestaPeticionVoto) error {
	// Realizada

	reply.VoteGranted = false
	reply.Term = nr.E.CurrentTerm
	if peticion.Term >= nr.E.CurrentTerm {
		nr.E.CurrentTerm = nr.E.CurrentTerm + 1
		var doyVoto bool = (nr.E.VotedFor == -1 || nr.E.VotedFor == peticion.CandidateId) && nr.E.CommitIndex <= peticion.LastLogIndex
		if doyVoto {
			nr.E.VotedFor = peticion.CandidateId // Empezamos a votar al nuevo candidato
			reply.VoteGranted = true
			reply.Term = nr.E.CurrentTerm
			//nos convertimos en seguidor
		}
	}

	return nil
}

// Ejemplo de código enviarPeticionVoto
//
// nodo int -- indice del servidor destino en nr.nodos[]
//
// args *RequestVoteArgs -- argumentos par la llamada RPC
//
// reply *RequestVoteReply -- respuesta RPC
//
// Los tipos de argumentos y respuesta pasados a CallTimeout deben ser
// los mismos que los argumentos declarados en el metodo de tratamiento
// de la llamada (incluido si son punteros
//
// Si en la llamada RPC, la respuesta llega en un intervalo de tiempo,
// la funcion devuelve true, sino devuelve false
//
// la llamada RPC deberia tener un timout adecuado.
//
// Un resultado falso podria ser causado por una replica caida,
// un servidor vivo que no es alcanzable (por problemas de red ?),
// una petición perdida, o una respuesta perdida
//
// Para problemas con funcionamiento de RPC, comprobar que la primera letra
// del nombre  todo los campos de la estructura (y sus subestructuras)
// pasadas como parametros en las llamadas RPC es una mayuscula,
// Y que la estructura de recuperacion de resultado sea un puntero a estructura
// y no la estructura misma.

func (nr *NodoRaft) enviarPeticionVoto(nodo int, args *ArgsPeticionVoto,
	reply *RespuestaPeticionVoto) bool {
	// Realizada
	nr.Mux.Lock()
	args.Term = nr.E.CurrentTerm
	args.CandidateId = nr.Yo
	args.LastLogIndex = nr.E.LastApplied
	args.LastLogTerm = nr.E.Log[args.LastLogIndex].Mandato

	// enviamos al nodo que nos pasan por parámetro la petición y recibimos su respuesta
	if nr.Nodos[nodo].CallTimeout("Envio peticion voto", &args, &reply, time.Duration(25)*time.Millisecond) == nil {
		// si es == a nil es que no ha habido error y se ha recibido respuesta
		if reply.Term > nr.E.CurrentTerm {
			nr.ConvertirseEnSeguidor(reply.Term)
		} else if reply.VoteGranted == true {
			nr.VotosRecibidos++
			// Si tenemos la mayoria de votos nos convertimos en lider
			if nr.VotosRecibidos > (len(nr.Nodos) / 2) {
				nr.ConvertirseEnLider()
			}
		}
		return true
	}
	nr.Mux.Unlock()

	return false
}

// ------------------------------- funciones appendEntries -------------------------------------------------//

// Lo envia el lider a los seguidores, puede ser solo un latido, o la petición de un cliente
type ArgAppendEntries struct {
	// Realizada
	Term         int          // Mandato del lider
	LeaderId     int          // Id del lider, para que los sefuidores puedan redirigir al cliente en caso de que les haga una solicitud
	PrevLogIndex int          // índice del último log realizado
	PrevLogTerm  int          // mandato del lider en el que se realizo el último registro en log
	Entries      []RegistroOp // log del lider para los seguidores (empty for heartbeat; may send more than one for efficiency)
	LeaderCommit int          // indice de la ultima entrada comprometida del lider
}

type Results struct {
	Term    int  // Mandato del seguidor para que se actualice el lider
	Success bool // true si se ha guardado el log que se recivio / true if follower contained entry matching prevLogIndex and prevLogTerm
}

// Metodo de tratamiento de llamadas RPC AppendEntries = latido te llega
func (nr *NodoRaft) AppendEntries(args *ArgAppendEntries,
	results *Results) error {
	// Realizada
	results.Success = true
	results.Term = nr.E.CurrentTerm
	// 1. Responda falso si termino < termino actual
	if args.Term < nr.E.CurrentTerm {
		results.Success = false
		results.Term = nr.E.CurrentTerm
	} else {
		// pasos dos y tres se realizarán en la práctica 5
		nr.IdLider = args.LeaderId
		// 4. Annadir entradas nuevas que aun no esten en el registro
		var i int
		for i = nr.E.LastApplied + 1; i < len(args.Entries); i++ {
			nr.E.Log = append(nr.E.Log, args.Entries[i])
			nr.E.LastApplied++
		}
		nr.E.LastApplied = i
		// 5. Si leaderCommit > commitIndex, commitIndex = min(leaderCommit, indice de la ultima entrada nueva)
		if args.LeaderCommit > nr.E.CommitIndex {
			if args.LeaderCommit < nr.E.LastApplied {
				nr.E.CommitIndex = args.LeaderCommit
			} else {
				nr.E.CommitIndex = nr.E.LastApplied
			}
		}
	}

	return nil
}

// El lider envia el AppendEntries = latido a los demas nodos
func (nr *NodoRaft) enviarAppendEntries(nodo int, args *ArgAppendEntries,
	reply *Results) bool {
	// Realizada
	nr.Mux.Lock()
	args.Term = nr.E.CurrentTerm
	args.LeaderId = nr.Yo
	args.LeaderCommit = nr.E.CommitIndex
	args.PrevLogIndex = nr.E.NextIndex[nodo] - 1
	args.PrevLogTerm = nr.E.Log[args.PrevLogIndex-1].Mandato

	// Creamos el vector de nuevas entradas a enviar (Entries) para ello hacemos
	// un bucle for que lee el vector de registro (Log) desde el indice (next)
	// ya que es en este en el que se guardaran las nuevas entradas a comprometer
	// La longitud de lo que se va a leer en Log es len(nr.E.Log)-nr.E.NextIndex[nodo],
	// ya que es la longitud total del Log del lider - el indice de las nuevas entradas
	// que se podran llegar a comprometer. Ejemplo :
	//             0   1   2  3  4
	// 		lider [3] [2] [] [] []		nextIndex=1; prevIndex=0; prevTerm=3; commitTerm=3
	//		s1 	  [3] []
	// 	len(nr.E.Log)-nr.E.NextIndex[nodo] => len(nr.E.Log) = 2
	//										nr.E.NextIndex[nodo] = 1
	//										2-1 = 1, Longitud de entries = 1, una entrada nueva

	for i := nr.E.NextIndex[nodo]; i < len(nr.E.Log)-nr.E.NextIndex[nodo]; i++ {
		args.Entries = append(args.Entries, nr.E.Log[i])
	}

	// enviamos al nodo que nos pasan por parámetro la petición y recibimos su respuesta
	if nr.Nodos[nodo].CallTimeout("Envio peticion voto", &args, &reply, time.Duration(25)*time.Millisecond) == nil {
		// si es == a nil es que no ha habido error y se ha recibido respuesta
		if reply.Term > nr.E.CurrentTerm {
			// pasamos a ser seguidor ;)
		} else if !reply.Success {
			// es false por tanto no se ha comprometido la entrada se trata el caso
		} else { // La entrada se ha registrado el en Log correctamente
			nr.NodosLogCorrecto++
			if nr.NodosLogCorrecto > len(nr.Nodos)/2 {
				// comprometemos la entrada ya que tenemos mayoria
				nr.NodosLogCorrecto = 0
				nr.E.CommitIndex = args.PrevLogIndex + 1
				nr.E.NextIndex[nodo]++
				nr.E.MatchIndex[nodo]++
			}
		}
		return true
	}
	nr.Mux.Unlock()

	return false
}

// ------------------------------- funciones cambio Roll --------------------------------------------//

func (nr *NodoRaft) ConvertirseEnSeguidor(mandato int) {
	nr.Roll = SEGUIDOR
	nr.E.CurrentTerm = mandato
	nr.E.VotedFor = -1
	// tiempo aleatorio entre 50 y 200 milisegundos
	tiempo := time.Duration(rand.Intn(50)+150) * time.Millisecond
	nr.TimerEleccion.Reset(tiempo)
	nr.TimerLatido.Stop()
}

func (nr *NodoRaft) ConvertirseEnLider() {
	nr.Roll = LIDER
	nr.IdLider = nr.Yo
	// Inicializar nextIndex
	for i := 0; i < len(nr.E.NextIndex); i++ {
		nr.E.NextIndex = append(nr.E.NextIndex, nr.E.LastApplied+1)
	}
	// Inicializar matchIndex
	for i := 0; i < len(nr.Nodos); i++ {
		nr.E.MatchIndex = append(nr.E.MatchIndex, 0)
	}
	// Preparamos los parámetros para enviar los latidos
	var reply Results
	var args ArgAppendEntries

	args.Term = nr.E.CurrentTerm
	args.LeaderId = nr.IdLider
	args.PrevLogIndex = nr.E.LastApplied
	args.PrevLogTerm = nr.E.Log[args.PrevLogIndex].Mandato
	args.Entries = nil
	args.LeaderCommit = nr.E.CommitIndex

	// enviar RPC de AppendEntries vacios iniciales
	for i := 0; i < len(nr.Nodos); i++ {
		if i != nr.Yo {
			go nr.enviarAppendEntries(i, &args, &reply)
		}
	}
	// gestion de timers
	nr.TimerEleccion.Stop()
	nr.TimerLatido.Reset(50)

}

func (nr *NodoRaft) ConvertirseEnCandidato() {
	nr.Mux.Lock()
	nr.Roll = CANDIDATO
	nr.E.CurrentTerm++ // votamos popr nosotros mismos
	nr.NodosLogCorrecto++
	nr.Mux.Unlock()
	// preparamos los parámetros para la solicitud del voto
	var reply RespuestaPeticionVoto
	var args ArgsPeticionVoto
	args.Term = nr.E.CurrentTerm
	args.CandidateId = nr.Yo
	args.LastLogIndex = nr.E.CommitIndex
	args.LastLogTerm = nr.E.Log[args.LastLogIndex].Mandato
	// solicitamos el voto
	for i := 0; i < len(nr.Nodos); i++ {
		go nr.enviarPeticionVoto(i, &args, &reply)
		// en la práctica 5 se tendra que comprobar el bool de respuesta para saber si ha llegado sin problemas
	}
	nr.Mux.Lock()
	// gestion de los timers para comenzar una nueva eleccion si es necesario
	tiempo := time.Duration(rand.Intn(50)+150) * time.Millisecond
	nr.TimerEleccion.Reset(tiempo)
	nr.Mux.Unlock()
}
