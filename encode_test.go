package keys

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseEncoding(t *testing.T) {
	enc, err := ParseEncoding("base1000")
	require.EqualError(t, err, "invalid encoding base1000")
	require.Equal(t, NoEncoding, enc)

	enc2, err := ParseEncoding("base64")
	require.NoError(t, err)
	require.Equal(t, Base64, enc2)
}

func TestEncode(t *testing.T) {
	s := MustEncode([]byte("🤓"), Base62)
	require.Equal(t, "4PCobb", s)

	s = MustEncode([]byte("🤓"), Base64)
	require.Equal(t, "8J+kkw==", s)

	s = MustEncode([]byte("🤓🤓🤓🤓🤓"), Saltpack)
	require.Equal(t, "YKecp8NtwMvKIdy lDKcKhWX0nGV.", s)

	s = MustEncode(bytes.Repeat([]byte{0x01}, 32), BIP39)
	require.Equal(t, "absurd amount doctor acoustic avoid letter advice cage absurd amount doctor acoustic avoid letter advice cage absurd amount doctor acoustic avoid letter advice comic", s)

	s = MustEncode(bytes.Repeat([]byte{0x01}, 32), Hex)
	require.Equal(t, "0101010101010101010101010101010101010101010101010101010101010101", s)

	s = MustEncode(bytes.Repeat([]byte{0x01}, 32), Base58)
	require.Equal(t, "1BfGRZL7c75qu5bFwXXjWpmRmz15rJ1q6oLzUX9GJk2c", s)

	s = MustEncode([]byte("test"), Base58)
	require.Equal(t, "3yZe7d", s)
}

func TestIsASCII(t *testing.T) {
	ok := IsASCII([]byte("ok"))
	require.True(t, ok)

	ok2 := IsASCII([]byte{0xFF})
	require.False(t, ok2)
}

func TestDecode(t *testing.T) {
	b := []byte{0x01, 0x02, 0x03, 0x04}
	s := "AQIDBA=="
	bout, err := Decode(s, Base64)
	require.NoError(t, err)
	require.Equal(t, b, bout)

	bout, err = Decode("YKecp8NtwMvKIdy lDKcKhWX0nGV.", Saltpack)
	require.NoError(t, err)
	require.Equal(t, []byte("🤓🤓🤓🤓🤓"), bout)
}

func TestHasUpper(t *testing.T) {
	ok := hasUpper("ok")
	require.False(t, ok)

	ok2 := hasUpper("Ok")
	require.True(t, ok2)
}
