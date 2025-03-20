package global

import "sync"

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
