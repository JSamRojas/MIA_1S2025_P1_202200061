package commands

import (
	"errors"
	"fmt"
	"regexp"
	util "server/Utilities"
	global "server/global"
	"strings"
)

type CAT struct {
	file_Path string
}

func Cat_Command(tokens []string) (*CAT, string, error) {

	cat := &CAT{}

	atributos := strings.Join(tokens, " ")

	lexic := regexp.MustCompile(`(?i)-file[1-9][0-9]*="[^"]+"|(?i)-file[1-9][0-9]*=[^\s]+`)

	found := lexic.FindAllString(atributos, -1)

	// Variable para almacenar el contenido de todos los archivos o el archivo a leer
	var Content strings.Builder

	for _, fun := range found {

		parametro := strings.SplitN(fun, "=", 2)
		if len(parametro) != 2 {
			return nil, "ERROR COMANDO CAT: formato de parametros invalido", fmt.Errorf("formato de parametro invalido; %s", fun)
		}

		key, value := strings.ToLower(parametro[0]), parametro[1]

		if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
			value = strings.Trim(value, "\"")
		}

		if strings.HasPrefix(key, "-file") {

			if value == "" {
				return nil, "ERROR COMANDO CAT: la ruta del archivo no puede estar vacia", errors.New("el nombre del archivo no puede estar vacio")
			}
			cat.file_Path = value
		} else {
			return nil, "ERROR COMANDO CAT: parametro invalido", fmt.Errorf("parametro no reconocido: %s", key)
		}

		if cat.file_Path == "" {
			return nil, "ERROR COMANDO CAT: el nombre del archivo es obligatorio", errors.New("el nombre del archivo es obligatorio")
		}

		// Leemos el contenido del archivo
		contenido, err := execute_Cat(cat)
		if err != nil {
			return nil, contenido, err
		}

		// Se concatena el contenido del archivo
		Content.WriteString(contenido + "\n")

	}

	return cat, "[comando cat] lectura realizada con exito\nCONTENIDO: \n" + Content.String(), nil

}

func execute_Cat(cat *CAT) (string, error) {

	/*
		Hay que leer el archivo que esta dentro de la ruta especificada
		El orden es el siguiente: Inode -> folderblock -> contenido
	*/

	// la ruta del archivo esta dentro del parametro file_Path de cat
	parent_Dirs, dest_Dirs := util.Get_Parent_Dirs(cat.file_Path)
	// Arreglo de carpetas o directorios padres
	//fmt.Println("\nDirectorios padres del archivo: ", parent_Dirs)
	// Nombre del archivo destino
	//fmt.Println("Directorio destino: ", dest_Dirs)

	// Obtener el Id de la particion donde esta logueado
	partition_Id := global.Get_id_Session()

	// obtenemos partes escenciales de la particion
	partition_superblock, _, partition_path, err := global.Get_superblock_from_part(partition_Id)
	if err != nil {
		return "", fmt.Errorf("[error comando cat] no se pudo obtener la particion montada: %v", err)
	}

	// si el arreglo de directorios padre esta vacio, solo se trabaja desde el inode root

	// si el nombre del directorio termina con .txt entonces es un archivo
	if !strings.HasSuffix(dest_Dirs, ".txt") {
		return "", errors.New("[error comando cat] el nombre del archivo a leer no es valido")
	}

	if len(parent_Dirs) == 0 {

		contenido, err := partition_superblock.Found_archive(partition_path, 0, dest_Dirs)
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

	for i := 0; i < len(parent_Dirs); i++ {

		// enviamos a buscar el directorio en la posicion i, para saber si existe
		next_inode, err := partition_superblock.Found_directory(partition_path, Inode_destino, parent_Dirs[i])
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

	contenido, err := partition_superblock.Found_archive(partition_path, Inode_destino, dest_Dirs)
	if err != nil {
		return "", err
	}

	return contenido, nil

}
