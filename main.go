package main

import (
	"flag"
	"log"
	"mime"
	"net"
	"os"
	"path/filepath"

	"github.com/autograde/aguis/ci"
	"github.com/autograde/aguis/envoy"
	"github.com/autograde/aguis/web"
	"github.com/autograde/aguis/web/auth"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	pb "github.com/autograde/aguis/ag"
	"github.com/autograde/aguis/database"

	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"google.golang.org/grpc"
)

func init() {
	mustAddExtensionType := func(ext, typ string) {
		if err := mime.AddExtensionType(ext, typ); err != nil {
			panic(err)
		}
	}

	// On Windows, mime types are read from the registry, which often has
	// outdated content types. This enforces that the correct mime types
	// are used on all platforms.
	mustAddExtensionType(".html", "text/html")
	mustAddExtensionType(".css", "text/css")
	mustAddExtensionType(".js", "application/javascript")
	mustAddExtensionType(".jsx", "application/javascript")
	mustAddExtensionType(".map", "application/json")
	mustAddExtensionType(".ts", "application/x-typescript")
}

func envString(env, fallback string) string {
	e := os.Getenv(env)
	if e == "" {
		return fallback
	}
	return e
}

func main() {
	var (
		httpAddr   = flag.String("http.addr", ":8081", "HTTP listen address")
		public     = flag.String("http.public", "public", "path to static files to serve")
		scriptPath = flag.String("script.path", "ci/scripts", "path to continuous integration scripts")
		dbFile     = flag.String("database.file", tempFile("ag.db"), "database file")
		baseURL    = flag.String("service.url", "localhost", "service base url")
		fake       = flag.Bool("provider.fake", false, "enable fake provider")
		grpcAddr   = flag.String("grpc.addr", ":9090", "gRPC listen address")
	)
	flag.Parse()

	cfg := zap.NewDevelopmentConfig()
	// database logging is only enabled if the LOGDB environment variable is set
	cfg = database.GormLoggerConfig(cfg)
	// add colorization
	cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	// we only want stack trace enabled for panic level and above
	logger, err := cfg.Build(zap.AddStacktrace(zapcore.PanicLevel))
	if err != nil {
		log.Fatalf("can't initialize logger: %v\n", err)
	}
	defer logger.Sync()

	db, err := database.NewGormDB("sqlite3", *dbFile, database.NewGormLogger(logger))
	if err != nil {
		log.Fatalf("can't connect to database: %v\n", err)
	}
	defer func() {
		if dbErr := db.Close(); dbErr != nil {
			log.Printf("error closing database: %v\n", dbErr)
		}
	}()

	// start envoy in a docker container; fetch envoy docker image if necessary
	go envoy.StartEnvoy(logger)

	// holds references for activated providers for current user token
	scms := auth.NewScms()
	bh := web.BaseHookOptions{
		BaseURL: *baseURL,
		Secret:  os.Getenv("WEBHOOK_SECRET"),
	}

	docker := ci.Docker{
		Endpoint: envString("DOCKER_HOST", "http://localhost:4243"),
		Version:  envString("DOCKER_VERSION", "1.30"),
	}

	agService := web.NewAutograderService(logger, db, scms, bh, &docker)
	go web.New(agService, *public, *httpAddr, *scriptPath, *fake)

	lis, err := net.Listen("tcp", *grpcAddr)
	if err != nil {
		log.Fatalf("failed to start tcp listener: %v\n", err)
	}
	opt := grpc.UnaryInterceptor(pb.Interceptor(logger))
	grpcServer := grpc.NewServer(opt)
	pb.RegisterAutograderServiceServer(grpcServer, agService)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to start grpc server: %v\n", err)
	}
}

func tempFile(name string) string {
	return filepath.Join(os.TempDir(), name)
}
