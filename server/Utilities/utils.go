package Utilities

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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

func Create_Parent_Dir(path string) error {
	dir := filepath.Dir(path)
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return fmt.Errorf("error al crear el directorio padre: %s", err)
	}
	return nil
}

// Funcion para obtener las carpetas padres y el directorio destino
func Get_Parent_Dirs(path string) ([]string, string) {

	/*
		Se estandariza el path, eliminando redundancias
		//home//user///docs → /home/user/docs
	*/
	path = filepath.Clean(path)

	// Dividir el path en partes
	directories := strings.Split(path, string(filepath.Separator))

	// Se crea una lista para almacenar las rutas de los directorios padre
	var parent_Dir []string

	// Se generarn las ruta de las carpetas padres, excluyendo el ultimo directorio
	for i := 1; i < len(directories)-1; i++ {
		parent_Dir = append(parent_Dir, directories[i])
	}

	// El ultimo elemento es el destino
	dest_Dir := directories[len(directories)-1]

	return parent_Dir, dest_Dir

}

func Get_File_Names(path string) (string, string) {
	dir := filepath.Dir(path)
	baseName := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	dotFileName := filepath.Join(dir, baseName+".dot")
	outputImage := path
	return dotFileName, outputImage
}
