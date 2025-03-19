package Structs

import (
	"encoding/binary"
	"os"
)

// Funcion para crear los BitMaps de inodos y bloques dentro del archivo especificado
func (superblock *SUPERBLOCK) Create_Bit_Maps(path string) error {

	/*
		BITMAP DE INODOS
	*/

	// Escribimos los BitMaps
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	/*
		Se continua con los Bitmap de inodos
		se mueve el puntero del archivo a la posicion especificada
	*/

	_, err = file.Seek(int64(superblock.Sb_bm_inode_start), 0)
	if err != nil {
		return err
	}

	// Se crea un buffer con la cantidad de "n" ceros ('0')
	buffer := make([]byte, superblock.Sb_free_inodes_count)
	for i := range buffer {
		buffer[i] = '0'
	}

	// Una vez escrito el buffer, lo escribimos en el archivo
	err = binary.Write(file, binary.LittleEndian, buffer)
	if err != nil {
		return err
	}

	/*
		BITMAP DE BLOQUES
	*/

	// Movemos el puntero del archivo a la posicion que se especifico
	_, err = file.Seek(int64(superblock.Sb_bm_block_start), 0)
	if err != nil {
		return err
	}

	// Se crea el buffer con los ceros '0'
	buffer = make([]byte, superblock.Sb_free_blocks_count)
	for i := range buffer {
		buffer[i] = 'O'
	}

	// Se escribe el buffer en el archivo binario
	err = binary.Write(file, binary.LittleEndian, buffer)
	if err != nil {
		return err
	}

	return nil

}

// Funcion para actualizar el Bitmap de inodos
func (superblock *SUPERBLOCK) Update_Inode_Bitmap(path string) error {

	//Abrimos el archivo del disco
	file, err := os.OpenFile(path, os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// Mover el puntero del archivo a la posicion del bitmap de Inodos
	_, err = file.Seek(int64(superblock.Sb_bm_inode_start)+int64(superblock.Sb_inodes_count), 0)
	if err != nil {
		return err
	}

	// Escribimos el bit dentro del archivo
	_, err = file.Write([]byte{'1'})
	if err != nil {
		return err
	}

	return nil

}

// Funcion para actualizar el Bitmap de Bloques de carpetas
func (superblock *SUPERBLOCK) Update_Block_Bitmap(path string) error {

	// Abrimos el archivo del disco
	file, err := os.OpenFile(path, os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// Mover el puntero de larchivo a la posicion del bitmap de bloques
	_, err = file.Seek(int64(superblock.Sb_bm_block_start)+int64(superblock.Sb_blocks_count), 0)
	if err != nil {
		return err
	}

	// Escribimos el bit dentro del archivo
	_, err = file.Write([]byte{'X'})
	if err != nil {
		return err
	}

	return nil

}
