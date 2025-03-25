package Structs

import (
	"errors"
	util "server/Utilities"
	"strings"
	"time"
)

// Funcion para crear el archivo de usuarios para el superblock (users.txt)
func (superblock *SUPERBLOCK) Create_UsersTXT(path string) error {

	// Se crea el INODE root o raiz

	root_Inode := &INODE{
		I_uid:   1,
		I_gid:   1,
		I_size:  0,
		I_atime: float32(time.Now().Unix()),
		I_ctime: float32(time.Now().Unix()),
		I_mtime: float32(time.Now().Unix()),
		I_block: [15]int32{superblock.Sb_blocks_count, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1},
		I_type:  [1]byte{'0'},
		I_perm:  [3]byte{'7', '7', '7'},
	}

	// Se serializa el Inodo raiz
	err := root_Inode.Serialize(path, int64(superblock.Sb_first_ino))
	if err != nil {
		return err
	}

	// Actualizamos el bitmap de inodos
	err = superblock.Update_Inode_Bitmap(path)
	if err != nil {
		return err
	}

	// Actualizamos los demas campos del superblock
	superblock.Sb_inodes_count++
	superblock.Sb_free_inodes_count--
	superblock.Sb_first_ino += superblock.Sb_inode_size

	// Creamos el bloque perteneciente al Inodo root o raiz
	root_Block := &FOLDERBLOCK{
		B_content: [4]FOLDERCONTENT{
			{B_name: [12]byte{'.'}, B_inodo: 0},
			{B_name: [12]byte{'.', '.'}, B_inodo: 0},
			{B_name: [12]byte{'-'}, B_inodo: -1},
			{B_name: [12]byte{'-'}, B_inodo: -1},
		},
	}

	// Actualizar el bitmap de bloques
	err = superblock.Update_Block_Bitmap(path)
	if err != nil {
		return nil
	}

	// Serializar el bloque del root
	err = root_Block.Serialize(path, int64(superblock.Sb_first_blo))
	if err != nil {
		return err
	}

	// Actualizamos los campos del superbloque relacionados al folderblock
	superblock.Sb_blocks_count++
	superblock.Sb_free_blocks_count--
	superblock.Sb_first_blo += superblock.Sb_block_size

	// Verificamos el Inode root
	//fmt.Println("\n INODE ROOT: ")
	//root_Inode.Print()

	// Verificamos el fileblock root
	//fmt.Println("\n FOLDERBLOCK ROOT: ")
	//root_Block.Print()

	/*
		Se crea el archivo userts.txt
	*/
	usersTXT := "1,G,root\n1,U,root,root,123\n"

	// Deserealizamos el nodo root
	err = root_Inode.Deserialize(path, int64(superblock.Sb_inode_start+0))
	if err != nil {
		return err
	}

	// Actualizamos el Inode root
	root_Inode.I_atime = float32(time.Now().Unix())

	// Serializamos el Inode root
	err = root_Inode.Serialize(path, int64(superblock.Sb_inode_start+0))
	if err != nil {
		return err
	}

	// Deserealizamos el folderblock de root
	err = root_Block.Deserialize(path, int64(superblock.Sb_block_start+0))
	if err != nil {
		return err
	}

	// Actualizamos el bloque de la carpeta root
	root_Block.B_content[2] = FOLDERCONTENT{B_name: [12]byte{'u', 's', 'e', 'r', 's', '.', 't', 'x', 't'}, B_inodo: superblock.Sb_inodes_count}

	// Serializamos el bloque del root

	err = root_Block.Serialize(path, int64(superblock.Sb_block_start+0))
	if err != nil {
		return err
	}

	// Creamos el Inode de users.txt
	users_Inode := &INODE{
		I_uid:   1,
		I_gid:   1,
		I_size:  int32(len(usersTXT)),
		I_atime: float32(time.Now().Unix()),
		I_ctime: float32(time.Now().Unix()),
		I_mtime: float32(time.Now().Unix()),
		I_block: [15]int32{superblock.Sb_blocks_count, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1},
		I_type:  [1]byte{'1'},
		I_perm:  [3]byte{'7', '7', '7'},
	}

	// Actualizamos el bitmap de inodos
	err = superblock.Update_Inode_Bitmap(path)
	if err != nil {
		return err
	}

	// Serializamos el inode del users.txt
	err = users_Inode.Serialize(path, int64(superblock.Sb_first_ino))
	if err != nil {
		return err
	}

	// Actualizamos de nuevo los parametros del superblock
	superblock.Sb_inodes_count++
	superblock.Sb_free_inodes_count--
	superblock.Sb_first_ino += superblock.Sb_inode_size

	// Creamos el bloque de users.txt
	users_Block := &FILEBLOCK{
		B_content: [64]byte{},
	}

	//Copiamos el contenido de usuarios dentro del bloque
	copy(users_Block.B_content[:], usersTXT)

	// Serializar el bloque de users.txt
	err = users_Block.Serialize(path, int64(superblock.Sb_first_blo))
	if err != nil {
		return err
	}

	// Actualizamos el bitmap de bloques
	err = superblock.Update_Block_Bitmap(path)
	if err != nil {
		return err
	}

	// Actualizamos el superblock
	superblock.Sb_blocks_count++
	superblock.Sb_free_blocks_count--
	superblock.Sb_first_blo += superblock.Sb_block_size

	// Verificar el Inode root
	//fmt.Println("\n INODE root Actualizado: ")
	//root_Inode.Print()

	// Verificar el folerblock del root
	//fmt.Println("\n FOLDERBLOCK root actualizado: ")
	//root_Block.Print()

	// Verificar el inode users.txt
	//fmt.Print("\n INODE users.txt: ")
	//users_Inode.Print()

	// Verificar el fileblock de users.txt
	//fmt.Print("\n FILEBLOCK de users.txt: ")
	//users_Block.Print()

	return nil

}

// Funcion para crear un Inode de tipo carpeta en un inode especifico
func (superblock *SUPERBLOCK) Create_Inode_as_Folder(path string, inode_Index int32, parents_Directories []string, dest_Directory string, create_Parents bool) error {

	// creamos una nueva instancia de inode
	inode := &INODE{}

	// deserializamos el inode que se paso por parametro
	err := inode.Deserialize(path, int64(superblock.Sb_inode_start+(inode_Index*superblock.Sb_inode_size)))
	if err != nil {
		return err
	}

	// verificar si el inode es de tipo carpeta
	if inode.I_type[0] == '1' {
		return errors.New("[error comando mkdir] la ruta especificada incluye un archivo, no una carpeta")
	}

	// iterar sobre cada bloque del inode
	for i := 0; i < len(inode.I_block); i++ {

		if inode.I_block[i] == -1 {

			if inode_Index == 0 {

				inode.I_block[i] = superblock.Sb_blocks_count

				new_folderblock := &FOLDERBLOCK{
					B_content: [4]FOLDERCONTENT{
						{B_name: [12]byte{'.'}, B_inodo: 0},
						{B_name: [12]byte{'.', '.'}, B_inodo: 0},
						{B_name: [12]byte{'-'}, B_inodo: -1},
						{B_name: [12]byte{'-'}, B_inodo: -1},
					},
				}

				// creamos una instancia de foldercontent
				content_folder := &FOLDERCONTENT{}

				// obtenemos el tercer campo del mismo (el primer espacio disponible)
				content_folder = &new_folderblock.B_content[2]

				// copiamos el nombre del archivo
				copy(content_folder.B_name[:], dest_Directory)

				// actualizamos el bitmap de bloques
				err = superblock.Update_Block_Bitmap(path)
				if err != nil {
					return err
				}

				// serializar el bloque
				err = new_folderblock.Serialize(path, int64(superblock.Sb_first_blo))
				if err != nil {
					return err
				}

				// actualizamos los campos del superbloque
				superblock.Sb_blocks_count++
				superblock.Sb_free_blocks_count--
				superblock.Sb_first_blo += superblock.Sb_block_size

				// Actualizamos el inode root
				inode.I_atime = float32(time.Now().Unix())

				// serializamos el inode root
				err = inode.Serialize(path, int64(superblock.Sb_block_start+(inode_Index*superblock.Sb_block_size)))
				if err != nil {
					return err
				}

				/*
					una vez creado el nuevo bloque y asignado al inode
					creamos el nuevo inode y lo guardaremos en el bloque
				*/

				new_Inode := &INODE{
					I_uid:   1,
					I_gid:   1,
					I_size:  0,
					I_atime: float32(time.Now().Unix()),
					I_ctime: float32(time.Now().Unix()),
					I_mtime: float32(time.Now().Unix()),
					I_block: [15]int32{superblock.Sb_blocks_count, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1},
					I_type:  [1]byte{'0'},
					I_perm:  [3]byte{'6', '6', '4'},
				}

				err = new_Inode.Serialize(path, int64(superblock.Sb_first_ino))
				if err != nil {
					return err
				}

				// actualizamos el bitmap de inodes
				err = superblock.Update_Inode_Bitmap(path)
				if err != nil {
					return err
				}

				// actualizamos los inodes del superblock
				superblock.Sb_inodes_count++
				superblock.Sb_free_inodes_count--
				superblock.Sb_first_ino += superblock.Sb_inode_size

				// creamos el bloque del inode recien creado

				newInode_folderblock := &FOLDERBLOCK{
					B_content: [4]FOLDERCONTENT{
						{B_name: [12]byte{'.'}, B_inodo: superblock.Sb_inodes_count - 1},
						{B_name: [12]byte{'.', '.'}, B_inodo: 0},
						{B_name: [12]byte{'-'}, B_inodo: -1},
						{B_name: [12]byte{'-'}, B_inodo: -1},
					},
				}

				// serializamos el bloque de la carpeta
				err = newInode_folderblock.Serialize(path, int64(superblock.Sb_first_blo))
				if err != nil {
					return err
				}

				// actualizar el bitmap de bloques
				err = superblock.Update_Block_Bitmap(path)
				if err != nil {
					return err
				}

				// actualizar el superbloque
				superblock.Sb_blocks_count++
				superblock.Sb_free_blocks_count--
				superblock.Sb_first_blo += superblock.Sb_block_size

				return nil

			}

			return nil
		}

		// se crea una instancia de floder block
		folderblock := &FOLDERBLOCK{}

		// deserializar el bloque segun el index
		err := folderblock.Deserialize(path, int64(superblock.Sb_block_start+(inode.I_block[i]*superblock.Sb_block_size)))
		if err != nil {
			return err
		}

		// iteramos sobre cada contenido del bloque, desde el index 2, porque los primeros dos indican al directorio padre y a si mismo
		for index_Content := 2; index_Content < len(folderblock.B_content); index_Content++ {

			// obtener el contenido del bloque
			content_block := folderblock.B_content[index_Content]

			// si las carpetas padre no estan vacias, se busca la carpeta mas cercana
			if len(parents_Directories) != 0 {

				// si el contenido esta vacio, se sale
				if content_block.B_inodo == -1 {
					break
				}

				// se obtiene la carpeta padre mas cercana
				first_dir := parents_Directories[0]

				// convertir B_name a string y eliminar los caracteres nulos
				content_name := strings.Trim(string(content_block.B_name[:]), "\x00")
				// convertir el directorio padre a string
				parent_dir_name := strings.Trim(first_dir, "\x00")
				// si el nombre del contenido coincide con el nombre de la carpeta
				if strings.EqualFold(content_name, parent_dir_name) {

					err := superblock.Create_Inode_as_Folder(path, content_block.B_inodo, util.Remove_At(parents_Directories, 0), dest_Directory, create_Parents)
					if err != nil {
						return err
					}
					return nil

				}

			} else {

				//fmt.Println("CREANDO CARPETA NUEVA")

				// si el apuntador al inode esta ocupado, continuamos al siguiente
				if content_block.B_inodo != -1 {
					continue
				}

				// actualizamos el contenido del bloque
				copy(content_block.B_name[:], dest_Directory)
				content_block.B_inodo = superblock.Sb_inodes_count

				// actualizamos el bloque
				folderblock.B_content[index_Content] = content_block

				// serializamos el bloque
				err = folderblock.Serialize(path, int64(superblock.Sb_block_start+(inode.I_block[i]*superblock.Sb_block_size)))
				if err != nil {
					return err
				}

				// creamos el inode de la carpeta
				folder_Inode := &INODE{

					I_uid:   1,
					I_gid:   1,
					I_size:  0,
					I_atime: float32(time.Now().Unix()),
					I_ctime: float32(time.Now().Unix()),
					I_mtime: float32(time.Now().Unix()),
					I_block: [15]int32{superblock.Sb_blocks_count, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1},
					I_type:  [1]byte{'0'},
					I_perm:  [3]byte{'6', '6', '4'},
				}

				err = folder_Inode.Serialize(path, int64(superblock.Sb_first_ino))
				if err != nil {
					return err
				}

				// actualizamos el bitmap de inodes
				err = superblock.Update_Inode_Bitmap(path)
				if err != nil {
					return err
				}

				// actualizamos el superblock
				superblock.Sb_inodes_count++
				superblock.Sb_free_inodes_count--
				superblock.Sb_first_ino += superblock.Sb_inode_size

				//creamos el bloque de la carpeta
				folderblock := &FOLDERBLOCK{
					B_content: [4]FOLDERCONTENT{
						{B_name: [12]byte{'.'}, B_inodo: content_block.B_inodo},
						{B_name: [12]byte{'.', '.'}, B_inodo: inode_Index},
						{B_name: [12]byte{'-'}, B_inodo: -1},
						{B_name: [12]byte{'-'}, B_inodo: -1},
					},
				}

				// serializamos el bloque de la carpeta
				err = folderblock.Serialize(path, int64(superblock.Sb_first_blo))
				if err != nil {
					return err
				}

				// actualizar el bitmap de bloques
				err = superblock.Update_Block_Bitmap(path)
				if err != nil {
					return err
				}

				// actualizar el superbloque
				superblock.Sb_blocks_count++
				superblock.Sb_free_blocks_count--
				superblock.Sb_first_blo += superblock.Sb_block_size

				return nil

			}

		}

	}

	return nil

}

// Funcion para crear bloques de carpetas
//func (superblock *SUPERBLOCK) Create_Block_as_Folder(path string, inode)
