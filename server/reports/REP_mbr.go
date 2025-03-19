package reports

import (
	"encoding/binary"
	"fmt"
	"os"
	"os/exec"
	estructuras "server/Structs"
	util "server/Utilities"
	"strings"
	"time"
)

func ReporteMBR(mbr *estructuras.MBR, path string, diskPath string) error {
	err := util.Create_Parent_Dir(path)
	if err != nil {
		return err
	}

	dotFileName, outputImage := util.Get_File_Names(path)

	dotIMG := fmt.Sprintf(`digraph G {
		node [shape=plaintext]
		tabla [label =<
		<table border="0" cellborder="1" cellspacing="0">
			<tr><td colspan="2" bgcolor="lightblue"> REPORTE MBR </td></tr>
			<tr><td bgcolor = "lightblue1"> MBR_SIZE </td><td> %d </td></tr>
			<tr><td bgcolor = "lightblue1"> MBR_DATE </td><td> %s </td></tr>
			<tr><td bgcolor = "lightblue1"> MBR_DISK_SIGNATURE </td><td> %d </td></tr>`, mbr.Mbr_size, time.Unix(int64(mbr.Mbr_date), 0), mbr.Mbr_signature_disk)

	// AGREGAR LAS PARTICIONES AL REPORTE
	for i, part := range mbr.Mbr_partitions {

		part_Name := strings.TrimRight(string(part.Partition_name[:]), "\x00")
		part_Status := rune(part.Partition_status[0])
		part_Type := rune(part.Partition_type[0])
		part_Fit := rune(part.Partition_fit[0])

		dotIMG += fmt.Sprintf(`
			<tr><td colspan="2" bgcolor="royalblue"> PARTICION %d </td></tr>
			<tr><td bgcolor = "dodgerblue"> PART_STATUS </td><td> %c </td></tr>
			<tr><td bgcolor = "dodgerblue"> PART_TYPE </td><td> %c </td></tr>
			<tr><td bgcolor = "dodgerblue"> PART_FIT </td><td> %c </td></tr>
			<tr><td bgcolor = "dodgerblue"> PART_START </td><td> %d </td></tr>
			<tr><td bgcolor = "dodgerblue"> PART_SIZE </td><td> %d </td></tr>
			<tr><td bgcolor = "dodgerblue"> PART_NAME </td><td> %s </td></tr>`, i+1, part_Status, part_Type, part_Fit, part.Partition_start, part.Partition_size, part_Name)

		/*

			Si la particion es extendida
			se buscan las particiones logicas y los EBRs

		*/

		if part_Type == 'E' {

			file, err := os.Open(diskPath) // Nos movemos al inicio de la particion extendida
			if err != nil {
				return fmt.Errorf("error al abrir el archivo del disco: %v", err)
			}
			defer file.Close()

			var ebr estructuras.EBR // Se crea una instancia de un EBR

			_, err = file.Seek(int64(part.Partition_start), 0) // Se mueve al inicio de la particion extendida y se lee el primer EBR
			if err != nil {
				return fmt.Errorf("error al moverse al inicio de la particion extendida: %v", err)
			}
			err = binary.Read(file, binary.LittleEndian, &ebr)
			if err != nil {
				return fmt.Errorf("error al leer el primer EBR: %v", err)
			}

			// Si el EBR no existe, entonces no tenemos particiones logicas
			if ebr.Partition_size == 0 {
				continue
			}

			// Reccorrer todos los EBRs
			for {
				// Se convierten los campos necesarios del EBR a string
				ebr_Name := strings.TrimRight(string(ebr.Partition_name[:]), "\x00")
				ebr_Status := rune(ebr.Partition_mount[0])
				ebr_Fit := rune(ebr.Partition_fit[0])

				// AGREGAR EL EBR AL REPORTE (con su particion logica)
				dotIMG += fmt.Sprintf(`
					<tr><td colspan="2" bgcolor="forestgreen"> PARTICION LOGICA </td></tr>
					<tr><td bgcolor="chartreuse2"> EBR_STATUS </td><td> %c </td></tr>
					<tr><td bgcolor="chartreuse2"> EBR_FIT </td><td> %c </td></tr>
					<tr><td bgcolor="chartreuse2"> EBR_START </td><td> %d </td></tr>
					<tr><td bgcolor="chartreuse2"> EBR_SIZE </td><td> %d </td></tr>
					<tr><td bgcolor="chartreuse2"> EBR_NEXT </td><td> %d </td></tr>
					<tr><td bgcolor="chartreuse2"> EBR_NAME </td><td> %s </td></tr>`, ebr_Status, ebr_Fit, ebr.Partition_start, ebr.Partition_size, ebr.Partition_next, ebr_Name)

				// Si no hay otra particion logica, se sale del ciclo
				if ebr.Partition_next == -1 {
					break
				}

				// Se mueve al siguiente EBR
				_, err = file.Seek(int64(ebr.Partition_next), 0)
				if err != nil {
					return fmt.Errorf("error al moverse al siguiente EBR: %v", err)
				}
				err = binary.Read(file, binary.LittleEndian, &ebr)
				if err != nil {
					return fmt.Errorf("error al leer el siguiente EBR: %v", err)
				}

			}

		}

	}

	dotIMG += `</table>>] }` // Cierre de la tabla

	file, err := os.Create(dotFileName) // Se crea el archivo .dot
	if err != nil {
		return fmt.Errorf("error al crear el archivo .dot: %v", err)
	}
	defer file.Close()

	_, err = file.WriteString(dotIMG) // Se escribe el contenido del archivo .dot
	if err != nil {
		return fmt.Errorf("error al escribir el archivo .dot: %v", err)
	}

	cmd := exec.Command("dot",
		"-Tpng",
		dotFileName,
		"-o",
		outputImage)

	err = cmd.Run() // Se ejecuta el comando dot
	if err != nil {
		return fmt.Errorf("error al ejecutar el comando dot: %v", err)
	}

	return nil

}
