package Structs

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
)

type FOLDERBLOCK struct {
	B_content [4]FOLDERCONTENT // 4 * 16 = 64 bytes
}

type FOLDERCONTENT struct {
	B_name  [12]byte
	B_inodo int32
}

// Funcion para escribir la estructura de FolderBlock dentro del archivo binario (o Disco)
func (folderblock *FOLDERBLOCK) Serialize(path string, offset int64) error {

	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// Mover el puntero del archivo a la posicion del offset
	_, err = file.Seek(offset, 0)
	if err != nil {
		return err
	}

	// Serializar el folder block directamente en el disco
	err = binary.Write(file, binary.LittleEndian, folderblock)
	if err != nil {
		return err
	}

	return nil
}

// Funcion para leer la estructura del folderblock desde un archivo binario
func (folderblock *FOLDERBLOCK) Deserialize(path string, offset int64) error {

	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	// Mover el puntero del archivo a la posicion del offset
	_, err = file.Seek(offset, 0)
	if err != nil {
		return err
	}

	// Obtener el size del fileblock
	fb_Size := binary.Size(folderblock)
	if fb_Size <= 0 {
		return fmt.Errorf("el size del folderblock es invalido: %d", fb_Size)
	}

	// Leemos solamente la cantidad de bytes que corresponden al size del folderblock
	buffer := make([]byte, fb_Size)
	_, err = file.Read(buffer)
	if err != nil {
		return err
	}

	// Deserealizar los bytes leidos en los campos del folderblock
	reader := bytes.NewReader(buffer)
	err = binary.Read(reader, binary.LittleEndian, folderblock)
	if err != nil {
		return err
	}

	return nil

}

// Funcion para imprimir los atributos del folder block
func (folderblock *FOLDERBLOCK) Print() {
	for i, content := range folderblock.B_content {
		name := string(content.B_name[:])
		fmt.Printf("Contenido %d: \n", i+1)
		fmt.Printf("	B_name: %s\n", name)
		fmt.Printf("	B_inodo: %d\n", content.B_inodo)
	}
}
