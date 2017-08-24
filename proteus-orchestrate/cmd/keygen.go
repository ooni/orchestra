package cmd

import (
	"os"
	"io/ioutil"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"fmt"

	"github.com/spf13/cobra"
)

var outputFile string

func keygen() {
	// pub, priv, err := ed25519.GenerateKey(rand.Reader)
	priv_key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		panic(err)
	}
	priv_key_bytes, err := x509.MarshalECPrivateKey(priv_key)
	if err != nil {
		panic(err)
	}
	pub_key_bytes, err := x509.MarshalPKIXPublicKey(&priv_key.PublicKey)
	if err != nil {
		panic(err)
	}
	pem_priv := pem.EncodeToMemory(
		&pem.Block{Type: "EC PRIVATE KEY", Bytes: priv_key_bytes},
	)
	//fmt.Println("pem_priv_block: ")
	//fmt.Print(string(pem_priv))

	pem_pub := pem.EncodeToMemory(
		&pem.Block{Type: "EC PUBLIC KEY", Bytes: pub_key_bytes},
	)
	//fmt.Println("pem_pub_block: ")
	//fmt.Print(string(pem_pub))
	ioutil.WriteFile(outputFile, pem_priv, 0600)
	ioutil.WriteFile(outputFile + ".pub", pem_pub, 0644)
}

// keygenCmd represents the keygen command
var keygenCmd = &cobra.Command{
	Use:   "keygen",
	Short: "Generate a keypair for use with proteus orchestration",
	Long: ``,
	Run: func(cmd *cobra.Command, args []string) {
		if _, err := os.Stat(outputFile); !os.IsNotExist(err) {
			// XXX add confirmation dialog
			fmt.Printf("WARNING: %s exists. Will overwrite\n", outputFile)
		}
		keygen()
	},
}

func init() {
	RootCmd.AddCommand(keygenCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	keygenCmd.PersistentFlags().StringVar(&outputFile, "f", "proteus-key", "Specify where to write the key to")
}
