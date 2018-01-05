package meta

import (
	"database/sql"
	//db package
	_ "github.com/go-sql-driver/mysql"
)

func createMaterialType(db *sql.DB, materialType *MaterialType) error {
	stm, err := db.Prepare("insert into materialtype(typename) values (?)")
	if err != nil {
		return err
	}
	_, err = stm.Exec(materialType.Name)
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
	_, err = stm.Exec(typename)
	if err != nil {
		return err
	}
	stm.Close()
	return nil
}
