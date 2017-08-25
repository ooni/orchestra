package cmd

import (
	"os"
	"strings"
	"io/ioutil"
	"crypto/rsa"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/thetorproject/proteus/proteus-orchestrate/orchestrate/keystore"
)

var outputFile string

func keygen(writePrivKey bool) (*rsa.PrivateKey, error) {
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}
	privKeyBytes := x509.MarshalPKCS1PrivateKey(privKey)
	pubKeyBytes, err := x509.MarshalPKIXPublicKey(&privKey.PublicKey)
	if err != nil {
		return nil, err
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
	ioutil.WriteFile(outputFile + ".pub", pemPub, 0644)
	return privKey, nil
}

func askForConfirm() bool {
	var response string
	_, err := fmt.Scanln(&response)
	if err != nil {
		panic(err)
	}
	if (strings.ToLower(string(response[0])) == "y") {
		return true
	}
	return false
}

// keygenCmd represents the keygen command
var keygenCmd = &cobra.Command{
	Use:   "keygen",
	Short: "Generate a keypair for use with proteus orchestration",
	Long: ``,
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
		privKey, err := keygen(true)
		if err != nil {
			fmt.Printf("failed to generate key pair: %v", err)
		}
		//sh, err := orchestrate.SetupHSM("/usr/local/lib/softhsm/libsofthsm2.so")
		err = keystore.AddKey("/usr/local/lib/softhsm/libsofthsm2.so", privKey)
		if err != nil {
			fmt.Printf("failed to add key: %v\n", err)
		}
		err = keystore.ListKeys("/usr/local/lib/softhsm/libsofthsm2.so")
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
	keygenCmd.PersistentFlags().StringVar(&outputFile, "f", "proteus-key", "Specify where to write the key to")
}
