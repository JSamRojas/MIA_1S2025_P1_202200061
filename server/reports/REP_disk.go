package reports

import (
	"encoding/binary"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	estructuras "server/Structs"
	util "server/Utilities"
	"strings"
)

func ReporteDISK(mbr *estructuras.MBR, path string, diskPath string) error {

	err := util.Create_Parent_Dir(path)
	if err != nil {
		return err
	}

	dotFileName, outputImage := util.Get_File_Names(path)

	const mbr_Size = 153.0

	total_Size := float64(mbr.Mbr_size)
	usable_Size := total_Size - mbr_Size

	disk_Name := filepath.Base(diskPath)

	dotIMG := fmt.Sprintf(`digraph G {
		labelloc="t";
		label = "Reporte de Disco: %s (Size: %d bytes)";
		node [shape=plaintext];
		
		tabla [label=<
		<table border="1" cellborder="1" cellspacing="0" cellpadding="10"
		bgcolor="#F9F9F9">
		<tr><td rowspan="2" bgcolor="#A95C68" border="1" color="black"><b>MBR</b><br/>%.2f%%</td>`, disk_Name, int64(total_Size), 0.00)

	var partition_Rows string
	var logical_Rows string
	extendend_Found := false
	space_Used := mbr_Size
	free_Space := usable_Size
	var free_Extended float64

	// RECORRER LAS PARTICIONES PRIMARIAS Y LAS EXTENDIDAS
	for _, part := range mbr.Mbr_partitions {
		if part.Partition_size == 0 {
			continue
		}

		part_Name := strings.TrimRight(string(part.Partition_name[:]), "\x00")
		part_Type := rune(part.Partition_type[0])
		part_Size := float64(part.Partition_size)
		part_Percentage := (part_Size / total_Size) * 100

		space_Used += part_Size
		free_Space -= part_Size

		// SI LA PARTICION ES EXTENDIDA

		if part_Type == 'E' {

			extendend_Found = true
			partition_Rows += fmt.Sprintf(`
			<td colspan="20" bgcolor="#F0E68C" border="1" color="black"><b>EXTENDIDA<br/>%.2f%% del Disco</b></td>`, part_Percentage)

			file, err := os.Open(diskPath) // Nos movemos al inicio de la particion extendida
			if err != nil {
				return fmt.Errorf("error al abrir el archivo del disco: %v", err)
			}
			defer file.Close()

			var ebr estructuras.EBR
			_, err = file.Seek(int64(part.Partition_start), 0)
			if err != nil {
				return fmt.Errorf("error al moverse al inicio de la particion extendida: %v", err)
			}
			err = binary.Read(file, binary.LittleEndian, &ebr)
			if err != nil {
				return fmt.Errorf("error al leer el primer EBR: %v", err)
			}

			/*

				Si el EBR no existe, entonces no tenemos particiones logicas
				en cambio, si existe, entonces se recorreran todos los EBRs
				y por ende las particiones logicas

			*/

			for {
				ebr_Name := strings.TrimRight(string(ebr.Partition_name[:]), "\x00")
				ebr_Size := float64(ebr.Partition_size)
				ebr_Percent := (ebr_Size / total_Size) * 100

				logical_Rows += fmt.Sprintf(`
				<td bgcolor="#D27D2D" border="1" color="black">EBR</td>
				<td bgcolor="#C2B280" border="1" color="black">Logica<br/>%s<br/>%.2f%% del Disco</td>`, ebr_Name, ebr_Percent)

				space_Used += ebr_Size + ebr_Size // Considerando un EBR y una particion logica

				if ebr.Partition_next == -1 {

					/*

						Se calcula el espacio libre dentro de la particion extendida y se agrega
						luego de la ultima particion logica

					*/

					free_Extended = float64(part.Partition_size) - float64(ebr.Partition_start-part.Partition_start) - ebr_Size
					if free_Extended > 0 {
						logical_Rows += fmt.Sprintf(`
						<td bgcolor="#E0E0E0" border="1" color="black"> Espacio Libre en extendida<br/>%.2f%% del Disco</td>`, (free_Extended/total_Size)*100)
					}
					break
				}

				_, err = file.Seek(int64(ebr.Partition_next), 0)
				if err != nil {
					return fmt.Errorf("error al moverse al siguiente EBR: %v", err)
				}
				err = binary.Read(file, binary.LittleEndian, &ebr)
				if err != nil {
					return fmt.Errorf("error al leer le siguiente EBR: %v", err)
				}
			}
		} else {
			if string(part_Type) != "0" {
				partition_Rows += fmt.Sprintf(`
				<td rowspan="2" bgcolor="#DAA06D" border="1" color="black"><b>%s</b><br/>%d bytes<br/>%.2f%% del Disco</td>`, part_Name, int64(part_Size), part_Percentage)
			}
		}

	}

	/*

		Una vez agregado todo el espacio utilizado
		por ultimo se coloca el espacio libre dentro del disco

	*/

	if free_Space > 0 {
		free_Space_Percent := (free_Space / total_Size) * 100
		partition_Rows += fmt.Sprintf(`
		<td rowspan="2" bgcolor="#E0E0E0" border="1" color="black">Espacio Libre <br/>%.2f%% del Disco</td>`, free_Space_Percent)
	}

	if !extendend_Found {
		dotIMG += partition_Rows + `</tr></table>>]; }`
	} else {
		dotIMG += partition_Rows + `</tr><tr>` + logical_Rows + `</tr></table>>]; }`
	}

	// GUARDAR EL CONTENIDO DEL DOT EN UN ARCHIVO
	file, err := os.Create(dotFileName)
	if err != nil {
		return fmt.Errorf("error al crear el archivo: %v", err)
	}
	defer file.Close()

	_, err = file.WriteString(dotIMG)
	if err != nil {
		return fmt.Errorf("error al escribir en el archivo: %v", err)
	}

	// Se ejecuta el comando de Graphviz para generar la imagen
	cmd := exec.Command("dot",
		"-Tpng",
		dotFileName,
		"-o",
		outputImage)

	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("error al ejecutar el comando Graphviz: %v", err)
	}

	return nil

}
