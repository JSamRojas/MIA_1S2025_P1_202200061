package global

import (
	"errors"
	estructuras "server/Structs"
)

var (
	MountedPartitions map[string]string = make(map[string]string)
)

// Funcion para obtener la particion y superbloque con el id especificado
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

// Funcion para obtener la particion montada con el id especificado
func Get_Mounted_Partition(id string) (*estructuras.PARTITION, string, error) {

	//Obtener el path de la particion montada

	// Imprimir los bytes del string
	//for i, b := range id {
	//	fmt.Printf("Índice %d: Caracter '%c' (Byte: %d)\n", i, b, b)
	//}

	//println(len(id))

	path := MountedPartitions[id]

	if path == "" {
		return nil, "", errors.New("id de particion invalido: la particion no esta montada")
	}

	var mbr estructuras.MBR

	// Deserealizar el MBR
	msg, err := mbr.DeserializeMBR(path)
	if err != nil {
		return nil, msg, err
	}

	// Buscar la particion con el id especificos
	part, err := mbr.GetPartitionByID(id)
	if part == nil {
		return nil, "", err
	}

	return part, path, nil

}

// Funcion para obtener el superblock de la particion montada con el id
func Get_superblock_from_part(id string) (*estructuras.SUPERBLOCK, *estructuras.PARTITION, string, error) {

	// Obtener el path de la particion montada
	path := MountedPartitions[id]
	if path == "" {
		return nil, nil, "", errors.New("la particion no esta montada")
	}

	// Se crea una instancia de mbr
	var mbr estructuras.MBR

	// Deserializar el MBR usando el archivo binario
	_, err := mbr.DeserializeMBR(path)
	if err != nil {
		return nil, nil, "", err
	}

	// Buscar la particion con el id especifico
	part, err := mbr.GetPartitionByID(id)
	if part == nil {
		return nil, nil, "", err
	}

	// Creamos una instancia de superblock
	var superblock estructuras.SUPERBLOCK

	// Deserializar la estructura superblock desde un archivo binario
	err = superblock.Deserialize(path, int64(part.Partition_start))
	if err != nil {
		return nil, nil, "", err
	}

	return &superblock, part, path, nil

}
