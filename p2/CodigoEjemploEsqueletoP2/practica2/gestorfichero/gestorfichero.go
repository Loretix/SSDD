package gestorfichero

import (
	"fmt"
	"os"
	"strconv"
)

func CheckError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
}

func LeerFichero(me int) string {
	ruta, err := os.Getwd()
	CheckError(err)
	leido, err := os.ReadFile(ruta + "/fichero" + strconv.Itoa(me) + ".txt")
	CheckError(err)
	datosLeidos := string(leido)
	return datosLeidos
}

func EscribirFichero(fragmento string, N int) {
	ruta, err := os.Getwd() // devuelve la ruta desde la que se ejecuta el programa
	CheckError(err)
	for i := 1; i <= N; i++ {
		fichero, err := os.OpenFile(ruta+"/fichero"+strconv.Itoa(i)+".txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		CheckError(err)
		fichero.WriteString(fragmento)
	}
}
