package commands

import (
	"errors"
	"fmt"
	"regexp"
	estructuras "server/Structs"
	"strconv"
	"strings"
)

func Fdisk_Command(tokens []string) (*estructuras.FDISK, string, error) {

	fdisk := &estructuras.FDISK{}

	atributos := strings.Join(tokens, " ")

	lexic := regexp.MustCompile(`(?i)-size=\d+|(?i)-unit=[bBkKmM]|(?i)-fit=[bBfF]{2}|(?i)-path="[^"]+"|(?i)-path=[^\s]+|(?i)-type=[pPeElL]|(?i)-name="[^"]+"|(?i)-name=[^\s]+`)

	found := lexic.FindAllString(atributos, -1)

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
				return nil, "ERROR: El tamaño de la particion debe ser un numero entero positivo", errors.New("el tamaño de la particion debe ser un numero entero positivo")
			}
			fdisk.Size = size

		case "-unit":

			value = strings.ToUpper(value)
			if value != "B" && value != "K" && value != "M" {
				return nil, "ERROR: la unidad de la particion debe ser K o M", errors.New("la unidad de la particion debe ser K o M")
			}
			fdisk.Unit = value

		case "-path":

			if value == "" {
				return nil, "ERROR: El path de la particion no puede ser vacio", errors.New("el path de la particion no puede ser vacio")
			}
			fdisk.Path = value

		case "-type":

			value = strings.ToUpper(value)
			if value != "P" && value != "E" && value != "L" {
				return nil, "ERROR: El tipo de la particion debe ser P, E o L", errors.New("el tipo de la particion debe ser P, E o L")
			}
			fdisk.Type = value

		case "-fit":

			value = strings.ToUpper(value)
			if value != "BF" && value != "FF" && value != "WF" {
				return nil, "ERROR: El ajuste de la particion debe ser BF, FF o WF", errors.New("el ajuste de la particion debe ser BF, FF o WF")
			}
			fdisk.Fit = value

		case "-name":

			if value == "" {
				return nil, "ERROR: El nombre de la particion no puede ser vacio", errors.New("el nombre de la particion no puede ser vacio")
			}
			fdisk.Name = value

		default:

			return nil, "ERROR: Parametro no reconocido", fmt.Errorf("parametro no reconocido: %s", key)

		}

	}

	if fdisk.Size == 0 {
		return nil, "ERROR: El tamaño de la particion no puede ser 0", errors.New("el tamaño de la particion no puede ser 0")
	}

	if fdisk.Path == "" {
		return nil, "ERROR: El path de la particion no puede ser vacio", errors.New("el path de la particion no puede ser vacio")
	}

	if fdisk.Name == "" {
		return nil, "ERROR: El nombre de la particion no puede ser vacio", errors.New("el nombre de la particion no puede ser vacio")
	}

	if fdisk.Unit == "" {
		fdisk.Unit = "M"
	}

	if fdisk.Fit == "" {
		fdisk.Fit = "WF"
	}

	if fdisk.Type == "" {
		fdisk.Type = "P"
	}

	msg, err := estructuras.Struct_FDISK(fdisk)
	if err != nil {
		fmt.Println("Error al crear la particion: ", err)
		return nil, msg, err
	}

	/*
		var mbrPrint estructuras.MBR

		msg, err1 := mbrPrint.DeserializeMBR(fdisk.Path)
		if err1 != nil {
			fmt.Println("Error al leer el MBR: ", err1)
			return nil, msg, err1
		}
	*/

	//mbrPrint.Print()
	//fmt.Println(" ")
	//mbrPrint.PrintPartitions()

	return fdisk, "FDISK: Particion creada correctamente", nil

}
