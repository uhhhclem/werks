package user

import (
	"testing"
)

const salt = "semolina pilchard"
const filename = "users.json"

func TestNew(t *testing.T) {
	var u *User
	var err error

	s := Init(salt, filename)

	u, err = s.New("nickname", "password")
	if err != nil {
		t.Errorf("%s", err)
		return
	}
	if s.LookupByNickname("Nickname") != u {
		t.Errorf("LookupByNickname failed.")
	}
	if s.LookupById(u.Id) != u {
		t.Errorf("LookupById failed.")
	}

	u, err = s.New("nickname", "password")
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

func TestCanonicalizePassword(t *testing.T) {
	input := []string {"a b", "  C!99;d", "EF G H_ "}
	expected := []string { "ab", "C!99;d", "EFGH_"}
	for i, _ := range input {
		a := canonicalizePassword(input[i])
		e := expected[i]
		if a != e {
			t.Errorf("Expected: %s, actual: %s", e, a)
		}
	}

}

func TestLogin(t *testing.T) {
	var nicknames = []string {"Alpha", "Beta", "Gamma", "Delta"}
	var passwords = []string {"Epsilon", "Omicron", "Omega", "Upsilon"}

	s := Init(salt, filename)

	for i, _ := range nicknames {
		s.New(nicknames[i], passwords[i])
	}

	var user *User
	var err error

	user, err = s.Login("..Alpha", "Epsilon")
	if user == nil {
		if err == nil {
			t.Errorf("Should have gotten an error.")
		} else {
			t.Errorf("%s", err)
		}
		return
	}

	user, err = s.Login("  alpha  ", "Slipshod")
	if user != nil {
		t.Errorf("Login should have failed.");
		return
	}
	if err != InvalidPasswordError {
		t.Errorf("%s", err)
		return
	}

	user, err = s.Login("Bogus", "Slipshod")
	if user != nil {
		t.Errorf("Login should have failed.");
		return
	}
	if err != InvalidNicknameError {
		t.Errorf("%s", err)
		return
	}

}

func TestSaveAndLoad(t *testing.T) {
	var nicknames = []string {"Alpha", "Beta", "Gamma", "Delta"}
	var passwords = []string {"Epsilon", "Omicron", "Omega", "Upsilon"}

	s := Init(salt, filename)

	for i, _ := range nicknames {
		s.New(nicknames[i], passwords[i])
	}
	for i, _ := range nicknames {
		if s.LookupByNickname(nicknames[i]) == nil {
			t.Errorf("LookupByNickname failed.")
			return
		}
	}
	err := s.SaveUsers()
	if err != nil {
		t.Errorf("%s", err)
	}

	s = Init(salt, filename)

	for i, _ := range nicknames {
		if s.LookupByNickname(nicknames[i]) != nil {
			t.Errorf("LookupByNickname shouldn't have found anything.")
			return
		}
	}

	err = s.LoadUsers()
	if err != nil {
		t.Errorf("%s", err)
	}

	for i, _ := range nicknames {
		if s.LookupByNickname(nicknames[i]) == nil {
			t.Errorf("LookupByNickname failed.")
			return
		}
	}

}
