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

type CHGRP struct {
	User  string
	Group string
}

func Chgrp_Command(tokens []string) (*CHGRP, string, error) {

	chgrp := &CHGRP{}

	// unir tokens en una sola cadena y luego dividir por espacios
	atributos := strings.Join(tokens, " ")
	// expresion regular para encontrar los parametros del comando
	lexic := regexp.MustCompile(`(?i)-user="[^"]+"|(?i)-user=[^\s]+|(?i)-grp="[^"]+"|(?i)-grp=[^\s]+`)
	// encontramos todas las coincidencia de la expresion regular
	found := lexic.FindAllString(atributos, -1)

	for _, fun := range found {

		parametro := strings.SplitN(fun, "=", 2)
		if len(parametro) != 2 {
			return nil, "", fmt.Errorf("[error comando chgrp] formato de parametro invalido: %s", fun)
		}

		key, value := strings.ToLower(parametro[0]), parametro[1]

		if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
			value = strings.Trim(value, "\"")
		}

		switch key {

		case "-user":

			if value == "" {
				return nil, "", errors.New("[error comando chgrp] el parametro user no puede estar vacio")
			}
			chgrp.User = value

		case "-grp":

			if value == "" {
				return nil, "", errors.New("[error comando chgrp] el parametro grp no puede estar vacio")
			}
			chgrp.Group = value

		default:
			return nil, "", fmt.Errorf("[error comando chgrp] parametro invalido: %s", key)
		}

	}

	if chgrp.User == "" {
		return nil, "", errors.New("[error comando chgrp] el parametro user no puede estar vacio")
	}

	if chgrp.Group == "" {
		return nil, "", errors.New("[error comando chgrp] el parametro grp no puede estar vacio")
	}

	msg, err := Change_Group(chgrp)
	if err != nil {
		return nil, msg, err
	}

	return chgrp, msg, nil

}

func Change_Group(chgrp *CHGRP) (string, error) {

	partition_Id := global.Get_id_Session()
	user := global.Get_user_Active(partition_Id)

	// verificar que sea el usuario root
	if user != "root" {
		return "", errors.New("[error comando chgrp] solamente el usuario root puede cambiar de grupo a los usuarios")
	}

	// obtener la particion con el id
	partition_superblock, _, partition_path, err := global.Get_superblock_from_part(partition_Id)
	if err != nil {
		return "", fmt.Errorf("[error comando chgrp] no se pudo obtener la particion montada: %v", err)
	}

	inode := &estructuras.INODE{}

	// Deserializar el inode root
	err = inode.Deserialize(partition_path, int64(partition_superblock.Sb_inode_start+(0*partition_superblock.Sb_inode_size)))
	if err != nil {
		return "", fmt.Errorf("[error comando chgrp] no se pudo obtener el inode root: %v", err)
	}

	// verificar que el primer inode es el 0
	if inode.I_block[0] == 0 {

		// moverse al bloque 0
		folderblock := &estructuras.FOLDERBLOCK{}

		err = folderblock.Deserialize(partition_path, int64(partition_superblock.Sb_block_start+(inode.I_block[0]*partition_superblock.Sb_block_size)))
		if err != nil {
			return "", fmt.Errorf("[error comando chgrp] error al obtener el bloque 0: %v", err)
		}

		// recorrer los contenidos del bloque 0 para buscar el archivo
		for _, contenido := range folderblock.B_content {

			name := strings.Trim(string(contenido.B_name[:]), "\x00")
			apuntador := contenido.B_inodo

			if name == "users.txt" {

				// movernos al inode que apunta el contenido
				err = inode.Deserialize(partition_path, int64(partition_superblock.Sb_inode_start+(apuntador*partition_superblock.Sb_inode_size)))
				if err != nil {
					return "", fmt.Errorf("[error comando chgrp] no se pudo obtener el inode users.txt: %v", err)
				}

				// verificar que el primer bloque sea 1
				if inode.I_block[0] == 1 {

					// variable para almacenar el contenido del archivo
					content_users := ""

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
							return "", fmt.Errorf("[error comando chgrp] no se pudo obtener el archivo de users.txt: %v", err)
						}

						// obtenemos el contenido de este bloque
						content_users += strings.Trim(string(fileblock.B_content[:]), "\x00")

					}

					// reemplazar los \r\n con \n para asegurar saltos de linea
					content_users = strings.ReplaceAll(content_users, "\r\n", "\n")

					// Dividir las lineas para obtener cada usuario o grupo
					lines := strings.Split(content_users, "\n")

					// booleana para saber si el grupo fue encontrado
					found_group := false

					// booleana para saber si el usuario existe
					found_user := false

					for _, line := range lines {
						if strings.TrimSpace(line) == "" {
							continue
						}

						values := strings.Split(line, ",")

						// Verificar si es un usuario y si coincide con el usuario que queremos cambiar
						if len(values) >= 3 && values[1] == "G" {
							if values[2] == chgrp.Group {

								if values[0] == "0" {
									return "", errors.New("[error comando chgrp] el grupo no existe")
								}

								found_group = true
							}
						}
					}

					if !found_group {
						return "", fmt.Errorf("[error comando chgrp] el grupo %s no existe", chgrp.Group)
					}

					for i, line := range lines {
						if strings.TrimSpace(line) == "" {
							continue
						}

						values := strings.Split(line, ",")

						// Verificar si es un usuario y si coincide con el usuario que queremos cambiar
						if len(values) >= 5 && values[1] == "U" && values[3] == chgrp.User {
							// Actualizar el grupo del usuario
							values[2] = chgrp.Group
							lines[i] = strings.Join(values, ",")
							found_user = true
						}

					}

					if !found_user {
						return "", fmt.Errorf("[error comando chgrp] el usuario %s no existe", chgrp.User)
					}

					// volvemos a unir las lineas
					content_users = strings.Join(lines, "\n")

					// separamos el contenido en pedazos
					new_Content := util.Split_into_Chunks(content_users)

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
							return "", fmt.Errorf("[error comando chgrp] no se pudo escribir el nuevo archivo users.txt: %v", err)
						}

					}

					return fmt.Sprintf("COMANDO CHGRP: el grupo del usuario %s ha sido cambiado por el grupo %s de manera exitosa", chgrp.User, chgrp.Group), nil

				}
			}
		}
	}
	return "[error comando chgrp] no se encontro el archivo users.txt", nil
}
