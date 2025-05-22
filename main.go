package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"runtime"
	"strconv"

	cmd "github.com/bloc-transfer-service/cmd"

	"github.com/gorilla/mux"
)

func main() {
	if len(os.Args) > 3 {
		//means the command has more arguments and should go to the init function
		fmt.Println("Arguments passed", os.Args)
		cmd.HandleInit()
		return
	}
	//initialize dbs
	cmd.Initialize()
	//initialize routing
	router := mux.NewRouter()
	//add middlewares
	handler := cmd.RegisterMiddleware(router)
	//add services
	cmd.RegisterServices(router)
	//start application
	numWorkers := runtime.NumCPU()
	environment := os.Getenv("ENVIRONMENT")
	fmt.Printf("____________________Environment: %s______________________ \n", environment)
	port := os.Getenv("PORT")
	listener, err := net.Listen("tcp", ":"+port) // All workers use the same listener
	if err != nil {
		log.Fatal("Error creating listener:", err)
	}
	portInt, err := strconv.Atoi(port)
	if err != nil {
		log.Fatal("Error converting port to integer:", err)
	}
	fmt.Printf("=============== Starting %d workers on port %d...======================== \n", numWorkers, portInt)
	for i := 0; i < numWorkers; i++ {
		go func() {
			fmt.Printf("Worker %d\n started", i)
			http.Serve(listener, handler)
		}()
	}
	select {} // Keep main running
}
