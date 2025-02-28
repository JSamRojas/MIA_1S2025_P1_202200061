package Structs

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

	// 0 = particion creada, 1 = particion activa, 2 = particion disponible
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
