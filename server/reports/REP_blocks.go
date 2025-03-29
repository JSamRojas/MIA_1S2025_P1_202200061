package reports

import (
	"fmt"
	"os"
	"os/exec"
	estructuras "server/Structs"
	util "server/Utilities"
	"strings"
)

func ReporteBLOCK(superblock *estructuras.SUPERBLOCK, diskPath string, path string) error {

	// crear la ruta si no existe
	err := util.Create_Parent_Dir(path)
	if err != nil {
		return err
	}

	// obtener el nombre base del archivo
	dotfile_Name, output_Img := util.Get_File_Names(path)

	// iniciar el contenido del DOT
	dot_Content := `digraph G {
		rankdir=LR;
		node [shape=plaintext]
	`

	var prevBlock_Index int32 = -1

	for i := int32(0); i < superblock.Sb_inodes_count; i++ {

		// creamos una instancia de inode
		inode := &estructuras.INODE{}

		// deserializamos el inodo
		err := inode.Deserialize(diskPath, int64(superblock.Sb_inode_start+(i*superblock.Sb_inode_size)))
		if err != nil {
			return err
		}

		// iterar sobre cada bloque del inodo
		for _, blockIndex := range inode.I_block {

			if blockIndex == -1 {
				break
			}

			// verificamos el tipo del bloque
			if inode.I_type[0] == '0' {

				// creamos una instancia de folderblock
				folderblock := &estructuras.FOLDERBLOCK{}

				// deserializamos el bloque
				err := folderblock.Deserialize(diskPath, int64(superblock.Sb_block_start+(blockIndex*superblock.Sb_block_size)))
				if err != nil {
					return err
				}

				block_Content := ""

				dot_Content += fmt.Sprintf(`block%d [label=<
					<table border="2" cellborder="1" cellspacing="0">
					<tr><td colspan="2" bgcolor="#21618C"><font color="white"> BLOQUE CARPETA %d </font></td></tr>
					`, blockIndex, blockIndex)

				content_index := 1

				// creamos el contenido del bloque
				for _, content := range folderblock.B_content {

					// eliminamos los caracteres nulos del nombre del bloque
					content_Name := strings.Trim(string(content.B_name[:]), "\x00")
					block_Content += fmt.Sprintf(`
						<tr><td colspan="2" bgcolor = "#6B9AFF"><font color="white"> CONTENIDO %d </font></td></tr>
						<tr><td bgcolor = "#21618C"><font color="white">B_name</font></td><td>%s</td></tr>
						<tr><td bgcolor = "#21618C"><font color="white">B_inode</font></td><td>%d</td></tr>
					`, content_index, content_Name, content.B_inodo)

					content_index++
				}

				dot_Content += block_Content + `</table>>];`

			} else if inode.I_type[0] == '1' {

				// creamos una instancia de fileblock
				fileblock := &estructuras.FILEBLOCK{}

				// deserializamos el bloque
				err := fileblock.Deserialize(diskPath, int64(superblock.Sb_block_start+(blockIndex*superblock.Sb_block_size)))
				if err != nil {
					return err
				}

				dot_Content += fmt.Sprintf(`block%d [label=<
					<table border="2" cellborder="1" cellspacing="0">
						<tr><td colspan="2" bgcolor="#21618C"><font color="white"> BLOQUE ARCHIVO %d </font></td></tr>
						<tr><td bgcolor = "#6b9AFF"><font color="white">b_content</font></td><td>%s</td></tr>
						</table>>];`, blockIndex, blockIndex, formatGraphvizString(strings.Trim(string(fileblock.B_content[:]), "\x00")))
			}

			// enlazar los bloques
			if prevBlock_Index != -1 {
				dot_Content += fmt.Sprintf("block%d -> block%d;\n", prevBlock_Index, blockIndex)
			}
			prevBlock_Index = blockIndex
		}
	}

	dot_Content += "}"

	// creamos el archivo
	dot_File, err := os.Create(dotfile_Name)
	if err != nil {
		return err
	}
	defer dot_File.Close()

	_, err = dot_File.WriteString(dot_Content)
	if err != nil {
		return err
	}

	// generamos el comando para la imagen
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

func formatGraphvizString(data string) string {
	result := ""
	for i, c := range data {
		result += string(c)
		if (i+1)%10 == 0 && i != len(data)-1 {
			result += "<br/>"
		}
	}
	return result
}
