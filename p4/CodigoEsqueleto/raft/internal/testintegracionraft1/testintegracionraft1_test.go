package testintegracionraft1

import (
	"fmt"

	//"log"
	//"crypto/rand"
	//"os"
	//"path/filepath"
	"strconv"
	"testing"
	"time"

	"raft/internal/comun/check"
	"raft/internal/comun/rpctimeout"
	"raft/internal/despliegue"
	"raft/internal/raft"
)

const (
	//hosts
	/*MAQUINA1 = "127.0.0.1"
	MAQUINA2 = "127.0.0.1"
	MAQUINA3 = "127.0.0.1"*/
	MAQUINA_LOCAL = "127.0.0.1"
	MAQUINA1      = "155.210.154.206"
	MAQUINA2      = "155.210.154.191"
	MAQUINA3      = "155.210.154.209"

	//puertos
	PUERTOREPLICA1 = "29008"
	PUERTOREPLICA2 = "29005"
	PUERTOREPLICA3 = "29006"

	//nodos replicas
	REPLICA1 = MAQUINA1 + ":" + PUERTOREPLICA1
	REPLICA2 = MAQUINA2 + ":" + PUERTOREPLICA2
	REPLICA3 = MAQUINA3 + ":" + PUERTOREPLICA3

	// paquete main de ejecutables relativos a PATH previo
	EXECREPLICA = "cmd/srvraft/main.go"

	// comandos completo a ejecutar en máquinas remota con ssh. Ejemplo :
	// 				cd $HOME/raft; go run cmd/srvraft/main.go 127.0.0.1:29001

	// Ubicar, en esta constante, nombre de fichero de vuestra clave privada local
	// emparejada con la clave pública en authorized_keys de máquinas remotas

	PRIVKEYFILE = "id_rsa"
)

// PATH de los ejecutables de modulo golang de servicio Raft
var PATH string = "/home/a799301/practicas/SSDD/p4/CodigoEsqueleto/raft"

// go run cmd/srvraft/main.go 0 127.0.0.1:29001 127.0.0.1:29002 127.0.0.1:29003
var EXECREPLICACMD string = "cd " + PATH + "; /usr/local/go/bin/go run " + EXECREPLICA

// TEST primer rango
func TestPrimerasPruebas(t *testing.T) { // (m *testing.M) {
	// <setup code>
	// Crear canal de resultados de ejecuciones ssh en maquinas remotas
	cfg := makeCfgDespliegue(t,
		3,
		[]string{REPLICA1, REPLICA2, REPLICA3},
		[]bool{true, true, true})

	// tear down code
	// eliminar procesos en máquinas remotas
	defer cfg.stop()

	// Run test sequence

	// Test1 : No debería haber ningun primario, si SV no ha recibido aún latidos
	t.Run("T1:soloArranqueYparada",
		func(t *testing.T) { cfg.soloArranqueYparadaTest1(t) })

	// Test2 : No debería haber ningun primario, si SV no ha recibido aún latidos
	t.Run("T2:ElegirPrimerLider",
		func(t *testing.T) { cfg.elegirPrimerLiderTest2(t) })

	// Test3: tenemos el primer primario correcto
	t.Run("T3:FalloAnteriorElegirNuevoLider",
		func(t *testing.T) { cfg.falloAnteriorElegirNuevoLiderTest3(t) })

	// Test4: Tres operaciones comprometidas en configuración estable
	t.Run("T4:tresOperacionesComprometidasEstable",
		func(t *testing.T) { cfg.tresOperacionesComprometidasEstable(t) })

}

// TEST primer rango
func TestAcuerdosConFallos(t *testing.T) { // (m *testing.M) {
	// <setup code>
	// Crear canal de resultados de ejecuciones ssh en maquinas remotas
	cfg := makeCfgDespliegue(t,
		3,
		[]string{REPLICA1, REPLICA2, REPLICA3},
		[]bool{true, true, true})

	// tear down code
	// eliminar procesos en máquinas remotas
	defer cfg.stop()

	// Test5: Se consigue acuerdo a pesar de desconexiones de seguidor
	t.Run("T5:AcuerdoAPesarDeDesconexionesDeSeguidor ",
		func(t *testing.T) { cfg.AcuerdoApesarDeSeguidor(t) })

	t.Run("T5:SinAcuerdoPorFallos ",
		func(t *testing.T) { cfg.SinAcuerdoPorFallos(t) })

	t.Run("T5:SometerConcurrentementeOperaciones ",
		func(t *testing.T) { cfg.SometerConcurrentementeOperaciones(t) })

}

// ---------------------------------------------------------------------
//
// Canal de resultados de ejecución de comandos ssh remotos
type canalResultados chan string

func (cr canalResultados) stop() {
	close(cr)

	// Leer las salidas obtenidos de los comandos ssh ejecutados
	for s := range cr {
		fmt.Println(s)
	}
}

// ---------------------------------------------------------------------
// Operativa en configuracion de despliegue y pruebas asociadas
type configDespliegue struct {
	t           *testing.T
	conectados  []bool
	numReplicas int
	nodosRaft   []rpctimeout.HostPort
	cr          canalResultados
}

// Crear una configuracion de despliegue
func makeCfgDespliegue(t *testing.T, n int, nodosraft []string,
	conectados []bool) *configDespliegue {
	cfg := &configDespliegue{}
	cfg.t = t
	cfg.conectados = conectados
	cfg.numReplicas = n
	cfg.nodosRaft = rpctimeout.StringArrayToHostPortArray(nodosraft)
	cfg.cr = make(canalResultados, 2000)

	return cfg
}

func (cfg *configDespliegue) stop() {
	cfg.stopDistributedProcesses()

	time.Sleep(50 * time.Millisecond)

	cfg.cr.stop()
}

// --------------------------------------------------------------------------
// FUNCIONES DE SUBTESTS

// Se pone en marcha una replica ?? - 3 NODOS RAFT
func (cfg *configDespliegue) soloArranqueYparadaTest1(t *testing.T) {
	t.Skip("SKIPPED soloArranqueYparadaTest1")

	fmt.Println(t.Name(), ".....................")

	cfg.t = t // Actualizar la estructura de datos de tests para errores

	// Poner en marcha replicas en remoto con un tiempo de espera incluido
	cfg.startDistributedProcesses()

	// Comprobar estado replica 0
	cfg.comprobarEstadoRemoto(0, 0, false, -1)

	// Comprobar estado replica 1
	cfg.comprobarEstadoRemoto(1, 0, false, -1)

	// Comprobar estado replica 2
	cfg.comprobarEstadoRemoto(2, 0, false, -1)

	// Parar réplicas almacenamiento en remoto
	cfg.stopDistributedProcesses()

	fmt.Println(".............", t.Name(), "Superado")
}

// Primer lider en marcha - 3 NODOS RAFT
func (cfg *configDespliegue) elegirPrimerLiderTest2(t *testing.T) {
	//t.Skip("SKIPPED ElegirPrimerLiderTest2")

	fmt.Println(t.Name(), ".....................")

	cfg.startDistributedProcesses()

	// Se ha elegido lider ?
	fmt.Printf("Probando lider en curso\n")
	cfg.pruebaUnLider(3)

	// Parar réplicas alamcenamiento en remoto
	cfg.stopDistributedProcesses() // Parametros

	fmt.Println(".............", t.Name(), "Superado")
}

// Fallo de un primer lider y reeleccion de uno nuevo - 3 NODOS RAFT
func (cfg *configDespliegue) falloAnteriorElegirNuevoLiderTest3(t *testing.T) {
	//t.Skip("SKIPPED FalloAnteriorElegirNuevoLiderTest3")

	fmt.Println(t.Name(), ".....................")

	cfg.startDistributedProcesses()

	fmt.Printf("Lider inicial\n")
	cfg.pruebaUnLider(3)

	// Desconectar lider
	cfg.desconectarLider()

	fmt.Printf("Comprobar nuevo lider\n")
	cfg.pruebaUnLider(3)

	// Parar réplicas almacenamiento en remoto
	cfg.stopDistributedProcesses() //parametros

	fmt.Println(".............", t.Name(), "Superado")
}

// 3 operaciones comprometidas con situacion estable y sin fallos - 3 NODOS RAFT
func (cfg *configDespliegue) tresOperacionesComprometidasEstable(t *testing.T) {
	//t.Skip("SKIPPED tresOperacionesComprometidasEstable")

	fmt.Println(t.Name(), ".....................")

	cfg.startDistributedProcesses()

	cfg.someterOperacion("escritura", 1)
	cfg.someterOperacion("escritura", 2)
	cfg.someterOperacion("escritura", 3)

	cfg.comprobarOperacionesComprometidas(3)

	// Parar réplicas almacenamiento en remoto
	cfg.stopDistributedProcesses() //parametros

	fmt.Println(".............", t.Name(), "Superado")
}

// Se consigue acuerdo a pesar de desconexiones de seguidor -- 3 NODOS RAFT
func (cfg *configDespliegue) AcuerdoApesarDeSeguidor(t *testing.T) {
	// t.Skip("SKIPPED AcuerdoApesarDeSeguidor")

	fmt.Println(t.Name(), ".....................")

	cfg.startDistributedProcesses()

	//  Obtener un lider y, a continuación desconectar una de los nodos Raft
	cfg.desconectarUnNodo()
	// Comprobar varios acuerdos con una réplica desconectada
	cfg.someterOperacion("escritura", 1)
	cfg.someterOperacion("escritura", 2)
	cfg.someterOperacion("escritura", 3)

	cfg.comprobarOperacionesComprometidas(3)
	fmt.Println("-", t.Name(), "Se han comprometido las entradas en los nodos vivos")

	// reconectar nodo Raft previamente desconectado y comprobar varios acuerdos
	cfg.conectarNodos()

	cfg.comprobarOperacionesComprometidas(3)
	fmt.Println("-", t.Name(), "Se han comprometido las entradas en todos los nodos")

	cfg.stopDistributedProcesses() //parametros

	fmt.Println(".............", t.Name(), "Superado")
}

// NO se consigue acuerdo al desconectarse mayoría de seguidores -- 3 NODOS RAFT
func (cfg *configDespliegue) SinAcuerdoPorFallos(t *testing.T) {
	// t.Skip("SKIPPED SinAcuerdoPorFallos")

	fmt.Println(t.Name(), ".....................")

	cfg.startDistributedProcesses()

	//  Obtener un lider y, a continuación desconectar 2 de los nodos Raft
	cfg.desconectarUnNodo()
	cfg.desconectarUnNodo()

	// Comprobar varios acuerdos con 2 réplicas desconectada
	cfg.someterOperacion("escritura", 1)
	cfg.someterOperacion("escritura", 2)
	cfg.someterOperacion("escritura", 3)

	// reconectar lo2 nodos Raft  desconectados y probar varios acuerdos
	cfg.comprobarOperacionesSinComprometer(3)

	// reconectar nodo Raft previamente desconectado y comprobar varios acuerdos
	cfg.conectarNodos()
	// comprobamos que las entradas se comprometen una vez reconectados los nodos
	cfg.comprobarOperacionesComprometidas(3)

	cfg.stopDistributedProcesses() //parametros

	fmt.Println(".............", t.Name(), "Superado")
}

// Se somete 5 operaciones de forma concurrente -- 3 NODOS RAFT
func (cfg *configDespliegue) SometerConcurrentementeOperaciones(t *testing.T) {
	// t.Skip("SKIPPED SometerConcurrentementeOperaciones")

	fmt.Println(t.Name(), ".....................")

	// un bucle para estabilizar la ejecucion

	// Obtener un lider y, a continuación someter una operacion
	// Someter 5  operaciones concurrentes

	cfg.startDistributedProcesses()

	cfg.someterOperacion("escritura", 1)
	cfg.someterOperacion("escritura", 2)
	cfg.someterOperacion("escritura", 3)
	cfg.someterOperacion("escritura", 4)
	cfg.someterOperacion("escritura", 5)

	// Comprobar estados de nodos Raft, sobre todo
	// el avance del mandato en curso e indice de registro de cada uno
	// que debe ser identico entre ellos

	cfg.comprobarOperacionesComprometidas(5)

	// Parar réplicas almacenamiento en remoto
	cfg.stopDistributedProcesses() //parametros

	fmt.Println(".............", t.Name(), "Superado")
}

// --------------------------------------------------------------------------
// FUNCIONES DE APOYO
// Comprobar que hay un solo lider
// probar varias veces si se necesitan reelecciones
func (cfg *configDespliegue) pruebaUnLider(numreplicas int) int {
	for iters := 0; iters < 10; iters++ {
		time.Sleep(500 * time.Millisecond)
		mapaLideres := make(map[int][]int)
		for i := 0; i < numreplicas; i++ {
			if cfg.conectados[i] {
				if _, mandato, eslider, _, _ := cfg.obtenerEstadoRemoto(i); eslider {
					mapaLideres[mandato] = append(mapaLideres[mandato], i)
				}
			}
		}

		ultimoMandatoConLider := -1
		for mandato, lideres := range mapaLideres {
			if len(lideres) > 1 {
				cfg.t.Fatalf("mandato %d tiene %d (>1) lideres",
					mandato, len(lideres))
			}
			if mandato > ultimoMandatoConLider {
				ultimoMandatoConLider = mandato
			}
		}

		if len(mapaLideres) != 0 {

			return mapaLideres[ultimoMandatoConLider][0] // Termina

		}
	}
	cfg.t.Fatalf("un lider esperado, ninguno obtenido")

	return -1 // Termina
}

func (cfg *configDespliegue) obtenerEstadoRemoto(
	indiceNodo int) (int, int, bool, int, error) {
	var reply raft.EstadoRemoto
	err := cfg.nodosRaft[indiceNodo].CallTimeout("NodoRaft.ObtenerEstadoNodo",
		raft.Vacio{}, &reply, 20*time.Millisecond)
	//fmt.Println(cfg.nodosRaft[indiceNodo])
	check.CheckError(err, "Error en llamada RPC ObtenerEstadoRemoto")

	return reply.IdNodo, reply.Mandato, reply.EsLider, reply.IdLider, err
}

// start  gestor de vistas; mapa de replicas y maquinas donde ubicarlos;
// y lista clientes (host:puerto)
func (cfg *configDespliegue) startDistributedProcesses() {
	//cfg.t.Log("Before starting following distributed processes: ", cfg.nodosRaft)

	for i, endPoint := range cfg.nodosRaft {
		despliegue.ExecMutipleHosts(EXECREPLICACMD+
			" "+strconv.Itoa(i)+" "+
			rpctimeout.HostPortArrayToString(cfg.nodosRaft),
			[]string{endPoint.Host()}, cfg.cr, PRIVKEYFILE)

		// dar tiempo para se establezcan las replicas
		time.Sleep(10000 * time.Millisecond)
	}

	// aproximadamente 500 ms para cada arranque por ssh en portatil
	time.Sleep(4000 * time.Millisecond)
}

func (cfg *configDespliegue) stopDistributedProcesses() {
	var reply raft.Vacio

	for _, endPoint := range cfg.nodosRaft {
		_ = endPoint.CallTimeout("NodoRaft.ParaNodo",
			raft.Vacio{}, &reply, 10*time.Millisecond)
		//check.CheckError(err, "Error en llamada RPC Para nodo")
	}
}

// Comprobar estado remoto de un nodo con respecto a un estado prefijado
func (cfg *configDespliegue) comprobarEstadoRemoto(idNodoDeseado int,
	mandatoDeseado int, esLiderDeseado bool, IdLiderDeseado int) {
	idNodo, _, esLider, idLider, err := cfg.obtenerEstadoRemoto(idNodoDeseado)
	check.CheckError(err, "Error en llamada RPC ObtenerEstadoRemoto")

	fmt.Println("Estado replica ", idNodoDeseado, idNodo, esLider, idLider)

	if idNodo != idNodoDeseado || esLider != esLiderDeseado || idLider != IdLiderDeseado {
		cfg.t.Fatalf("Estado incorrecto en replica %d en subtest %s",
			idNodoDeseado, cfg.t.Name())
	}

}

// ----------------------------- nuestras funciones  ---------------------------------------

func (cfg *configDespliegue) desconectarLider() {
	//time.Sleep(1000 * time.Millisecond)
	var reply raft.EstadoRemoto
	for _, endPoint := range cfg.nodosRaft {
		// buscamos el lider
		err := endPoint.CallTimeout("NodoRaft.ObtenerEstadoNodo",
			raft.Vacio{}, &reply, 20*time.Millisecond)
		check.CheckError(err, "Error en llamada RPC ObtenerEstadoNodo")

		// Si es lider paramos ese nodo
		if reply.EsLider {
			err = endPoint.CallTimeout("NodoRaft.ParaNodo",
				raft.Vacio{}, &reply, 20*time.Millisecond)
			check.CheckError(err, "Error en llamada RPC Para nodo")
		}
	}
	// Dar tiempo para encontrar nuevo lider
	//time.Sleep(1000 * time.Millisecond)

}

func (cfg *configDespliegue) desconectarUnNodo() {
	time.Sleep(1000 * time.Millisecond)
	var reply raft.EstadoRemoto
	continua := true
	for _, endPoint := range cfg.nodosRaft {
		// buscamos el lider
		err := endPoint.CallTimeout("NodoRaft.ObtenerEstadoNodo",
			raft.Vacio{}, &reply, 20*time.Millisecond)
		check.CheckError(err, "Error en llamada RPC ObtenerEstadoNodo")

		// Si es lider paramos ese nodo
		if !reply.EsLider && continua {
			continua = false
			err = endPoint.CallTimeout("NodoRaft.ParaNodo",
				raft.Vacio{}, &reply, 20*time.Millisecond)
			check.CheckError(err, "Error en llamada RPC Para nodo")
		}
	}

}

func (cfg *configDespliegue) conectarNodos() {
	time.Sleep(1000 * time.Millisecond)
	var reply raft.EstadoRemoto
	for i, endPoint := range cfg.nodosRaft {
		// buscamos el lider
		err := endPoint.CallTimeout("NodoRaft.ObtenerEstadoNodo",
			raft.Vacio{}, &reply, 20*time.Millisecond)
		check.CheckError(err, "Error en llamada RPC ObtenerEstadoNodo")

		// Si es lider paramos ese nodo
		if err != nil {
			despliegue.ExecMutipleHosts(EXECREPLICACMD+
				" "+strconv.Itoa(i)+" "+
				rpctimeout.HostPortArrayToString(cfg.nodosRaft),
				[]string{endPoint.Host()}, cfg.cr, PRIVKEYFILE)
		}
	}

}

func (cfg *configDespliegue) someterOperacion(operacion string, clave int) {
	// creamos el parámetro a enviar
	var args raft.TipoOperacion
	args.Operacion = operacion
	args.Clave = clave

	var reply raft.EstadoRemoto

	for _, endPoint := range cfg.nodosRaft {
		// buscamos el lider
		err := endPoint.CallTimeout("NodoRaft.ObtenerEstadoNodo",
			raft.Vacio{}, &reply, 20*time.Millisecond)

		// Si es lider le solicitamos someter la op
		if reply.EsLider && err == nil {
			err = endPoint.CallTimeout("NodoRaft.SometerOperacionRaft",
				raft.Vacio{}, &reply, 20*time.Millisecond)
			check.CheckError(err, "Error en llamada RPC Someter Operacion")
		}
	}
}

func (cfg *configDespliegue) comprobarOperacionesComprometidas(solucionCorrecta int) {
	// damos tiempo a que se comprometan las entradas
	time.Sleep(1000 * time.Millisecond)
	var mandato int = 0
	var reply raft.EntradasComprometidas
	for _, endPoint := range cfg.nodosRaft {

		err := endPoint.CallTimeout("NodoRaft.NumEntradasComprometidas",
			raft.Vacio{}, &reply, 20*time.Millisecond)
		// aseguramos que el número de entradas comprometidas por el nodo es correcto
		if err == nil {
			if reply.Comprometidas != solucionCorrecta {
				cfg.t.Fatalf("Numero de entradas comprometidas %d, deberia ser: %d",
					reply.Comprometidas, solucionCorrecta)
			}

			if mandato == 0 {
				mandato = reply.Mandato
			} else if reply.Mandato != mandato {
				cfg.t.Fatalf("Los mandatos no coinciden")
			}
		} else {
			fmt.Println("El  nodo ", endPoint, " esta muerto")
		}

	}
}

func (cfg *configDespliegue) comprobarOperacionesSinComprometer(solucionCorrecta int) {
	// damos tiempo a que se comprometan las entradas
	time.Sleep(1000 * time.Millisecond)
	var reply raft.EntradasSinComprometer
	for _, endPoint := range cfg.nodosRaft {

		err := endPoint.CallTimeout("NodoRaft.NumEntradasSinComprometer",
			raft.Vacio{}, &reply, 20*time.Millisecond)
		// aseguramos que el número de entradas comprometidas por el nodo es correcto
		if err == nil {
			if reply.SinComprometer != solucionCorrecta {
				cfg.t.Fatalf("Numero de entradas sin comprometer %d, deberia ser: %d",
					reply.SinComprometer, solucionCorrecta)
			}

		} else {
			fmt.Println("El  nodo ", endPoint, " esta muerto")
		}

	}
}
