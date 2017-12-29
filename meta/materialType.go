package meta

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"strings"
)

func createMaterialType(db *sql.DB, materialType *MaterialType) error {
	stm, err := db.Prepare("insert into materialtype(typename) values (?)")
	if err != nil {
		return err
	}
	rs, err := stm.Exec(materialType.Name)
	if err != nil {
		return err
	}
	stm.Close()
	return nil
}

func deleteMaterialType(db *sql.DB, typename string) error {
	stm, err := db.Prepare("delete from materialtype where typename=?")
	if err != nil {
		return err
	}
	rs, err := stm.Exec(typename)
	if err != nil {
		return err
	}
	stm.Close()
	return nil
}
