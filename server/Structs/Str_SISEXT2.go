package Structs

import "time"

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

	return nil

}
