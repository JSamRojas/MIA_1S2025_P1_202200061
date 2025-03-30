package reports

import (
	"os"
	"os/exec"
	estructuras "server/Structs"
	util "server/Utilities"
)

func ReporteTREE(superblock *estructuras.SUPERBLOCK, diskPath string, path string) error {

	// creamos las carpetas padre del path, si no existen
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
	contenido_tree, err := Execute_tree(superblock, diskPath)
	if err != nil {
		return err
	}

	// agregamos el contenido del tree
	dot_Content += contenido_tree

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

func Execute_tree(sb *estructuras.SUPERBLOCK, path string) (string, error) {

	// funcion intermedia que servira para llamar a la recursiva

	content, err := Recursive_Tree(sb, 0)
	if err != nil {
		return "", err
	}

	return content, nil

}

func Recursive_Tree(sb *estructuras.SUPERBLOCK, number_struct int) (string, error) {

}
