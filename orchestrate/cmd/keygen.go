package cmd

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/ooni/orchestra/orchestrate/orchestrate/keystore"
)

const PKS11LibPath = "/usr/local/lib/libykcs11.dylib"

var outputFile string

func newCertificate(commonName string, startTime, endTime time.Time) (*x509.Certificate, error) {
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)

	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, fmt.Errorf("failed to generate new certificate: %v", err)
	}

	return &x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName: commonName,
		},
		NotBefore: startTime,
		NotAfter:  endTime,

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageCodeSigning},
		BasicConstraintsValid: true,
	}, nil
}

func keygen(writePrivKey bool) (*rsa.PrivateKey, []byte, error) {
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}
	privKeyBytes := x509.MarshalPKCS1PrivateKey(privKey)
	pubKeyBytes, err := x509.MarshalPKIXPublicKey(&privKey.PublicKey)
	if err != nil {
		return nil, nil, err
	}
	pemPriv := pem.EncodeToMemory(
		&pem.Block{Type: "RSA PRIVATE KEY", Bytes: privKeyBytes},
	)
	pemPub := pem.EncodeToMemory(
		&pem.Block{Type: "RSA PUBLIC KEY", Bytes: pubKeyBytes},
	)
	if writePrivKey == true {
		ioutil.WriteFile(outputFile, pemPriv, 0600)
	}

	// We need to have a certificate mapped to the key, otherwise it will not be
	// added to the yubikey.
	// The expiry date for this certificate is set to 25 years in the future.
	startTime := time.Now()
	template, err := newCertificate("OONI Operator", startTime, startTime.AddDate(25, 0, 0))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create the certificate template: %v", err)
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, template, template, &privKey.PublicKey, privKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create the certificate: %v", err)
	}

	ioutil.WriteFile(outputFile+".pub", pemPub, 0644)
	return privKey, certBytes, nil
}

func askForConfirm() bool {
	var response string
	_, err := fmt.Scanln(&response)
	if err != nil {
		panic(err)
	}
	if strings.ToLower(string(response[0])) == "y" {
		return true
	}
	return false
}

// keygenCmd represents the keygen command
var keygenCmd = &cobra.Command{
	Use:   "keygen",
	Short: "Generate a keypair for use for probe orchestration",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		if _, err := os.Stat(outputFile); !os.IsNotExist(err) {
			// XXX add confirmation dialog
			fmt.Printf("WARNING: %s exists\n", outputFile)
			fmt.Printf("overwrite? (y/n) ")
			if askForConfirm() == false {
				fmt.Println("ok quiting...")
				return
			}
			fmt.Println("overwriting...")
		}
		privKey, certBytes, err := keygen(true)
		if err != nil {
			fmt.Printf("failed to generate key pair: %v", err)
		}
		//sh, err := orchestrate.SetupHSM("/usr/local/lib/softhsm/libsofthsm2.so")
		err = keystore.AddKey(PKS11LibPath, privKey, certBytes)
		if err != nil {
			fmt.Printf("failed to add key: %v\n", err)
		}
		err = keystore.ListKeys(PKS11LibPath)
		if err != nil {
			fmt.Printf("failed to list keys: %v\n", err)
		}
	},
}

func init() {
	RootCmd.AddCommand(keygenCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	keygenCmd.PersistentFlags().StringVar(&outputFile, "f", "orchestrate-key", "Specify where to write the key to")
}
