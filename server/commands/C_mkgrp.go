package commands

import (
	"errors"
	"fmt"
	"regexp"
	estructuras "server/Structs"
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
	partition_superblock, _, partition_path, err := global.Get_superblock_from_part(partition_id)
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

				// verificar que el primer inode este en 1
				if inode.I_block[0] == 1 {

					fileblock := &estructuras.FILEBLOCK{}

					err = fileblock.Deserialize(partition_path, int64(partition_superblock.Sb_block_start+(inode.I_block[0]*partition_superblock.Sb_block_size)))
					if err != nil {
						return "ERROR COMANDO MKGRP: no se pudo obtener el archivo users.txt", fmt.Errorf("[error comando mkgrp] no se pudo obtener el archivo de users.txt: %v", err)
					}

					// obtenemos el contenido del archivo users.txt
					contenido := strings.Trim(string(fileblock.B_content[:]), "\x00")

					// reemplazamos \r\n con saltos \n normales
					contenido = strings.ReplaceAll(contenido, "\r\n", "\n")

					// dividir en lineas para obtener los usuarios o grupos
					lines := strings.Split(contenido, "\n")

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
					contenido += NewGroup_line

					// escribimos el contenido actualizado en el bloque
					copy(fileblock.B_content[:], contenido)

					// Guardamos los cambios en el archivo

					err = fileblock.Serialize(partition_path, int64(partition_superblock.Sb_block_start+(inode.I_block[0]*partition_superblock.Sb_block_size)))
					if err != nil {
						return "ERROR COMANDO MKGRP: no se pudo escribir el nuevo archivo de users.txt", fmt.Errorf("[error comando mkgrp] no se pudo escribir el nuevo archivo users.txt: %v", err)
					}
					fmt.Println("-----------CREAR GRUPO--------------")
					fileblock.Print()
					fmt.Println("------------------------------------")
					return "COMANDO MKGRP: grupo " + mkgrp.Name + " creado correctamente", nil
				}
			}
		}
	}
	return "", nil
}
