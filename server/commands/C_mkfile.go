package commands

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type MKFILE struct {
	Path string
	R    bool
	Size int
	Cont string
}

func Mkfile_Command(tokens []string) (*MKFILE, string, error) {

	// creamos una nueva instancia de mkfile
	mkfile := &MKFILE{
		R: false,
	}

	// unimos todos los tokens en una sola cadena y luego se divide por espacios
	atributos := strings.Join(tokens, " ")
	// expresion regular para encontrar los parametros del comando
	lexic := regexp.MustCompile(`(?i)-path="[^"]+"|(?i)-path=[^\s]+|(?i)-r|-size=\d+|(?i)-cont="[^"]+"|(?i)-cont=[^\s]+`)
	// encuentra todas las coincidencias de la expresion regular en la cadena de argumentos
	found := lexic.FindAllString(atributos, -1)

	// verificar que todos los tokens fueron reconocidos por la expresion
	for _, fun := range found {

		if strings.EqualFold(fun, "-r") { // Verificar si es el parámetro -r
			mkfile.R = true
			continue
		}

		parametro := strings.SplitN(fun, "=", 2)
		if len(parametro) != 2 {
			return nil, "", fmt.Errorf("[error comando mkfile] formato de parametro invalido: %s", fun)
		}

		key, value := strings.ToLower(parametro[0]), parametro[1]

		if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
			value = strings.Trim(value, "\"")
		}

		switch key {

		case "-path":

			if value == "" {
				return nil, "", errors.New("[error comando mkfile] el path no puede estar vacio")
			}
			mkfile.Path = value

		case "-size":

			size, err := strconv.Atoi(value)
			if err != nil || size <= 0 {
				return nil, "", errors.New("[error comando mkfile] el parametro de size no puede ser 0 o menor")
			}
			mkfile.Size = size

		case "-cont":

			if value == "" {
				mkfile.Cont = ""
			}

		default:
			return nil, "", fmt.Errorf("[error comando mkfile] parametro desconocido: %s", key)
		}

	}

	if mkfile.Path == "" {
		return nil, "", errors.New("[error comando mkfile] el path no puede estar vacio")
	}

	if mkfile.Cont == "" && mkfile.Size == 0 {
		return nil, "", errors.New("[error comando mkfile] los parametros cont y size, no pueden estar vacios al mismo tiempo")
	}

	return mkfile, "", nil

}
