package cmd

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/hellais/jwt-go"
	"github.com/spf13/cobra"
	"github.com/thalesignite/crypto11"
	"golang.org/x/crypto/ssh/terminal"
)

var privKeyPath string

type OrchestraClaims struct {
	Foo string `json:"foo"`
	jwt.StandardClaims
}

func signLocal(claims OrchestraClaims) {
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

func signHSM(claims OrchestraClaims) {
	config := &crypto11.PKCS11Config{
		Path:        hsmConfig.LibPath,
		Pin:         hsmConfig.UserPin,
		TokenSerial: hsmConfig.TokenSerial,
	}
	_, err := crypto11.Configure(config)
	if err != nil {
		fmt.Println("Failed to config")
		panic(err)
	}

	key, err := crypto11.FindKeyPair([]byte{11}, nil)
	if err != nil {
		fmt.Println("Failed to find keypair")
		panic(err)
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS512, claims)
	ss, err := token.SignedString(key)
	if err != nil {
		fmt.Println("Failed to sign")
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
		if terminal.IsTerminal(0) {
			// Example:
			// base64.b64encode(json.dumps({"iss": "art", "exp": 15000, "foo": "bar"}))
			// eyJpc3MiOiAiYXJ0IiwgImZvbyI6ICJiYXIiLCAiZXhwIjogMTUwMDB9
			fmt.Println("I require some pipe")
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
		fmt.Println("I will sign: ")
		fmt.Print(string(jsonBytes))
		fmt.Println("Press the yubikey button")
		claims := OrchestraClaims{}

		err = json.Unmarshal(jsonBytes, &claims)
		if err != nil {
			panic(err)
		}

		signHSM(claims)
	},
}

func init() {
	RootCmd.AddCommand(signCmd)
	addOperatorConfig(signCmd)
}
