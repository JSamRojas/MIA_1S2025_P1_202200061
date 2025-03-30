package reports

import (
	"fmt"
	"os"
	"os/exec"
	estructuras "server/Structs"
	util "server/Utilities"
	"time"
)

func ReporteSB(superblock *estructuras.SUPERBLOCK, diskPath string, path string) error {

	// creamos las carpetas padres
	err := util.Create_Parent_Dir(path)
	if err != nil {
		return err
	}

	// obtenemos los nombres base de los archivos
	dotfile_Name, output_Img := util.Get_File_Names(path)

	// definimos el contenido de la tabla
	dot_Content := fmt.Sprintf(`digraph G{
		node [shape=plaintext]
		tabla [label=<
		<table border="2" cellborder="1" cellspacing="0">
			<tr><td colspan="2" bgcolor="#663399"><font color="white">REPORTE SUPERBLOQUE </font></td></tr>
			<tr><td bgcolor="#9370DB">Filesytem_Type</td><td>%d</td></tr>
			<tr><td bgcolor="#9370DB">Inodes_Count</td><td>%d</td></tr>
			<tr><td bgcolor="#9370DB">Blocks_Count</td><td>%d</td></tr>
			<tr><td bgcolor="#9370DB">Free_Inodes_Count</td><td>%d</td></tr>
			<tr><td bgcolor="#9370DB">Free_Blocks_Count</td><td>%d</td></tr>
			<tr><td bgcolor="#9370DB">Mount_Time</td><td>%s</td></tr>
			<tr><td bgcolor="#9370DB">Unmount_Time</td><td>%s</td></tr>
			<tr><td bgcolor="#9370DB">Mount_Count</td><td>%d</td></tr>
			<tr><td bgcolor="#9370DB">Magic</td><td>%d</td></tr>
			<tr><td bgcolor="#9370DB">Inode_Size</td><td>%d</td></tr>
			<tr><td bgcolor="#9370DB">Block_Size</td><td>%d</td></tr>
			<tr><td bgcolor="#9370DB">First_Inode</td><td>%d</td></tr>
			<tr><td bgcolor="#9370DB">First_Block</td><td>%d</td></tr>
			<tr><td bgcolor="#9370DB">Bitmap_Inode_Start</td><td>%d</td></tr>
			<tr><td bgcolor="#9370DB">Bitmap_Block_Start</td><td>%d</td></tr>
			<tr><td bgcolor="#9370DB">Inode_Start</td><td>%d</td></tr>
			<tr><td bgcolor="#9370DB">Block_Start</td><td>%d</td></tr>
		</table>
		>]}`, superblock.Sb_filesystem_type, superblock.Sb_inodes_count, superblock.Sb_blocks_count, superblock.Sb_free_inodes_count, superblock.Sb_free_blocks_count, time.Unix(int64(superblock.Sb_mtime), 0).Format(time.RFC3339), time.Unix(int64(superblock.Sb_umtime), 0).Format(time.RFC3339), superblock.Sb_mnt_count, superblock.Sb_magic, superblock.Sb_inode_size, superblock.Sb_block_size, superblock.Sb_first_ino, superblock.Sb_first_blo, superblock.Sb_bm_inode_start, superblock.Sb_bm_block_start, superblock.Sb_inode_start, superblock.Sb_block_start)

	// creamos el archivo Dot
	file, err := os.Create(dotfile_Name)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(dot_Content)
	if err != nil {
		return fmt.Errorf("[error comando rep -sb] no se pudo generar el reporte: %v", err)
	}

	// ejecutamos el comando de graphviz
	cmd := exec.Command("dot",
		"-Tjpg",
		dotfile_Name,
		"-o",
		output_Img)

	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("[error comando rep -sb] no se pudo ejecutar el comando: %v", err)
	}

	return nil
}
