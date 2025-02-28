package Structs

import (
	"fmt"
	"os"
	"path/filepath"
	util "server/Utilities"
)

type MKDISK struct {
	Size int
	Unit string
	Fit  string
	Path string
}

func Struct_MKDISK(disk *MKDISK) (string, error) {

	zBytes, err := util.ConvertBytes(disk.Size, disk.Unit)

	if err != nil {
		fmt.Println("Error al convertir el size: ", err)
		return "Error: No se pudo convertir el size del disco", err
	}

	var msg string

	msg, err = MakeDisk(disk, zBytes)
	if err != nil {
		fmt.Println("Error al crear el disco: ", err)
		return msg, err
	}

	msg, err = CreateMBR(disk, zBytes)
	if err != nil {
		fmt.Println("Error al crear el MBR: ", err)
		return msg, err
	}

	return "", nil
}

func MakeDisk(disk *MKDISK, sizeB int) (string, error) {

	err := os.MkdirAll(filepath.Dir(disk.Path), os.ModePerm)

	if err != nil {
		fmt.Println("Error al crear la carpeta: ", err)
		return "Error: No se pudo crear la carpeta", err
	}

	file, err := os.Create(disk.Path)
	if err != nil {
		fmt.Println("Error al crear el archivo: ", err)
		return "Error: No se pudo crear el archivo", err
	}

	defer file.Close()

	buffer := make([]byte, 1024*1024)
	for sizeB > 0 {
		wSize := len(buffer)
		if sizeB < wSize {
			wSize = sizeB
		}
		if _, err := file.Write(buffer[:wSize]); err != nil {
			return "Error: Error al escribir el disco", err
		}
		sizeB -= wSize
	}
	return "", nil

}
