package meta

import (
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"strings"
	//db package
	_ "github.com/go-sql-driver/mysql"
)

//StoreConfig : db configuration info
type StoreConfig struct {
	StoreType       string `json:"storeType"`
	MysqlDatasource string `json:"dataSourceName,omitempty"`
}

func testDBConnect() {
	db, err := connectMySQLDB()
	if err != nil {
		panic(err.Error())
	}
	var materialType *MaterialType
	materialType.Name = "Potato"
	createMaterialType(db, materialType)
}

func connectMySQLDB() (*sql.DB, error) {
	storeConfig, _ := readStoreConfig("./mysqlStoreConfig.json")
	db, err := sql.Open(storeConfig.StoreType, storeConfig.MysqlDatasource)
	if err != nil {
		return nil, err
	}
	return db, err
}

func readStoreConfig(configFileLocation string) (*StoreConfig, error) {
	var storeConfig StoreConfig
	raw, err := ioutil.ReadFile(configFileLocation)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(raw, &storeConfig)
	return &storeConfig, err
}

func createMaterial(db *sql.DB, material *Material) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	//get aliasid, if not exist in alias table, insert
	row := db.QueryRow("select id from alias where materialname=?", material.Name)
	var aliasID int64 = -1
	err = row.Scan(&aliasID)
	if err != nil {
		if err == sql.ErrNoRows {
			//insert into alias
			stm, err := tx.Prepare("insert into alias (materialname) values (?)")
			if err != nil {
				return err
			}
			defer stm.Close()
			rs, err := stm.Exec(material.Name)
			if err != nil {
				return err
			}
			//get aliasid which was just inserted
			aliasID, err = rs.LastInsertId()
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}
	//insert alias into alias table
	allAlias := strings.Split(material.Alias, ";")
	for i := 0; i < len(allAlias); i++ {
		row := db.QueryRow("select id from alias where materialname=?", allAlias[i])
		var tmp int
		err := row.Scan(&tmp)
		if err != nil {
			if err == sql.ErrNoRows {
				//insert into alias
				stm, err := tx.Prepare("insert into alias (materialname) values (?)")
				if err != nil {
					return err
				}
				defer stm.Close()
				_, err = stm.Exec(allAlias[i])
				if err != nil {
					return err
				}
			}
		}
	}

	//get typeid
	row = db.QueryRow("select id from materialtype where typename=?", material.Type)
	var typeID int64 = -1
	err = row.Scan(&typeID)
	if err != nil {
		return err
	}

	//inset into material
	stm, err := tx.Prepare("insert into material (aliasid, typeid, description) values (?, ?, ?)")
	if err != nil {
		return err
	}
	defer stm.Close()
	_, err = stm.Exec(aliasID, typeID, material.Description)
	if err != nil {
		return err
	}

	//insert into tagsofmaterial
	if len(material.Tags) != 0 {
		allTag := strings.Split(material.Tags, ";")
		for i := 0; i < len(allTag); i++ {
			//get tagid
			row := db.QueryRow("select id from tag where name=?", allTag[i])
			var tagID int64 = -1
			err := row.Scan(&tagID)
			if err != nil {
				return err
			}
			//insert into tagsofmaterial
			stm, err := tx.Prepare("insert into tagsofmaterial (aliasid, tagid) values (?, ?)")
			if err != nil {
				return err
			}
			defer stm.Close()
			_, err = stm.Exec(aliasID, tagID)
			if err != nil {
				return err
			}
		}
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

func updateMaterial(db *sql.DB, material *Material) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	//get typeid
	row := db.QueryRow("select id from materialtype where typename=?", material.Type)
	var typeID int64 = -1
	err = row.Scan(&typeID)
	if err != nil {
		return err
	}
	//get aliasid
	row = db.QueryRow("select id from alias where materialname=?", material.Name)
	var aliasID int64 = -1
	err = row.Scan(&aliasID)
	if err != nil {
		return err
	}
	//update material
	stm, err := tx.Prepare("update material set description=?, typeid=? where aliasid=?")
	if err != nil {
		return err
	}
	defer stm.Close()
	_, err = stm.Exec(material.Description, typeID, aliasID)
	if err != nil {
		return err
	}
	//update tagsofmaterial,delete existing records, add new ones
	//delete
	stm, err = tx.Prepare("delete from tagsofmaterial where aliasid=?")
	if err != nil {
		return err
	}
	defer stm.Close()
	_, err = stm.Exec(aliasID)
	if err != nil {
		return err
	}
	//add
	if len(material.Tags) != 0 {
		allTag := strings.Split(material.Tags, ";")
		for i := 0; i < len(allTag); i++ {
			//get tagid
			row := db.QueryRow("select id from tag where name=?", allTag[i])
			var tagID int64 = -1
			err := row.Scan(&tagID)
			if err != nil {
				return err
			}
			//insert into tagsofmaterial
			stm, err := tx.Prepare("insert into tagsofmaterial (aliasid, tagid) values (?, ?)")
			if err != nil {
				return err
			}
			defer stm.Close()
			_, err = stm.Exec(aliasID, tagID)
			if err != nil {
				return err
			}
		}
	}

	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

func deleteMaterial(db *sql.DB, materialName string) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	//get aliasid
	row := db.QueryRow("select id from alias where materialname=?", materialName)
	var aliasID int64 = -1
	err = row.Scan(&aliasID)
	if err != nil {
		return err
	}
	//delete material
	stm, err := tx.Prepare("delete from material where aliasid=?")
	if err != nil {
		return err
	}
	defer stm.Close()
	_, err = stm.Exec(aliasID)
	if err != nil {
		return err
	}
	//delete tagofmaterial
	stm, err = tx.Prepare("delete from tagsofmaterial where aliasid=?")
	if err != nil {
		return err
	}
	defer stm.Close()
	_, err = stm.Exec(aliasID)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

func getMaterial(db *sql.DB, materialName string) (*Material, error) {
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	//get  type & description
	row := db.QueryRow("select material.description, materialtype.typename  from material, materialtype, alias where alias.materialname=? and alias.id=material.aliasid and material.typeid=materialtype.id", materialName)
	var description string
	var typename string
	err = row.Scan(&description, &typename)
	if err != nil {
		return nil, err
	}
	//get all alias
	rows, _ := db.Query("select materialname from alias where id=(select id from alias where materialname=?)", materialName)
	defer rows.Close()
	var allalias string
	for rows.Next() {
		var aliastmp string
		err = rows.Scan(&aliastmp)
		if err != nil {
			return nil, err
		}
		if strings.Compare(materialName, aliastmp) != 0 {
			if allalias == "" {
				allalias = aliastmp
			} else {
				allalias += ";" + aliastmp
			}
		}
	}
	//get all tag
	rows, _ = db.Query("select tag.name from tag,tagsofmaterial,alias where alias.materialname=? and alias.id=tagsofmaterial.aliasid and tagsofmaterial.tagid=tag.id", materialName)
	defer rows.Close()
	var alltag string
	for rows.Next() {
		var tagtmp string
		err = rows.Scan(&tagtmp)
		if err != nil {
			return nil, err
		}
		if alltag == "" {
			alltag = tagtmp
		} else {
			alltag += ";" + tagtmp
		}
	}
	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	var material Material
	material.Description = description
	material.Type = typename
	material.Name = materialName
	material.Tags = alltag
	material.Alias = allalias
	return &material, err
}
