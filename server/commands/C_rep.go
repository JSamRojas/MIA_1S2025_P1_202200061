package commands

import (
	"errors"
	"fmt"
	"regexp"
	global "server/global"
	reportes "server/reports"
	"strings"
)

type REP struct {
	Name      string
	Path      string
	Id        string
	Path_file string
}

func Rep_Command(tokens []string) (*REP, string, error) {

	reporte := &REP{}

	atributos := strings.Join(tokens, " ")

	lexic := regexp.MustCompile(`(?i)-id=[^\s]+|(?i)-path="[^"]+"|(?i)-path=[^\s]+|(?i)-name=[^\s]+|(?i)-path_file_ls="[^"]+"|(?i)-path_file_ls=[^\s]+`)

	found := lexic.FindAllString(atributos, -1)

	for _, fun := range found {

		parametro := strings.SplitN(fun, "=", 2)
		if len(parametro) != 2 {
			return nil, "ERROR: Parametro invalido", fmt.Errorf(("parametro invalido: %s"), fun)
		}

		key, value := strings.ToLower(parametro[0]), parametro[1]

		if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
			value = strings.Trim(value, "\"")
		}

		switch key {

		case "-name":

			validNames := []string{"mbr", "disk", "inode", "block", "bm_inode", "bm_block", "sb", "file", "ls"}

			if !contains(validNames, value) {
				return nil, "ERROR: nombre de reporte invalido: " + value, errors.New("nombre de reporte invalido: " + value)
			}
			reporte.Name = value

		case "-id":

			if value == "" {
				return nil, "ERROR: id invalido, es de caracter obligatorio", errors.New("id invalido, es de caracter obligatorio")
			}
			reporte.Id = value

		case "-path":

			if value == "" {
				return nil, "ERROR: path invalido, es de caracter obligatorio", errors.New("path invalido, es de caracter obligatorio")
			}
			reporte.Path = value

		case "path_file_ls":

			reporte.Path_file = value

		default:
			return nil, "ERROR: Parametro invalido", fmt.Errorf(("parametro invalido: %s"), key)
		}
	}

	if reporte.Name == "" || reporte.Id == "" || reporte.Path == "" {
		return nil, "ERROR: Faltan parametros obligatorios", errors.New("faltan parametros obligatorios")
	}

	msg, err := Get_type_report(reporte)
	if err != nil {
		return nil, msg, err
	}

	return reporte, "COMANDO REP: Reporte realizado con exito", nil

}

func contains(list []string, value string) bool {
	for _, v := range list {
		if v == value {
			return true
		}
	}
	return false
}

func Get_type_report(reporte *REP) (string, error) {

	mbrREP, sbREP, diskpathREP, err := global.Get_essential_rep(reporte.Id)
	if err != nil {
		return "", err
	}

	switch reporte.Name {

	case "disk":

		err = reportes.ReporteDISK(mbrREP, reporte.Path, diskpathREP)
		if err != nil {
			fmt.Printf("ERROR: %s\n", err)
		}

	case "mbr":

		err = reportes.ReporteMBR(mbrREP, reporte.Path, diskpathREP)
		if err != nil {
			fmt.Printf("ERROR: %s\n", err)
		}

	case "inode":
		err = reportes.ReporteINODE(sbREP, diskpathREP, reporte.Path)
		if err != nil {
			fmt.Printf("ERROR: %s\n", err)
		}

	}

	return "", nil

}
