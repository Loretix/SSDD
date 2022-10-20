package gestorfichero

import (
	"fmt"
	"os"
)

func CheckError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
}

func LeerFichero() string {
	ruta, err := os.Getwd()
	CheckError(err)
	leido, err := os.ReadFile(ruta + "/fichero.txt")
	CheckError(err)
	datosLeidos := string(leido)
	return datosLeidos
}

func EscribirFichero(fragmento string) {
	ruta, err := os.Getwd() // devuelve la ruta desde la que se ejecuta el programa
	CheckError(err)
	os.WriteFile(ruta+"/fichero.txt", []byte(fragmento), 0777)
}
