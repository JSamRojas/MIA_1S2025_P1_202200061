package reports

import (
	"fmt"
	estructuras "server/Structs"
	util "server/Utilities"
)

func ReporteSB(superblock *estructuras.SUPERBLOCK, path string, diskPath string) error {

	// creamos las carpetas padres
	err := util.Create_Parent_Dir(path)
	if err != nil {
		return err
	}

	// obtenemos los nombres base de los archivos
	dotfile_Name, output_Img := util.Get_File_Names(path)

	// definimos el contenido de la tabla
	dot_Content := fmt.Sprintf(`diagraph G{
		node [shape=plaintext]
		tabla [label=<
		<table border="2" cellborder="1" cellspacing="0">
			<tr><td colspan="2" bgcolor="#663399"><font color="white">REPORTE SUPERBLOQUE </font></td></tr>
			<tr><td bgcolor="#9370DB">File

	`)

}
