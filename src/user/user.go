package user

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"uuid"
)

const salt = "clytemNestra hoTcakces"

var users []*User
var usersByNickname map[string]*User
var usersById map[string]*User

type User struct {
	Id string 	`json:"id"`
	Nickname string `json:"nickname"`
	Pwhash string `json:"pwhash"`
}

func New(nickname string, password string) (*User, error) {

	id, err := uuid.GenUUID()
	if err != nil {
		return nil, err
	}

	u := &User {
		Id: id,
		Nickname: nickname,
		Pwhash: hashPassword(password)}
	return u, nil
}

func hashPassword(password string) string {

	h := sha1.New()
	io.WriteString(h, salt)
	io.WriteString(h, password)

	return fmt.Sprintf("%x", h.Sum(nil))
}

func loadUsers(filename string) error {
	var err error
	var f *os.File
	var fi os.FileInfo
	var b []byte

	fi, err = os.Stat(filename)
	if err != nil {
		return err
	}

	b = make([]byte, fi.Size())
	f, err = os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Read(b)
	if err != nil {
		return err
	}

	err = json.Unmarshal(b, &users)
	if err != nil {
		return err
	}

	return nil

}

func saveUsers(filename string) error {
	var b []byte
	var f *os.File
	var err error

	b, err = json.Marshal(users)
	if err != nil {
		return err
	}
	f, err = os.Create(filename)
	if err != nil {
		return err
	}
	_, err = f.Write(b)
	if err != nil {
		return err
	}
	return nil
}
