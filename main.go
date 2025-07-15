package main

import (
	"context"
	"log"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/mong0520/linebot-ptt-set/bots"
	"github.com/mong0520/linebot-ptt-set/models"
	"github.com/mong0520/linebot-ptt-set/utils"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var logger *log.Logger
var meta = &models.Model{}
var logRoot = "logs"

func initDB(dbURI string, enableSSL bool) {
	logger.Printf("Connecting to MongoDB: %s (SSL: %t)", dbURI, enableSSL)
	
	// Set client options
	clientOptions := options.Client().ApplyURI(dbURI)
	clientOptions.SetConnectTimeout(30 * time.Second)
	clientOptions.SetServerSelectionTimeout(30 * time.Second)
	clientOptions.SetMaxPoolSize(10)
	
	if enableSSL {
		clientOptions.SetTLSConfig(nil) // Use default TLS config
	}

	// Create context
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Retry connection logic
	var client *mongo.Client
	var err error
	maxRetries := 5
	
	for i := 0; i < maxRetries; i++ {
		logger.Printf("Attempting to connect to MongoDB (attempt %d/%d)...", i+1, maxRetries)
		
		client, err = mongo.Connect(ctx, clientOptions)
		if err == nil {
			// Test connection
			pingCtx, pingCancel := context.WithTimeout(context.Background(), 10*time.Second)
			err = client.Ping(pingCtx, nil)
			pingCancel()
			
			if err == nil {
				logger.Println("Successfully connected to MongoDB!")
				break
			}
			client.Disconnect(ctx)
		}
		
		logger.Printf("Connection failed: %v. Retrying in 3 seconds...", err)
		time.Sleep(3 * time.Second)
	}
	
	if err != nil {
		logger.Fatalf("Unable to connect to MongoDB after %d attempts: %v", maxRetries, err)
	}
	
	// Set database and collections
	database := client.Database("ptt")
	meta.MongoClient = client
	meta.Database = database
	meta.Collection = database.Collection("set")
	meta.CollectionUserFavorite = database.Collection("users")
	
	logger.Println("Database connection established successfully")
}

func main() {
	logFile, err := initLogFile()
	if err != nil {
		log.Fatalf("Failed to initialize log file: %v", err)
	}
	defer logFile.Close()

	// Initialize logger
	logger = utils.GetLogger(logFile)
	meta.Log = logger
	
	// Get environment variables
	dbURI := os.Getenv("MongoDBURI")
	if dbURI == "" {
		logger.Fatalln("MongoDBURI environment variable is required")
	}
	
	dbWithSSL, _ := strconv.ParseBool(os.Getenv("MongoDBSSL"))
	
	logger.Println("Starting PTT SET Line Bot...")
	logger.Printf("MongoDB URI: %s", dbURI)
	logger.Printf("SSL Enabled: %t", dbWithSSL)
	
	// Initialize database
	logger.Println("Initializing database connection...")
	initDB(dbURI, dbWithSSL)
	logger.Println("Database initialization completed")

	// Initialize Line Bot
	logger.Println("Initializing Line Bot...")
	bots.InitLineBot(meta, bots.ModeHTTP, "", "")
	logger.Println("Application exited")
}

func initLogFile() (logFile *os.File, err error) {
	logfilename := "pttset.log"
	logFileName := path.Base(logfilename)
	logFilePath := path.Join(logRoot, logFileName)
	
	// Ensure log directory exists
	if _, err := os.Stat(logRoot); os.IsNotExist(err) {
		if err := os.MkdirAll(logRoot, 0755); err != nil {
			return nil, err
		}
		log.Printf("Created log directory: %s", logRoot)
	}
	
	// Create or open log file
	logFile, err = os.OpenFile(logFilePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}
	
	log.Printf("Log file initialized: %s", logFilePath)
	return logFile, nil
}