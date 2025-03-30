package reports

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	estructuras "server/Structs"
	util "server/Utilities"
	global "server/global"
	"strings"
	"time"
)

func ReporteLS(superblock *estructuras.SUPERBLOCK, path string, folder_path string) error {

	// creamos las carpetas padre si no existen
	err := util.Create_Parent_Dir(path)
	if err != nil {
		return err
	}

	// obtenemos los nombres base de los archivos
	dotfile_Name, output_Img := util.Get_File_Names(path)

	dot_Content := ""

	dot_Content += `
	digraph G {
    rankdir=LR;
    node [shape=plaintext];

    label=<
        <table border="2" cellborder="1" cellspacing="0">
            <tr>
                <td bgcolor = "#800000"><font color="white"> Permisos </font></td>
                <td bgcolor = "#800000"><font color="white"> Owner </font></td>
                <td bgcolor = "#800000"><font color="white"> Grupo </font></td>
                <td bgcolor = "#800000"><font color="white"> Size (en bytes) </font></td>
                <td bgcolor = "#800000"><font color="white"> Fecha y Hora </font></td>
                <td bgcolor = "#800000"><font color="white"> Tipo </font></td>
                <td bgcolor = "#800000"><font color="white"> Name </font></td>
            </tr>
	`

	// obtenemos el cotenido de la tabla
	tables_content, err := Execute_ls(folder_path, superblock)
	if err != nil {
		return err
	}

	dot_Content += tables_content

	dot_Content += `
		</table>>

	}
	`
	// creamos el archivo Dot
	file, err := os.Create(dotfile_Name)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(dot_Content)
	if err != nil {
		return fmt.Errorf("[error comando rep -ls] no se pudo generar el reporte: %v", err)
	}

	// ejecutamos el comando de graphviz
	cmd := exec.Command("dot",
		"-Tjpg",
		dotfile_Name,
		"-o",
		output_Img)

	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("[error comando rep -ls] no se pudo ejecutar el comando: %v", err)
	}

	return nil

}

func Execute_ls(path_Folder string, superblock_partition *estructuras.SUPERBLOCK) (string, error) {

	// obtenemos las carpetas padres y el archivo destino
	parent_Directories, destine_Directory := util.Get_Parent_Dirs(path_Folder)

	// obtenemos el id de la particion que se encuentra montada
	partition_Id := global.Get_id_Session()

	// obtenemos el superbloque y otros parametros de la particion
	_, _, partition_path, err := global.Get_superblock_from_part(partition_Id)
	if err != nil {
		return "", fmt.Errorf("[error comando rep -ls] no se pudo obtener la particion montada: %v", err)
	}

	// si el arreglo de directorios padre esta vacio, solo se trabaja desde el inode root

	// si el nombre del directorio termina con .txt entonces es un archivo
	if strings.HasSuffix(destine_Directory, ".txt") {
		return "", errors.New("[error comando rep -ls] el nombre del archivo a leer no es valido")
	}

	if len(parent_Directories) == 0 {

		if destine_Directory == "" {

			// funcion para generar la tabla del contenido de la carpeta
			contenido_tabla, err := Get_ls_content(superblock_partition, partition_path, 0)
			if err != nil {
				return "", err
			}

			return contenido_tabla, nil

		} else {

			// si el destino no esta vacio, obtengo el numero del inodo del directorio destino
			inode_Dest, err := superblock_partition.Found_directory(partition_path, 0, destine_Directory)
			if err != nil {
				return "", err
			}

			if inode_Dest != int32(-1) {

				// funcion para generar la tabla del contenido de la carpeta
				contenido_tabla, err := Get_ls_content(superblock_partition, partition_path, inode_Dest)
				if err != nil {
					return "", err
				}

				return contenido_tabla, nil

			} else {
				return "", errors.New("[error comando rep -ls] una de las carpetas del path no existe")
			}

		}
	}

	/*
		Si el arreglo de directorios padre no esta vacio, iteramos uno por uno,
		si el directorio existe, entonces lo obtenemos y pasamos al siguiente y si no existe, entonces mostramos error
	*/

	Inode_destino := int32(0)

	for i := 0; i < len(parent_Directories); i++ {

		// enviamos a buscar el directorio en la posicion i, para saber si existe
		next_inode, err := superblock_partition.Found_directory(partition_path, Inode_destino, parent_Directories[i])
		// si hay un error, lo devolvemos
		if err != nil {
			return "", err
		}

		// si el valor de la variable es un -1, entonces el directorio no existe, por ende, devolvemos error
		if next_inode == int32(-1) {

			return "", errors.New("[error comando rep -ls] uno de los directorios de la ruta no existe")

		} else {

			/*
				si no devuelve un -1, significa que encontro el inode de la siguiente carpeta, por ende se lo asignamos a Inode_destino y mantenemos la continuidad
			*/
			Inode_destino = next_inode

		}

	}

	// una vez encontrados los directorios padres, obtenemos el numero de inodo de la carpeta destino

	inodeFinal, err := superblock_partition.Found_directory(partition_path, Inode_destino, destine_Directory)
	if err != nil {
		return "", err
	}

	// llamada a funcion para obtener la tabla del contenido de la carpeta
	contenido_tabla, err := Get_ls_content(superblock_partition, partition_path, inodeFinal)
	if err != nil {
		return "", err
	}

	return contenido_tabla, nil

}

func Get_ls_content(sb *estructuras.SUPERBLOCK, path string, inode_Index int32) (string, error) {

	// creamos una instancia de inode
	inode := estructuras.INODE{}

	// deserializamos el inode
	err := inode.Deserialize(path, int64(sb.Sb_inode_start+(inode_Index*sb.Sb_inode_size)))
	if err != nil {
		return "", err
	}

	// variable para almacenar el contenido del archivo
	content_table := ""

	// iteramos sobre cada bloque del inodo
	for i := 0; i < len(inode.I_block); i++ {

		// si ya no hay mas bloques disponibles, regresamos
		if inode.I_block[i] == -1 {
			return content_table, nil
		}

		// creamos una instancia de folderblock
		folderblock := &estructuras.FOLDERBLOCK{}

		// deserializamos el bloque
		err := folderblock.Deserialize(path, int64(sb.Sb_block_start+(inode.I_block[i]*sb.Sb_block_size)))
		if err != nil {
			return "", err
		}

		// recorremos el cotenido de los bloques desde la posicion 3, que es la que contiene apuntadores a contenido
		for index_content := 2; index_content < len(folderblock.B_content); index_content++ {

			// obtener el contenido del bloque
			content := folderblock.B_content[index_content]

			// si el contenido esta vacio, retornamos el contenido porque significa que ya no hay mas directorios que registrar
			if content.B_inodo == -1 {
				return content_table, nil
			}

			// convertimos el nombre del bloque y eliminamos caracteres nulos
			block_Name := strings.Trim(string(content.B_name[:]), "\x00")

			// creamos una instancia de inode
			inode_interno := estructuras.INODE{}

			// deserializamos el inode para obtener ciertos valores necesarios
			err := inode_interno.Deserialize(path, int64(sb.Sb_inode_start+(content.B_inodo*sb.Sb_inode_size)))
			if err != nil {
				return "", err
			}

			// variable para guardar los permisos del archivo o carpeta
			permisos := ""

			if inode_interno.I_perm[0] == '7' {
				permisos = "-rwxrwxrwx"
			} else {
				permisos = "-rw-rw-r--"
			}

			// variable para guardar el tipo de inodo
			inode_type := ""

			if inode_interno.I_type[0] == '0' {
				inode_type = "Carpeta"
			} else {
				inode_type = "Archivo"
			}

			// obtenemos el nombre del grupo y del usuario que crearon la carpeta/archivo
			usrName, grpName, err := global.Get_user_group(inode_interno.I_uid, inode_interno.I_gid)
			if err != nil {
				return "", err
			}

			content_table += fmt.Sprintf(`
			<tr>
				<td>%s</td>
				<td>%s</td>
				<td>%s</td>
				<td>%d</td>
				<td>%s</td>
				<td>%s</td>
				<td>%s</td>
			</tr>`, permisos, usrName, grpName, inode_interno.I_size, time.Unix(int64(inode_interno.I_atime), 0).Format(time.RFC3339), inode_type, block_Name)

		}

	}

	return "", nil

}
