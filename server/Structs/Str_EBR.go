package Structs

type EBR struct {
	Partition_mount [1]byte
	Partition_fit   [1]byte
	Partition_start int32
	Partition_size  int32
	Partition_next  int32
	Partition_name  [16]byte
}
