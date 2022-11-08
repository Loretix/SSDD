Campos existentes:

    // constantes: 
    - IntNOINICIALIZADO         // Constante para fijar valor entero no inicializado
    - kEnableDebugLogs          // Aseguraros de poner kEnableDebugLogs a false antes de la entrega
    - kLogToStdout              // true: stdout por terminal | false: stdout a fichero 
    - kLogOutputDir             // directorio al q se redirige el log
    - VotosRecibidos = 0        // para contar los votos que se reciben 
    //NOTA: se pone a cero al pasar a ser seguidor, y al pasar a ser lider

    // Estados 
	- LIDER     = "lider"
	- CANDIDATO = "candidato"
	- SEGUIDOR  = "seguidor"

    // structs:
    - TipoOperacion
        Operacion string        // La operaciones posibles son "leer" y "escribir"
        Clave     string        // indice del mandato????
        Valor     string        // en el caso de la lectura Valor = ""
    
    // A medida que el nodo Raft conoce las operaciones de las  entradas de registro
    // comprometidas, envía un AplicaOperacion, con cada una de ellas, al canal
    // "canalAplicar" (funcion NuevoNodo) de la maquina de estados
    - AplicaOperacion
        Indice    int           // en la entrada de registro
        Operacion TipoOperacion

    // struct que representa la información de cada entrada del
    // registro de operaciones (logs)
    - RegistroOp 
        Comando string // en la entrada de registro
        Mandato int    // mandato

    - NodoRaft
        Mux sync.Mutex
        // Host:Port de todos los nodos (réplicas) Raft, en mismo orden
        Nodos   []rpctimeout.HostPort 
        Yo      int             // indice de este nodos en campo array "nodos"
        IdLider int 
        Logger *log.Logger	    // Utilización opcional de este logger para depuración
        E Estado                // se define debajo
        Roll string             //  Roll (SEGUIDOR, LIDER, CANDIDATO)


    - Estado 
        CurrentTerm int         // mandato actual 
        VotedFor int            // ultimo candidato votado, -1 en caso de que no este asignado
        Log []RegistroOp        // el registro interno de acda nodo        
        CommitIndex int         // indice de la última petición comprometida
        LastApplied int         // índice de la última petición guardada (no tiene por que estar comprometida)
        // respecto a las peticiones de los clientes
        NextIndex []int         //indice de la sig entrada de reg para enviar
        MatchIndex []int        //indice de la entrada de reg mas alta

    - ArgsPeticionVoto
        Term         int         // mandato del candidato
        CandidateId  int         // id del cantidado que solicita la votacion
        LastLogIndex int         // ultimo valor del indice del log del candidato
        LastLogTerm  int         // mandato en el que se registro el ultimo indice del log

    - RespuestaPeticionVoto
    	Term        int          // currentTerm, for candidate to update itself
    	VoteGranted bool         // TRUE si recivimos el voto, false en caso contrario

    - ArgAppendEntries
    	Term         int          // Mandato del lider
        LeaderId     int          // Id del lider, para que los sefuidores puedan redirigir al cliente en caso de que les haga una solicitud
        PrevLogIndex int          // índice del último log realizado
        PrevLogTerm  int          // mandato del lider en el que se realizo el último registro en log
        Entries      []RegistroOp // log del lider para los seguidores (empty for heartbeat; may send more than one for efficiency)
        LeaderCommit int          // indice de la ultima entrada comprometida del lider

    - Results
        Term    int  // Mandato del seguidor para que se actualice el lider
        Success bool // true si se ha guardado el log que se recivio / true if follower contained entry matching prevLogIndex and prevLogTerm

FUNCIONES:
    inicializarEstado(nr *NodoRaft)
            // parámetros                                                                       // Devuelve     
    NuevoNodo(nodos []rpctimeout.HostPort, yo int,	canalAplicarOperacion chan AplicaOperacion) *NodoRaft 
            // llama a inicializarEstado
    //parámetros
    (nr *NodoRaft) para() // Metodo Para() utilizado cuando no se necesita mas al nodo
    //parámetro                     // devuelve
    (nr *NodoRaft) obtenerEstado() (int, int, bool, int)

    // Si somos el lider registramos la petición del cliente
    //parámetro                     // parámetro              // return 
    (nr *NodoRaft) someterOperacion(operacion TipoOperacion) (int, int, bool, int, string)

    // Devuleve el mensaje al candidato que ha solicitado el voto
        // parámetro            // parámetros                                               // return
    func (nr *NodoRaft) PedirVoto(peticion *ArgsPeticionVoto, reply *RespuestaPeticionVoto) error

    // recibes el latido del lider y respondes
    // parámetros               // parámetro                               // return
    (nr *NodoRaft) AppendEntries(args *ArgAppendEntries, results *Results) error

    // Enviamos la petición del voto al resto de los candidatos y nos votamos 
    // parámetros               // parámetro                                                          // return
    (nr *NodoRaft) enviarPeticionVoto(nodo int, args *ArgsPeticionVoto,reply *RespuestaPeticionVoto) bool

    // envia  un mensaje a un nodo concreto y gestiona el timeout
    // además recoge la respuesta del nodo 
    (hp HostPort) CallTimeout(serviceMethod string, args interface{},
							reply interface{}, timeout time.Duration) error 

    // El lider envia el latido a los demas nodos
    func (nr *NodoRaft) enviarAppendEntries(nodo int, args *ArgAppendEntries, reply *Results) bool 