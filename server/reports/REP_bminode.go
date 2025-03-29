package reports

import (
	"fmt"
	"os"
	estructuras "server/Structs"
	util "server/Utilities"
	"strings"
)

func ReporteBMINODE(superblock *estructuras.SUPERBLOCK, diskPath string, path string) error {

	// crear las carpetas padre si no existen
	err := util.Create_Parent_Dir(path)
	if err != nil {
		return err
	}

	// abrir el archivo del disco
	file, err := os.Open(diskPath)
	if err != nil {
		return fmt.Errorf("[error comando report -bm_inode] no se pudo abrir el disco: %v", err)
	}
	defer file.Close()

	// calculamos el numero total de inodos
	total_inodes := superblock.Sb_inodes_count + superblock.Sb_free_inodes_count

	// obtenemos el contenido del bitmap
	var Inodes_bitmap strings.Builder

	for i := int32(0); i < total_inodes; i++ {

		// establecemos el puntero donde se localiza el archivo
		_, err := file.Seek(int64(superblock.Sb_bm_inode_start+i), 0)
		if err != nil {
			return fmt.Errorf("[error comando report -bm_inode] no se pudo leer el archivo del inode: %v", err)
		}

		// leemos el byte
		char := make([]byte, 1)
		_, err = file.Read(char)
		if err != nil {
			return fmt.Errorf("[error comando report -bm_inode] no se pudo leer el byte del archivo: %v", err)
		}

		// agregamos el caracter al contenido del bitmap
		Inodes_bitmap.WriteByte(char[0])

		// agregamos un caracter de nueva linea cada 20 caracteres (20 inodes)
		if (i+1)%20 == 0 {
			Inodes_bitmap.WriteString("\n")
		}
	}

	// creamos el archivo txt
	txt_File, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("[error comando report -bm_inode] no se pudo crear el archivo txt: %v", err)
	}
	defer txt_File.Close()

	// escribimos el contenido del bitmap en el archivo
	_, err = txt_File.WriteString(Inodes_bitmap.String())
	if err != nil {
		return fmt.Errorf("[error comando report -bm_inode] error al escribir en el archivo txt: %v", err)
	}

	return nil

}
