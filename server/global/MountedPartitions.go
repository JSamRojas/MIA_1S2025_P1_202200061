package global

import (
	"errors"
	estructuras "server/Structs"
)

var MountedPartitions map[string]string = make(map[string]string)

func Get_essential_rep(id string) (*estructuras.MBR, *estructuras.SUPERBLOCK, string, error) {

	path := MountedPartitions[id]
	if path == "" {
		return nil, nil, "", errors.New("no se encontro la particion montada")
	}

	var mbr estructuras.MBR

	_, err := mbr.DeserializeMBR(path)
	if err != nil {
		return nil, nil, "", err
	}

	partition, err := mbr.GetPartitionByID(id)
	if partition == nil {
		return nil, nil, "", err
	}

	var sb estructuras.SUPERBLOCK

	err = sb.Deserialize(path, int64(partition.Partition_start))
	if err != nil {
		return nil, nil, "", err
	}

	return &mbr, &sb, path, nil

}

func Get_Mounted_Partition(id string) (*estructuras.PARTITION, string, error) {

	//Obtener el path de la particion montada
	path := MountedPartitions[id]
	if path == "" {
		return nil, "", errors.New("id de particion invalido: la particion no esta montada")
	}

	var mbr estructuras.MBR

	// Deserealizar el MBR
	msg, err := mbr.DeserializeMBR(id)
	if err != nil {
		return nil, msg, err
	}

	// Buscar la particion con el id especifico
	part, err := mbr.GetPartitionByID(id)
	if part == nil {
		return nil, "", err
	}

	return part, path, nil

}
