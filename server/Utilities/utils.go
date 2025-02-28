package Utilities

import (
	"errors"
	"fmt"
	"os"
)

func ConvertBytes(size int, unit string) (int, error) {

	switch unit {
	case "B":
		return size, nil
	case "K":
		return size * 1024, nil
	case "M":
		return size * 1048576, nil
	default:
		return 0, errors.New("error: Unidad no reconocida")
	}

}

func RemoveDiskFile(path string) (string, error) {
	err := os.Remove(path)
	if err != nil {
		return "", fmt.Errorf("error al eliminar el disco: %s", err)
	}

	return "RMDISK: Disco eliminado correctamente", nil

}
