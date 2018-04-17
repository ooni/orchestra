package cmd

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/apex/log"
	"github.com/hellais/jwt-go"
	"github.com/ooni/orchestra/orchestrate/orchestrate/keystore"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"
)

var privKeyPath string

func signLocal(claims keystore.OrchestraClaims) {
	keyPEM, err := ioutil.ReadFile(privKeyPath)
	if err != nil {
		panic(err)
	}
	privKey, err := jwt.ParseRSAPrivateKeyFromPEM(keyPEM)
	if err != nil {
		panic(err)
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS512, claims)
	ss, err := token.SignedString(privKey)
	if err != nil {
		panic(err)
	}
	fmt.Println("Signed claims: ")
	fmt.Printf("%v", ss)
	fmt.Println("")
}

// signCmd represents the sign command
var signCmd = &cobra.Command{
	Use:   "sign",
	Short: "Used to sign orchestration commands",
	Long:  `Usually this command is run with the output given from the orchestra web interface`,
	Run: func(cmd *cobra.Command, args []string) {
		initHSMConfig()
		log.Debug("Running signCmd")
		if terminal.IsTerminal(0) {
			// Example:
			log.Error(`You need to pipe me data.
	For example:

	$ python -c 'import json,base64;print(base64.b64encode(json.dumps({"iss": "art", "exp": 15000, "foo": "bar"})))'
	eyJpc3MiOiAiYXJ0IiwgImZvbyI6ICJiYXIiLCAiZXhwIjogMTUwMDB9
	$ echo "eyJpc3MiOiAiYXJ0IiwgImZvbyI6ICJiYXIiLCAiZXhwIjogMTUwMDB9" | ooni-orchestrate sign
`)
			return
		}
		inBytes, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			panic(err)
		}
		jsonBytes, err := base64.StdEncoding.DecodeString(string(inBytes))
		if err != nil {
			panic(err)
		}
		log.WithField("json_bytes", string(jsonBytes)).Info("Signing")
		log.Warn("Press the yubikey button")
		claims := keystore.OrchestraClaims{}

		err = json.Unmarshal(jsonBytes, &claims)
		if err != nil {
			panic(err)
		}

		ks := keystore.NewKeyStore(hsmConfig)
		signedClaims, err := ks.SignClaims(claims)
		if err != nil {
			panic(err)
		}
		log.WithField("signed_string", signedClaims).Info("Signed claims")
	},
}

func init() {
	RootCmd.AddCommand(signCmd)
}
