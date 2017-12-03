package main

import (
	"database/sql"
	"fmt"
	"strconv"
	"time"

	_ "github.com/mattn/go-sqlite3"
	uuid "github.com/satori/go.uuid"
	"github.com/philhug/bitwarden-client-go/bitwarden"
)

type DB struct {
	db *sql.DB
}

func (db *DB) init() error {
	query1 := "CREATE TABLE \"accounts\" ( `id` INTEGER, `name` TEXT, `email` TEXT UNIQUE, `masterPasswordHash` NUMERIC, `masterPasswordHint` TEXT, `key` TEXT, 'refreshtoken' TEXT, PRIMARY KEY(id) );"
	query2 := "CREATE TABLE \"ciphers\" ( `id` INTEGER PRIMARY KEY AUTOINCREMENT, `type` INTEGER, `revisiondate` INTEGER, `data` BLOB, `owner` INTEGER );"
	query3 := "CREATE TABLE \"folders\" (`id`	TEXT,	`name`	TEXT,	`revisiondate`	INTEGER,	`owner`	INTEGER, PRIMARY KEY(id))"
	stmt1, err := db.db.Prepare(query1)
	if err != nil {
		return err
	}

	_, err = stmt1.Exec()
	if err != nil {
		return err
	}

	stmt2, err := db.db.Prepare(query2)
	if err != nil {
		return err
	}

	_, err = stmt2.Exec()
	if err != nil {
		return err
	}

	stmt3, err := db.db.Prepare(query3)
	if err != nil {
		return err
	}

	_, err = stmt3.Exec()
	if err != nil {
		return err
	}
	return err
}

func (db *DB) open() error {
	var err error
	db.db, err = sql.Open("sqlite3", "db")
	return err
}

func (db *DB) close() {
	db.db.Close()
}

func (db *DB) getCiphers(owner string) ([]bitwarden.Cipher, error) {
	iowner, err := strconv.ParseInt(owner, 10, 64)
	if err != nil {
		return nil, err
	}

	var ciphers []bitwarden.Cipher
	query := "SELECT id, type, revisiondate, data FROM ciphers WHERE owner = $1"
	rows, err := db.db.Query(query, iowner)

	var iid int
	var revDate int64
	var blob []byte
	for rows.Next() {
		ciph := bitwarden.Cipher{
			Favorite:            false,
			Edit:                true,
			OrganizationUseTotp: false,
			Attachments:         nil,
		}

		err := rows.Scan(&iid, &ciph.Type, &revDate, &blob)
		if err != nil {
			return nil, err
		}
		err = ciph.UnMarshalData(blob)
		if err != nil {
			return nil, err
		}
		ciph.Id = strconv.Itoa(iid)
		ciph.RevisionDate = &bitwarden.Time{time.Unix(revDate, 0)}

		ciphers = append(ciphers, ciph)
	}

	if len(ciphers) < 1 {
		ciphers = make([]bitwarden.Cipher, 0) // Make an empty slice if there are none or android app will crash
	}
	return ciphers, err
}

func (db *DB) newCipher(ciph bitwarden.Cipher, owner string) (bitwarden.Cipher, error) {
	iowner, err := strconv.ParseInt(owner, 10, 64)
	if err != nil {
		return bitwarden.Cipher{}, err
	}

	ciph.RevisionDate = &bitwarden.Time{time.Now()}

	stmt, err := db.db.Prepare("INSERT INTO ciphers(type, revisiondate, data, owner) values(?,?,?, ?)")
	if err != nil {
		return ciph, err
	}

	data, err := ciph.MarshalData()
	if err != nil {
		return ciph, err
	}

	res, err := stmt.Exec(ciph.Type, ciph.RevisionDate.Unix(), data, iowner)
	if err != nil {
		return ciph, err
	}

	lID, err := res.LastInsertId()
	ciph.Id = fmt.Sprintf("%v", lID)

	return ciph, nil

}

// Important to check that the owner is correct before an update!
func (db *DB) updateCipher(newData bitwarden.Cipher, owner string, ciphID string) error {
	iowner, err := strconv.ParseInt(owner, 10, 64)
	if err != nil {
		return err
	}

	iciphID, err := strconv.ParseInt(ciphID, 10, 64)
	if err != nil {
		return err
	}

	stmt, err := db.db.Prepare("UPDATE ciphers SET type=$1, revisiondate=$2, data=$3 WHERE id=$4 AND owner=$5")
	if err != nil {
		return err
	}

	bdata, err := newData.MarshalData()
	if err != nil {
		return err
	}

	_, err = stmt.Exec(newData.Type, time.Now().Unix(), bdata, iciphID, iowner)
	if err != nil {
		return err
	}

	return nil
}

// Important to check that the owner is correct before an update!
func (db *DB) deleteCipher(owner string, ciphID string) error {
	iowner, err := strconv.ParseInt(owner, 10, 64)
	if err != nil {
		return err
	}

	iciphID, err := strconv.ParseInt(ciphID, 10, 64)
	if err != nil {
		return err
	}

	stmt, err := db.db.Prepare("DELETE from ciphers WHERE id=$1 AND owner=$2")
	if err != nil {
		return err
	}

	_, err = stmt.Exec(iciphID, iowner)
	if err != nil {
		return err
	}
	return nil
}

func (db *DB) addAccount(acc bitwarden.Account) error {
	stmt, err := db.db.Prepare("INSERT INTO accounts(name, email, masterPasswordHash, masterPasswordHint, key, refreshtoken) values(?,?,?,?,?, ?)")
	if err != nil {
		return err
	}

	_, err = stmt.Exec(acc.Name, acc.Email, acc.MasterPasswordHash, acc.MasterPasswordHint, acc.Key, "")
	if err != nil {
		return err
	}

	return nil
}

func (db *DB) updateAccountInfo(acc bitwarden.Account) error {
	id, err := strconv.ParseInt(acc.Id, 10, 64)
	if err != nil {
		return err
	}

	stmt, err := db.db.Prepare("UPDATE accounts SET refreshtoken=$1 WHERE id=$2")
	if err != nil {
		return err
	}

	_, err = stmt.Exec(acc.RefreshToken, id)
	if err != nil {
		return err
	}

	return nil
}

func (db *DB) getAccount(username string) (bitwarden.Account, error) {
	var row *sql.Row
	acc := bitwarden.Account{}
	if username != "" {
		query := "SELECT * FROM accounts WHERE email = $1"
		row = db.db.QueryRow(query, username)
	}

	var iid int
	err := row.Scan(&iid, &acc.Name, &acc.Email, &acc.MasterPasswordHash, &acc.MasterPasswordHint, &acc.Key, &acc.RefreshToken)
	if err != nil {
		return acc, err
	}

	acc.Id = strconv.Itoa(iid)

	return acc, nil
}

func (db *DB) addFolder(name string, owner string) (bitwarden.Folder, error) {
	iowner, err := strconv.ParseInt(owner, 10, 64)
	if err != nil {
		return bitwarden.Folder{}, err
	}

	folder := bitwarden.Folder{
		Id:           uuid.NewV4().String(),
		Name:         name,
		Object:       "folder",
		RevisionDate: &bitwarden.Time{time.Now()},
	}

	stmt, err := db.db.Prepare("INSERT INTO folders(id, name, revisiondate, owner) values(?,?,?, ?)")
	if err != nil {
		return bitwarden.Folder{}, err
	}

	_, err = stmt.Exec(folder.Id, folder.Name, folder.RevisionDate.Unix(), iowner)
	if err != nil {
		return bitwarden.Folder{}, err
	}

	return folder, nil
}

func (db *DB) getFolders(owner string) ([]bitwarden.Folder, error) {
	iowner, err := strconv.ParseInt(owner, 10, 64)
	if err != nil {
		return nil, err
	}

	var folders []bitwarden.Folder
	query := "SELECT id, name, revisiondate FROM folders WHERE owner = $1"
	rows, err := db.db.Query(query, iowner)
	if err != nil {
		return nil, err
	}

	var revDate int64
	for rows.Next() {
		f := bitwarden.Folder{}
		err := rows.Scan(&f.Id, &f.Name, &revDate)
		if err != nil {
			return nil, err
		}
		f.RevisionDate = &bitwarden.Time{time.Unix(revDate, 0)}

		folders = append(folders, f)
	}

	if len(folders) < 1 {
		folders = make([]bitwarden.Folder, 0) // Make an empty slice if there are none or android app will crash
	}
	return folders, err
}
