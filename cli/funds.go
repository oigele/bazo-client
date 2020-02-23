package cli

import (
	"crypto/ecdsa"
	"errors"
	"fmt"
	"github.com/oigele/bazo-client/network"
	"github.com/oigele/bazo-client/util"
	"github.com/oigele/bazo-miner/crypto"
	"github.com/oigele/bazo-miner/p2p"
	"github.com/oigele/bazo-miner/protocol"
	"github.com/urfave/cli"
	"log"
)

type fundsArgs struct {
	header			int
	fromWalletFile	string
	toWalletFile	string
	toAddress		string
	multisigFile	string
	amount			uint64
	fee				uint64
	txcount		    int
}

type massFundsArgs struct {
	header			int
	fromWalletFile	string
	toWalletFile	string
	toAddress		string
	multisigFile	string
	amount			uint64
	fee				uint64
	transactions    int
}

func GetFundsCommand(logger *log.Logger) cli.Command {
	return cli.Command {
		Name:	"funds",
		Usage:	"send funds from one account to another",
		Action:	func(c *cli.Context) error {
			args := &fundsArgs{
				header: 		c.Int("header"),
				fromWalletFile: c.String("from"),
				toWalletFile: 	c.String("to"),
				toAddress: 		c.String("toAddress"),
				multisigFile: 	c.String("multisig"),
				amount: 		c.Uint64("amount"),
				fee: 			c.Uint64("fee"),
				txcount:		c.Int("txcount"),
			}

			return sendFunds(args, logger)
		},
		Flags:	[]cli.Flag {
			cli.IntFlag {
				Name: 	"header",
				Usage: 	"header flag",
				Value:	0,
			},
			cli.StringFlag {
				Name: 	"from",
				Usage: 	"load the sender's private key from `FILE`",
			},
			cli.StringFlag {
				Name: 	"to",
				Usage: 	"load the recipient's public key from `FILE`",
			},
			cli.StringFlag {
				Name: 	"toAddress",
				Usage: 	"the recipient's 128 byze public address",
			},
			cli.Uint64Flag {
				Name: 	"amount",
				Usage:	"specify the amount to send",
			},
			cli.Uint64Flag {
				Name: 	"fee",
				Usage:	"specify the fee",
				Value: 	1,
			},
			cli.IntFlag {
				Name: 	"txcount",
				Usage:	"the sender's current transaction counter",
			},
			cli.StringFlag {
				Name: 	"multisig",
				Usage: 	"load multi-signature server’s private key from `FILE`",
			},
		},
	}
}


func GetMassFundsCommand(logger *log.Logger) cli.Command {
	return cli.Command {
		Name:	"massfunds",
		Usage:	"send mass funds from one account to another for testing purposes",
		Action:	func(c *cli.Context) error {
			args := &massFundsArgs{
				header: 		c.Int("header"),
				fromWalletFile: c.String("from"),
				toWalletFile: 	c.String("to"),
				toAddress: 		c.String("toAddress"),
				multisigFile: 	c.String("multisig"),
				amount: 		c.Uint64("amount"),
				fee: 			c.Uint64("fee"),
				transactions:	c.Int("transactions"),
			}

			return sendMassFunds(args, logger)
		},
		Flags:	[]cli.Flag {
			cli.IntFlag {
				Name: 	"header",
				Usage: 	"header flag",
				Value:	0,
			},
			cli.StringFlag {
				Name: 	"from",
				Usage: 	"load the sender's private key from `FILE`",
			},
			cli.StringFlag {
				Name: 	"to",
				Usage: 	"load the recipient's public key from `FILE`",
			},
			cli.StringFlag {
				Name: 	"toAddress",
				Usage: 	"the recipient's 128 byze public address",
			},
			cli.Uint64Flag {
				Name: 	"amount",
				Usage:	"specify the amount to send",
			},
			cli.Uint64Flag {
				Name: 	"fee",
				Usage:	"specify the fee",
				Value: 	1,
			},
			cli.IntFlag {
				Name: 	"transactions",
				Usage:	"number of transactions to be sent",
			},
			cli.StringFlag {
				Name: 	"multisig",
				Usage: 	"load multi-signature server’s private key from `FILE`",
			},
		},
	}
}

func sendFunds(args *fundsArgs, logger *log.Logger) error {
	err := args.ValidateInput()
	if err != nil {
		return err
	}

	fromPrivKey, err := crypto.ExtractECDSAKeyFromFile(args.fromWalletFile)
	if err != nil {
		return err
	}

	var toPubKey *ecdsa.PublicKey
	if len(args.toWalletFile) == 0 {
		if len(args.toAddress) == 0 {
			return errors.New(fmt.Sprintln("No recipient specified"))
		} else {
			if len(args.toAddress) != 128 {
				return errors.New(fmt.Sprintln("Invalid recipient address"))
			}

			runes := []rune(args.toAddress)
			pub1 := string(runes[:64])
			pub2 := string(runes[64:])

			toPubKey, err = crypto.GetPubKeyFromString(pub1, pub2)
			if err != nil {
				return err
			}
		}
	} else {
		toPubKey, err = crypto.ExtractECDSAPublicKeyFromFile(args.toWalletFile)
		if err != nil {
			return err
		}
	}

	var multisigPrivKey *ecdsa.PrivateKey
	if len(args.multisigFile) > 0 {
		multisigPrivKey, err = crypto.ExtractECDSAKeyFromFile(args.multisigFile)
		if err != nil {
			return err
		}
	} else {
		multisigPrivKey = fromPrivKey
	}

	fromAddress := crypto.GetAddressFromPubKey(&fromPrivKey.PublicKey)
	toAddress := crypto.GetAddressFromPubKey(toPubKey)

	tx, err := protocol.ConstrFundsTx(
		byte(args.header),
		uint64(args.amount),
		uint64(args.fee),
		uint32(args.txcount),
		protocol.SerializeHashContent(fromAddress),
		protocol.SerializeHashContent(toAddress),
		fromPrivKey,
		multisigPrivKey,
		nil)

	if err != nil {
		logger.Printf("%v\n", err)
		return err
	}

	for _, committee := range util.CommitteesIpPortSlice {
		if err := network.SendTx(committee, tx, p2p.FUNDSTX_BRDCST); err != nil {
			logger.Printf("%v\n", err)
		} else {
			logger.Printf("Transaction successfully sent to IpPort:%s\nTxHash: %x%v", committee, tx.Hash(), tx)
		}
	}

	return nil
}

func sendMassFunds(args *massFundsArgs, logger *log.Logger) error {
	err := args.ValidateInput()
	if err != nil {
		return err
	}

	fromPrivKey, err := crypto.ExtractECDSAKeyFromFile(args.fromWalletFile)
	if err != nil {
		return err
	}

	var toPubKey *ecdsa.PublicKey
	if len(args.toWalletFile) == 0 {
		if len(args.toAddress) == 0 {
			return errors.New(fmt.Sprintln("No recipient specified"))
		} else {
			if len(args.toAddress) != 128 {
				return errors.New(fmt.Sprintln("Invalid recipient address"))
			}

			runes := []rune(args.toAddress)
			pub1 := string(runes[:64])
			pub2 := string(runes[64:])

			toPubKey, err = crypto.GetPubKeyFromString(pub1, pub2)
			if err != nil {
				return err
			}
		}
	} else {
		toPubKey, err = crypto.ExtractECDSAPublicKeyFromFile(args.toWalletFile)
		if err != nil {
			return err
		}
	}

	var multisigPrivKey *ecdsa.PrivateKey
	if len(args.multisigFile) > 0 {
		multisigPrivKey, err = crypto.ExtractECDSAKeyFromFile(args.multisigFile)
		if err != nil {
			return err
		}
	} else {
		multisigPrivKey = fromPrivKey
	}

	fromAddress := crypto.GetAddressFromPubKey(&fromPrivKey.PublicKey)
	toAddress := crypto.GetAddressFromPubKey(toPubKey)

	logger.Printf("Start send mass funds")
	for i := 1; i<= args.transactions ; i++ {

		tx, err := protocol.ConstrFundsTx(
			byte(args.header),
			uint64(args.amount),
			uint64(args.fee),
			uint32(1),
			protocol.SerializeHashContent(fromAddress),
			protocol.SerializeHashContent(toAddress),
			fromPrivKey,
			multisigPrivKey,
			nil)

		if err != nil {
			logger.Printf("%v\n", err)
			return err
		}

		for _, committee := range util.CommitteesIpPortSlice {
			if err := network.SendTx(committee, tx, p2p.FUNDSTX_BRDCST); err != nil {
				logger.Printf("%v\n", err)
			} else {
				logger.Printf("Transaction successfully sent to IpPort:%s\nTxHash: %x", committee, tx.Hash())
			}
		}
	}
	logger.Printf("End send mass funds")

	return nil
}

func (args fundsArgs) ValidateInput() error {
	if len(args.fromWalletFile) == 0 {
		return errors.New("argument missing: from")
	}

	if args.txcount < 0 {
		return errors.New("invalid argument: txcnt must be >= 0")
	}

	if len(args.toWalletFile) == 0 && len(args.toAddress) == 0 {
		return errors.New("argument missing: to or toAddess")
	}

	if len(args.toWalletFile) == 0 && len(args.toAddress) != 128 {
		return errors.New("invalid argument: toAddress")
	}

	if args.fee <= 0 {
		return errors.New("invalid argument: fee must be > 0")
	}

	if args.amount <= 0 {
		return errors.New("invalid argument: amount must be > 0")
	}

	return nil
}

func (args massFundsArgs) ValidateInput() error {
	if len(args.fromWalletFile) == 0 {
		return errors.New("argument missing: from")
	}

	if args.transactions < 0 {
		return errors.New("invalid argument: txcnt must be >= 0")
	}

	if len(args.toWalletFile) == 0 && len(args.toAddress) == 0 {
		return errors.New("argument missing: to or toAddess")
	}

	if len(args.toWalletFile) == 0 && len(args.toAddress) != 128 {
		return errors.New("invalid argument: toAddress")
	}

	if args.fee <= 0 {
		return errors.New("invalid argument: fee must be > 0")
	}

	if args.amount <= 0 {
		return errors.New("invalid argument: amount must be > 0")
	}

	return nil
}
