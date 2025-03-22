package commands

import (
	"errors"
	"fmt"
	"regexp"
	estructuras "server/Structs"
	util "server/Utilities"
	global "server/global"
	"strconv"
	"strings"
)

type MKGRP struct {
	Name string
}

func Mkgrp_Command(tokens []string) (*MKGRP, string, error) {

	mkgrp := &MKGRP{}

	// Unimos los tokens en una sola cadena y luego los dividimos por espacios
	atributos := strings.Join(tokens, " ")
	// Expresion regular para encontrar los parametros del comando
	lexic := regexp.MustCompile(`(?i)-name="[^"]+"|(?i)-name=[^\s]+`)
	// Encontramos todas las coincidencias de la expresion regular en la cadena del comando
	found := lexic.FindAllString(atributos, -1)

	for _, fun := range found {

		parametro := strings.SplitN(fun, "=", 2)
		if len(parametro) != 2 {
			return nil, "ERROR COMANDO MKGRP: parametro invalido", fmt.Errorf("[error comando mkgrp] parametro invalido: %s", fun)
		}

		key, value := strings.ToLower(parametro[0]), parametro[1]

		if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
			value = strings.Trim(value, "\"")
		}

		switch key {

		case "-name":

			if value == "" {
				return nil, "ERROR COMANDO MKGRP: el nombre del grupo no puede estar vacio", errors.New("[error comando mkgrp] el nombre no puede estar vacio")
			}
			mkgrp.Name = value

		default:
			return nil, "ERROR COMANDO MKGRP: parametro desconocido", fmt.Errorf("[error comando mkgrp] parametro desconocido: %s", key)
		}
	}

	if mkgrp.Name == "" {
		return nil, "ERROR COMANDO MKGRP: el nombre no puede estar vacio", errors.New("[error comando mkgrp] el nombre no puede estar vacio")
	}

	msg, err := Create_Mkgrp(mkgrp)
	if err != nil {
		return nil, msg, err
	}

	return mkgrp, msg, nil

}

func Create_Mkgrp(mkgrp *MKGRP) (string, error) {

	partition_id := global.Get_id_Session()

	user := global.Get_user_Active(partition_id)

	// verificar que el usuario sea el root
	if user != "root" {
		return "ERROR COMANDO MKGRP: solamente el usuario root puede crear grupos", errors.New("[error comando mkgrp] solamente el usuario root puede crear grupos")
	}

	// obtenemos la particion y el superblock de la misma donde se creara el grupo
	partition_superblock, mounted_partition, partition_path, err := global.Get_superblock_from_part(partition_id)
	if err != nil {
		return "ERROR COMANDO MKGRP: no se pudo obtener la particion para crear el grupo", fmt.Errorf("[error comando mkgrp] no se pudo obtener la particion para crear el grupo: %v", err)
	}

	inode := &estructuras.INODE{}

	// Deserializamos el Inode root
	err = inode.Deserialize(partition_path, int64(partition_superblock.Sb_inode_start+(0*partition_superblock.Sb_inode_size)))
	if err != nil {
		return "ERROR COMANDO MKGRP: no se pudo acceder al inode root", fmt.Errorf("[error comando mkgrp] no se pudo aceder al inode root: %v", err)
	}

	// verificar que el primer inode este en 0
	if inode.I_block[0] == 0 {

		folderblock := &estructuras.FOLDERBLOCK{}

		err = folderblock.Deserialize(partition_path, int64(partition_superblock.Sb_block_start+(inode.I_block[0]*partition_superblock.Sb_block_size)))
		if err != nil {
			return "ERROR COMANDO MKGRP: no se pudo obtener el primer bloque", fmt.Errorf("[error comando mkgrp] no se pudo obtener el primer bloque: %v", err)
		}

		// recorremos el contenido del bloque 0
		for _, contenido := range folderblock.B_content {
			name := strings.Trim(string(contenido.B_name[:]), "\x00")
			apuntador := contenido.B_inodo
			if name == "users.txt" {

				// nos movemos al inode del archivo users.txt
				err = inode.Deserialize(partition_path, int64(partition_superblock.Sb_inode_start+(apuntador*partition_superblock.Sb_inode_size)))
				if err != nil {
					return "ERROR COMANDO MKGRP: no se pudo obtener el inode de users.txt", fmt.Errorf("[error comando mkgrp] no se pudo obtener el inode de users.txt: %v", err)
				}

				// variable para almacenar el contenido del archivo
				content_users := ""

				// verificar que el primer inode este en 1
				if inode.I_block[0] == 1 {

					// for para recorrer todos los bloques que contiene el archivo
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
						fileblock := &estructuras.FILEBLOCK{}

						err = fileblock.Deserialize(partition_path, int64(partition_superblock.Sb_block_start+(inode.I_block[0]*partition_superblock.Sb_block_size)))
						if err != nil {
							return "ERROR COMANDO MKGRP: no se pudo obtener el archivo users.txt", fmt.Errorf("[error comando mkgrp] no se pudo obtener el archivo de users.txt: %v", err)
						}


						// obtenemos el contenido del archivo users.txt
						contenido := strings.Trim(string(fileblock.B_content[:]), "\x00")
					*/

					// reemplazamos \r\n con saltos \n normales
					content_users = strings.ReplaceAll(content_users, "\r\n", "\n")

					// dividir en lineas para obtener los usuarios o grupos
					lines := strings.Split(content_users, "\n")

					// vairable para almacenar el ultimo numero de grupo
					last_Group := 0

					// recorremos cada linea del archivo
					for _, line := range lines {
						if strings.TrimSpace(line) == "" {
							continue
						}

						values := strings.Split(line, ",")

						// verificamos que sea un grupo y obtenemos el numero del mismo
						if len(values) >= 3 && values[1] == "G" {
							G_number, err := strconv.Atoi(values[0])
							if err == nil && G_number > last_Group {
								last_Group = G_number
							}

							if G_number == 0 {
								last_Group += 1
							}
						}
					}

					// incrementamos el maximo numero de grupo
					last_Group += 1

					// creamos la nueva linea del archivo
					NewGroup_line := fmt.Sprintf("%d,G,%s\n", last_Group, mkgrp.Name)

					// agregamos la nueva linea
					content_users += NewGroup_line

					// obtenemos el contenido partido en chunks (64 bytes)
					new_Content := util.Split_into_Chunks(content_users)

					/*
						pos := 0
						for _, content := range new_Content {
							pos++
							fmt.Println("posicion " + string(rune(pos)) + " " + content)
						}
					*/

					//fmt.Println("-----------CREAR GRUPO--------------")

					// ciclo para recorrer el arreglo de contenidos
					for i := 0; i < len(new_Content); i++ {

						/*
							validamos de que el bloque ya estuviera contemplado
							si no, le agregamos el numero que le corresponde y
							actualizamos el bitmap de bloquues
						*/
						if inode.I_block[i] == -1 {
							// le asignamos su numero de bloque segun le toque
							inode.I_block[i] = partition_superblock.Sb_blocks_count

							// actualizamos el bitmap de bloques
							err = partition_superblock.Update_Block_Bitmap(partition_path)
							if err != nil {
								return "ERROR COMANDO MKGRP: no se pudo escribir el nuevo archivo de users.txt", fmt.Errorf("[error comando mkgrp] no se pudo escribir el nuevo archivo users.txt: %v", err)
							}

							// actualizamos el superbloque
							partition_superblock.Sb_blocks_count++
							partition_superblock.Sb_free_blocks_count--
							partition_superblock.Sb_first_blo += partition_superblock.Sb_block_size
						}

						// creamos el bloque del archivo
						fileblock := &estructuras.FILEBLOCK{
							B_content: [64]byte{},
						}

						// copiamos el texto que le corresponde
						copy(fileblock.B_content[:], new_Content[i])

						//serializamos el bloque
						err = fileblock.Serialize(partition_path, int64(partition_superblock.Sb_block_start+(inode.I_block[i]*partition_superblock.Sb_block_size)))
						if err != nil {
							return "ERROR COMANDO MKGRP: no se pudo escribir el nuevo archivo de users.txt", fmt.Errorf("[error comando mkgrp] no se pudo escribir el nuevo archivo users.txt: %v", err)
						}

						//fileblock.Print()

					}

					// serializamos el inode users.txt por si ocupo otro bloque
					err = inode.Serialize(partition_path, int64(partition_superblock.Sb_inode_start+(apuntador*partition_superblock.Sb_inode_size)))
					if err != nil {
						return "ERROR COMANDO MKGRP: no se pudo escribir los cambios en la particion", fmt.Errorf("[error comando mkgrp] no se pudo escribir los cambios en la particion: %v", err)
					}

					// serializamos el superbloque por si el archivo users.txt ocupo otro bloque
					err = partition_superblock.Serialize(partition_path, int64(mounted_partition.Partition_start))
					if err != nil {
						return "ERROR COMANDO MKGRP: no se pudo escribir los cambios en la particion", fmt.Errorf("[error comando mkgrp] no se pudo escribir los cambios en la particion: %v", err)
					}

					//fmt.Println("------------------------------------")
					return "COMANDO MKGRP: grupo " + mkgrp.Name + " creado correctamente", nil
				}
			}
		}
	}
	return "", nil
}
