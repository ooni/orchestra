package keystore

import (
	"bytes"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"syscall"

	"github.com/miekg/pkcs11"
	"golang.org/x/crypto/ssh/terminal"
)

const (
	// KeymodeNone means that no touch or PIN is required to sign with the yubikey
	KeymodeNone = 0
	// KeymodeTouch means that only touch is required to sign with the yubikey
	KeymodeTouch = 1
	// KeymodePinOnce means that the pin entry is required once the first time to sign with the yubikey
	KeymodePinOnce = 2
	// KeymodePinAlways means that pin entry is required every time to sign with the yubikey
	KeymodePinAlways = 4
)

var (
	yubikeyKeymode = KeymodeTouch | KeymodePinAlways
)

// SetupHSM will initialize a pkcs11 context based on the path to the pkcs11
// library
func SetupHSM(libPath string) (*pkcs11.Ctx, pkcs11.SessionHandle, error) {
	if libPath == "" {
		return nil, 0, errors.New("libPath is empty")
	}
	p := pkcs11.New(libPath)
	if err := p.Initialize(); err != nil {
		return nil, 0, fmt.Errorf("found library %s, but initialize error %s", libPath, err.Error())
	}

	slots, err := p.GetSlotList(true)
	if err != nil {
		defer p.Destroy()
		defer p.Finalize()
		return nil, 0, fmt.Errorf("loaded library %s, failed to list slots %s", libPath, err)
	}
	if len(slots) < 1 {
		defer p.Destroy()
		defer p.Finalize()
		return nil, 0, fmt.Errorf("loaded library %s, but no slots found", libPath)
	}

	fmt.Printf("Slot list: \n")
	for _, slot := range slots {
		fmt.Printf("- #%d\n", slot)
	}
	fmt.Printf("Using slot #%d\n", slots[0])
	session, err := p.OpenSession(slots[0], pkcs11.CKF_SERIAL_SESSION|pkcs11.CKF_RW_SESSION)
	if err != nil {
		defer p.Destroy()
		defer p.Finalize()
		defer p.CloseSession(session)
		return nil, 0, fmt.Errorf("loaded library, but failed to start session %s", err)
	}
	return p, session, nil
}

// LoginPrompt will show an interactive login prompt to receive the HSM pin
func LoginPrompt(ctx *pkcs11.Ctx, session pkcs11.SessionHandle, userFlag uint) error {
	var (
		pinType string
	)
	maxAttempts := 2
	if userFlag == pkcs11.CKU_SO {
		pinType = "SO"
	} else {
		pinType = "User"
	}
	for attempts := 0; attempts <= maxAttempts; attempts++ {
		fmt.Printf("Enter your %s pin: ", pinType)
		pinBytes, err := terminal.ReadPassword(int(syscall.Stdin))
		if err != nil {
			return err
		}
		err = ctx.Login(session, userFlag, string(pinBytes))
		if err == nil {
			fmt.Printf("\nLogged in as %s\n", pinType)
			return nil
		}
		fmt.Println("Wrong ping. Try again.")
	}
	return errors.New("Incorrect attempts exceeded")
}

func encodeByteSlice(in ...[]byte) []byte {
	l := 0
	for _, v := range in {
		l += len(v)
	}
	if l > 4294967295 {
		panic(fmt.Errorf("input byte slice is too long"))
	}
	out := make([]byte, 4+l)
	binary.BigEndian.PutUint32(out, uint32(l))

	start := 4 + copy(out[4:], in[0])
	if len(in) > 1 {
		for _, v := range in[1:] {
			copy(out[start:], v)
		}
	}
	return out
}

func getKeyID(privKey *rsa.PrivateKey) (string, error) {
	key, ok := privKey.Public().(*rsa.PublicKey)
	if !ok {
		return "", errors.New("unsupported type")
	}
	buf := bytes.NewBuffer(nil)
	buf.Write(encodeByteSlice([]byte("rsa")))
	e := make([]byte, 4)
	binary.BigEndian.PutUint32(e, uint32(key.E))
	buf.Write(encodeByteSlice(bytes.TrimLeft(e, "\x00")))
	buf.Write(encodeByteSlice([]byte{0}, key.N.Bytes()))

	digest := sha256.Sum256(buf.Bytes())
	keyID := hex.EncodeToString(digest[:])

	return keyID, nil
}

// AddKey will add a private key to the device
func AddKey(libPath string, privKey *rsa.PrivateKey, certBytes []byte) error {
	/*
		keyID, err := getKeyID(privKey)
		if err != nil {
			return err
		}
	*/
	keyID := 11

	ctx, session, err := SetupHSM(libPath)
	if err != nil {
		return err
	}
	defer ctx.Destroy()
	defer ctx.Finalize()
	defer ctx.CloseSession(session)

	if err = LoginPrompt(ctx, session, pkcs11.CKU_SO); err != nil {
		return err
	}

	defer ctx.Logout(session)
	// XXX check if the key is already on the token

	privTemplate := []*pkcs11.Attribute{
		// Taken from: http://docs.oasis-open.org/pkcs11/pkcs11-base/v2.40/os/pkcs11-base-v2.40-os.html#_toc416959720
		pkcs11.NewAttribute(pkcs11.CKA_CLASS, pkcs11.CKO_PRIVATE_KEY),
		pkcs11.NewAttribute(pkcs11.CKA_KEY_TYPE, pkcs11.CKK_RSA),
		pkcs11.NewAttribute(pkcs11.CKA_TOKEN, true),
		pkcs11.NewAttribute(pkcs11.CKA_LABEL, []byte("orchestrate-key")),

		pkcs11.NewAttribute(pkcs11.CKA_ID, keyID),
		pkcs11.NewAttribute(pkcs11.CKA_PUBLIC_EXPONENT, big.NewInt(int64(privKey.PublicKey.E)).Bytes()), // XXX this is a big ghetto
		pkcs11.NewAttribute(pkcs11.CKA_PRIME_1, privKey.Primes[0].Bytes()),
		pkcs11.NewAttribute(pkcs11.CKA_PRIME_2, privKey.Primes[1].Bytes()),
		pkcs11.NewAttribute(pkcs11.CKA_EXPONENT_1, privKey.Precomputed.Dp.Bytes()),
		pkcs11.NewAttribute(pkcs11.CKA_EXPONENT_2, privKey.Precomputed.Dq.Bytes()),
		pkcs11.NewAttribute(pkcs11.CKA_COEFFICIENT, privKey.Precomputed.Qinv.Bytes()),
		pkcs11.NewAttribute(pkcs11.CKA_VENDOR_DEFINED, yubikeyKeymode),
	}
	certTemplate := []*pkcs11.Attribute{
		pkcs11.NewAttribute(pkcs11.CKA_CLASS, pkcs11.CKO_CERTIFICATE),
		pkcs11.NewAttribute(pkcs11.CKA_VALUE, certBytes),
		pkcs11.NewAttribute(pkcs11.CKA_ID, keyID),
	}
	_, err = ctx.CreateObject(session, certTemplate)
	if err != nil {
		return fmt.Errorf("error importing: %v", err)
	}

	_, err = ctx.CreateObject(session, privTemplate)
	if err != nil {
		return fmt.Errorf("error importing key: %v", err)
	}
	return nil
}

// ListKeys will list all the keys on the device
func ListKeys(libPath string) error {
	fmt.Println("Listing keys")
	ctx, session, err := SetupHSM(libPath)
	if err != nil {
		return err
	}
	defer ctx.Destroy()
	defer ctx.Finalize()
	defer ctx.CloseSession(session)

	if err = LoginPrompt(ctx, session, pkcs11.CKU_SO); err != nil {
		return err
	}

	findTemplate := []*pkcs11.Attribute{
		pkcs11.NewAttribute(pkcs11.CKA_KEY_TYPE, pkcs11.CKK_RSA),
	}
	attrTemplate := []*pkcs11.Attribute{
		pkcs11.NewAttribute(pkcs11.CKA_KEY_TYPE, []byte{0}),
	}

	if err = ctx.FindObjectsInit(session, findTemplate); err != nil {
		return err
	}
	objs, _, err := ctx.FindObjects(session, 100)
	if err != nil {
		return err
	}

	if err = ctx.FindObjectsFinal(session); err != nil {
		return err
	}
	fmt.Printf("Len: %d\n", len(objs))
	if len(objs) != 1 {
		fmt.Printf("Len: %d\n", len(objs))
	}

	attr, err := ctx.GetAttributeValue(session, objs[0], attrTemplate)
	if err != nil {
		fmt.Printf("Failed to get Attribute for: %v (%v)\n", objs[0], err)
		return err
	}
	for _, a := range attr {
		fmt.Printf("%v: %v\n", a.Type, a.Value)
	}

	return nil
}
