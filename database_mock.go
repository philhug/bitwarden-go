package main

import (
	_ "github.com/mattn/go-sqlite3"
	"github.com/philhug/bitwarden-client-go/bitwarden"
)

// mock database used for testing
type mockDB struct {
	username     string
	password     string
	refreshToken string
}

func (db *mockDB) init() error {
	return nil
}

func (db *mockDB) open() error {
	return nil
}

func (db *mockDB) close() {
}

func (db *mockDB) updateAccountInfo(acc bitwarden.Account) error {
	return nil
}

func (db *mockDB) getCiphers(owner string) ([]bitwarden.Cipher, error) {
	return nil, nil
}

func (db *mockDB) newCipher(ciph bitwarden.Cipher, owner string) (bitwarden.Cipher, error) {
	return bitwarden.Cipher{}, nil

}

func (db *mockDB) updateCipher(newData bitwarden.Cipher, owner string, ciphID string) error {
	return nil
}

func (db *mockDB) deleteCipher(owner string, ciphID string) error {
	return nil
}

func (db *mockDB) addAccount(acc bitwarden.Account) error {
	return nil
}

func (db *mockDB) getAccount(username string) (bitwarden.Account, error) {
	return bitwarden.Account{Email: db.username, MasterPasswordHash: db.password, RefreshToken: db.refreshToken}, nil
}

func (db *mockDB) addFolder(name string, owner string) (bitwarden.Folder, error) {
	return bitwarden.Folder{}, nil
}

func (db *mockDB) getFolders(owner string) ([]bitwarden.Folder, error) {
	return nil, nil
}
