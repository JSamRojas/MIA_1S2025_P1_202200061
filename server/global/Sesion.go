package global

import (
	"errors"
	estructuras "server/Structs"
	"strconv"
	"strings"
	"sync"
)

var sessions_Map = make(map[string]map[string]bool)
var mutex sync.Mutex

// Funcion para obtener el ID de la particion que tiene una sesion activa
func Get_id_Session() string {
	mutex.Lock()
	defer mutex.Unlock()

	for part_Id, users := range sessions_Map {
		for _, active := range users {
			if active {
				return part_Id
			}
		}
	}
	return ""
}

// Funcion para verificar si un usuario tiene un sesion activa dentro de una particion
func Is_session_Active(part_Id string) bool {
	mutex.Lock()
	defer mutex.Unlock()

	if users, exists := sessions_Map[part_Id]; exists {
		for _, active := range users {
			if active {
				return true
			}
		}
	}
	return false
}

// Funcion para iniciar sesion en una particion (se marca como sesion activa de un usuario dentro de una particion)
func Activate_session(partition_Id string, user string) {
	mutex.Lock()
	defer mutex.Unlock()

	// Si la particion no existe, entonces se crea un nuevo mapa de usuarios para la misma
	if _, exists := sessions_Map[partition_Id]; !exists {
		sessions_Map[partition_Id] = make(map[string]bool)
	}

	// Marcar el usuario como activo en la particion
	sessions_Map[partition_Id][user] = true

}

// Funcion para cerrar las sesiones activas (desactivar las particiones)
func Desactivate_session() {
	mutex.Lock()
	defer mutex.Unlock()

	for partition_Id := range sessions_Map {
		for user := range sessions_Map[partition_Id] {
			sessions_Map[partition_Id][user] = false
		}
	}
}

// Funcion para verificar si un usuario tiene una sesion activa en una particion
func Is_any_session_Active() bool {
	mutex.Lock()
	defer mutex.Unlock()

	for _, users := range sessions_Map {
		for _, active := range users {
			if active {
				return true
			}
		}
	}
	return false
}

// Funcion para obtener el usuario activo
func Get_user_Active(partition_Id string) string {
	mutex.Lock()
	defer mutex.Unlock()

	if users, exists := sessions_Map[partition_Id]; exists {
		for user, active := range users {
			if active {
				return user
			}
		}
	}

	return ""
}

// Funcion para saber si el usuario esta activo o no
func User_is_Active(content []string, user string) bool {

	for _, line := range content {

		values := strings.Split(line, ",")

		if len(values) >= 3 && values[1] == "U" && values[3] == user {
			if values[0] == "0" {
				return false
			} else {
				return true
			}
		}

	}
	return false

}

// Funcion para obtener el id del usuario activo y el id del grupo al que pertenece
func Get_userid_groupid() (int32, int32, error) {

	partition_Id := Get_id_Session()
	user := Get_user_Active(partition_Id)

	// verificar si es el usuario root
	if user == "root" {
		return int32(1), int32(1), nil
	}

	superblock_partition, _, partition_path, err := Get_superblock_from_part(partition_Id)
	if err != nil {
		return int32(-1), int32(-1), err
	}

	// creamos una instancia de inode
	inode := estructuras.INODE{}

	// deserializamos el inode root
	err = inode.Deserialize(partition_path, int64(superblock_partition.Sb_inode_start+(0*superblock_partition.Sb_inode_size)))
	if err != nil {
		return int32(-1), int32(-1), err
	}

	// verificamos que le bloque en la primera posicion, si sea el 0
	if inode.I_block[0] == 0 {

		// creamos una instancia de folderblock
		folderblock := &estructuras.FOLDERBLOCK{}

		// deserializamos el primer bloque del inode root
		err = folderblock.Deserialize(partition_path, int64(superblock_partition.Sb_block_start+(inode.I_block[0]*superblock_partition.Sb_block_size)))
		if err != nil {
			return int32(-1), int32(-1), err
		}

		// recorremos el contenido del bloque 0
		for _, contenido := range folderblock.B_content {

			name := strings.Trim(string(contenido.B_name[:]), "\x00")
			apuntador := contenido.B_inodo

			if name == "users.txt" {

				// moverme al inode que apunta el contenido
				err = inode.Deserialize(partition_path, int64(superblock_partition.Sb_inode_start+(apuntador*superblock_partition.Sb_inode_size)))

				if err != nil {
					return int32(-1), int32(-1), err
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
							return int32(-1), int32(-1), err
						}

						// obtenemos el contenido de este bloque
						content_users += strings.Trim(string(fileblock.B_content[:]), "\x00")

					}

					// reemplazar los \r\n con \n para asegurar saltos de linea
					content_users = strings.ReplaceAll(content_users, "\r\n", "\n")

					// Dividir las lineas para obtener cada usuario o grupo
					lines := strings.Split(content_users, "\n")

					// variable para guardar el nombre del grupo
					grpName := ""

					// variables para saber si el usuario y el grupo fueron encontrados
					usrNumber := int32(-1)
					grpNumber := int32(-1)

					// ciclo para buscar el id del usuario
					for _, line := range lines {

						// omitimos las lineas en blanco
						if strings.TrimSpace(line) == "" {
							continue
						}

						values := strings.Split(line, ",")

						// verificar si es un usuario y si el nombre del usuario coincide
						if len(values) >= 3 && values[1] == "U" && values[3] == user {
							// convertimos el id de usuario a integer
							usr_Number, err := strconv.ParseInt(values[0], 10, 32)
							if err != nil {
								return int32(-1), int32(-1), err
							}

							usrNumber = int32(usr_Number)

							// guardamos el nombre del grupo para buscarlo despues
							grpName = values[2]
							break
						}
					}

					// ciclo para buscar el id del grupo, ya con el nombre
					for _, line := range lines {

						// omitimos las lineas en blanco
						if strings.TrimSpace(line) == "" {
							continue
						}

						values := strings.Split(line, ",")

						// verificar si es un grupo y si el nombre del grupo coincide
						if len(values) >= 3 && values[1] == "G" && values[2] == grpName {
							// convertimos el id del grupo a integer
							grp_Number, err := strconv.ParseInt(values[0], 10, 32)
							if err != nil {
								return int32(-1), int32(-1), err
							}

							grpNumber = int32(grp_Number)

						}

					}

					if grpNumber == -1 && usrNumber == -1 {
						return int32(-1), int32(-1), errors.New("error al obtener el id del usuario activo y el id de su grupo")
					}

					return usrNumber, grpNumber, nil
				}
			}
		}
	}

	return int32(-1), int32(-1), err
}
