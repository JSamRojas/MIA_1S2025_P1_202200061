package Structs

import "fmt"

type PARTITION struct {
	Partition_status [1]byte
	Partition_type   [1]byte
	Partition_fit    [1]byte
	Partition_start  int32
	Partition_size   int32
	Partition_name   [16]byte
	Partition_number int32
	Partition_id     [4]byte
}

func (p *PARTITION) CreatePartition(partStart, partSize int, partType, partFit, partName string) {

	// 0 = particion creada, 1 = particion montada, 2 = particion disponible
	p.Partition_status[0] = '0'

	// Byte del inicio de la particion
	p.Partition_start = int32(partStart)

	// Tamaño de la particion
	p.Partition_size = int32(partSize)

	// Se asigna el tipo de particion
	if len(partType) > 0 {
		p.Partition_type[0] = partType[0]
	}

	// Se asigna el ajuste de la particion
	if len(partFit) > 0 {
		p.Partition_fit[0] = partFit[0]
	}

	// Se asigna el nombre de la particion
	copy(p.Partition_name[:], partName)

}

func (p *PARTITION) MountPartition(number int, id string) error {

	p.Partition_status[0] = '1'

	copy(p.Partition_id[:], id)

	return nil

}

func (p *PARTITION) Print() {

	fmt.Printf("---------- PARTITION ----------\n")
	fmt.Printf("Partition_status: %c\n", p.Partition_status[0])
	fmt.Printf("Partition_type: %c\n", p.Partition_type[0])
	fmt.Printf("Partition_fit: %c\n", p.Partition_fit[0])
	fmt.Printf("Partition_start: %d\n", p.Partition_start)
	fmt.Printf("Partition_size: %d\n", p.Partition_size)
	fmt.Printf("Partition_name: %s\n", p.Partition_name)
	fmt.Printf("Partition_number: %d\n", p.Partition_number)
	fmt.Printf("Partition_id: %s\n", p.Partition_id)
	fmt.Printf("---------- END PARTITION ----------\n")

}
