package Structs

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"time"
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

// Funcion para crear una carpeta o carpetas, dentro del sistema de archivos
func (sb *SUPERBLOCK) Create_Folder(path string, parents_Directories []string, destine_Directory string, create_Parents bool) error {

	// si el arreglo de directorios padre esta vacio, solo se trabaja desde el inode root
	if len(parents_Directories) == 0 {
		return sb.Create_Inode_as_Folder(path, 0, parents_Directories, destine_Directory, create_Parents)
	}

	// iteramos sobre cada inode ya que se necesita buscar el inode padre
	for i := int32(0); i < int32(sb.Sb_inodes_count); i++ {
		err := sb.Create_Inode_as_Folder(path, i, parents_Directories, destine_Directory, create_Parents)
		if err != nil {
			return err
		}
	}

	return nil

}

// Funcion para imprimir los valores del superblock
func (sb *SUPERBLOCK) Print() {

	mount_Time := time.Unix(int64(sb.Sb_mtime), 0)
	unmount_Time := time.Unix(int64(sb.Sb_umtime), 0)

	fmt.Printf("Filesystem type: %d\n", sb.Sb_filesystem_type)
	fmt.Printf("Inodes Count: %d\n", sb.Sb_inodes_count)
	fmt.Printf("Blocks Count: %d\n", sb.Sb_blocks_count)
	fmt.Printf("Free Inodes Count: %d\n", sb.Sb_free_inodes_count)
	fmt.Printf("Free Blocks Count: %d\n", sb.Sb_free_blocks_count)
	fmt.Printf("Mount time: %s\n", mount_Time.Format(time.RFC3339))
	fmt.Printf("UnMount time: %s\n", unmount_Time.Format(time.RFC3339))
	fmt.Printf("Mount Count: %d\n", sb.Sb_mnt_count)
	fmt.Printf("Magic: %d\n", sb.Sb_magic)
	fmt.Printf("Inode Size: %d\n", sb.Sb_inode_size)
	fmt.Printf("Block Size: %d\n", sb.Sb_block_size)
	fmt.Printf("First Inode: %d\n", sb.Sb_first_ino)
	fmt.Printf("First Block: %d\n", sb.Sb_first_blo)
	fmt.Printf("Bitmap Inode Start: %d\n", sb.Sb_bm_inode_start)
	fmt.Printf("Bitmap Block Start: %d\n", sb.Sb_bm_block_start)
	fmt.Printf("Inode Start: %d\n", sb.Sb_inode_start)
	fmt.Printf("Block Start: %d\n", sb.Sb_block_start)
}

// Funcion para imprimir los inodes del superblock
func (sb *SUPERBLOCK) Print_Inodes(path string) error {

	fmt.Println("\nInodos\n----------------")

	for i := int32(0); i < sb.Sb_inodes_count; i++ {
		inode := &INODE{}
		err := inode.Deserialize(path, int64(sb.Sb_inode_start+(i*sb.Sb_inode_size)))
		if err != nil {
			return err
		}
		fmt.Printf("\nInodo %d:\n", i)
		inode.Print()
	}
	return nil

}
