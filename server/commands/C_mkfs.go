package commands

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"regexp"
	estructuras "server/Structs"
	"server/global"
	"strings"
	"time"
)

type MKFS struct {
	Id   string
	Type string
}

func Mkfs_Command(tokens []string) (*MKFS, string, error) {

	mkfs := &MKFS{}

	atributos := strings.Join(tokens, " ")

	lexic := regexp.MustCompile(`(?i)-id=[^\s]+|(?i)-type=[^\s]+`)

	found := lexic.FindAllString(atributos, -1)

	for _, fun := range found {

		parametro := strings.SplitN(fun, "=", 2)
		if len(parametro) != 2 {
			return nil, "ERROR: formato de parametros invalido", fmt.Errorf("formato de parametros invalido: %s", fun)
		}
		key, value := strings.ToLower(parametro[0]), parametro[1]

		if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
			value = strings.Trim(value, "\"")
		}

		switch key {

		case "-id":
			if value == "" {
				return nil, "ERROR COMANDO MKFS: parametro id vacio", errors.New("error comando mkfs: parametro id vacio")
			}
			mkfs.Id = value

		case "-type":
			if value != "full" {
				return nil, "ERROR COMANDO MKFS: parametro type desconocido", fmt.Errorf("error comando mkfs: parametro type desconocido: %s", key)
			}
			mkfs.Type = value

		default:
			return nil, "ERROR COMANDO MKFS: parametro desconocido", fmt.Errorf("error comando mkfs: parametro desconocido: %s", key)
		}

	}

	if mkfs.Id == "" {
		return nil, "ERROR COMANDO MKFS: el parametro id es obligatorio", errors.New("error comando mkfs: el parametro id es obligatorio")
	}

	if mkfs.Type == "" {
		mkfs.Type = "full"
	}

	err := Create_MKFS(mkfs)
	if err != nil {
		fmt.Println("Error: ", err)
	}

	return mkfs, "COMANDO MKFS: particion formateada con exito", nil

}

func Create_MKFS(mkfs *MKFS) error {

	mounted_Part, part_Path, err := global.Get_Mounted_Partition(mkfs.Id)
	if err != nil {
		return err
	}

	// calculamos el valor de n
	n_Value := calculate_N(mounted_Part)

	// Se crea el superbloque
	super_Block := Create_SuperBlock(mounted_Part, n_Value)

	// Se crean los bitmaps
	err = super_Block.Create_Bit_Maps(part_Path)
	if err != nil {
		return err
	}

	// Se crea el archivo users.txt
	err = super_Block.Create_UsersTXT(part_Path)
	if err != nil {
		return err
	}

	// Serializar el superblock
	err = super_Block.Serialize(part_Path, int64(mounted_Part.Partition_start))
	if err != nil {
		return err
	}

	return nil

}

func calculate_N(part *estructuras.PARTITION) int32 {

	/*

		n = (Size_partition - Size_SuperBlock)/(4 + size_inodes + (3*Size_fileBlock))

	*/

	numerador := int(part.Partition_size) - binary.Size(estructuras.SUPERBLOCK{})
	denominador := 4 + binary.Size(estructuras.INODE{}) + (3 * binary.Size(estructuras.FILEBLOCK{}))
	n := math.Floor(float64(numerador) / float64(denominador))

	return int32(n)

}

func Create_SuperBlock(part *estructuras.PARTITION, n_Value int32) *estructuras.SUPERBLOCK {

	// Se calculan los punteros de las estructuras
	bm_inode_start := part.Partition_start + int32(binary.Size(estructuras.SUPERBLOCK{}))

	// n_Value indica la cantidad de inodos para ser representado en un bitmap
	bm_block_start := bm_inode_start + n_Value

	// indica la cantidad de bloques y se multiplica por 3 porque se tiene 3 tipos de bloques
	inode_start := bm_block_start + (3 * n_Value)

	// El valor de n indica la cantidad de inodos, pero en este caso, indica la cantidad de estructuras Inode
	block_start := inode_start + (int32(binary.Size(estructuras.INODE{})) * n_Value)

	// Se crea el nuevo superbloque

	super_block := &estructuras.SUPERBLOCK{

		Sb_filesystem_type:   2,
		Sb_inodes_count:      0,
		Sb_blocks_count:      0,
		Sb_free_inodes_count: int32(n_Value),
		Sb_free_blocks_count: int32(n_Value * 3),
		Sb_mtime:             float32(time.Now().Unix()),
		Sb_umtime:            float32(time.Now().Unix()),
		Sb_mnt_count:         1,
		Sb_magic:             0xEF53,
		Sb_inode_size:        int32(binary.Size(estructuras.INODE{})),
		Sb_block_size:        int32(binary.Size(estructuras.FILEBLOCK{})),
		Sb_first_ino:         inode_start,
		Sb_first_blo:         block_start,
		Sb_bm_inode_start:    bm_inode_start,
		Sb_bm_block_start:    bm_block_start,
		Sb_inode_start:       inode_start,
		Sb_block_start:       block_start,
	}

	return super_block

}
