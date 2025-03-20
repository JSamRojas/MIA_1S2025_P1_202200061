package commands

import (
	"fmt"
	global "server/global"
)

func Logout_Command(tokens []string) (string, string, error) {

	if len(tokens) != 0 {
		return "", "", fmt.Errorf("numero de comandos incorrecto")
	}

	if !global.Is_any_session_Active() {
		return "", "ERROR COMANDO LOGOUT: No hay ninguna sesion activa en ninguna particion", nil
	}

	global.Desactivate_session()
	return "", "COMANDO LOGOUT: se realizo con exito, todas las sesiones fueron cerradas", nil

}
