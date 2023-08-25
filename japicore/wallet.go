package japicore

import (
	"fmt"
	"github.com/JackalLabs/jackalgo/handlers/file_io_handler"
	"github.com/JackalLabs/jackalgo/handlers/wallet_handler"
	"os"
)

func InitWalletSession() (*wallet_handler.WalletHandler, *file_io_handler.FileIoHandler) {
	seed := os.Getenv("JHTTP_SEED")
	if len(seed) == 0 {
		panic("No Seed Provided!")
	}
	rpc := os.Getenv("JHTTP_RPC")
	if len(rpc) == 0 {
		rpc = "https://jackal-testnet-rpc.polkachu.com:443"
	}
	chainid := os.Getenv("JHTTP_CHAIN")
	if len(chainid) == 0 {
		chainid = "lupulella-2"
	}
	operatingRoot := os.Getenv("JHTTP_OP_ROOT")
	if len(operatingRoot) == 0 {
		operatingRoot = "JAPI"
	}

	wallet, err := wallet_handler.NewWalletHandler(
		seed, //slim odor fiscal swallow piece tide naive river inform shell dune crunch canyon ten time universe orchard roast horn ritual siren cactus upon forum
		rpc,
		chainid)

	if err != nil {
		panic(err)
	}

	fileIo, err := file_io_handler.NewFileIoHandler(wallet.WithGas("500000"))
	if err != nil {
		panic(err)
	}

	_, err = fileIo.DownloadFolder(fmt.Sprintf("s/%s", operatingRoot))
	if err != nil {
		_, err = fileIo.GenerateInitialDirs([]string{operatingRoot})
		if err != nil {
			panic(err)
		}
	}

	return wallet, fileIo
}
