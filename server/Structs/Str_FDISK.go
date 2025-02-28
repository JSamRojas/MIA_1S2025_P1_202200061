package Structs

import (
	"encoding/binary"
	"fmt"
	"os"
	util "server/Utilities"
)

type FDISK struct {
	Size int
	Unit string
	Path string
	Type string
	Fit  string
	Name string
}

func Struct_FDISK(fdisk *FDISK) (string, error) {

	sizeBytes, err := util.ConvertBytes(fdisk.Size, fdisk.Unit)
	if err != nil {
		fmt.Println("Error al convertir el size: ", err)
		return "ERROR: No se pudo convertir el size de la particion", err
	}

	var msg string

	switch fdisk.Type {

	case "P":
		msg, err = C_primaryPartition(fdisk, sizeBytes)
		if err != nil {
			fmt.Println("Error al crear la particion primaria: ", err)
			return msg, err
		}

	case "E":
		msg, err = C_extendedPartition(fdisk, sizeBytes)
		if err != nil {
			fmt.Println("Error al crear la particion extendida: ", err)
			return msg, err
		}

	case "L":
		msg, err = C_logicalPartition(fdisk, sizeBytes)
		if err != nil {
			fmt.Println("Error al crear la particion logica: ", err)
			return msg, err
		}

	}

	return "", nil

}

func C_primaryPartition(fdisk *FDISK, sizeBytes int) (string, error) {

	var part_mbr MBR

	msg, err := part_mbr.DeserializeMBR(fdisk.Path)
	if err != nil {
		return msg, fmt.Errorf("ERROR: No se pudo leer el MBR del disco: %s", err)
	}

	primaryParts := 0
	for _, partition := range part_mbr.Mbr_partitions {
		if partition.Partition_status[0] != '2' {
			if partition.Partition_type[0] == 'P' || partition.Partition_type[0] == 'E' {
				primaryParts++
			}
		}
	}

	if primaryParts >= 4 {
		return "ERROR: Ya existen 4 particiones primarias", fmt.Errorf("ya existen 4 particiones primarias")
	}

	if sizeBytes > int(part_mbr.Mbr_size) {
		return "ERROR: El tamaño de la particion es mayor al tamaño disponible en el disco", fmt.Errorf("el tamaño de la particion es mayor al tamaño disponible en el disco")
	}

	newParticion, startParticion, indexParticion, msg := part_mbr.GetFirstPartitionAvaible()

	if newParticion == nil {
		return msg, fmt.Errorf("no se pudo obtener una particion disponible")
	}

	newParticion.CreatePartition(startParticion, sizeBytes, fdisk.Type, fdisk.Fit, fdisk.Name)

	part_mbr.Mbr_partitions[indexParticion] = *newParticion

	msg, err = part_mbr.SerializeMBR(fdisk.Path)
	if err != nil {
		return msg, fmt.Errorf("ERROR: No se pudo escribir el MBR en el disco: %s", err)
	}

	return "", nil

}

func C_extendedPartition(fdisk *FDISK, sizeBytes int) (string, error) {

	var part_mbr MBR

	msg, err := part_mbr.DeserializeMBR(fdisk.Path)
	if err != nil {
		return msg, fmt.Errorf("ERROR: No se pudo leer el MBR del disco: %s", err)
	}

	extendedExists := 0
	for _, partition := range part_mbr.Mbr_partitions {
		if partition.Partition_status[0] != '2' {
			if partition.Partition_type[0] == 'E' {
				extendedExists++
			}
		}
	}

	if extendedExists > 0 {
		return "ERROR: Ya existe una particion extendida", fmt.Errorf("ya existe una particion extendida")
	}

	if sizeBytes > int(part_mbr.Mbr_size) {
		return "ERROR: El tamaño de la particion es mayor al tamaño disponible en el disco", fmt.Errorf("el tamaño de la particion es mayor al tamaño disponible en el disco")
	}

	newParticion, startParticion, indexParticion, msg := part_mbr.GetFirstPartitionAvaible()
	if newParticion == nil {
		return msg, fmt.Errorf("no se pudo obtener una particion disponible")
	}

	newParticion.CreatePartition(startParticion, sizeBytes, fdisk.Type, fdisk.Fit, fdisk.Name)

	part_mbr.Mbr_partitions[indexParticion] = *newParticion

	msg, err = part_mbr.SerializeMBR(fdisk.Path)

	if err != nil {
		return msg, fmt.Errorf("ERROR: No se pudo escribir el MBR en el disco: %s", err)
	}

	return "", nil

}

func C_logicalPartition(fdisk *FDISK, sizeBytes int) (string, error) {

	var part_mbr MBR

	msg, err := part_mbr.DeserializeMBR(fdisk.Path)
	if err != nil {
		return msg, fmt.Errorf("ERROR: No se pudo leer el MBR del disco: %s", err)
	}

	var extendedPartition *PARTITION
	for _, partition := range part_mbr.Mbr_partitions {
		if partition.Partition_type[0] == 'E' {
			extendedPartition = &partition
			break
		}
	}

	if extendedPartition == nil {
		return "ERROR: No existe una particion extendida", fmt.Errorf("no existe una particion extendida")
	}

	if sizeBytes > int(extendedPartition.Partition_size) {
		return "ERROR: El tamaño de la particion logica es mayor al tamaño disponible en la particion extendida", fmt.Errorf("el tamaño de la particion logica es mayor al tamaño disponible en la particion extendida")
	}

	// Se abre el archivo del disco
	file, err := os.OpenFile(fdisk.Path, os.O_RDWR, 0644)
	if err != nil {
		return "ERROR: No se pudo abrir el archivo del disco", err
	}
	defer file.Close()

	// Se posiciona al inicio de la particion extendida
	_, err = file.Seek(int64(extendedPartition.Partition_start), 0)
	if err != nil {
		return "ERROR: No se pudo posicionar en la particion extendida", err
	}

	var ebr EBR
	err = binary.Read(file, binary.LittleEndian, &ebr)

	if err != nil || ebr.Partition_size == 0 {

		// Si no existe un EBR en la particion extendida, se crea uno

		ebr = EBR{

			Partition_mount: [1]byte{'0'},
			Partition_fit:   [1]byte{fdisk.Fit[0]},
			// El EBR empieza en el inicio de la particion extendida
			Partition_start: extendedPartition.Partition_start,
			// El size de la EBR es el size de la particion logica
			Partition_size: int32(sizeBytes),
			// No hay otro EBR
			Partition_next: -1,
		}
		copy(ebr.Partition_name[:], []byte(fdisk.Name))

		_, err = file.Seek(int64(extendedPartition.Partition_start), 0)
		if err != nil {
			return "ERROR: No se pudo posicionar al inicio de la particion extendida", err
		}

		err = binary.Write(file, binary.LittleEndian, &ebr)
		if err != nil {
			return "ERROR: No se pudo escribir el EBR en la particion extendida", err
		}

		logicalStart := extendedPartition.Partition_start + int32(binary.Size(ebr))

		var logicalPartition PARTITION

		logicalPartition.CreatePartition(int(logicalStart), sizeBytes, fdisk.Type, fdisk.Fit, fdisk.Name)

		logicalPartition.Partition_id = extendedPartition.Partition_id

		_, err = file.Seek(int64(logicalStart), 0)
		if err != nil {
			return "ERROR: No se pudo posicionar al inicio de la particion logica", err
		}

		err = binary.Write(file, binary.LittleEndian, &logicalPartition)
		if err != nil {
			return "ERROR: No se pudo escribir la particion logica", err
		}

		msg, err = part_mbr.SerializeMBR(fdisk.Path)
		if err != nil {
			return msg, err
		}

		return "FDISK: Particion logica y EBR creados correctamente", nil

	}

	// Si ya existe un EBR en la particion extendida, se busca el ultimo EBR

	for ebr.Partition_next != -1 {
		_, err = file.Seek(int64(ebr.Partition_next), 0)
		if err != nil {
			return "ERROR: No se pudo posicionar al siguiente EBR", err
		}
		err = binary.Read(file, binary.LittleEndian, &ebr)
		if err != nil {
			return "ERROR: No se pudo leer el siguiente EBR", err
		}
	}

	newEBRstart := ebr.Partition_start + ebr.Partition_size + int32(binary.Size(ebr))

	ebr.Partition_next = newEBRstart

	_, err = file.Seek(int64(ebr.Partition_start), 0)
	if err != nil {
		return "ERROR: No se pudo posicionar en el anterior EBR para actualizarlo", err
	}

	err = binary.Write(file, binary.LittleEndian, &ebr)
	if err != nil {
		return "ERROR: No se pudo actualizar el EBR anterior con el nuevo", err
	}

	ebrNew := EBR{
		Partition_mount: [1]byte{'0'},
		Partition_fit:   [1]byte{fdisk.Fit[0]},
		Partition_start: newEBRstart,
		Partition_size:  int32(sizeBytes),
		Partition_next:  -1,
	}
	copy(ebrNew.Partition_name[:], []byte(fdisk.Name))

	_, err = file.Seek(int64(newEBRstart), 0)
	if err != nil {
		return "ERROR: No se pudo posicionar en el nuevo EBR", err
	}

	logicalStart := newEBRstart + int32(binary.Size(ebrNew))

	var logicalPartition PARTITION

	logicalPartition.CreatePartition(int(logicalStart), sizeBytes, fdisk.Type, fdisk.Fit, fdisk.Name)

	logicalPartition.Partition_id = extendedPartition.Partition_id

	_, err = file.Seek(int64(newEBRstart), 0)
	if err != nil {
		return "ERROR: No se pudo posicionar al inicio de la particion logica", err
	}

	err = binary.Write(file, binary.LittleEndian, &logicalPartition)

	if err != nil {
		return "ERROR: No se pudo escribir la particion logica", err
	}

	msg, err = part_mbr.SerializeMBR(fdisk.Path)
	if err != nil {
		return msg, err
	}

	return "FDISK: EBR creado correctamente", nil

}
