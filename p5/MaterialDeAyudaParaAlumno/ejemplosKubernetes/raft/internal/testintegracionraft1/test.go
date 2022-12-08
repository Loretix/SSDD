package main

import (
	"fmt"

	//"log"
	//"crypto/rand"
	//"os"
	//"path/filepath"
	"time"

	"raft/internal/comun/check"
	"raft/internal/comun/rpctimeout"
	"raft/internal/raft"
)

const (

	//nodos replicas
	REPLICA1 = "ss-0.ss-service.default.svc.cluster.local:6000"
	REPLICA2 = "ss-1.ss-service.default.svc.cluster.local:6000"
	REPLICA3 = "ss-2.ss-service.default.svc.cluster.local:6000"

)


// TEST primer rango
func main() { // (m *testing.M) {
	// <setup code>
	// Crear canal de resultados de ejecuciones ssh en maquinas remotas
	cfg := makeCfgDespliegue(3,
		[]string{REPLICA1, REPLICA2, REPLICA3},
		[]bool{true, true, true})

	// tear down code
	// eliminar procesos en máquinas remotas
	defer cfg.stop()

	// Run test sequence 
	time.Sleep(20000 * time.Millisecond)
	// Test2 : No debería haber ningun primario, si SV no ha recibido aún latidos
	cfg.elegirPrimerLiderTest2()

	// Test3: tenemos el primer primario correcto
	cfg.falloAnteriorElegirNuevoLiderTest3() 

	// Test4: Tres operaciones comprometidas en configuración estable
	cfg.tresOperacionesComprometidasEstable()

	// Test5: Se consigue acuerdo a pesar de desconexiones de seguidor
	cfg.AcuerdoApesarDeSeguidor() 
	time.Sleep(5000 * time.Millisecond)
	// Test6: NO se consigue acuerdo al desconectarse mayoría de seguidores -- 3 NODOS RAFT
	cfg.SinAcuerdoPorFallos() 
	time.Sleep(5000 * time.Millisecond)
	// Test7: Se somete 5 operaciones de forma concurrente -- 3 NODOS RAFT
	cfg.SometerConcurrentementeOperaciones() 

}

// ---------------------------------------------------------------------
//
// Canal de resultados de ejecución de comandos ssh remotos
//type canalResultados chan string

// ---------------------------------------------------------------------
// Operativa en configuracion de despliegue y pruebas asociadas
type configDespliegue struct {
	conectados  []bool
	numReplicas int
	nodosRaft   []rpctimeout.HostPort
}

// Crear una configuracion de despliegue
func makeCfgDespliegue(n int, nodosraft []string,
	conectados []bool) *configDespliegue {
	cfg := &configDespliegue{}
	cfg.conectados = conectados
	cfg.numReplicas = n
	cfg.nodosRaft = rpctimeout.StringArrayToHostPortArray(nodosraft)

	return cfg
}

func (cfg *configDespliegue) stop() {
	cfg.stopDistributedProcesses()

	time.Sleep(10000 * time.Millisecond)

}

// --------------------------------------------------------------------------
// FUNCIONES DE SUBTESTS

// TEST 2: Primer lider en marcha - 3 NODOS RAFT
func (cfg *configDespliegue) elegirPrimerLiderTest2() {

	fmt.Println(".....................")

	//cfg.startDistributedProcesses()

	// Se ha elegido lider ?
	fmt.Printf("Probando lider en curso\n")
	cfg.pruebaUnLider(3)

	// Parar réplicas alamcenamiento en remoto
	cfg.stopDistributedProcesses() // Parametros

	fmt.Println("............. Test 2 Superado")
	time.Sleep(20000 * time.Millisecond)
}

// TEST 3: Fallo de un primer lider y reeleccion de uno nuevo - 3 NODOS RAFT
func (cfg *configDespliegue) falloAnteriorElegirNuevoLiderTest3(){

	fmt.Println(".....................")

	//cfg.startDistributedProcesses()

	fmt.Printf("Lider inicial\n")
	cfg.pruebaUnLider(3)

	// Desconectar lider
	cfg.desconectarLider()

	fmt.Printf("Comprobar nuevo lider\n")
	cfg.pruebaUnLider(3)

	fmt.Printf("Paramos el proceso distribuido\n")
	// Parar réplicas almacenamiento en remoto
	cfg.stopDistributedProcesses() //parametros

	fmt.Println("............. Test 3 Superado")
	time.Sleep(20000 * time.Millisecond)
}

// TEST 4: 3 operaciones comprometidas con situacion estable y sin fallos - 3 NODOS RAFT
func (cfg *configDespliegue) tresOperacionesComprometidasEstable() {

	fmt.Println(".....................")

	//cfg.startDistributedProcesses()
	fmt.Println("Sometemos las entradas")
	cfg.someterOperacion("escritura", 1)
	cfg.someterOperacion("escritura", 2)
	cfg.someterOperacion("escritura", 3)

	fmt.Println("Comprobamos que se han sometido bien")
	cfg.comprobarOperacionesComprometidas(3)

	// Parar réplicas almacenamiento en remoto
	cfg.stopDistributedProcesses() //parametros

	fmt.Println("............. Test 4 Superado")
	time.Sleep(20000 * time.Millisecond)
}

// TEST 5: Se consigue acuerdo a pesar de desconexiones de seguidor -- 3 NODOS RAFT
func (cfg *configDespliegue) AcuerdoApesarDeSeguidor(){
	fmt.Println(".....................")

	//cfg.startDistributedProcesses()

	//  Obtener un lider y, a continuación desconectar una de los nodos Raft
	fmt.Println("Desconectamos un nodo")
	cfg.matarUnNodo()
	// Comprobar varios acuerdos con una réplica desconectada
	cfg.someterOperacion("escritura", 1)
	cfg.someterOperacion("escritura", 2)
	cfg.someterOperacion("escritura", 3)

	cfg.comprobarOperacionesComprometidas(3)
	fmt.Println("- Se han comprometido las entradas en los nodos vivos")

	// reconectar nodo Raft previamente desconectado y comprobar varios acuerdos
	fmt.Println("- Reconectamos el nodo")
	//cfg.conectarNodos()
	cfg.comprobarOperacionesComprometidas(3)
	fmt.Println("- Se han comprometido las entradas en todos los nodos")

	cfg.stopDistributedProcesses() //parametros

	fmt.Println("............. Test 5 Superado")
	time.Sleep(20000 * time.Millisecond)
}

// TEST 6: NO se consigue acuerdo al desconectarse mayoría de seguidores -- 3 NODOS RAFT
func (cfg *configDespliegue) SinAcuerdoPorFallos() {

	fmt.Println(".....................")

	//cfg.startDistributedProcesses()
	//  Obtener un lider y, a continuación desconectar 2 de los nodos Raft
	cfg.matarUnNodo()
	cfg.matarUnNodo()

	// Comprobar varios acuerdos con 2 réplicas desconectada
	cfg.someterOperacion("escritura", 1)
	cfg.someterOperacion("escritura", 2)
	cfg.someterOperacion("escritura", 3)
	// reconectar lo2 nodos Raft  desconectados y probar varios acuerdos
	cfg.comprobarOperacionesSinComprometer(3)
	// reconectar nodo Raft previamente desconectado y comprobar varios acuerdos
	//cfg.conectarNodos()
	// comprobamos que las entradas se comprometen una vez reconectados los nodos
	cfg.comprobarOperacionesComprometidas(3)
	cfg.stopDistributedProcesses() //parametros

	fmt.Println(".............", "Test 6 Superado")
	time.Sleep(20000 * time.Millisecond)
}

// TEST 7: Se somete 5 operaciones de forma concurrente -- 3 NODOS RAFT
func (cfg *configDespliegue) SometerConcurrentementeOperaciones() {
	//t.Skip("SKIPPED SometerConcurrentementeOperaciones")


	// un bucle para estabilizar la ejecucion

	// Obtener un lider y, a continuación someter una operacion
	// Someter 5  operaciones concurrentes

	//cfg.startDistributedProcesses()
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

	fmt.Println(".............",  "Test 7 Superado")
	time.Sleep(20000 * time.Millisecond)
}

// --------------------------------------------------------------------------
// FUNCIONES DE APOYO
// Comprobar que hay un solo lider
// probar varias veces si se necesitan reelecciones
func (cfg *configDespliegue) pruebaUnLider(numreplicas int) int {
	for iters := 0; iters < 10; iters++ {
		time.Sleep(1000 * time.Millisecond)
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
				fmt.Println("mandato", mandato, "tiene", len(lideres), "(>1) lideres")
			}
			if mandato > ultimoMandatoConLider {
				ultimoMandatoConLider = mandato
			}
		}

		if len(mapaLideres) != 0 {

			return mapaLideres[ultimoMandatoConLider][0] // Termina

		}
	}
	fmt.Println("un lider esperado, ninguno obtenido")

	return -1 // Termina
}

func (cfg *configDespliegue) obtenerEstadoRemoto(
	indiceNodo int) (int, int, bool, int, error) {
	var reply raft.EstadoRemoto
	err := cfg.nodosRaft[indiceNodo].CallTimeout("NodoRaft.ObtenerEstadoNodo",
		raft.Vacio{}, &reply, 5000*time.Millisecond)
	//fmt.Println(cfg.nodosRaft[indiceNodo])
	//check.CheckError(err, "Error en llamada RPC ObtenerEstadoRemoto")

	return reply.IdNodo, reply.Mandato, reply.EsLider, reply.IdLider, err
}

func (cfg *configDespliegue) stopDistributedProcesses() {
	var reply raft.Vacio

	for _, endPoint := range cfg.nodosRaft {
		err := endPoint.CallTimeout("NodoRaft.ParaNodo",
			raft.Vacio{}, &reply, 1000*time.Millisecond)
		check.CheckError(err, "Error en llamada RPC Para nodo")
	}
}

// Comprobar estado remoto de un nodo con respecto a un estado prefijado
func (cfg *configDespliegue) comprobarEstadoRemoto(idNodoDeseado int,
	mandatoDeseado int, esLiderDeseado bool, IdLiderDeseado int) {
	idNodo, _, esLider, idLider, err := cfg.obtenerEstadoRemoto(idNodoDeseado)
	check.CheckError(err, "Error en llamada RPC ObtenerEstadoRemoto")

	fmt.Println("Estado replica ", idNodoDeseado, idNodo, esLider, idLider)

	if idNodo != idNodoDeseado || esLider != esLiderDeseado || idLider != IdLiderDeseado {
		fmt.Println("Estado incorrecto en replica", idNodoDeseado, " en subtest")
	}
}

// ----------------------------- nuestras funciones  ---------------------------------------

func (cfg *configDespliegue) desconectarLider() {
	time.Sleep(20000 * time.Millisecond)
	var reply raft.EstadoRemoto
	for i, endPoint := range cfg.nodosRaft {
		// buscamos el lider
		_, _, esLider, _, err := cfg.obtenerEstadoRemoto(i)
		check.CheckError(err, "Error en llamada RPC ObtenerEstadoRemoto")

		// Si es lider paramos ese nodo
		if esLider {
			err := endPoint.CallTimeout("NodoRaft.DormirNodo",
				raft.Vacio{}, &reply, 1000*time.Millisecond)
			check.CheckError(err, "Error en llamada RPC Para nodo")

			fmt.Println("Lider parado", endPoint)
		}
	}
	// Dar tiempo para encontrar nuevo lider
	time.Sleep(20000 * time.Millisecond)

}

func (cfg *configDespliegue) matarUnNodo() {
	time.Sleep(20000 * time.Millisecond)
	var reply raft.EstadoRemoto
	continua := true
	for i, endPoint := range cfg.nodosRaft {
		// buscamos el lider
		_, _, esLider, _, err := cfg.obtenerEstadoRemoto(i)
		//check.CheckError(err, "Error en llamada RPC ObtenerEstadoRemoto")

		fmt.Println("Me pienso si lo mato ", esLider, i)
		// Si no es lider paramos ese nodo
		fmt.Println(!esLider, "&&", continua,"&&", err, "--------------------------------------" )
		
		if !esLider && continua && err == nil {
			continua = false
			fmt.Println("Me muero", endPoint)
			err := endPoint.CallTimeout("NodoRaft.ParaNodo",
				raft.Vacio{}, &reply, 1000*time.Millisecond)
			check.CheckError(err, "Error en llamada RPC Para nodo")
		}
	}
	time.Sleep(20000 * time.Millisecond)
}
/*
func (cfg *configDespliegue) conectarNodos() {
	time.Sleep(1000 * time.Millisecond)
	var reply raft.EstadoRemoto
	for i, endPoint := range cfg.nodosRaft {
		// buscamos el lider
		err := endPoint.CallTimeout("NodoRaft.ObtenerEstadoNodo",
			raft.Vacio{}, &reply, 1000*time.Millisecond)
		//check.CheckError(err, "Error en llamada RPC ObtenerEstadoNodo")

		if err != nil {
			despliegue.ExecMutipleHosts(EXECREPLICACMD+
				" "+strconv.Itoa(i)+" "+
				rpctimeout.HostPortArrayToString(cfg.nodosRaft),
				[]string{endPoint.Host()}, cfg.cr, PRIVKEYFILE)

			fmt.Println("Conectamos", endPoint)

		}
	}

}
*/

func (cfg *configDespliegue) someterOperacion(operacion string, clave int) {
	// creamos el parámetro a enviar
	time.Sleep(20000 * time.Millisecond)

	var args raft.TipoOperacion
	args.Operacion = operacion
	args.Clave = clave

	var reply raft.EstadoRemoto

	for i, endPoint := range cfg.nodosRaft {
		// buscamos el lider
		_, _, esLider, _, err := cfg.obtenerEstadoRemoto(i)
		fmt.Println("Someter operacion", esLider, err)
		// Si es lider le solicitamos someter la op
		if esLider && err == nil {
			err = endPoint.CallTimeout("NodoRaft.SometerOperacionRaft",
				raft.Vacio{}, &reply, 1000*time.Millisecond)
			check.CheckError(err, "Error en llamada RPC Someter Operacion")
		}
	}
	time.Sleep(20000 * time.Millisecond)

}

func (cfg configDespliegue) comprobarOperacionesComprometidas(solucionCorrecta int) {
	// damos tiempo a que se comprometan las entradas
	time.Sleep(20000 * time.Millisecond)
	var mandato int = 0
	var reply raft.EntradasComprometidas
	for _, endPoint := range cfg.nodosRaft {

		err := endPoint.CallTimeout("NodoRaft.NumEntradasComprometidas",
			raft.Vacio{}, &reply, 1000*time.Millisecond)

		//check.CheckError(err, "se jodio")
		// aseguramos que el número de entradas comprometidas por el nodo es correcto
		if err == nil {
			if reply.Comprometidas != solucionCorrecta {
				fmt.Println("Numero de entradas comprometidas, ", reply.Comprometidas," deberia ser: ", solucionCorrecta)
			}

			if mandato == 0 {
				mandato = reply.Mandato
			} else if reply.Mandato != mandato {
				fmt.Println("Los mandatos no coinciden")
			}
		} else {
			fmt.Println("El  nodo ", endPoint, " esta muerto")
		}

	}
	time.Sleep(20000 * time.Millisecond)

}

func (cfg *configDespliegue) comprobarOperacionesSinComprometer(solucionCorrecta int) {
	// damos tiempo a que se comprometan las entradas
	time.Sleep(20000 * time.Millisecond)
	var reply raft.EntradasSinComprometer
	for _, endPoint := range cfg.nodosRaft {

		err := endPoint.CallTimeout("NodoRaft.NumEntradasSinComprometer",
			raft.Vacio{}, &reply, 1000*time.Millisecond)
		// aseguramos que el número de entradas comprometidas por el nodo es correcto
		if err == nil {
			if reply.SinComprometer != solucionCorrecta {
				fmt.Println("Numero de entradas sin comprometer", reply.SinComprometer, "deberia ser:", solucionCorrecta)
			}

		} else {
			fmt.Println("El  nodo ", endPoint, " esta muerto")
		}

	}
	time.Sleep(20000 * time.Millisecond)

}
