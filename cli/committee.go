package cli

import (
	"errors"
	"github.com/oigele/bazo-client/network"
	"github.com/oigele/bazo-client/util"
	"github.com/oigele/bazo-miner/crypto"
	"github.com/oigele/bazo-miner/p2p"
	"github.com/oigele/bazo-miner/protocol"
	"github.com/urfave/cli"
	"log"
)

type committeeArgs struct {
	header			int
	fee				uint64
	walletFile		string
	committee		string
	rootWallet		string
}

func GetCommitteeCommand(logger *log.Logger) cli.Command {
	return cli.Command {
		Name:	"committee",
		Usage:	"join the committee",
		Action:	func(c *cli.Context) error {
			args := &committeeArgs{
				header: 		c.Int("header"),
				fee: 			c.Uint64("fee"),
				walletFile: 	c.String("to"),
				committee: 		c.String("committee"),
				rootWallet:		c.String("rootwallet"),
			}
			return sendCommittee(args, logger)
		},
		Flags:	[]cli.Flag {
			cli.IntFlag {
				Name: 	"header",
				Usage: 	"header flag",
				Value:	0,
			},
			cli.Uint64Flag {
				Name: 	"fee",
				Usage: 	"specify the fee",
				Value:   1,
			},
			cli.StringFlag {
			Name: 	"wallet",
			Usage: 	"load validator's public key from `FILE`",
			Value: 	"wallet.txt",
			},
			cli.StringFlag {
				Name: 	"committee",
				Usage: 	"load committee's committee key from `FILE`",
				Value: 	"committee.txt",
			},
			cli.StringFlag{
				Name:  "rootwallet",
				Usage: "load root's public private key from `FILE`",
			},
		},
	}
}

func sendCommittee(args *committeeArgs, logger *log.Logger) error {
	err := args.ValidateInput()
	if err != nil {
		return err
	}

	rootPrivKey, err := crypto.ExtractECDSAKeyFromFile(args.rootWallet)
	if err != nil {
		return err
	}

	committeePrivKey, err := crypto.ExtractRSAKeyFromFile(args.committee)
	if err != nil {
		return err
	}

	validatorPubKey, err := crypto.ExtractECDSAPublicKeyFromFile(args.walletFile)
	if err != nil {
		logger.Printf("%v\n", err)
		return err
	}

	committeeAddress := crypto.GetAddressFromPubKey(validatorPubKey)


	tx, err := protocol.ConstrCommitteeTx(byte(args.header), uint64(args.fee), true, committeeAddress, rootPrivKey, &committeePrivKey.PublicKey)
	if err != nil {
		return err
	}

	if err := network.SendTx(util.Config.BootstrapIpport, tx, p2p.COMMITTEETX_BRDCST); err != nil {
		logger.Printf("%v\n", err)
		return err
	} else {
		logger.Printf("Transaction successfully sent to network:\nTxHash: %x%v", tx.Hash(), tx)
	}

	return nil


}




func (args committeeArgs) ValidateInput() error {
	if args.fee <= 0 {
		return errors.New("invalid argument: fee must be > 0")
	}
	if len(args.walletFile) == 0 {
		return errors.New("argument missing: wallet")
	}
	if len(args.committee) == 0 {
		return errors.New("argument missing: committee")
	}
	if len(args.rootWallet) == 0 {
		return errors.New("argument missing: root wallet")
	}

	return nil
}
