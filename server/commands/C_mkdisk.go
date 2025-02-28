package commands

import (
	"errors"
	"fmt"
	"regexp"
	estructuras "server/Structs"
	"strconv"
	"strings"
)

func Mkdisk_Command(tokens []string) (*estructuras.MKDISK, string, error) {

	disk := &estructuras.MKDISK{}

	atributos := strings.Join(tokens, " ")

	lexic := regexp.MustCompile(`(?i)-size=\d+|(?i)-unit=[kKmM]|(?i)-fit=[bBfFwW]{2}|(?i)-path="[^"]+"|(?i)-path=[^\s]+`)

	found := lexic.FindAllString(atributos, -1)

	if len(found) != len(tokens) {
		for _, token := range tokens {
			if !lexic.MatchString(token) {
				return nil, "", fmt.Errorf("ERROR: Parametro no reconocido: %s en comando MKDISK", token)
			}
		}
	}

	for _, fun := range found {
		parametro := strings.SplitN(fun, "=", 2)
		if len(parametro) != 2 {
			return nil, "ERROR: Parametro invalido", fmt.Errorf(("parametro invalido: %s"), fun)
		}
		key, value := strings.ToLower(parametro[0]), parametro[1]

		if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
			value = strings.Trim(value, "\"")
		}

		switch key {
		case "-size":
			size, err := strconv.Atoi(value)
			if err != nil || size <= 0 {
				return nil, "ERROR: La capacidad del disco debe ser un numero entero positivo", errors.New("la capacidad del disco debe ser un numero entero positivo")
			}
			disk.Size = size

		case "-unit":
			value = strings.ToUpper(value)
			if value != "K" && value != "M" {
				return nil, "ERROR: la unidad del disco debe ser K o M", errors.New("la unidad del disco debe ser K o M")
			}
			disk.Unit = value

		case "-fit":
			value = strings.ToUpper(value)
			if value != "BF" && value != "FF" && value != "WF" {
				return nil, "ERROR: El ajuste del disco debe ser BF, FF o WF", errors.New("el ajuste del disco debe ser BF, FF o WF")
			}
			disk.Fit = value

		case "-path":
			if value == "" {
				return nil, "ERROR: El path del disco no puede ser vacio", errors.New("el path del disco no puede ser vacio")
			}
			disk.Path = value

		default:
			return nil, "ERROR: Parametro no reconocido", fmt.Errorf("parametro no reconocido: %s", key)
		}
	}

	if disk.Size == 0 {
		return nil, "ERROR: La capacidad del disco no puede ser 0", errors.New("la capacidad del disco no puede ser 0")
	}

	if disk.Path == "" {
		return nil, "ERROR: El path del disco no puede ser vacio", errors.New("el path del disco no puede ser vacio")
	}

	if disk.Unit == "" {
		disk.Unit = "M"
	}

	if disk.Fit == "" {
		disk.Fit = "FF"
	}

	msg, err := estructuras.Struct_MKDISK(disk)
	if err != nil {
		fmt.Println("Error al crear el disco: ", err)
		return nil, msg, err
	}

	return disk, "MKDISK: Disco creado con exito", nil

}
