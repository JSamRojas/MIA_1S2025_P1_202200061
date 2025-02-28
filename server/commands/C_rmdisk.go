package commands

import (
	"errors"
	"fmt"
	"regexp"
	util "server/Utilities"
	"strings"
)

type RMDISK struct {
	pathdisk string
}

func Rmdisk_Command(tokens []string) (*RMDISK, string, error) {

	rm := &RMDISK{}

	atributos := strings.Join(tokens, " ")

	lexic := regexp.MustCompile(`(?i)-path="[^"]+"|(?i)-path=[^\s]+`)

	found := lexic.FindAllString(atributos, -1)

	for _, fu := range found {

		parametro := strings.SplitN(fu, "=", 2)
		if len(parametro) != 2 {
			return nil, "ERROR: Parametro invalido", fmt.Errorf("ERROR: Parametro invalido: %s", fu)
		}

		key, value := strings.ToLower(parametro[0]), parametro[1]

		if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
			value = strings.Trim(value, "\"")
		}

		switch key {
		case "-path":
			if value == "" {
				return nil, "ERROR: El path del disco no puede estar vacio", errors.New("ERROR: El path del disco no puede estar vacio")
			}
			rm.pathdisk = value
		default:
			return nil, "ERROR: Parametro no reconocido", fmt.Errorf("ERROR: Parametro no reconocido: %s", key)
		}

	}

	if rm.pathdisk == "" {
		return nil, "ERROR: Faltan parametros obligatorios (parametro -path)", errors.New("ERROR: Faltan parametros obligatorios")
	}

	msg, err := util.RemoveDiskFile(rm.pathdisk)
	if err != nil {
		return nil, "ERROR: No se pudo eliminar el disco", err
	}
	return rm, msg, nil

}
