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

type RMGRP struct {
	Name string
}

func Rmgrp_Command(tokens []string) (*RMGRP, string, error) {

	rmgrp := &RMGRP{}

	// Unimos los tokens en una sola cadena y los dividimos por espacios
	atributos := strings.Join(tokens, " ")

	// expresion regular que se usa para encontrar los parametros
	lexic := regexp.MustCompile(`(?i)-name="[^"]+"|(?i)-name=[^\s]+`)

	// encontramos todas las coincidencias de la expresion regular
	found := lexic.FindAllString(atributos, -1)

	for _, fun := range found {

		parametro := strings.SplitN(fun, "=", 2)
		if len(parametro) != 2 {
			return nil, "ERROR COMANDO RMGRP: formato de parametros invalido", fmt.Errorf("[error comando rmgrp] formato de parametro invalido: %s", fun)
		}

		key, value := strings.ToLower(parametro[0]), parametro[1]

		if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
			value = strings.Trim(value, "\"")
		}

		switch key {

		case "-name":

			if value == "" {
				return nil, "ERROR COMANDO RMGRP: el nombre del grupo no puede estar vacio", errors.New("[error comando rmgrp] el nombre del grupo no puede estar vacio")
			}
			rmgrp.Name = value

		default:
			return nil, "ERROR COMANDO RMGRP: parametro invalido", fmt.Errorf("[error comando rmgrp] parametro invalido: %s", key)
		}
	}

	/*

		if rmgrp.Name == "" {
			return nil, "ERROR COMANDO RMGRP: el nombre del grupo no puede estar vacio", errors.New("[error comando rmgrp] el nombre del grupo no puede estar vacio")
		}

	*/

	msg, err := Remove_group(rmgrp)
	if err != nil {
		return nil, msg, err
	}

	return rmgrp, msg, nil

}

func Remove_group(rmgrp *RMGRP) (string, error) {

	partition_Id := global.Get_id_Session()

	user := global.Get_user_Active(partition_Id)

	// verificar que sea el usuario root
	if user != "root" {
		return "ERROR COMANDO RMGRP: solamente el usuario root puede eliminar grupos", errors.New("[error comando rmgrp] solamente el usuario root puede eliminar grupos")
	}

	// obtener la particion con el id donde se realizara la eliminacion
	partition_superblock, _, partition_path, err := global.Get_superblock_from_part(partition_Id)
	if err != nil {
		return "ERROR COMANDO RMGRP: no se pudo obtener la particion", fmt.Errorf("[error comando rmgrp] no se pudo obtener la particion: %v", err)
	}

	inode := &estructuras.INODE{}

	// Deserializamos el inode root
	err = inode.Deserialize(partition_path, int64(partition_superblock.Sb_inode_start+(0*partition_superblock.Sb_inode_size)))
	if err != nil {
		return "ERROR COMANDO RMGRP: no se pudo obtener el inode root", fmt.Errorf("[error comando rmgrp] no se pudo obtener el inode root: %v", err)
	}

	// verificar si el primer inode esta en 0
	if inode.I_block[0] == 0 {

		folderblock := &estructuras.FOLDERBLOCK{}

		err = folderblock.Deserialize(partition_path, int64(partition_superblock.Sb_block_start+(inode.I_block[0]*partition_superblock.Sb_block_size)))
		if err != nil {
			return "ERROR COMANDO RMGRP: no se pudo obtener el bloque 0", fmt.Errorf("[error comando rmgrp] no se pudo obtener el bloque 0: %v", err)
		}

		// recorrer el contenido del bloque 0
		for _, contenido := range folderblock.B_content {
			name := strings.Trim(string(contenido.B_name[:]), "\x00")
			apuntador := contenido.B_inodo
			if name == "users.txt" {

				// moverme al inode que apunta al contenido
				err = inode.Deserialize(partition_path, int64(partition_superblock.Sb_inode_start+(apuntador*partition_superblock.Sb_inode_size)))
				if err != nil {
					return "ERROR COMANDO RMGRP: no se pudo obtener el inode de users.txt", fmt.Errorf("[error comando rmgrp] no se pudo obtener el inode users.txt: %v", err)
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

						err = fileblock.Deserialize(partition_path, int64(partition_superblock.Sb_block_start+(block*partition_superblock.Sb_block_size)))
						if err != nil {
							return "ERROR COMANDO MKGRP: no se pudo obtener el archivo users.txt", fmt.Errorf("[error comando mkgrp] no se pudo obtener el archivo de users.txt: %v", err)
						}

						// obtenemos el contenido de este bloque
						content_users += strings.Trim(string(fileblock.B_content[:]), "\x00")

					}

					/*
						// moverme al bloque 1
						fileblock := &estructuras.FILEBLOCK{}

						err = fileblock.Deserialize(partition_path, int64(partition_superblock.Sb_block_start+(inode.I_block[0]*partition_superblock.Sb_block_size)))
						if err != nil {
							return "ERROR COMANDO RMGRP: no se pudo obtener el fileblock de users.txt", fmt.Errorf("[error comando rmgrp] no se pudo obtener el fileblock de users.txt: %v", err)
						}

						// obtener el contenido del archivo users.txt
						contenido := strings.Trim(string(fileblock.B_content[:]), "\x00")

					*/

					// reemplazar los \r\n con \n para asegurar saltos de linea
					content_users = strings.ReplaceAll(content_users, "\r\n", "\n")

					// Dividir las lineas para obtener cada usuario o grupo
					lines := strings.Split(content_users, "\n")

					// booleana para saber si encontro el grupo
					found_group := false

					// arreglo para el nuevo archivo (sin el grupo eliminado)
					var newUsers []string

					// recorrer cada linea del archivo users.txt
					for _, line := range lines {

						if strings.TrimSpace(line) == "" {
							continue
						}

						values := strings.Split(line, ",")

						// verificar si es un grupo y si el nombre del grupo coincide
						if len(values) >= 3 && values[1] == "G" && values[2] == rmgrp.Name {

							if values[0] == "0" {
								return "[error comand rmgrp] no se puede eliminar un grupo que no existe", nil
							}

							values[0] = "0"
							rm_group := strings.Join(values, ",")
							newUsers = append(newUsers, rm_group)
							found_group = true
							continue
						}

						// agregamos la linea al nuevo contenido si no es el grupo a eliminar
						newUsers = append(newUsers, line)
					}

					// si no se encontro el grupo, se devuelve un error
					if !found_group {
						return "ERROR COMANDO RMGRP: no se encontro el grupo a eliminar", nil
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
						err = fileblock.Serialize(partition_path, int64(partition_superblock.Sb_block_start+(inode.I_block[i]*partition_superblock.Sb_block_size)))
						if err != nil {
							return "ERROR COMANDO RMGRP: no se pudo escribir el nuevo archivo de users.txt", fmt.Errorf("[error comando mkgrp] no se pudo escribir el nuevo archivo users.txt: %v", err)
						}

					}

					//fmt.Println("--------------BORRAR GRUPO-------------")
					//fileblock.Print()
					//fmt.Println("---------------------------------------")

					return "COMANDO RMGRP: grupo " + rmgrp.Name + " eliminado correctamente", nil
				}
			}
		}
	}
	return "", nil
}
