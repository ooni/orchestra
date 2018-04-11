package cmd

import (
	"crypto/rsa"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/hellais/jwt-go"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"
)

var pubKeyPath string

func verify(tokenString string, pubKey *rsa.PublicKey) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return pubKey, nil
	})
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		fmt.Println("TOKEN IS VALID")
		fmt.Printf("Check passed: %v\n", claims.VerifyIssuer("art", true))
		fmt.Printf("Issuer      : %v\n", claims["iss"])
		fmt.Printf("Foo         : %v\n", claims["foo"])
		fmt.Printf("Expiry      : %v\n", claims["exp"])
	} else {
		fmt.Println("ERR: TOKEN IS INVALID")
		fmt.Println(err)
	}
}

// verifyCmd represents the verify command
var verifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "Used to verify that a signed token is valid",
	Long:  `Takes from stdin the signed token and says if the token is valid or not`,
	Run: func(cmd *cobra.Command, args []string) {
		if terminal.IsTerminal(0) {
			fmt.Println("I require some pipe")
			return
		}
		inBytes, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			panic(err)
		}
		pubKeyBytes, err := ioutil.ReadFile(pubKeyPath)
		if err != nil {
			panic(err)
		}
		pubKey, err := jwt.ParseRSAPublicKeyFromPEM(pubKeyBytes)
		if err != nil {
			panic(err)
		}
		verify(strings.TrimSpace(string(inBytes)), pubKey)
	},
}

func init() {
	RootCmd.AddCommand(verifyCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	verifyCmd.PersistentFlags().StringVar(&pubKeyPath, "f", "proteus-key.pub", "The path to the public key path")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// verifyCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}
