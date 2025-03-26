package commands

import (
	"errors"
	"fmt"
	"regexp"
	estructuras "server/Structs"
	util "server/Utilities"
	global "server/global"
	"strings"
)

type MKDIR struct {
	Path    string
	Parents bool
}

func Mkdir_Command(tokens []string) (*MKDIR, string, error) {

	mkdir := &MKDIR{
		Parents: false,
	}

	// separar todos los tokens en una sola cadena y se dividen por espacios
	atributos := strings.Join(tokens, " ")

	// expresion regular para encontrar los parametros del comando
	lexic := regexp.MustCompile(`(?i)-path="[^"]+"|(?i)-path=[^\s]+|-p`)

	// encontramos todas las coincidencias con las expresion regular
	found := lexic.FindAllString(atributos, -1)

	for _, fun := range found {

		if strings.EqualFold(fun, "-p") { // Verificar si es el parámetro -p
			mkdir.Parents = true
			continue
		}

		parametro := strings.SplitN(fun, "=", 2)
		if len(parametro) != 2 {
			return nil, "", fmt.Errorf("[error comando mkdir] formato de parametro invalido: %s", fun)
		}

		key, value := strings.ToLower(parametro[0]), parametro[1]

		if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
			value = strings.Trim(value, "\"")
		}

		switch key {

		case "-path":

			if value == "" {
				return nil, "", errors.New("[error comando mkdir] el path no puede estar vacio")
			}
			mkdir.Path = value

		default:
			return nil, "", fmt.Errorf("[error comando mkdir] parametro desconocido: %s", key)
		}
	}

	if mkdir.Path == "" {
		return nil, "", errors.New("[error comando mkdir] el path no puede estar vacio")
	}

	msg, err := Execute_mkdir(mkdir)
	if err != nil {
		return nil, msg, err
	}

	return mkdir, "COMANDO MKDIR: directorios realizados con exito, ruta " + mkdir.Path + " creada con exito", nil

}

func Execute_mkdir(mkdir *MKDIR) (string, error) {

	// verificar si existe una sesion activa y obtener el id enlazado
	partition_Id := global.Get_id_Session()
	if partition_Id == "" {
		return "", errors.New("[error comando mkdir] no hay ninguna sesion activa")
	}

	partition_superblock, partition_mounted, partition_path, err := global.Get_superblock_from_part(partition_Id)
	if err != nil {
		return "", fmt.Errorf("[error comando mkdir] no se pudo obtener el superblock de la partiion: %v", err)
	}

	// se crea el directorio
	msg := ""
	msg, err = Create_directory(mkdir.Path, partition_superblock, partition_path, partition_mounted, mkdir.Parents)
	if err != nil {
		return msg, err
	}

	return "", nil

}

func Create_directory(DirectPath string, partition_superblock *estructuras.SUPERBLOCK, partition_path string, partition_mounted *estructuras.PARTITION, create_Parents bool) (string, error) {

	parent_Directories, dest_Directory := util.Get_Parent_Dirs(DirectPath)
	fmt.Println("\nDirectorios padre: ", parent_Directories)
	fmt.Println("Directorio destino: ", dest_Directory)

	usrActive, grpActive, err := global.Get_userid_groupid()
	if err != nil {
		return "", err
	}

	// crear el directorio segun el path definido
	err = partition_superblock.Create_Folder(partition_path, parent_Directories, dest_Directory, create_Parents, usrActive, grpActive)
	if err != nil {
		return "", err
	}

	//partition_superblock.Print_Inodes(partition_path)
	partition_superblock.Print_blocks(partition_path)

	err = partition_superblock.Serialize(partition_path, int64(partition_mounted.Partition_start))
	if err != nil {
		return "", fmt.Errorf("[error comando mkdir] no se pudo serializar el superblock: %v", err)
	}

	return "", nil

}
