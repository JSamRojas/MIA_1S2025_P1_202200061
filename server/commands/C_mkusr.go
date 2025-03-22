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

type MKUSR struct {
	User     string
	Password string
	Group    string
}

func Mkusr_Command(tokens []string) (*MKUSR, string, error) {

	mkusr := &MKUSR{}

	// Unimos los tokens en una sola cadena y luego los dividimos por espacios
	atributos := strings.Join(tokens, " ")
	// expresion regular para encontrar los parametros del comando
	lexic := regexp.MustCompile(`(?i)-user="[^"]+"|(?i)-user=[^\s]+|(?i)-pass="[^"]+"|(?i)-pass=[^\s]+|(?i)-grp="[^"]+"|(?i)-grp=[^\s]+`)
	// encuentra todas las coincidencias de la expresion regular en la cadena
	found := lexic.FindAllString(atributos, -1)

	for _, fun := range found {

		parametro := strings.SplitN(fun, "=", 2)
		if len(parametro) != 2 {
			return nil, "ERROR COMANDO MKUSR: formato de parametro invalido", fmt.Errorf("[error comando mkusr] formato de parametro invalido: %s", fun)
		}

		key, value := strings.ToLower(parametro[0]), parametro[1]

		if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
			value = strings.Trim(value, "\"")
		}

		switch key {

		case "-user":
			if value == "" {
				return nil, "ERROR COMANDO MKUSR: el parametro user no puede estar vacio", errors.New("[error comando mkusr] el parametro user no puede estar vacio")
			}
			mkusr.User = value

		case "-pass":
			if value == "" {
				return nil, "ERROR COMANDO MKUSR: el parametro pass no puede estar vacio", errors.New("[error comando mkusr] el parametro user no puede estar vacio")
			}
			mkusr.Password = value

		case "-grp":
			if value == "" {
				return nil, "ERROR COMANDO MKUSR: el parametro grp no puede estar vacio", errors.New("[error comando mkusr] el parametro grp no puede estar vacio")
			}
			mkusr.Group = value

		default:
			return nil, "ERROR COMANDO MKUSR: parametro invalido", fmt.Errorf("[error comando mkusr] parametro invalido: %s", key)
		}
	}

	if len(mkusr.User) > 10 {
		return nil, "ERROR COMANDO MKUSR: el usuario no debe ser mayor a 10 caracteres", errors.New("[error comando mkusr] el usuario no debe ser mayor a 10 caracteres")
	}

	if len(mkusr.Password) > 10 {
		return nil, "ERROR COMANDO MKUSR: la password no debe ser mayor a 10 caracteres", errors.New("[error comando mkusr] la password no debe ser mayor a 10 caracteres")
	}

	if len(mkusr.Group) > 10 {
		return nil, "ERROR COMANDO MKUSR: el grupo no debe ser mayor a 10 caracteres", errors.New("[error comando mkusr] el grupo no debe ser mayor a 10 caracteres")
	}

	msg, err := Make_usr(mkusr)
	if err != nil {
		return nil, msg, err
	}

	return mkusr, msg, err

}

func Make_usr(mkusr *MKUSR) (string, error) {

	partition_Id := global.Get_id_Session()

	user := global.Get_user_Active(partition_Id)

	// verificar que el usuario sea el root
	if user != "root" {
		return "ERROR COMANDO MKUSR: solamente el usuario root puede crear usuarios", errors.New("[error comando mkusr] solamente el usuario root puede crear usuarios")
	}

	//obtener la particion con el id donde se realiza la creacion del usuario
	partition_superblock, mounted_partition, partition_path, err := global.Get_superblock_from_part(partition_Id)
	if err != nil {
		return "ERROR COMANDO MKUSR: no se pudo obtener la particion", fmt.Errorf("[error comando mkusr] no se pudo obtener la particion: %v", err)
	}

	inode := &estructuras.INODE{}

	// Deserializar el inode root
	err = inode.Deserialize(partition_path, int64(partition_superblock.Sb_inode_start+(0*partition_superblock.Sb_inode_size)))
	if err != nil {
		return "ERROR COMANDO MKUSR: no se pudo obtener el inode root", fmt.Errorf("[error comando mkusr] no se pudo obtener el inode root: %v", err)
	}

	// verificar que el primer inode sea el 0
	if inode.I_block[0] == 0 {

		folderblock := &estructuras.FOLDERBLOCK{}

		err = folderblock.Deserialize(partition_path, int64(partition_superblock.Sb_block_start+(inode.I_block[0]*partition_superblock.Sb_block_size)))
		if err != nil {
			return "ERROR COMANDO MKUSR: no se pudo acceder al bloque 0", fmt.Errorf("[error comando mkusr] no se pudo acceder al bloque 0: %v", err)
		}

		// recorrer los contenidos del bloque 0
		for _, contenido := range folderblock.B_content {
			name := strings.Trim(string(contenido.B_name[:]), "\x00")
			apuntador := contenido.B_inodo
			if name == "users.txt" {

				err = inode.Deserialize(partition_path, int64(partition_superblock.Sb_inode_start+(apuntador*partition_superblock.Sb_inode_size)))
				if err != nil {
					return "ERROR COMANDO MKUSR: no se pudo obtener el inode users.txt", fmt.Errorf("[error comando mkusr] no se pudo obtener el inode users.txt: %v", err)
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
							return "ERROR COMANDO MKUSR: no se pudo obtener el archivo users.txt", fmt.Errorf("[error comando mkusr] no se pudo obtener el archivo de users.txt: %v", err)
						}

						// obtenemos el contenido de este bloque
						content_users += strings.Trim(string(fileblock.B_content[:]), "\x00")

					}

					/*

							fileblock := &estructuras.FILEBLOCK{}

							err = fileblock.Deserialize(partition_path, int64(partition_superblock.Sb_block_start+(inode.I_block[0]*partition_superblock.Sb_block_size)))
							if err != nil {
								return "ERROR COMANDO MKUSR: no se pudo obtener el bloque del archivo users.txt", fmt.Errorf("[error comando mkusr] no se pudo obtener el bloque de archivo users.txt: %v", err)
							}



						// obtener el contenido del archivo users.txt
						contenido := strings.Trim(string(fileblock.B_content[:]), "\x00")

					*/

					// reemplazar \r\n con \n para asegurar saltos de linea uniformes
					content_users = strings.ReplaceAll(content_users, "\r\n", "\n")

					// dividir en lineas para obtener cada usuario o grupo
					lines := strings.Split(content_users, "\n")

					// vairable para obtener el ultimo numero de usuario
					maxUsr := 0

					// variable para saber si el grupo que ingreso, existe
					found_group := false

					// recorrer linea por linea el archivo
					for _, line := range lines {
						if strings.TrimSpace(line) == "" {
							continue
						}

						values := strings.Split(line, ",")

						// verificar si es un usuario y obtener el numero del usuario
						if len(values) >= 5 && values[1] == "U" {
							user_number, err := strconv.Atoi(values[0])
							if err == nil && user_number > maxUsr {
								maxUsr = user_number
							}

							if values[3] == mkusr.User {
								return "ERROR COMANDO MKUSR: ya existe un usuario con este nombre", errors.New("[error comando mkusr] ya existe un usuario con este nombre")
							}

							if user_number == 0 {
								maxUsr += 1
							}

						}

						// verificamos que el grupo si este activo
						if len(values) <= 3 && values[1] == "G" {

							if values[2] == mkusr.Group && values[0] == "0" {
								return "ERROR COMANDO MKUSR: el grupo de este usuario no existe", errors.New("[errores comando mkusr] el grupo de este usuario no existe")
							}

							if values[2] == mkusr.Group {
								found_group = true
							}

						}

					}

					if !found_group {
						return "ERROR COMANDO MKUSR: el grupo de este usuario no existe", errors.New("[errores comando mkusr] el grupo de este usuario no existe")
					}

					// incrementar el numero de usuario para el nuevo usuario
					newUsr_number := maxUsr + 1

					// formatear la nueva linea del usuario
					newUsr_line := fmt.Sprintf("%d,U,%s,%s,%s\n", newUsr_number, mkusr.Group, mkusr.User, mkusr.Password)

					// agregamos el nuevo usuario al contenido
					content_users += newUsr_line

					new_Content := util.Split_into_Chunks(content_users)

					//fmt.Println("-------------CREAR USUARIO--------------")

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
								return "ERROR COMANDO MKUSR: no se pudo escribir el nuevo archivo de users.txt", fmt.Errorf("[error comando mkusr] no se pudo escribir el nuevo archivo users.txt: %v", err)
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
							return "ERROR COMANDO MKUSR: no se pudo escribir el nuevo archivo de users.txt", fmt.Errorf("[error comando mkusr] no se pudo escribir el nuevo archivo users.txt: %v", err)
						}

						//fileblock.Print()

					}

					// serializamos el inode users.txt por si ocupo otro bloque
					err = inode.Serialize(partition_path, int64(partition_superblock.Sb_inode_start+(apuntador*partition_superblock.Sb_inode_size)))
					if err != nil {
						return "ERROR COMANDO MKUSR: no se pudo escribir los cambios en la particion", fmt.Errorf("[error comando mkusr] no se pudo escribir los cambios en la particion: %v", err)
					}

					// serializamos el superbloque por si el archivo users.txt ocupo otro bloque
					err = partition_superblock.Serialize(partition_path, int64(mounted_partition.Partition_start))
					if err != nil {
						return "ERROR COMANDO MKGRP: no se pudo escribir los cambios en la particion", fmt.Errorf("[error comando mkgrp] no se pudo escribir los cambios en la particion: %v", err)
					}

					//fmt.Println("----------------------------------------")
					return "COMANDO MKUSR: usuario " + mkusr.User + " creado con exito", nil
				}
			}
		}
	}
	return "", nil
}
