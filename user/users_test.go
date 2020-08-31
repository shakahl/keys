package user_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/docs"
	"github.com/keys-pub/keys/request"
	"github.com/keys-pub/keys/tsutil"
	"github.com/keys-pub/keys/user"
	"github.com/stretchr/testify/require"
)

func TestCheckNoUsers(t *testing.T) {
	sk := keys.NewEdX25519KeyFromSeed(testSeed(0x01))
	sc := keys.NewSigchain(sk.ID())

	req := request.NewMockRequestor()
	clock := tsutil.NewTestClock()
	ds := docs.NewMem()
	scs := keys.NewSigchains(ds)
	users := user.NewUsers(ds, scs, user.Requestor(req), user.Clock(clock))

	res, err := users.CheckSigchain(context.TODO(), sc)
	require.NoError(t, err)
	require.Equal(t, 0, len(res))

	rk := keys.GenerateEdX25519Key()
	res, err = users.Update(context.TODO(), rk.ID())
	require.NoError(t, err)
	require.Equal(t, 0, len(res))
}

func TestCheckFailure(t *testing.T) {
	req := request.NewMockRequestor()
	clock := tsutil.NewTestClock()
	ds := docs.NewMem()
	scs := keys.NewSigchains(ds)
	users := user.NewUsers(ds, scs, user.Requestor(req), user.Clock(clock))

	msg := "BEGIN MESSAGE.HWNhu0mATP1TJvQ 2MsM6UREvrdpmJL mlr4taMzxi0olt7 nV35Vkco9gjJ3wyZ0z9hiq2OxrlFUT QVAdNgSZPX3TCKq 6Xr2MZHgg6PbuKB KKAcQRbMCMprx0eQ9AAmF37oSytfuD ekFhesy6sjWc4kJ XA4C6PAxTFwtO14 CEXTYQyBxGH2CYAsm4w2O9xq9TNTZw lo0e7ydqx99UXE8 Qivwr0VNs5.END MESSAGE."
	req.SetResponse("https://mobile.twitter.com/boboloblaw/status/1259188857846632448", []byte(msg))

	usr := &user.User{
		Name:    "gabriel",
		KID:     keys.ID("kex1d69g7mzjjn8cfm3ssdr9u8z8mh2d35cvjzsrwrndt4d006uhh69qyx2k5x"),
		Seq:     1,
		Service: "twitter",
		URL:     "https://twitter.com/boboloblaw/status/1259188857846632448",
	}
	result := users.RequestVerify(context.TODO(), usr)
	require.Equal(t, usr.Name, result.User.Name)
	require.Equal(t, result.Status, user.StatusFailure)
	require.Equal(t, result.Err, "path invalid (name mismatch) for url https://twitter.com/boboloblaw/status/1259188857846632448")
}

func TestSigchainUsersUpdate(t *testing.T) {
	// user.SetLogger(user.NewLogger(user.DebugLevel))
	// link.SetLogger(link.NewLogger(link.DebugLevel))

	b := []byte(`{".sig":"5NUJkMad0hNC6Xy3bJGmTHkaDjRIH6IWWpLdwf2qrrZI2NNEHb8+Hf4YxDgTcEA/Q5FsUJxkslrksVCRBYTEAw==","data":"eyJrIjoia2V4MWQ2OWc3bXpqam44Y2ZtM3NzZHI5dTh6OG1oMmQzNWN2anpzcndybmR0NGQwMDZ1aGg2OXF5eDJrNXgiLCJuIjoiZ2FicmllbCIsInNxIjoxLCJzciI6InR3aXR0ZXIiLCJ1IjoiaHR0cHM6Ly90d2l0dGVyLmNvbS9nYWJyaWVsL3N0YXR1cy8xMjU5MTg4ODU3ODQ2NjMyNDQ4In0=","kid":"kex1d69g7mzjjn8cfm3ssdr9u8z8mh2d35cvjzsrwrndt4d006uhh69qyx2k5x","seq":1,"ts":1589049007370,"type":"user"}`)

	kid := keys.ID("kex1d69g7mzjjn8cfm3ssdr9u8z8mh2d35cvjzsrwrndt4d006uhh69qyx2k5x")
	sc := keys.NewSigchain(kid)

	var st keys.Statement
	err := json.Unmarshal(b, &st)
	require.NoError(t, err)

	err = sc.Add(&st)
	require.NoError(t, err)

	clock := tsutil.NewTestClock()
	ds := docs.NewMem()
	scs := keys.NewSigchains(ds)
	req := request.NewMockRequestor()
	users := user.NewUsers(ds, scs, user.Requestor(req), user.Clock(clock))

	msg := "BEGIN MESSAGE.HWNhu0mATP1TJvQ 2MsM6UREvrdpmJL mlr4taMzxi0olt7 nV35Vkco9gjJ3wyZ0z9hiq2OxrlFUT QVAdNgSZPX3TCKq 6Xr2MZHgg6PbuKB KKAcQRbMCMprx0eQ9AAmF37oSytfuD ekFhesy6sjWc4kJ XA4C6PAxTFwtO14 CEXTYQyBxGH2CYAsm4w2O9xq9TNTZw lo0e7ydqx99UXE8 Qivwr0VNs5.END MESSAGE."
	req.SetResponse("https://mobile.twitter.com/gabriel/status/1259188857846632448", []byte(msg))

	err = scs.Save(sc)
	require.NoError(t, err)

	res, err := users.Update(context.TODO(), kid)
	require.NoError(t, err)
	require.Equal(t, 1, len(res))
	require.Equal(t, user.StatusOK, res[0].Status)
}

func TestSigchainRevokeUpdate(t *testing.T) {
	// user.SetLogger(user.NewLogger(user.DebugLevel))
	clock := tsutil.NewTestClock()
	ds := docs.NewMem()
	scs := keys.NewSigchains(ds)
	req := request.NewMockRequestor()
	users := user.NewUsers(ds, scs, user.Requestor(req), user.Clock(clock))

	sk := keys.GenerateEdX25519Key()
	kid := sk.ID()
	sc := keys.NewSigchain(kid)

	// User #1
	testTwitterSigchain(t, sk, "gabriel", sc, scs, req, clock)

	res, err := users.Update(context.TODO(), kid)
	require.NoError(t, err)
	require.Equal(t, 1, len(res))
	require.Equal(t, user.StatusOK, res[0].Status)

	// Revoke
	_, err = sc.Revoke(1, sk)
	require.NoError(t, err)
	err = scs.Save(sc)
	require.NoError(t, err)
	// Don't update here to test revoke + new statement updates correctly

	// Re-add user
	testTwitterSigchain(t, sk, "gabriel", sc, scs, req, clock)

	res, err = users.Update(context.TODO(), kid)
	require.NoError(t, err)
	require.Equal(t, 1, len(res))
	require.Equal(t, user.StatusOK, res[0].Status)
	require.Equal(t, "gabriel", res[0].User.Name)
	require.Equal(t, "twitter", res[0].User.Service)

	// User #2 (echo)
	testEchoSigchain(t, sk, "g", sc, scs, clock)

	res, err = users.Update(context.TODO(), kid)
	require.NoError(t, err)
	require.Equal(t, 2, len(res))
	require.Equal(t, user.StatusOK, res[0].Status)
	require.Equal(t, "gabriel", res[0].User.Name)
	require.Equal(t, "twitter", res[0].User.Service)
	require.Equal(t, user.StatusOK, res[1].Status)
	require.Equal(t, "g", res[1].User.Name)
	require.Equal(t, "echo", res[1].User.Service)
}

func TestCheckForExisting(t *testing.T) {
	var err error

	clock := tsutil.NewTestClock()
	req := request.NewMockRequestor()
	ds := docs.NewMem()
	scs := keys.NewSigchains(ds)
	users := user.NewUsers(ds, scs, user.Requestor(req), user.Clock(clock))

	sk1 := keys.NewEdX25519KeyFromSeed(testSeed(0x01))
	sc1 := keys.NewSigchain(sk1.ID())
	testEchoSigchain(t, sk1, "alice", sc1, scs, clock)
	kid, err := users.CheckForExisting(context.TODO(), sc1)
	require.NoError(t, err)
	require.Empty(t, kid)
	err = scs.Save(sc1)
	require.NoError(t, err)
	_, err = users.Update(context.TODO(), sk1.ID())
	require.NoError(t, err)

	sk2 := keys.NewEdX25519KeyFromSeed(testSeed(0x02))
	sc2 := keys.NewSigchain(sk2.ID())
	testEchoSigchain(t, sk2, "alice", sc2, scs, clock)
	kid, err = users.CheckForExisting(context.TODO(), sc2)
	require.NoError(t, err)
	require.Equal(t, kid, sk1.ID())

}
