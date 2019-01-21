package keystore

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"math/big"
	"syscall"

	"github.com/apex/log"
	jwt "github.com/hellais/jwt-go"
	"github.com/miekg/pkcs11"
	"github.com/thalesignite/crypto11"
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

// HSMConfig is the configuration for the Hardware Security Module
type HSMConfig struct {
	TokenSerial string
	UserPin     string
	SOPin       string
	LibPath     string
	KeyID       int
}

// KeyStore look bellow for methods to use with the KeyStore
type KeyStore struct {
	config *HSMConfig
}

// NewKeyStore creation
func NewKeyStore(config *HSMConfig) *KeyStore {
	return &KeyStore{
		config: config,
	}
}

// SessionContext stores the context, session pair
type SessionContext struct {
	Ctx     *pkcs11.Ctx
	Session pkcs11.SessionHandle
}

// CloseContext by destroying and finalizing the context
func (sc *SessionContext) CloseContext() {
	sc.Ctx.Destroy()
	sc.Ctx.Finalize()
}

// Close the context & the session
func (sc *SessionContext) Close() {
	sc.CloseContext()
	sc.Ctx.CloseSession(sc.Session)
}

// Logout of a session
func (sc *SessionContext) Logout() {
	sc.Ctx.Logout(sc.Session)
}

// MaybeLoginPrompt will show a login prompt only if the pin settings are unset
func (sc *SessionContext) MaybeLoginPrompt(config *HSMConfig, userFlag uint) error {
	log.Debug("MaybeloginPrompt")
	var pinStr string

	if userFlag == pkcs11.CKU_SO {
		pinStr = config.SOPin
	} else {
		pinStr = config.UserPin
	}

	if pinStr != "" {
		log.Debug("sc.Ctx.Login")
		return sc.Ctx.Login(sc.Session, userFlag, pinStr)
	}
	return sc.LoginPrompt(userFlag)
}

// LoginPrompt will show an interactive login prompt
func (sc *SessionContext) LoginPrompt(userFlag uint) error {
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
		err = sc.Ctx.Login(sc.Session, userFlag, string(pinBytes))
		if err == nil {
			fmt.Printf("\nLogged in as %s\n", pinType)
			return nil
		}
		fmt.Println("Wrong ping. Try again.")
	}
	return errors.New("Incorrect attempts exceeded")
}

// NewSessionContext creates a new session, context pair
func (ks *KeyStore) NewSessionContext() (*SessionContext, error) {
	sc := new(SessionContext)

	if ks.config.LibPath == "" {
		return sc, errors.New("libPath is empty")
	}
	sc.Ctx = pkcs11.New(ks.config.LibPath)
	if err := sc.Ctx.Initialize(); err != nil {
		return sc, fmt.Errorf("found library %s, but initialize error %s", ks.config.LibPath, err.Error())
	}
	slots, err := sc.Ctx.GetSlotList(true)
	if err != nil {
		defer sc.CloseContext()
		return sc, fmt.Errorf("loaded library %s, failed to list slots %s", ks.config.LibPath, err)
	}
	if len(slots) < 1 {
		defer sc.CloseContext()
		return sc, fmt.Errorf("loaded library %s, but no slots found", ks.config.LibPath)
	}

	log.Debug("Slot list")
	for _, slot := range slots {
		log.Debugf("- #%d", slot)
	}
	log.Debugf("Using slot #%d", slots[0])
	sc.Session, err = sc.Ctx.OpenSession(slots[0], pkcs11.CKF_SERIAL_SESSION|pkcs11.CKF_RW_SESSION)
	if err != nil {
		defer sc.Close()
		return sc, fmt.Errorf("loaded library, but failed to start session %s", err)
	}
	return sc, nil
}

// AddKey will add a private key and public key to the HSM token
func (ks *KeyStore) AddKey(privKey *rsa.PrivateKey, certBytes []byte) error {
	sc, err := ks.NewSessionContext()
	if err != nil {
		log.WithError(err).Error("Failed to create new session")
		return err
	}

	defer sc.Close()

	if err = sc.MaybeLoginPrompt(ks.config, pkcs11.CKU_SO); err != nil {
		return err
	}

	defer sc.Logout()

	// XXX check if the key is already on the token
	privTemplate := []*pkcs11.Attribute{
		// Taken from: http://docs.oasis-open.org/pkcs11/pkcs11-base/v2.40/os/pkcs11-base-v2.40-os.html#_toc416959720
		pkcs11.NewAttribute(pkcs11.CKA_CLASS, pkcs11.CKO_PRIVATE_KEY),
		pkcs11.NewAttribute(pkcs11.CKA_KEY_TYPE, pkcs11.CKK_RSA),
		pkcs11.NewAttribute(pkcs11.CKA_TOKEN, true),
		pkcs11.NewAttribute(pkcs11.CKA_LABEL, []byte("orchestrate-key")),

		pkcs11.NewAttribute(pkcs11.CKA_ID, ks.config.KeyID),
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
		pkcs11.NewAttribute(pkcs11.CKA_ID, ks.config.KeyID),
	}

	log.Debug("sc.Ctx.CreateObject")
	_, err = sc.Ctx.CreateObject(sc.Session, certTemplate)
	if err != nil {
		return fmt.Errorf("error importing: %v", err)
	}

	log.Debug("sc.Ctx.CreateObject(privTemplate)")
	_, err = sc.Ctx.CreateObject(sc.Session, privTemplate)
	if err != nil {
		return fmt.Errorf("error importing key: %v", err)
	}
	return nil
}

// ListKeys will list all the keys stored on the device
func (ks *KeyStore) ListKeys() error {
	fmt.Println("Listing keys")
	sc, err := ks.NewSessionContext()
	if err != nil {
		log.WithError(err).Error("Failed to create new session")
		return err
	}

	defer sc.Close()

	ctx := sc.Ctx
	session := sc.Session

	if err = sc.MaybeLoginPrompt(ks.config, pkcs11.CKU_SO); err != nil {
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
	fmt.Printf("Obj count: %d\n", len(objs))

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

// OrchestraClaims are the claims to be signed
type OrchestraClaims struct {
	ProbeCC  []string               `json:"probe_cc"`
	TestName string                 `json:"test_name"`
	Schedule string                 `json:"schedule"`
	Args     map[string]interface{} `json:"args"`
	jwt.StandardClaims
}

// SignClaims will sign some probe orchestration claims
func (ks *KeyStore) SignClaims(claims OrchestraClaims) (string, error) {
	config := &crypto11.PKCS11Config{
		Path:        ks.config.LibPath,
		Pin:         ks.config.UserPin,
		TokenSerial: ks.config.TokenSerial,
	}
	_, err := crypto11.Configure(config)
	if err != nil {
		log.WithError(err).Error("Failed to config")
		return "", err
	}

	// XXX I am not actually sure this is the correct ID to use
	key, err := crypto11.FindKeyPair([]byte(string(ks.config.KeyID)), nil)
	if err != nil {
		log.WithError(err).Error("Failed to find keypair")
		return "", err
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS512, claims)
	ss, err := token.SignedString(key)
	if err != nil {
		log.WithError(err).Error("Failed to sign")
		return "", err
	}
	return ss, nil
}
