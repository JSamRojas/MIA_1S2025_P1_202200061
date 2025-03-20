package Structs

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"time"
)

type INODE struct {
	I_uid   int32
	I_gid   int32
	I_size  int32
	I_atime float32
	I_ctime float32
	I_mtime float32
	I_block [15]int32
	I_type  [1]byte
	I_perm  [3]byte
}

// Funcion para escribir la estructura del Inode dentro de un archivo binario (o Disco)
func (inode *INODE) Serialize(path string, offset int64) error {

	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// Mover el puntero del archivo al offset especificado
	_, err = file.Seek(offset, 0)
	if err != nil {
		return err
	}

	// Se escribe la estructura Inode dentro del archivo binario
	err = binary.Write(file, binary.LittleEndian, inode)
	if err != nil {
		return err
	}

	return nil

}

// Funcion para leer el contenido de la estructura Inode desde un archivo binario
func (inode *INODE) Deserialize(path string, offset int64) error {

	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	// Movemos el puntero del archivo a la posicion del offset
	_, err = file.Seek(offset, 0)
	if err != nil {
		return err
	}

	// Obtener el size de la estructura Inode
	inode_Size := binary.Size(inode)
	if inode_Size <= 0 {
		return fmt.Errorf("size del inode invalido: %d", inode_Size)
	}

	// Leemos solo la cantidad de bytes que corresponden al size de la estructura Inode
	buffer := make([]byte, inode_Size)
	_, err = file.Read(buffer)
	if err != nil {
		return err
	}

	// Deserializamos los bytes leidos y los guardamos en la estructura inode
	reader := bytes.NewReader(buffer)
	err = binary.Read(reader, binary.LittleEndian, inode)
	if err != nil {
		return err
	}

	return nil

}

// Funcion para imprimir los atributos del Inode
func (inode *INODE) Print() {

	atime := time.Unix(int64(inode.I_atime), 0)
	ctime := time.Unix(int64(inode.I_ctime), 0)
	mtime := time.Unix(int64(inode.I_mtime), 0)

	fmt.Printf("I_uid: %d\n", inode.I_uid)
	fmt.Printf("I_gid: %d\n", inode.I_gid)
	fmt.Printf("I_size: %d\n", inode.I_size)
	fmt.Printf("I_atime: %s\n", atime.Format(time.RFC3339))
	fmt.Printf("I_ctime: %s\n", ctime.Format(time.RFC3339))
	fmt.Printf("I_mtime: %s\n", mtime.Format(time.RFC3339))
	fmt.Printf("I_block: %v\n", inode.I_block)
	fmt.Printf("I_type: %s\n", string(inode.I_type[:]))
	fmt.Printf("I_perm: %s\n", string(inode.I_perm[:]))

}
