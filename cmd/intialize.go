package cmd

import (
	"io"
	"time"

	"github.com/bloc-transfer-service/config/database"
	"github.com/bloc-transfer-service/middleware"

	"net/http"

	transactionInit "github.com/bloc-transfer-service/pkg/transactions/init"
	transferInit "github.com/bloc-transfer-service/pkg/transfers/init"

	"log"
	"os"

	sentrynegroni "github.com/getsentry/sentry-go/negroni"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/solarwinds/apm-go/swo"
	"github.com/urfave/negroni"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"gopkg.in/natefinch/lumberjack.v2"
)

// HandleInit handles the initialization of the project
func Initialize() {
	godotenv.Load()
	//connect to database
	database.ConnectDB()
	//connect to redis
	database.ConnectRedis()
	//initialize logger
	InitializeLogger()

	//HandleOtel()
}

// initialize daily log rotation
func InitializeLogger() {
	logFile := &lumberjack.Logger{
		Filename:   "logs/app-" + time.Now().Format("2006-01-02") + ".log",
		MaxSize:    10, // megabytes
		MaxBackups: 7,
		MaxAge:     28, //days
		Compress:   true,
	}
	// Create a multi-writer (log to file & console)
	multiWriter := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(multiWriter)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.Println("Logger initialized!")
}

// register middleare and default routing
func RegisterMiddleware(router *mux.Router) *negroni.Negroni {
	n := negroni.Classic()
	n.Use(sentrynegroni.New(sentrynegroni.Options{Repanic: true}))
	n.Use(negroni.HandlerFunc(middleware.Secure().HandlerFuncWithNext))
	n.Use(middleware.Cors())
	otelWrappedRouter := otelhttp.NewHandler(router, "Bloc API", otelhttp.WithSpanNameFormatter(func(operation string, r *http.Request) string {
		// This allows dynamic naming of the span based on the route.
		// Adjust according to your needs.
		return r.Method + " " + r.URL.Path
	}))
	n.UseHandler(otelWrappedRouter)
	router.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		writer.Write([]byte("Bloc transfer service is up and running..."))
	})

	return n
}

// register services
func RegisterServices(router *mux.Router) {
	//register services here
	// accountInit.InitAccount(router)
	// customerInit.InitCustomer(router)
	transactionInit.InitTransaction(router)
	transferInit.InitTransfer(router)
}

func HandleOtel() {
	// Initialize the SolarWinds APM library
	cb, err := swo.Start(
		// Optionally add service-level resource attributes
		semconv.ServiceVersion("v0.0.1"),
		attribute.String("environment", "testing"),
	)
	if err != nil {
		// Handle error
	}
	// This function returned from 'Start()' will tell the apm library to
	// shut down, often deferred until the end of 'main()'.
	defer cb()
}
