package Structs

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"strings"
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
func (sb *SUPERBLOCK) Create_Folder(path string, parents_Directories []string, destine_Directory string, create_Parents bool, content []string, sizeFile int, usrActive int32, grpActive int32) error {

	// si el arreglo de directorios padre esta vacio, solo se trabaja desde el inode root

	if len(parents_Directories) == 0 {

		// enviamos a buscar el directorio a crear, para saber si existe
		next_inode, err := sb.Found_directory(path, 0, destine_Directory)
		// si hay un error, lo devolvemos
		if err != nil {
			return err
		}
		// si el valor de la variable es un -1, significa que el directorio no existe, por ende, hay que crearlo
		if next_inode == int32(-1) {

			// si el nombre del directorio termina con .txt entonces es un archivo
			if strings.HasSuffix(destine_Directory, ".txt") {
				err := sb.Create_Inode_as_File(path, 0, destine_Directory, sizeFile, content, usrActive, grpActive)

				if err != nil {
					return err
				}

			} else {
				err := sb.Create_Inode_as_Folder(path, 0, destine_Directory, usrActive, grpActive)

				if err != nil {
					return err
				}

			}

		}

		return nil

	}

	/*
		Si el arreglo de directorios padre no esta vacio, iteramos uno por uno,
		si el directorio existe, entonces lo obtenemos y pasamos al siguiente y si no existe, lo creamos y luego pasamos al siguiente
	*/

	Inode_destino := int32(0)

	for i := 0; i < len(parents_Directories); i++ {

		// enviamos a buscar el directorio en la posicion i, para saber si existe
		next_inode, err := sb.Found_directory(path, Inode_destino, parents_Directories[i])
		// si hay un error, lo devolvemos
		if err != nil {
			return err
		}

		//fmt.Println("VALOR NEXT_INODE: ", next_inode)

		// si el valor de la variable es un -1, significa que el directorio no existe, por ende, hay que crearlo
		if next_inode == int32(-1) {
			if create_Parents {
				err := sb.Create_Inode_as_Folder(path, Inode_destino, parents_Directories[i], usrActive, grpActive)

				if err != nil {
					return err
				}

				/*
					ya que creamos una carpeta (inode) y es el inmediato siguiente al que teniamos, simplemente sumamos 1 a Inode_destino y mantenemos la continuidad
				*/
				Inode_destino = sb.Sb_inodes_count - 1

			} else {
				return errors.New("[error comando mkdir/mkfile] uno de los directorios de la ruta no existe")
			}
		} else {
			/*
				si no devuelve un -1, significa que encontro el inode de la siguiente carpeta, por ende se lo asignamos a Inode_destino y mantenemos la continuidad
			*/
			Inode_destino = next_inode
		}

	}

	//fmt.Println("VALOR NEXT_INODE: ", Inode_destino)

	/*
		una vez creados los directorios padres, podemos crear el directorio destino
	*/

	// si el nombre del directorio termina con .txt entonces es un archivo
	if strings.HasSuffix(destine_Directory, ".txt") {

		err := sb.Create_Inode_as_File(path, Inode_destino, destine_Directory, sizeFile, content, usrActive, grpActive)

		if err != nil {
			return err
		}

	} else {

		err := sb.Create_Inode_as_Folder(path, Inode_destino, destine_Directory, usrActive, grpActive)

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

// Funcion para imprimir los bloques del superblock
func (sb *SUPERBLOCK) Print_blocks(path string) error {

	// Imprimir bloques
	fmt.Println("\nBloques\n----------------")
	// Iterar sobre cada inodo
	for i := int32(0); i < sb.Sb_inodes_count; i++ {
		inode := &INODE{}
		// Deserializar el inodo
		err := inode.Deserialize(path, int64(sb.Sb_inode_start+(i*sb.Sb_inode_size)))
		if err != nil {
			return err
		}
		// Iterar sobre cada bloque del inodo (apuntadores)
		for _, blockIndex := range inode.I_block {
			// Si el bloque no existe, salir
			if blockIndex == -1 {
				break
			}
			// Si el inodo es de tipo carpeta
			if inode.I_type[0] == '0' {
				block := &FOLDERBLOCK{}
				// Deserializar el bloque
				err := block.Deserialize(path, int64(sb.Sb_block_start+(blockIndex*sb.Sb_block_size))) // 64 porque es el tamaño de un bloque
				if err != nil {
					return err
				}
				// Imprimir el bloque
				fmt.Printf("\nBloque %d:\n", blockIndex)
				block.Print()
				continue

				// Si el inodo es de tipo archivo
			} else if inode.I_type[0] == '1' {
				block := &FILEBLOCK{}
				// Deserializar el bloque
				err := block.Deserialize(path, int64(sb.Sb_block_start+(blockIndex*sb.Sb_block_size))) // 64 porque es el tamaño de un bloque
				if err != nil {
					return err
				}
				// Imprimir el bloque
				fmt.Printf("\nBloque %d:\n", blockIndex)
				block.Print()
				continue
			}

		}
	}

	return nil
}

// Funcion para recorrer los bloques de un inode en busca de una carpeta, si la encuentra se devuelve el numero, si no, devuelve un -1
func (sb *SUPERBLOCK) Found_directory(path string, inode_Index int32, directory string) (int32, error) {

	// creamos una instancia de inode
	inode := INODE{}

	err := inode.Deserialize(path, int64(sb.Sb_inode_start+(inode_Index*sb.Sb_inode_size)))
	if err != nil {
		return int32(-1), err
	}

	if inode.I_type[0] == '1' {
		return int32(-1), errors.New("[error] uno de los directorios de la ruta era un archivo y no una carpeta")
	}

	// iteramos sobre cada bloque del inodo
	for i := 0; i < len(inode.I_block); i++ {

		// si el bloque no existe, nos salimos
		if inode.I_block[i] == -1 {
			break
		}

		// creamos una instancia de folderblock
		folder := &FOLDERBLOCK{}

		// deserializamos el bloque
		err := folder.Deserialize(path, int64(sb.Sb_block_start+(inode.I_block[i]*sb.Sb_block_size)))
		if err != nil {
			return int32(-1), err
		}

		for index_content := 2; index_content < len(folder.B_content); index_content++ {

			// obtener el contenido del bloque
			content := folder.B_content[index_content]

			// si el contenido esta vacio, retornamos un -1 para indicar que no fue encontrado el directorio
			if content.B_inodo == -1 {
				break
			}

			// convertimos el nombre del bloque y eliminamos caracteres nulos
			block_Name := strings.Trim(string(content.B_name[:]), "\x00")
			// convertimos el nombre del directorio y eliminamos caracteres nulos
			dir_Name := strings.Trim(directory, "\x00")

			// si el nombre del contenido coincide con el nombre de la carpeta, entonces devolvemos el numero de inodo al que apunta
			if strings.EqualFold(block_Name, dir_Name) {
				return int32(content.B_inodo), nil
			}

		}

	}
	return int32(-1), nil
}

// Funcion para recorrer los bloques de un inodo, buscando un archivo y devolver su contenido
func (sb *SUPERBLOCK) Found_archive(path string, inode_Index int32, dest_Dirs string) (string, error) {

	// creamos una instancia de inode
	inode := INODE{}

	err := inode.Deserialize(path, int64(sb.Sb_inode_start+(inode_Index*sb.Sb_inode_size)))
	if err != nil {
		return "", err
	}

	// variable para almacenar el contenido del archivo
	content := ""
	inode_Archive := -1

	// iteramos sobre cada bloque del inodo
	for i := 0; i < len(inode.I_block); i++ {

		// si el bloque no existe, nos salimos
		if inode.I_block[i] == -1 {
			return "", errors.New("[error comando cat] el archivo que se busca, no existe")
		}

		// creamos una instancia de folderblock
		folderblock := &FOLDERBLOCK{}

		// deserializamos el bloque
		err := folderblock.Deserialize(path, int64(sb.Sb_block_start+(inode.I_block[i]*sb.Sb_block_size)))
		if err != nil {
			return "", err
		}

		for index_content := 2; index_content < len(folderblock.B_content); index_content++ {

			// obtener el contenido del bloque
			content := folderblock.B_content[index_content]

			// si el contenido esta vacio, retornamos un -1 para indicar que no fue encontrado el directorio
			if content.B_inodo == -1 {
				return "", errors.New("[error comando cat] el archivo que se busca, no existe")
			}

			// convertimos el nombre del bloque y eliminamos caracteres nulos
			block_Name := strings.Trim(string(content.B_name[:]), "\x00")
			// convertimos el nombre del directorio y eliminamos caracteres nulos
			dir_Name := strings.Trim(dest_Dirs, "\x00")

			// si el nombre del contenido coincide con el nombre de la carpeta, entonces devolvemos el numero de inodo al que apunta
			if strings.EqualFold(block_Name, dir_Name) {
				inode_Archive = int(content.B_inodo)
				break
			}

		}

		if inode_Archive != -1 {
			break
		}

	}

	err = inode.Deserialize(path, int64(sb.Sb_inode_start+(int32(inode_Archive)*sb.Sb_inode_size)))
	if err != nil {
		return "", nil
	}

	// iteramos sobre los bloques del inodo
	for i := 0; i < len(inode.I_block); i++ {

		if inode.I_block[i] == -1 {
			return content, nil
		}

		// creamos una instancia de fileblock
		fileblock := &FILEBLOCK{}

		// deserializamos el bloque
		err = fileblock.Deserialize(path, int64(sb.Sb_block_start+(inode.I_block[i]*sb.Sb_block_size)))
		if err != nil {
			return "", err
		}

		// agregamos el contenido del bloque a la variable que lo acumula
		content += strings.Trim(string(fileblock.B_content[:]), "\x00")

	}

	return "", errors.New("[error comando cat] el archivo que se busca, no existe")

}
