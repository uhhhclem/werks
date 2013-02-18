package user

import (
	"testing"
)

const salt = "semolina pilchard"
const filename = "users.json"

func TestNew(t *testing.T) {
	var u *User
	var err error

	Init(salt, filename)

	u, err = New("nickname", "password")
	if err != nil {
		t.Errorf("%s", err)
		return
	}
	if LookupByNickname("Nickname") != u {
		t.Errorf("LookupByNickname failed.")
	}
	if LookupById(u.Id) != u {
		t.Errorf("LookupById failed.")
	}

	u, err = New("nickname", "password")
	if err == nil {
		t.Errorf("Created user with a duplicate nickname.")
	}
}

func TestCanonicalizeNickname(t *testing.T) {
	input := []string {"a b", "C!99;d", "EFGH_"}
	expected := []string { "ab", "c99d", "efgh"}
	for i, _ := range input {
		a := canonicalizeNickname(input[i])
		e := expected[i]
		if a != e {
			t.Errorf("Expected: %s, actual: %s", e, a)
		}
	}
}

func TestSaveAndLoad(t *testing.T) {
	var nicknames = []string {"Alpha", "Beta", "Gamma", "Delta"}
	var passwords = []string {"Epsilon", "Omicron", "Omega", "Upsilon"}

	Init(salt, filename)

	for i, _ := range nicknames {
		New(nicknames[i], passwords[i])
	}
	for i, _ := range nicknames {
		if LookupByNickname(nicknames[i]) == nil {
			t.Errorf("LookupByNickname failed.")
			return
		}
	}
	err := SaveUsers()
	if err != nil {
		t.Errorf("%s", err)
	}

	Init(salt, filename)

	for i, _ := range nicknames {
		if LookupByNickname(nicknames[i]) != nil {
			t.Errorf("LookupByNickname shouldn't have found anything.")
			return
		}
	}

	err = LoadUsers()
	if err != nil {
		t.Errorf("%s", err)
	}

	for i, _ := range nicknames {
		if LookupByNickname(nicknames[i]) == nil {
			t.Errorf("LookupByNickname failed.")
			return
		}
	}

}
