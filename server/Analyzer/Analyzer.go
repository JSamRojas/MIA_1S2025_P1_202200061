package Analyzer

import (
	"fmt"
	//"os"
	//"os/exec"
	commands "server/commands"
	"strings"
)

func Analyzer(inputs []string) ([]string, []string) {
	var results []string
	var errors []string

	// Si no se proporciona ningún comando, se devuelve un error
	if len(inputs) == 0 {
		errors = append(errors, "No se proporcionó ningún comando")
		return results, errors
	}

	// Se confirma si el comando no es una linea en blanco o un comentario
	// Si lo es, se agrega a los resultados y se devuelve
	input := strings.TrimSpace(inputs[0])
	if input == "" || strings.HasPrefix(input, "#") {
		results = append(results, "\n"+input+"\n")
		return results, errors
	}

	tokens := strings.Fields(input)
	if len(tokens) == 0 {
		errors = append(errors, "No se proporciono ningun comando valido")
		return results, errors
	}

	tokens[0] = strings.ToLower(tokens[0])
	var msg string
	var err error

	switch tokens[0] {
	case "mkdisk":
		_, msg, err = commands.Mkdisk_Command(tokens[1:])

	case "rmdisk":
		_, msg, err = commands.Rmdisk_Command(tokens[1:])

	case "fdisk":
		_, msg, err = commands.Fdisk_Command(tokens[1:])

	case "mount":
		_, msg, err = commands.Mount_Command(tokens[1:])

	case "mounted":
		msg, err = commands.Mounted_Command(tokens[1:])

	case "mkfs":
		_, msg, err = commands.Mkfs_Command(tokens[1:])

	case "cat":
		_, msg, err = commands.Cat_Command(tokens[1:])

	case "login":
		_, msg, err = commands.Login_Command(tokens[1:])

	case "logout":
		_, msg, err = commands.Logout_Command(tokens[1:])

	case "rep":
		_, msg, err = commands.Rep_Command(tokens[1:])

	default:
		err = fmt.Errorf("comando no reconocido: %s", tokens[0])
	}

	if err != nil {
		errors = append(errors, err.Error())
	} else {
		results = append(results, msg)
	}

	return results, errors

}
