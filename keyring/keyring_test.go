package keyring

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestKeyring(t *testing.T) {
	kr, err := NewKeyring("Keys.pubTest")
	require.NoError(t, err)
	defer func() { _ = kr.Reset() }()

	testKeyring(t, kr)
}

func testKeyring(t *testing.T, kr Keyring) {
	salt := bytes.Repeat([]byte{0x01}, 32)
	auth, err := NewPasswordAuth("password123", salt)
	require.NoError(t, err)

	err = kr.Unlock(auth)
	require.NoError(t, err)

	item, err := kr.Get("abc")
	require.NoError(t, err)
	require.Nil(t, item)

	// Set/Get "abc"
	err = kr.Set(NewItem("abc", NewStringSecret("password"), "type1"))
	require.NoError(t, err)

	item, err = kr.Get("abc")
	require.NoError(t, err)
	require.NotNil(t, item)
	require.Equal(t, "abc", item.ID)
	require.Equal(t, []byte("password"), item.Secret().Data)

	has, err := kr.Exists("abc")
	require.NoError(t, err)
	require.True(t, has)

	has2, err := kr.Exists("xyz")
	require.NoError(t, err)
	require.False(t, has2)

	// Set (update)
	err = kr.Set(NewItem("abc", NewStringSecret("newpassword"), ""))
	require.NoError(t, err)

	item, err = kr.Get("abc")
	require.NoError(t, err)
	require.NotNil(t, item)
	require.Equal(t, "abc", item.ID)
	require.Equal(t, []byte("newpassword"), item.Secret().Data)

	// Set (hidden)
	err = kr.Set(NewItem(".ck", NewStringSecret("password"), ""))
	require.NoError(t, err)

	// Set "xyz"
	err = kr.Set(NewItem("xyz", NewStringSecret("xpassword"), "type2"))
	require.NoError(t, err)

	// List
	items, err := kr.List(nil)
	require.NoError(t, err)
	require.Equal(t, 2, len(items))
	require.Equal(t, items[0].ID, "abc")
	require.Equal(t, items[1].ID, "xyz")

	items2, err := kr.List(&ListOpts{Type: "type2"})
	require.NoError(t, err)
	require.Equal(t, 1, len(items2))
	require.Equal(t, items2[0].ID, "xyz")

	// IDs
	ids, err := kr.IDs("")
	require.NoError(t, err)
	require.Equal(t, 2, len(ids))
	require.Equal(t, ids[0], "abc")
	require.Equal(t, ids[1], "xyz")

	ids2, err := kr.IDs("a")
	require.NoError(t, err)
	require.Equal(t, 1, len(ids2))
	require.Equal(t, ids2[0], "abc")

	// Delete
	ok, err := kr.Delete("abc")
	require.NoError(t, err)
	require.True(t, ok)

	item3, err := kr.Get("abc")
	require.NoError(t, err)
	require.Nil(t, item3)

	has3, err := kr.Exists("abc")
	require.NoError(t, err)
	require.False(t, has3)

	ok2, err := kr.Delete("abc")
	require.NoError(t, err)
	require.False(t, ok2)
}

func TestReset(t *testing.T) {
	kr, err := NewKeyring("Keys.pubTest")
	require.NoError(t, err)
	defer func() { _ = kr.Reset() }()

	testReset(t, kr)
}

func testReset(t *testing.T, kr Keyring) {
	salt := bytes.Repeat([]byte{0x01}, 32)
	auth, err := NewPasswordAuth("password123", salt)
	require.NoError(t, err)

	err = kr.Unlock(auth)
	require.NoError(t, err)

	err = kr.Set(NewItem("key1", NewStringSecret("password"), ""))
	require.NoError(t, err)

	salt, err = kr.Salt()
	require.NoError(t, err)

	salt2, err := kr.Salt()
	require.NoError(t, err)
	require.Equal(t, salt, salt2)

	reerr := kr.Reset()
	require.NoError(t, reerr)

	err = kr.Set(NewItem("key1", NewStringSecret("password"), ""))
	require.EqualError(t, err, "keyring is locked")

	authed, err := kr.Authed()
	require.NoError(t, err)
	require.False(t, authed)

	salt3, err := kr.Salt()
	require.NoError(t, err)
	require.NotEqual(t, salt, salt3)

	auth2, err := NewPasswordAuth("newpassword123", salt)
	require.NoError(t, err)
	err = kr.Unlock(auth2)
	require.NoError(t, err)

	require.NotEqual(t, auth.Key(), auth2.Key())
}

func TestUnlock(t *testing.T) {
	kr, err := NewKeyring("Keys.pubTest")
	require.NoError(t, err)
	defer func() { _ = kr.Reset() }()
	testUnlock(t, kr)
}

func testUnlock(t *testing.T, kr Keyring) {
	err := kr.Set(NewItem("key1", NewStringSecret("password"), ""))
	require.EqualError(t, err, "keyring is locked")

	_, err = kr.Get("key1")
	require.EqualError(t, err, "keyring is locked")

	_, err = kr.List(nil)
	require.EqualError(t, err, "keyring is locked")

	key := bytes32(bytes.Repeat([]byte{0x01}, 32))
	err = kr.Unlock(NewKeyAuth(key))
	require.NoError(t, err)

	err = kr.Set(NewItem("key1", NewStringSecret("password"), ""))
	require.NoError(t, err)

	err = kr.Lock()
	require.NoError(t, err)

	_, err = kr.List(nil)
	require.EqualError(t, err, "keyring is locked")

	ids, err := kr.IDs("")
	require.NoError(t, err)
	require.Equal(t, []string{"key1"}, ids)
}

func TestSetErrors(t *testing.T) {
	kr, err := NewKeyring("Keys.pubTest")
	require.NoError(t, err)
	defer func() { _ = kr.Reset() }()
	key := bytes32(bytes.Repeat([]byte{0x01}, 32))
	err = kr.Unlock(NewKeyAuth(key))
	require.NoError(t, err)

	err = kr.Set(NewItem("", NewSecret(nil), ""))
	require.EqualError(t, err, "no id")
}

func TestReserved(t *testing.T) {
	kr, err := NewKeyring("Keys.pubTest")
	require.NoError(t, err)
	defer func() { _ = kr.Reset() }()
	testReserved(t, kr)
}

func testReserved(t *testing.T, kr Keyring) {
	key := bytes32(bytes.Repeat([]byte{0x01}, 32))
	err := kr.Unlock(NewKeyAuth(key))
	require.NoError(t, err)

	_, err = kr.Get("#key")
	require.EqualError(t, err, "keyring id prefix reserved #key")
	_, err = kr.Get("#salt")
	require.EqualError(t, err, "keyring id prefix reserved #salt")

	err = kr.Set(NewItem("#key", NewSecret(nil), ""))
	require.EqualError(t, err, "keyring id prefix reserved #key")
	err = kr.Set(NewItem("#salt", NewSecret(nil), ""))
	require.EqualError(t, err, "keyring id prefix reserved #salt")
}