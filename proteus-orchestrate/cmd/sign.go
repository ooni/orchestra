// Copyright © 2017 NAME HERE <EMAIL ADDRESS>
//

package cmd

import (
	"os"
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"fmt"

	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"
	"gopkg.in/dgrijalva/jwt-go.v3"
)

var privKeyPath string

type ProteusClaims struct {
    Foo string `json:"foo"`
    jwt.StandardClaims
}

func sign(claims ProteusClaims) {
	keyPEM, err := ioutil.ReadFile(privKeyPath)
	if err != nil {
		panic(err)
	}
	privKey, err := jwt.ParseECPrivateKeyFromPEM(keyPEM)
	if err != nil {
		panic(err)
	}
	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
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
	Long: `Usually this command is run with the output given from the proteus web interface`,
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
		fmt.Println("")
		claims := ProteusClaims{}

		err = json.Unmarshal(jsonBytes, &claims)
		if err != nil {
			panic(err)
		}
		sign(claims)
	},
}

func init() {
	RootCmd.AddCommand(signCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// signCmd.PersistentFlags().String("foo", "", "A help for foo")

	signCmd.PersistentFlags().StringVar(&privKeyPath, "f", "proteus-key", "Specify where to read private key from")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// signCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}
