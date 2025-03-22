package commands

import (
	"errors"
	"fmt"
	"regexp"
	estructuras "server/Structs"
	util "server/Utilities"
	global "server/global"
	"strings"
)

type CAT struct {
	file_Path string
}

func Cat_Command(tokens []string) (*CAT, string, error) {

	cat := &CAT{}

	atributos := strings.Join(tokens, " ")

	lexic := regexp.MustCompile(`(?i)-file[1-9][0-9]*="[^"]+"|(?i)-file[1-9][0-9]*=[^\s]+`)

	found := lexic.FindAllString(atributos, -1)

	// Variable para almacenar el contenido de todos los archivos o el archivo a leer
	var Content strings.Builder

	for _, fun := range found {

		parametro := strings.SplitN(fun, "=", 2)
		if len(parametro) != 2 {
			return nil, "ERROR COMANDO CAT: formato de parametros invalido", fmt.Errorf("formato de parametro invalido; %s", fun)
		}

		key, value := strings.ToLower(parametro[0]), parametro[1]

		if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
			value = strings.Trim(value, "\"")
		}

		if strings.HasPrefix(key, "-file") {

			if value == "" {
				return nil, "ERROR COMANDO CAT: la ruta del archivo no puede estar vacia", errors.New("el nombre del archivo no puede estar vacio")
			}
			cat.file_Path = value
		} else {
			return nil, "ERROR COMANDO CAT: parametro invalido", fmt.Errorf("parametro no reconocido: %s", key)
		}

		if cat.file_Path == "" {
			return nil, "ERROR COMANDO CAT: el nombre del archivo es obligatorio", errors.New("el nombre del archivo es obligatorio")
		}

		// Leemos el contenido del archivo
		contenido, err := execute_Cat(cat)
		if err != nil {
			return nil, contenido, err
		}

		// Se concatena el contenido del archivo
		Content.WriteString(contenido + "\n")

	}

	return cat, "COMANDO CAT: lectura realizada con exito\n CONTENIDO: \n" + Content.String(), nil

}

func execute_Cat(cat *CAT) (string, error) {

	/*
		Hay que leer el archivo que esta dentro de la ruta especificada
		El orden es el siguiente: Inode -> folderblock -> contenido
	*/

	// la ruta del archivo esta dentro del parametro file_Path de cat
	parent_Dirs, dest_Dirs := util.Get_Parent_Dirs(cat.file_Path)
	// Arreglo de carpetas o directorios padres
	fmt.Println("\nDirectorios padres del archivo: ", parent_Dirs)
	// Nombre del archivo destino
	fmt.Println("Directorio destino: ", dest_Dirs)

	// Obtener el Id de la particion donde esta logueado
	id_Part := global.Get_id_Session()

	// Primero accedemos al superblock para obtener el Inode root, y posteriormente el Inode del archivo
	part_super_block, part, part_path, err := global.Get_superblock_from_part(id_Part)

	if err != nil {
		return "ERROR COMANDO CAT: no se pudo obtener la particion montada", fmt.Errorf("error al obtener la particion montada: %v", err)
	}

	inode := &estructuras.INODE{}

	err = inode.Deserialize(part_path, int64(part_super_block.Sb_inode_start+(0*part_super_block.Sb_inode_size)))
	if err != nil {
		return "ERROR COMANDO CAT: erro al obtener el inode root", fmt.Errorf("error al obtener el inodo root: %v", err)
	}

	// recorrer los bloques del inode root
	for _, block := range inode.I_block {

		if block != -1 {

			//part_super_block.Print()

			/*
				verificar sobre los bloques del inode
				recorriendolos hasta encontrar el que contiene el directorio
			*/

			folderblock := &estructuras.FOLDERBLOCK{}

			err = folderblock.Deserialize(part_path, int64(part_super_block.Sb_block_start+(block*part_super_block.Sb_block_size)))
			if err != nil {
				return "ERROR COMANDO CAT: no se pudo obtener el bloque", fmt.Errorf("error al obtener el bloque: %v", err)
			}

			// Se recorre el contenido del bloque
			for _, content := range folderblock.B_content {

				if (strings.Trim(string(content.B_name[:]), "\x00") == "users.txt") && (strings.Trim(string(content.B_name[:]), "\x00") == dest_Dirs) {

					// Moverse hasta el inode que apunte al bloque
					err = inode.Deserialize(part_path, int64(part_super_block.Sb_inode_start+(content.B_inodo*part_super_block.Sb_inode_size)))
					if err != nil {
						return "ERROR COMANDO CAT: no se pudo obtener el inodo ", fmt.Errorf("error al obtener el inodo: %v", err)
					}

					// recorremos los bloques del inode para obtener el contenido
					fileblock := &estructuras.FILEBLOCK{}
					output := ""
					for _, block := range inode.I_block {
						if block != -1 {
							err = fileblock.Deserialize(part_path, int64(part_super_block.Sb_block_start+(block*part_super_block.Sb_block_size)))
							if err != nil {
								return "ERROR COMANDO CAT: no se pudo obtener el bloque", fmt.Errorf("error al obtene el bloque: %v", err)
							}
							output += strings.Trim(string(fileblock.B_content[:]), "\x00")
						}
					}
					return output, nil
				}

				if content.B_inodo != -1 && content.B_inodo != 0 && strings.Trim(string(content.B_name[:]), "\x00") != "." && strings.Trim(string(content.B_name[:]), "\x00") != ".." && strings.Trim(string(content.B_name[:]), "\x00") != "users.txt" {

					for i := 0; i < len(parent_Dirs); i++ {

						if strings.Trim(string(content.B_name[:]), "\x00") == parent_Dirs[i] {

							// vamos al inode que apunta al bloque
							err = inode.Deserialize(part_path, int64(part_super_block.Sb_inode_start+(content.B_inodo*part_super_block.Sb_inode_size)))
							if err != nil {
								return "ERROR COMANDO CAT: no se pudo obtener el inodo", fmt.Errorf("error al obtener el inodo: %v", err)
							}
							msg := ""
							msg, err = recursive_Block(inode, part_super_block, part_path, parent_Dirs, dest_Dirs)
							if err != nil {
								return msg, err
							}
							return msg, nil

						}

						if strings.Trim(string(content.B_name[:]), "\x00") == dest_Dirs {

							//Nos movemos al inode que apunta el bloque
							err = inode.Deserialize(part_path, int64(part_super_block.Sb_inode_start+(content.B_inodo*part_super_block.Sb_inode_size)))
							if err != nil {
								return "ERROR COMANDO CAT: no se pudo obtener el inode", fmt.Errorf("error al obtener el inode: %v", err)
							}

							// recorremos los bloques del inode para obtener el archivo
							fileblock := &estructuras.FILEBLOCK{}
							output := ""

							for _, block := range inode.I_block {
								if block != -1 {
									err = fileblock.Deserialize(part_path, int64(part_super_block.Sb_block_start+(block*part_super_block.Sb_block_size)))
									if err != nil {
										return "ERROR COMANDO CAT: no se pudo obtener el bloque", fmt.Errorf("error al obtener el bloque: %v", err)
									}

									// eliminamos los caracteres nulos
									output += strings.Trim(string(fileblock.B_content[:]), "\x00")
								}
							}
							return output, nil

						}

					}

					if strings.Trim(string(content.B_name[:]), "\x00") == dest_Dirs {

						// nos movemos al inode que apunta el bloque
						err = inode.Deserialize(part_path, int64(part_super_block.Sb_inode_start)+(int64(content.B_inodo*part_super_block.Sb_inode_size)))
						if err != nil {
							return "ERROR COMANDO CAT: no se pudo obtener el bloque", fmt.Errorf("error al obtener el bloque: %v", err)
						}

						// recorremos los bloques del inode para obtener el contenido del archivo
						fileblock := &estructuras.FILEBLOCK{}
						output := ""
						for _, block := range inode.I_block {
							if block != -1 {
								err = fileblock.Deserialize(part_path, int64(part_super_block.Sb_block_start+(block*part_super_block.Sb_block_size)))
								if err != nil {
									return "ERROR COMANDO CAT: no se pudo obtener el bloque", fmt.Errorf("error al obtener el bloque: %v", err)
								}

								// eliminamos los caracteres nulos
								output += strings.Trim(string(fileblock.B_content[:]), "\x00")
							}
						}
						return output, nil
					}
				}
			}
		}
	}

	err = part_super_block.Serialize(part_path, int64(part.Partition_start))
	if err != nil {
		return "ERROR COMANDO CAT: no se pudo serializar el superbloque de la particion", fmt.Errorf("error la serializar el superbloque de la particion: %v", err)
	}

	return "", nil

}

// Funcion recursiva para analizar los bloques de un inodo
func recursive_Block(inode *estructuras.INODE, partition_superblock *estructuras.SUPERBLOCK, partition_path string, parent_Dirs []string, dest_Dir string) (string, error) {

	/*
		-Se verifica en los bloques del inode
		-Se recorren hasta encontrar el bloque que contiene la ruta para llegar al  archivo
	*/

	folderblock := &estructuras.FOLDERBLOCK{}

	// recorremos los bloques del inode root
	for _, block := range inode.I_block {

		if block != -1 {
			err := folderblock.Deserialize(partition_path, int64(partition_superblock.Sb_block_start+(block*partition_superblock.Sb_block_size)))
			if err != nil {
				return "ERROR COMANDO CAT: error al obtener el bloque recursivo", fmt.Errorf("error al obtener el bloque recursivo: %v", err)
			}

			// recorremos el contenido de los bloques
			for _, content := range folderblock.B_content {

				if content.B_inodo != -1 && content.B_inodo != 0 {

					for i := 0; i < len(parent_Dirs); i++ {

						if strings.Trim(string(content.B_name[:]), "\x00") == parent_Dirs[i] {

							fmt.Println("Directorio encontrado RECURSIVA:", parent_Dirs[i])

							// Nos movemos al inode que apunta el bloque
							err = inode.Deserialize(partition_path, int64(partition_superblock.Sb_bm_inode_start+(content.B_inodo*partition_superblock.Sb_inode_size)))
							if err != nil {
								return "ERROR COMANDO CAT: error al obtener el inode recursivo", fmt.Errorf("error al obtener el inode recursivo: %v", err)
							}

							/*
								Si ya llegamos al ultimo directorio padre
								entones el inode que apunta el bloque, tiene el bloque que contiene el archivo
							*/

							if i == len(parent_Dirs)-1 {
								msg := ""
								msg, err = recursive_Block(inode, partition_superblock, partition_path, parent_Dirs, dest_Dir)
								if err != nil {
									return msg, err
								}
								return msg, nil
							} else {

								/*
									Si no, entonces significa que tenemos que seguir buscando en otro bloque del inode
								*/
								msg := ""
								msg, err = recursive_Block(inode, partition_superblock, partition_path, parent_Dirs, dest_Dir)
								if err != nil {
									return msg, err
								}
								return msg, nil

							}

						}

						if strings.Trim(string(content.B_name[:]), "\x00") == dest_Dir {

							fmt.Println("Archivo encontrado RECURSIVA: ", dest_Dir)

							// moverse al inode que apunta el bloque
							err = inode.Deserialize(partition_path, int64(partition_superblock.Sb_inode_start+(content.B_inodo*partition_superblock.Sb_inode_size)))
							if err != nil {
								return "ERROR COMANDO CAT: no se pudo mover al inode recursivo", fmt.Errorf("error al moverse al inode: %v", err)
							}

							// recorrer los bloques del inode para obtener el contenido del archivo
							fileblock := &estructuras.FILEBLOCK{}
							output := ""
							for _, block := range inode.I_block {

								if block != -1 {

									err = fileblock.Deserialize(partition_path, int64(partition_superblock.Sb_block_start+(block*partition_superblock.Sb_block_size)))
									if err != nil {
										return "ERROR COMANDO CAT: no se pudo obtener el bloque recursivo", fmt.Errorf("error al obtener el bloque: %v", err)
									}

									output += strings.Trim(string(fileblock.B_content[:]), "\x00")

								}
							}
							return output, nil
						}
					}
				}
			}
		}
	}

	return "COMANDO CAT: archivo no encontrado", nil
}
