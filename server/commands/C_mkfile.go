package commands

import (
	"errors"
	"fmt"
	"io/ioutil"
	"regexp"
	estructuras "server/Structs"
	util "server/Utilities"
	global "server/global"
	"strconv"
	"strings"
)

type MKFILE struct {
	Path string
	R    bool
	Size int
	Cont string
}

func Mkfile_Command(tokens []string) (*MKFILE, string, error) {

	// creamos una nueva instancia de mkfile
	mkfile := &MKFILE{
		R: false,
	}

	// unimos todos los tokens en una sola cadena y luego se divide por espacios
	atributos := strings.Join(tokens, " ")
	// expresion regular para encontrar los parametros del comando
	lexic := regexp.MustCompile(`(?i)-path="[^"]+"|(?i)-path=[^\s]+|(?i)-r|-size=\d+|(?i)-cont="[^"]+"|(?i)-cont=[^\s]+`)
	// encuentra todas las coincidencias de la expresion regular en la cadena de argumentos
	found := lexic.FindAllString(atributos, -1)

	// verificar que todos los tokens fueron reconocidos por la expresion
	for _, fun := range found {

		if strings.EqualFold(fun, "-r") { // Verificar si es el parámetro -r
			mkfile.R = true
			continue
		}

		parametro := strings.SplitN(fun, "=", 2)
		if len(parametro) != 2 {
			return nil, "", fmt.Errorf("[error comando mkfile] formato de parametro invalido: %s", fun)
		}

		key, value := strings.ToLower(parametro[0]), parametro[1]

		if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
			value = strings.Trim(value, "\"")
		}

		switch key {

		case "-path":

			if value == "" {
				return nil, "", errors.New("[error comando mkfile] el path no puede estar vacio")
			}
			mkfile.Path = value

		case "-size":

			size, err := strconv.Atoi(value)
			if err != nil || size <= 0 {
				return nil, "", errors.New("[error comando mkfile] el parametro de size no puede ser 0 o menor")
			}
			mkfile.Size = size

		case "-cont":

			//fmt.Println("RUTA INGRESADA: ", value)

			if value == "" {
				mkfile.Cont = ""
			}
			mkfile.Cont = value

		default:
			return nil, "", fmt.Errorf("[error comando mkfile] parametro desconocido: %s", key)
		}

	}

	if mkfile.Path == "" {
		return nil, "", errors.New("[error comando mkfile] el path no puede estar vacio")
	}

	if mkfile.Cont == "" && mkfile.Size == 0 {
		return nil, "", errors.New("[error comando mkfile] los parametros cont y size, no pueden estar vacios al mismo tiempo")
	}

	// creamos el archivo con los parametros indicados
	msg, err := Make_file(mkfile)
	if err != nil {
		return nil, msg, err
	}

	return mkfile, fmt.Sprintf("COMANDO MKFILE: archivo creado con exito en la ruta: %s", mkfile.Path), nil

}

func Make_file(mkfile *MKFILE) (string, error) {

	// obtenemos el id de la particion donde estan logueadas
	partition_Id := global.Get_id_Session()
	if partition_Id == "" {
		return "", errors.New("[error comando mkfile] no hay ninguna sesion activa")
	}

	// obtenemos todo lo necesario sobre la particion para trabajar
	partition_superblock, mounted_partition, partition_path, err := global.Get_superblock_from_part(partition_Id)
	if err != nil {
		return "", fmt.Errorf("[error comando mkfile] no se pudo obtener el superblock de la particion: %v", err)
	}

	if mkfile.Cont == "" {
		mkfile.Cont = Get_random_Content(mkfile.Size)
	} else {

		// leemos el archivo directo de la computadora
		content, err := ioutil.ReadFile(mkfile.Cont)
		if err != nil {
			return "", fmt.Errorf("[error comando mkfile] no se pudo leer el archivo de la computadora: %v", err)
		}

		// asigamos el contenido del archivo al parametro correspondiente
		mkfile.Cont = string(content)

		// asignamos el size del archivo recien obtenido
		mkfile.Size = len(mkfile.Cont)

	}

	fmt.Println("\nContenido del archivo: ", mkfile.Cont)

	// creamos el archivo correspondiente
	msg := ""
	msg, err = Create_File(mkfile, partition_superblock, partition_path, mounted_partition)
	if err != nil {
		return msg, fmt.Errorf("[error comando mkfile] no se pudo crear el archivo: %v", err)
	}

	return "", nil

}

func Create_File(mkfile *MKFILE, partition_superblock *estructuras.SUPERBLOCK, partition_path string, mounted_partition *estructuras.PARTITION) (string, error) {

	fmt.Println("\nCreando archivo: ", mkfile.Path)

	parentDirs, destDir := util.Get_Parent_Dirs(mkfile.Path)
	fmt.Println("\nDirectorios padres:", parentDirs)
	fmt.Println("Directorio destino:", destDir)

	// obtener el contenido por partes para los bloque de 64 bytes
	content_parts := util.Split_into_Chunks(mkfile.Cont)
	fmt.Println("\nChunks del contenido:", content_parts)

	usrActive, grpActive, err := global.Get_userid_groupid()
	if err != nil {
		return "", err
	}

	// crear el directorio segun el path definido
	err = partition_superblock.Create_Folder(partition_path, parentDirs, destDir, mkfile.R, content_parts, mkfile.Size, usrActive, grpActive)
	if err != nil {
		return "", err
	}

	//partition_superblock.Print_Inodes(partition_path)
	partition_superblock.Print_blocks(partition_path)

	err = partition_superblock.Serialize(partition_path, int64(mounted_partition.Partition_start))
	if err != nil {
		return "", fmt.Errorf("[error comando mkfile] no se pudo serializar el superblock: %v", err)
	}

	return "", nil

}

// Funcion para obtener el contenido random del archivo, segun el size ingresado
func Get_random_Content(size int) string {
	result := make([]byte, size)
	for i := 0; i < size; i++ {
		result[i] = '0' + byte(i%10)
	}
	return string(result)
}
