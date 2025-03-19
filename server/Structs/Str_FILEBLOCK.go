package Structs

import (
	"encoding/binary"
	"os"
)

type FILEBLOCK struct {
	B_content [64]byte
}

func (fileblock *FILEBLOCK) Serialize(path string, offset int64) error {

	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// MOver el puntero del archivo a la posicion del offset
	_, err = file.Seek(offset, 0)
	if err != nil {
		return err
	}

	// Serializar la estructura FileBlock firectamente en el archivo
	err = binary.Write(file, binary.LittleEndian, fileblock)
	if err != nil {
		return err
	}

	return nil

}
