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

type MOUNT struct {
	Path string
	Name string
	List string
}

func Mount_Command(tokens []string) (*MOUNT, string, error) {

	mount := &MOUNT{}

	atributos := strings.Join(tokens, " ")

	lexic := regexp.MustCompile(`(?i)-path="[^"]+"|(?i)-path=[^\s]+|(?i)-name="[^"]+"|(?i)-name=[^\s]+`)

	found := lexic.FindAllString(atributos, -1)

	for _, fu := range found {

		parametro := strings.SplitN(fu, "=", 2)
		if len(parametro) != 2 {
			return nil, "ERROR: Parametro invalido", fmt.Errorf(("parametro invalido: %s"), fu)
		}

		key, value := strings.ToLower(parametro[0]), parametro[1]

		if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
			value = strings.Trim(value, "\"")
		}

		switch key {

		case "-path":

			if value == "" {
				return nil, "ERROR: El path de la particion no puede ser vacio", errors.New("el path de la particion no puede ser vacio")
			}
			mount.Path = value

		case "-name":

			if value == "" {
				return nil, "ERROR: El nombre de la particion no puede ser vacio", errors.New("el nombre de la particion no puede ser vacio")
			}
			mount.Name = value

		default:

			return nil, "ERROR: Parametro invalido", fmt.Errorf(("parametro invalido: %s"), fu)

		}
	}

	if mount.Path == "" {
		return nil, "ERROR: El path de la particion no puede ser vacio", errors.New("el path de la particion no puede ser vacio")
	}
	if mount.Name == "" {
		return nil, "ERROR: El nombre de la particion no puede ser vacio", errors.New("el nombre de la particion no puede ser vacio")
	}

	msg, err := MountP(mount)
	if err != nil {
		fmt.Println("Error en mount: ", err)
		return nil, msg, err
	}

	return mount, "MOUNT: Montaje de la particion realizado con exito", nil

}

func MountP(mount *MOUNT) (string, error) {

	var mbr estructuras.MBR

	msg, err := mbr.DeserializeMBR(mount.Path)
	if err != nil {
		return msg, fmt.Errorf("error leyendo el MBR del disco: %s", err)
	}

	partition, indexPartition, msg := mbr.GetPartitionByName(mount.Name, mount.Path)

	if partition == nil {
		return msg, fmt.Errorf("no se encontro la particion con el nombre: %s", mount.Name)
	}

	if partition.Partition_type[0] == 'E' || partition.Partition_type[0] == 'L' {
		return "ERROR: No se puede montar una particion extendida o logica", errors.New("no se puede montar una particion extendida o logica")
	}

	if partition.Partition_status[0] == '1' {
		return "ERROR: La particion ya esta montada", errors.New("la particion ya esta montada")
	}

	mbr.UpdatePartitionNumber()

	partition = &mbr.Mbr_partitions[indexPartition]

	id, msg, err := GetPartitionId(mount, int(partition.Partition_number))

	if err != nil {
		return msg, fmt.Errorf("error obteniendo id de la particion: %s", err)
	}

	global.MountedPartitions[id] = mount.Path

	partition.MountPartition(indexPartition, id)

	mbr.Mbr_partitions[indexPartition] = *partition

	msg, err = mbr.SerializeMBR(mount.Path)
	if err != nil {
		return msg, fmt.Errorf("error escribiendo el MBR del disco: %s", err)
	}

	//mbr.PrintPartitions()

	return "", nil

}

func GetPartitionId(mount *MOUNT, indexPartition int) (string, string, error) {

	letra, err := util.GetLetra(mount.Path)
	if err != nil {
		fmt.Println("Error obteniendo letra: ", err)
		return "", "ERROR: Error obteniendo letra", err
	}

	idPartition := fmt.Sprintf("%s%d%s", util.Carnet, indexPartition, letra)

	return idPartition, "id obtenido correctamente", nil

}
