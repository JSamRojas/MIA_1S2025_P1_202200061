package Utilities

import (
	"errors"
	"fmt"
	"os"
)

const Carnet string = "61" //202200061

var pathToLetter = make(map[string]string)

var Alfabeto = []string{
	"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M", "N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z"}

var nextLetterIndex = 0

func ConvertBytes(size int, unit string) (int, error) {

	switch unit {
	case "B":
		return size, nil
	case "K":
		return size * 1024, nil
	case "M":
		return size * 1048576, nil
	default:
		return 0, errors.New("error: Unidad no reconocida")
	}

}

func RemoveDiskFile(path string) (string, error) {
	err := os.Remove(path)
	if err != nil {
		return "", fmt.Errorf("error al eliminar el disco: %s", err)
	}

	return "RMDISK: Disco eliminado correctamente", nil

}

func GetLetra(path string) (string, error) {

	if _, exist := pathToLetter[path]; !exist {
		if nextLetterIndex < len(Alfabeto) {
			pathToLetter[path] = Alfabeto[nextLetterIndex]
			nextLetterIndex++
		} else {
			return "No hay letras disponibles, demasiados discos", errors.New("no hay letras disponibles, demasiados discos")
		}
	}
	return pathToLetter[path], nil

}
