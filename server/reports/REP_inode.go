package reports

import (
	"fmt"
	"os"
	"os/exec"
	estructuras "server/Structs"
	util "server/Utilities"
	"time"
)

func ReporteINODE(superblock *estructuras.SUPERBLOCK, diskPath string, path string) error {

	// creamos las carpetas padre del path, si es que no existen
	err := util.Create_Parent_Dir(path)
	if err != nil {
		return err
	}

	// obtenemos los nombres del archivo
	dotfile_Name, output_Img := util.Get_File_Names(path)

	// Iniciamos con la escritura del reporte
	dot_Content := `digraph G {
		rankdir=LR;
		node [shape=plaintext]	
	`
	// iteramos sobre los inodes del disco
	for i := int32(0); i < superblock.Sb_inodes_count; i++ {

		// creamos una instancia de inode
		inode := &estructuras.INODE{}

		// deserializamos el inode
		err := inode.Deserialize(diskPath, int64(superblock.Sb_inode_start+(i*superblock.Sb_inode_size)))
		if err != nil {
			return err
		}

		// convertimos los tiempos a string
		atime := time.Unix(int64(inode.I_atime), 0).Format(time.RFC3339)
		ctime := time.Unix(int64(inode.I_ctime), 0).Format(time.RFC3339)
		mtime := time.Unix(int64(inode.I_mtime), 0).Format(time.RFC3339)

		// Completamos el contenido de cada Inode
		dot_Content += fmt.Sprintf(`INODE%d [label=<
			<table border="2" cellborder="1" cellspacing="0">
				<tr><td colspan="2" bgcolor = "#743060"><font color="white"> INODE NUMERO %d </font></td></tr>
				<tr><td bgcolor = "#743060"><font color="white">I_uid</font></td><td>%d</td></tr>
				<tr><td bgcolor = "#743060"><font color="white">I_gid</font></td><td>%d</td></tr>
				<tr><td bgcolor = "#743060"><font color="white">I_size</font></td><td>%d</td></tr>
				<tr><td bgcolor = "#743060"><font color="white">I_atime</font></td><td>%s</td></tr>
				<tr><td bgcolor = "#743060"><font color="white">I_ctime</font></td><td>%s</td></tr>
				<tr><td bgcolor = "#743060"><font color="white">I_mtime</font></td><td>%s</td></tr>
				<tr><td bgcolor = "#743060"><font color="white">I_type</font></td><td>%c</td></tr>
				<tr><td bgcolor = "#743060"><font color="white">I_perm</font></td><td>%s</td></tr>
				<tr><td colspan="2" bgcolor = "#AE6F9B"><font color="white">BLOQUES DIRECTOS</font></td></tr>`, i, i, inode.I_uid, inode.I_gid, inode.I_size, atime, ctime, mtime, rune(inode.I_type[0]), string(inode.I_perm[:]))

		// agregamos los bloques directos a la tabla del inode
		for j, block := range inode.I_block {

			if j == 12 {
				break
			}
			dot_Content += fmt.Sprintf(`<tr><td bgcolor = "#743060"><font color="white">%d</font></td><td>%d</td></tr>`, j+1, block)

		}

		// agregamos los bloques indirectos a la tabla (13, 14, 15)
		dot_Content += fmt.Sprintf(`
			<tr><td bgcolor = "#AE6F9B" colspan="2"><font color="white"> BLOQUE INDIRECTO</font></td></tr>
			<tr><td bgcolor = "#743060"><font color="white">%d</font></td><td>%d</td></tr>
			<tr><td bgcolor = "#AE6F9B" colspan="2"><font color="white"> BLOQUE INDIRECTO DOBLE</font></td></tr>
			<tr><td bgcolor = "#743060"><font color="white">%d</font></td><td>%d</td></tr>
			<tr><td bgcolor = "#AE6F9B" colspan="2"><font color="white"> BLOQUE INDIRECTO TRIPLE</font></td></tr>
			<tr><td bgcolor = "#743060"><font color="white">%d</font></td><td>%d</td></tr>
			</table>>];`, 13, inode.I_block[12], 14, inode.I_block[13], 15, inode.I_block[14])

		// agregamos el enlace al siguiente inode, si es el ultimo, entonces ya no
		if i < superblock.Sb_inodes_count-1 {
			dot_Content += fmt.Sprintf("INODE%d -> INODE%d;\n", i, i+1)
		}

	}

	// cerramos el archivo Dot
	dot_Content += "}"

	// creamos el archivo Dot
	dot_File, err := os.Create(dotfile_Name)
	if err != nil {
		return err
	}
	defer dot_File.Close()

	// escribimos el contenido en el archivo
	_, err = dot_File.WriteString(dot_Content)
	if err != nil {
		return err
	}

	// generamos el archivo con graphviz
	cmd := exec.Command("dot",
		"-Tjpg",
		dotfile_Name,
		"-o",
		output_Img)

	err = cmd.Run()
	if err != nil {
		return err
	}

	return nil
}
