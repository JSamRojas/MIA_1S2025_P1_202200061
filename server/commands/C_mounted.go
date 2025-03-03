package commands

import (
	global "server/global"
)

func Mounted_Command(tokens []string) (string, error) {

	var partitions string = "PARTICIONES MONTADAS: "

	if len(global.MountedPartitions) == 0 {
		return "NO HAY PARTICIONES MONTADAS", nil
	}

	for mounted := range global.MountedPartitions {

		partitions += mounted + " "

	}

	return partitions, nil

}
