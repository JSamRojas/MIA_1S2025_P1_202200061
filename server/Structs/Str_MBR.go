package Structs

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"
)

type MBR struct {
	Mbr_size           int32
	Mbr_date           float32
	Mbr_signature_disk int32
	Mbr_disk_fit       [1]byte
	Mbr_partitions     [4]PARTITION
}

func CreateMBR(disk *MKDISK, sizeB int) (string, error) {

	var fByte byte

	switch disk.Fit {
	case "BF":
		fByte = 'B'
	case "FF":
		fByte = 'F'
	case "WF":
		fByte = 'W'
	default:
		fmt.Println("ERROR: Ajuste no reconocido")
		return "ERROR: Ajuste no reconocido", nil
	}

	mbr := &MBR{
		Mbr_size:           int32(sizeB),
		Mbr_date:           float32(time.Now().Unix()),
		Mbr_signature_disk: rand.Int31(),
		Mbr_disk_fit:       [1]byte{fByte},
		Mbr_partitions: [4]PARTITION{
			{Partition_status: [1]byte{'2'}, Partition_type: [1]byte{'0'}, Partition_fit: [1]byte{'0'}, Partition_start: -1, Partition_size: -1, Partition_name: [16]byte{'0'}, Partition_number: 0, Partition_id: [4]byte{'0'}},
			{Partition_status: [1]byte{'2'}, Partition_type: [1]byte{'0'}, Partition_fit: [1]byte{'0'}, Partition_start: -1, Partition_size: -1, Partition_name: [16]byte{'0'}, Partition_number: 0, Partition_id: [4]byte{'0'}},
			{Partition_status: [1]byte{'2'}, Partition_type: [1]byte{'0'}, Partition_fit: [1]byte{'0'}, Partition_start: -1, Partition_size: -1, Partition_name: [16]byte{'0'}, Partition_number: 0, Partition_id: [4]byte{'0'}},
			{Partition_status: [1]byte{'2'}, Partition_type: [1]byte{'0'}, Partition_fit: [1]byte{'0'}, Partition_start: -1, Partition_size: -1, Partition_name: [16]byte{'0'}, Partition_number: 0, Partition_id: [4]byte{'0'}},
		},
	}

	msg, err := mbr.SerializeMBR(disk.Path)

	if err != nil {
		fmt.Println("Error al serializar el MBR: ", err)
		return msg, err
	}

	//mbr.Print()

	return "", nil

}

func (mbr *MBR) SerializeMBR(path string) (string, error) {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return "ERROR: No se pudo abrir el archivo al intentar serializarlo", err
	}
	defer file.Close()

	err = binary.Write(file, binary.LittleEndian, mbr)
	if err != nil {
		return "ERROR: No se pudo escribir el MBR en el archivo", err
	}

	return "", nil
}

func (mbr *MBR) DeserializeMBR(path string) (string, error) {

	file, err := os.Open(path)

	if err != nil {
		return "ERROR: No se pudo deserializar el MBR", err
	}

	defer file.Close()

	mbrSize := binary.Size(mbr)
	if mbrSize <= 0 {
		return "ERROR: No se pudo obtener el tamaño del MBR", fmt.Errorf("no se pudo obtener el tamaño del MBR: %d", mbrSize)
	}

	buffer := make([]byte, mbrSize)
	_, err = file.Read(buffer)
	if err != nil {
		return "ERROR: No se pudo leer el archivo a deserealizar", err
	}

	reader := bytes.NewReader(buffer)
	err = binary.Read(reader, binary.LittleEndian, mbr)

	if err != nil {
		return "ERROR: No se pudo deserializar el MBR", err
	}

	return "", nil

}

func (mbr *MBR) GetFirstPartitionAvaible() (*PARTITION, int, int, string) {

	offset := binary.Size(mbr)

	for i := 0; i < len(mbr.Mbr_partitions); i++ {

		if mbr.Mbr_partitions[i].Partition_start == -1 {

			return &mbr.Mbr_partitions[i], offset, i, ""

		} else {

			offset += int(mbr.Mbr_partitions[i].Partition_size)

		}
	}
	return nil, -1, -1, ""
}

func (mbr *MBR) GetPartitionByName(name string, path string) (*PARTITION, int, string) {

	for i, partition := range mbr.Mbr_partitions {

		partition_name := strings.Trim(string(partition.Partition_name[:]), "\x00")

		inputName := strings.Trim(name, "\x00")

		if strings.EqualFold(partition_name, inputName) {
			return &partition, i, ""
		}

	}

	return nil, -1, "No se encontro la particion con el nombre: " + name

}

func (mbr *MBR) GetPartitionByID(id string) (*PARTITION, error) {

	for i := 0; i < len(mbr.Mbr_partitions); i++ {
		partitionID := strings.Trim(string(mbr.Mbr_partitions[i].Partition_id[:]), "\x00")
		inputID := strings.Trim(id, "\x00")
		if strings.EqualFold(partitionID, inputID) {
			return &mbr.Mbr_partitions[i], nil
		}
	}
	return nil, errors.New("No se encontro la particion con el id: " + id)
}

func (mbr *MBR) UpdatePartitionNumber() {

	number := 1

	for i := 0; i < len(mbr.Mbr_partitions); i++ {

		partition := &mbr.Mbr_partitions[i]

		if partition.Partition_status[0] != 0 && partition.Partition_type[0] == 'P' {
			partition.Partition_number = int32(number)
			number++
		} else if partition.Partition_type[0] == 'E' {
			partition.Partition_number = 0
		}
	}

}

func (mbr *MBR) Print() {

	creationTime := time.Unix(int64(mbr.Mbr_date), 0)

	diskFit := rune(mbr.Mbr_disk_fit[0])

	fmt.Printf("---------- MBR ----------\n")
	fmt.Printf("MBR Size: %d\n", mbr.Mbr_size)
	fmt.Printf("Creation Date %s\n", creationTime.Format("02-01-2006 15:04:05"))
	fmt.Printf("Disk Signature: %d\n", mbr.Mbr_signature_disk)
	fmt.Printf("Disk Fit: %c\n", diskFit)
	fmt.Printf("---------- END MBR ----------\n")
}

func (mbr *MBR) PrintPartitions() {
	for i, partition := range mbr.Mbr_partitions {

		partition_status := rune(partition.Partition_status[0])
		partition_type := rune(partition.Partition_type[0])
		partition_fit := rune(partition.Partition_fit[0])

		partition_name := string(partition.Partition_name[:])

		partition_id := string(partition.Partition_id[:])

		if partition_status == '2' {
			continue
		}

		fmt.Printf("---------- PARTITION %d ----------\n", i+1)
		fmt.Printf("Partition_status: %c\n", partition_status)
		fmt.Printf("Partition_type: %c\n", partition_type)
		fmt.Printf("Partition_fit: %c\n", partition_fit)
		fmt.Printf("Partition_start: %d\n", partition.Partition_start)
		fmt.Printf("Partition_size: %d\n", partition.Partition_size)
		fmt.Printf("Partition_name: %s\n", partition_name)
		fmt.Printf("Partition_number: %d\n", partition.Partition_number)
		fmt.Printf("Partition_id: %s\n", partition_id)
		fmt.Printf("---------- END PARTITION %d ----------\n", i+1)

	}
}
