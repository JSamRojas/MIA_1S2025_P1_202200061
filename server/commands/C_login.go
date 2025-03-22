package commands

import (
	"errors"
	"fmt"
	"regexp"
	estructuras "server/Structs"
	global "server/global"
	"strings"
)

type LOGIN struct {
	User     string
	Password string
	Id       string
}

func Login_Command(tokens []string) (*LOGIN, string, error) {

	login := &LOGIN{}

	atributos := strings.Join(tokens, " ")

	lexic := regexp.MustCompile(`(?i)-user=[^\s]+|(?i)-pass=[^\s]+|(?i)-id=[^\s]+`)

	found := lexic.FindAllString(atributos, -1)

	for _, fun := range found {

		parametro := strings.SplitN(fun, "=", 2)
		if len(parametro) != 2 {
			return nil, "ERROR COMANDO LOGIN: parametros invalidos", fmt.Errorf("formato de parametros invalido: %s", fun)
		}

		key, value := strings.ToLower(parametro[0]), parametro[1]

		if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
			value = strings.Trim(value, "\"")
		}

		switch key {

		case "-user":

			if value == "" {
				return nil, "ERROR COMANDO LOGIN: el usuario no puede estar vacio", errors.New("el usuario del login no puede estar vacio")
			}
			login.User = value

		case "-pass":

			if value == "" {
				return nil, "ERROR COMANDO LOGIN: la password no puede estar vacia", errors.New("la password no puede estar vacia")
			}
			login.Password = value

		case "-id":

			if value == "" {
				return nil, "ERROR COMANDO LOGIN: el id de la particion no puede estar vacio", errors.New("el id de la particion no puede estar vacio")
			}
			login.Id = value

		default:
			return nil, "ERROR COMANDO LOGIN: parametro desconocido", fmt.Errorf("parametro desconocido: %s", key)
		}

	}

	if login.User == "" || login.Password == "" || login.Id == "" {
		return nil, "ERROR COMANDO LOGIN: faltan parametros obligatorios", errors.New("faltan parametros obligatorios")
	}

	msg, err := make_Login(login)
	if err != nil {
		return nil, msg, err
	}

	return login, "COMANDO LOGIN: realizado con exito " + msg, nil

}

func make_Login(login *LOGIN) (string, error) {

	// Tenemos que ir al archivo users.txt y buscar la pass y el user

	// obtenemos la particion donde se realizara el login
	partition_superblock, _, partition_path, err := global.Get_superblock_from_part(login.Id)
	if err != nil {
		return "ERROR COMANDO LOGIN: no se pudo obtener la particion para realizar el login", fmt.Errorf("error al obtener la particion para el login: %v", err)
	}

	inode := &estructuras.INODE{}

	err = inode.Deserialize(partition_path, int64(partition_superblock.Sb_inode_start+(0*partition_superblock.Sb_inode_size)))
	if err != nil {
		return "ERROR COMANDO LOGIN: no se pudo obtener el root inode", fmt.Errorf("error al obtener el root inode: %v", err)
	}

	// verificar que el primer inode este en 0
	if inode.I_block[0] == 0 {

		// nos movemos al folderblock 0
		folderblock := &estructuras.FOLDERBLOCK{}

		err = folderblock.Deserialize(partition_path, int64(partition_superblock.Sb_block_start+(inode.I_block[0]*partition_superblock.Sb_block_size)))
		if err != nil {
			return "ERROR COMANDO LOGIN: no se pudo obtener el bloque 0", fmt.Errorf("error al tratar de obtener el bloque 0: %v", err)
		}

		// recorremos el contenido del bloque 0
		for _, contenido := range folderblock.B_content {
			name := strings.Trim(string(contenido.B_name[:]), "\x00")
			apuntador := contenido.B_inodo
			if name == "users.txt" {

				// movernos al inode que apunta el contenido
				err = inode.Deserialize(partition_path, int64(partition_superblock.Sb_inode_start+(apuntador*partition_superblock.Sb_inode_size)))
				if err != nil {
					return "ERROR COMANDO LOGIN: no se pudo obtener el inode de users.txt", fmt.Errorf("no se pudo obtener el inode de users.txt: %v", err)
				}

				// variable para almacenar el contenido del archivo
				content_users := ""

				// verificamos que el primer i node este en 1
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
							return "ERROR COMANDO LOGIN: no se pudo obtener el archivo users.txt", fmt.Errorf("[error comando login] no se pudo obtener el archivo de users.txt: %v", err)
						}

						// obtenemos el contenido de este bloque
						content_users += strings.Trim(string(fileblock.B_content[:]), "\x00")

					}

					/*
						El fileblock tiene 64 bytes en donde se guarda lo siguiente:

						1,G,root
						1,U,root,root,123

						donde:
						GUI, TIPO, GRUPO
						UID, TIPO, GRUPO, USUARIO, CONTRASEÑA
					*/

					// reemplazamos \r\n con \n para asegurar saltos de linea correctos
					content_users = strings.ReplaceAll(content_users, "\r\n", "\n")

					// Eliminamos los saltos de linea
					credentials := strings.Split(content_users, "\n")

					for _, user := range credentials {

						values := strings.Split(user, ",")
						if len(values) >= 5 && values[1] == "U" {
							if values[3] == login.User && values[4] == login.Password {

								group_active := Group_is_Active(credentials, values[2])

								if !group_active {
									return "ERROR COMANDO LOGIN: el grupo de este usuario no existe o esta inactivo", errors.New("[errores comando login] el grupo de este usuario no existe o esta inactivo")
								}

								if global.Is_session_Active(login.Id) {
									msg := "YA HAY UNA SESION ACTIVA DENTRO DE ESTA PARTICION, PRIMERO DEBE HACER LOGOUT EN " + login.Id
									return msg, nil
								} else {
									global.Activate_session(login.Id, login.User)
									msg := "USUARIO Y PASSWORD CORRECTOS SESION INICIADA EN " + login.Id + " CON EL USUARIO " + login.User
									return msg, nil
								}
							}
						}
					}
					return "USUARIO Y/O PASSWORD INCORRECTOS", nil
				}
			}
		}
	}
	return "", nil
}

func Group_is_Active(content []string, group string) bool {

	for _, line := range content {

		values := strings.Split(line, ",")

		if len(values) >= 3 && values[1] == "G" && values[2] == group {
			if values[0] == "0" {
				return false
			} else {
				return true
			}
		}

	}
	return false
}
