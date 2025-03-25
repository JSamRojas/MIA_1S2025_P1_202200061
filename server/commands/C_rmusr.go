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

type RMUSR struct {
	User string
}

func Rmusr_Command(tokens []string) (*RMUSR, string, error) {

	rmusr := &RMUSR{}

	// Unimos los tokens en una sola cadena y luego se dividen por espacios
	atributos := strings.Join(tokens, " ")
	// expresion regular para encontrar los parametros del comando
	lexic := regexp.MustCompile(`(?i)-user="[^"]+"|(?i)-user=[^\s]+`)
	// encontramos todas las coincidencias de la expresion regular
	found := lexic.FindAllString(atributos, -1)

	for _, fun := range found {

		parametro := strings.SplitN(fun, "=", 2)
		if len(parametro) != 2 {
			return nil, "", fmt.Errorf("[error comando rmusr] formato de parametro invalido: %s", fun)
		}

		key, value := strings.ToLower(parametro[0]), parametro[1]

		if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
			value = strings.Trim(value, "\"")
		}

		switch key {

		case "-user":
			if value == "" {
				return nil, "", errors.New("[error comando rmusr] el nombre del usuario no puede estar vacio ")
			}
			rmusr.User = value

		default:
			return nil, "", errors.New("[error comando rmusr] parametro desconocido")
		}

	}

	if rmusr.User == "" {
		return nil, "", errors.New("[error comando rmusr] el nombre del usuario no puede estar vacio ")
	}

	msg, err := Remove_user(rmusr)
	if err != nil {
		return nil, msg, err
	}

	return rmusr, msg, nil

}

func Remove_user(rmusr *RMUSR) (string, error) {

	partition_id := global.Get_id_Session()

	user := global.Get_user_Active(partition_id)

	// verificar que el usuario sea el root
	if user != "root" {
		return "", errors.New("[error comando rmusr] solamente el usuario root puede remover usuarios")
	}

	// obtener multiples parametros de la particion logueada
	superblock_partition, _, partition_path, err := global.Get_superblock_from_part(partition_id)
	if err != nil {
		return "", fmt.Errorf("[error comando rmusr] error al obtener la partition montada: %v", err)
	}

	inode := &estructuras.INODE{}

	// Deserializamos el root inode
	err = inode.Deserialize(partition_path, int64(superblock_partition.Sb_inode_start+(0*superblock_partition.Sb_inode_size)))
	if err != nil {
		return "", fmt.Errorf("[error comando rmusr] error al obtener el inode root: %v", err)
	}

	// verificar que el primer inode este en 0
	if inode.I_block[0] == 0 {

		// moverme al bloque 0
		folderblock := &estructuras.FOLDERBLOCK{}

		err = folderblock.Deserialize(partition_path, int64(superblock_partition.Sb_block_start+(inode.I_block[0]*superblock_partition.Sb_block_size)))
		if err != nil {
			return "", fmt.Errorf("[error comando rmusr] error al obtener el bloque 0: %v", err)
		}

		// recorrer el contenido del bloque 0
		for _, contenido := range folderblock.B_content {

			name := strings.Trim(string(contenido.B_name[:]), "\x00")
			apuntador := contenido.B_inodo
			if name == "users.txt" {

				// moverme al inode que apunta el contenido
				err = inode.Deserialize(partition_path, int64(superblock_partition.Sb_inode_start+(apuntador*superblock_partition.Sb_inode_size)))

				if err != nil {
					return "", fmt.Errorf("[error comando rmusr] no se pudo obtener el inode de users.txt: %v", err)
				}

				// variable para almacenar el contenido del archivo
				content_users := ""

				// verificar que el primer inode sea 1
				if inode.I_block[0] == 1 {

					// ciclo para recorrer todos los bloques que contiene el archivo
					for _, block := range inode.I_block {

						/*
							si el bloque tiene un -1, significa que no esta en uso
							por ende no tiene contenido, salimos del bucle
						*/

						if block == -1 {
							break
						}

						fileblock := &estructuras.FILEBLOCK{}

						err = fileblock.Deserialize(partition_path, int64(superblock_partition.Sb_block_start+(block*superblock_partition.Sb_block_size)))
						if err != nil {
							return "ERROR COMANDO MKGRP: no se pudo obtener el archivo users.txt", fmt.Errorf("[error comando mkgrp] no se pudo obtener el archivo de users.txt: %v", err)
						}

						// obtenemos el contenido de este bloque
						content_users += strings.Trim(string(fileblock.B_content[:]), "\x00")

					}

					// reemplazar los \r\n con \n para asegurar saltos de linea
					content_users = strings.ReplaceAll(content_users, "\r\n", "\n")

					// Dividir las lineas para obtener cada usuario o grupo
					lines := strings.Split(content_users, "\n")

					// booleana para saber si encontro el usuario
					found_user := false

					// arreglo para el nuevo archivo (sin el usuario eliminado)
					var newUsers []string

					// recorrer cada linea del archivo users.txt
					for _, line := range lines {

						if strings.TrimSpace(line) == "" {
							continue
						}

						values := strings.Split(line, ",")

						// verificar si es un usuario y si el nombre del usuario coincide
						if len(values) >= 3 && values[1] == "U" && values[3] == rmusr.User {

							if values[0] == "0" {
								return "[error comando rmusr] no se puede eliminar este usuario porque no existe", nil
							}

							values[0] = "0"
							rm_user := strings.Join(values, ",")
							newUsers = append(newUsers, rm_user)
							found_user = true
							continue
						}

						// agregamos la linea al nuevo contenido si no es el usuario a eliminar
						newUsers = append(newUsers, line)
					}

					// si no se encontro el usuario, se devuelve un error
					if !found_user {
						return "ERROR COMANDO RMUSR: no se encontro el usuario a eliminar", nil
					}

					// unir las lineas nuevas en el contenido
					new_File := strings.Join(newUsers, "\n") + "\n"

					new_Content := util.Split_into_Chunks(new_File)

					// ciclo para recrorrer el arreglo de contenidos
					for i := 0; i < len(new_Content); i++ {

						// creamos un bloque de archivos
						fileblock := &estructuras.FILEBLOCK{
							B_content: [64]byte{},
						}

						// copiamos el texto que le corresponde
						copy(fileblock.B_content[:], new_Content[i])

						// serializamos el bloque
						err = fileblock.Serialize(partition_path, int64(superblock_partition.Sb_block_start+(inode.I_block[i]*superblock_partition.Sb_block_size)))
						if err != nil {
							return "ERROR COMANDO RMUSR: no se pudo escribir el nuevo archivo de users.txt", fmt.Errorf("[error comando mkusr] no se pudo escribir el nuevo archivo users.txt: %v", err)
						}

					}

					return "COMANDO RMUSR: Usuario " + rmusr.User + " eliminado correctamente", nil

				}

			}

		}

	}

	return "", nil

}
