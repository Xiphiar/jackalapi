package http

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/JackalLabs/jackalgo/handlers/file_io_handler"
	"github.com/JackalLabs/jackalgo/handlers/wallet_handler"
	"github.com/julienschmidt/httprouter"
	"github.com/rs/cors"
)

func start(port int, get Handlers, post Handlers, queue *Queue, fileIo *file_io_handler.FileIoHandler) {
	router := httprouter.New()
	handler := cors.Default().Handler(router)

	for getKey, getFunc := range get {
		router.GET(getKey, func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
			getFunc(w, r, ps, queue, fileIo)
		})
	}

	for postKey, postFunc := range post {
		router.POST(postKey, func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
			postFunc(w, r, ps, queue, fileIo)
		})
	}

	fmt.Printf("🌍 Started Jackal API: http://0.0.0.0:%d\n", port)
	err := http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", port), handler)
	if err != nil {
		fmt.Println(err)
		return
	}

	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("Server Closed\n")
		return
	} else if err != nil {
		fmt.Printf("error starting server: %s\n", err)
		os.Exit(1)
	}
}

func StartServer(Gets Handlers, Posts Handlers, initalDirs []string) {

	seed := os.Getenv("JHTTP_SEED")
	rpc := os.Getenv("JHTTP_RPC")
	if len(rpc) == 0 {
		rpc = "https://jackal-testnet-rpc.polkachu.com:443"
	}
	chainid := os.Getenv("JHTTP_CHAIN")
	if len(chainid) == 0 {
		chainid = "lupulella-2"
	}
	port := os.Getenv("JHTTP_PORT")
	if len(port) == 0 {
		port = "3535"
	}

	portNum, err := strconv.ParseInt(port, 10, 64)
	if err != nil {
		panic(err)
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

	_, err = fileIo.GenerateInitialDirs(initalDirs)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Starting server with account: %s\n", wallet.GetAddress())

	queue := NewQueue()
	go queue.Listen()

	start(int(portNum), Gets, Posts, queue, fileIo)

}
