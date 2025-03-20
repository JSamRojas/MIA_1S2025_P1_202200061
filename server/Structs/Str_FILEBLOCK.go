package Structs

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
)

type FILEBLOCK struct {
	B_content [64]byte
}

// Funcion para escribir la estructura del fileblock dentro del archivo binario (o Disco)
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

// Funcion para leer la estructura fileblock haciendo uso de un archivo binario
func (fileblock *FILEBLOCK) Deserialize(path string, offset int64) error {

	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	// Mover el apuntador a la posicion del offset
	_, err = file.Seek(offset, 0)
	if err != nil {
		return err
	}

	// Obtener el size de la estructura fileblock
	fb_Size := binary.Size(fileblock)
	if fb_Size <= 0 {
		return fmt.Errorf("size de fileblock invalidao: %d", fb_Size)
	}

	// Leemos la cantidad de bytes correspondiente al fileblock desde la posicion del offset
	buffer := make([]byte, fb_Size)
	_, err = file.Read(buffer)
	if err != nil {
		return err
	}

	// Deserealizar los bytes leidos en la estructura del fileblock
	reader := bytes.NewReader(buffer)
	err = binary.Read(reader, binary.LittleEndian, fileblock)
	if err != nil {
		return err
	}

	return nil

}

// Funcion para imprimir el contenido del fileblock
func (fileblock *FILEBLOCK) Print() {
	fmt.Printf("%s", fileblock.B_content)
}
