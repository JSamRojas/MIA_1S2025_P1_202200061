package reports

import (
	"errors"
	"fmt"
	"os"
	estructuras "server/Structs"
	util "server/Utilities"
	global "server/global"
	"strings"
)

func ReporteFILE(superblock *estructuras.SUPERBLOCK, path string, file_Path string) error {

	// creamos las carpetas padre si no existen
	err := util.Create_Parent_Dir(path)
	if err != nil {
		return err
	}

	txtFile, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("[error comando rep -file] no se pudo crear el archivo .txt: %v", err)
	}
	defer txtFile.Close()

	// obtenemos el contenido del archivo
	content, err := File_content_rep(file_Path)
	if err != nil {
		return fmt.Errorf("[error comando rep -file] no se pudo obtener el contenido del archivo: %v", err)
	}

	// escribimos el contenido en el archivo
	_, err = txtFile.WriteString(content)
	if err != nil {
		return fmt.Errorf("[error comando rep -file] no se pudo escribir el cotenido en el archivo: %v", err)
	}

	return nil

}

func File_content_rep(path_File string) (string, error) {

	// obtenemos las carpetas padres y el archivo destino
	parent_Directories, destine_Directory := util.Get_Parent_Dirs(path_File)

	// obtenemos el id de la particion que se encuentra montada
	partition_Id := global.Get_id_Session()

	// obtenemos el superbloque y otros parametros de la particion
	superblock_partition, _, partition_path, err := global.Get_superblock_from_part(partition_Id)
	if err != nil {
		return "", fmt.Errorf("[error comando rep -file] no se pudo obtener la particion montada: %v", err)
	}

	// si el arreglo de directorios padre esta vacio, solo se trabaja desde el inode root

	// si el nombre del directorio termina con .txt entonces es un archivo
	if !strings.HasSuffix(destine_Directory, ".txt") {
		return "", errors.New("[error comando rep -file] el nombre del archivo a leer no es valido")
	}

	if len(parent_Directories) == 0 {

		contenido, err := superblock_partition.Found_archive(partition_path, 0, destine_Directory)
		if err != nil {
			return "", err
		}

		return contenido, nil

	}

	/*
		Si el arreglo de directorios padre no esta vacio, iteramos uno por uno,
		si el directorio existe, entonces lo obtenemos y pasamos al siguiente y si no existe, entonces mostramos error
	*/

	Inode_destino := int32(0)

	for i := 0; i < len(parent_Directories); i++ {

		// enviamos a buscar el directorio en la posicion i, para saber si existe
		next_inode, err := superblock_partition.Found_directory(partition_path, Inode_destino, parent_Directories[i])
		// si hay un error, lo devolvemos
		if err != nil {
			return "", err
		}

		// si el valor de la variable es un -1, entonces el directorio no existe, por ende, devolvemos error
		if next_inode == int32(-1) {

			return "", errors.New("[error comando cat] uno de los directorios de la ruta no existe")

		} else {

			/*
				si no devuelve un -1, significa que encontro el inode de la siguiente carpeta, por ende se lo asignamos a Inode_destino y mantenemos la continuidad
			*/
			Inode_destino = next_inode

		}

	}

	// una vez encontrados los directorios y obtenido el inodo donde se encuentra el archivo, procedemos a leer su contenido

	contenido, err := superblock_partition.Found_archive(partition_path, Inode_destino, destine_Directory)
	if err != nil {
		return "", err
	}

	return contenido, nil

}
