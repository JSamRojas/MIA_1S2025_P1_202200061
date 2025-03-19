package Structs

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
)

type SUPERBLOCK struct {
	Sb_filesystem_type   int32
	Sb_inodes_count      int32
	Sb_blocks_count      int32
	Sb_free_blocks_count int32
	Sb_free_inodes_count int32
	Sb_mtime             float32
	Sb_umtime            float32
	Sb_mnt_count         int32
	Sb_magic             int32
	Sb_inode_size        int32
	Sb_block_size        int32
	Sb_first_ino         int32
	Sb_first_blo         int32
	Sb_bm_inode_start    int32
	Sb_bm_block_start    int32
	Sb_inode_start       int32
	Sb_block_start       int32
}

// Funcion para escribir la estructura del superblock dentro de un archivo binario
func (sb *SUPERBLOCK) Serialize(path string, offset int64) error {

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

	// Serializamos la estructura del superblock dentro del archivo
	err = binary.Write(file, binary.LittleEndian, sb)
	if err != nil {
		return err
	}

	return nil
}

// Funcion para leer la estructura superblock desde un archivo binario
func (sb *SUPERBLOCK) Deserialize(path string, offset int64) error {

	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Seek(offset, 0)
	if err != nil {
		return err
	}

	sb_Size := binary.Size(sb)
	if sb_Size <= 0 {
		return fmt.Errorf("size del superblock invalido: %d", sb_Size)
	}

	buffer := make([]byte, sb_Size)
	_, err = file.Read(buffer)
	if err != nil {
		return err
	}

	reader := bytes.NewReader(buffer)
	err = binary.Read(reader, binary.LittleEndian, sb)
	if err != nil {
		return err
	}

	return nil

}
