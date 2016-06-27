package cmd

import (
	"encoding/base64"
	"fmt"

	"github.com/codegangsta/cli"

	"github.com/containerops/dockyard/utils/aes"
)

// CmdEncrypt is cmd for encrypting password
var CmdEncrypt = cli.Command{
	Name:        "encrypt",
	Usage:       "encrypt password",
	Description: "dockyard encrypt is the tool for encrypting password.",
	Action:      runEncrypt,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "plain",
			Usage: "plain text.",
		},
	},
}

func runEncrypt(c *cli.Context) {
	plain := c.String("plain")
	if len(plain) <= 0 {
		fmt.Printf("Plain can't be null\n")
	} else {
		if cipher, err := aes.Encrypt([]byte(plain), aes.AESKey); err != nil {
			fmt.Printf("Encrypt error:%s\n", err.Error())
		} else {
			fmt.Printf("%s\n", base64.StdEncoding.EncodeToString(cipher))
		}
	}
}
