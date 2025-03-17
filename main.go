package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"

	"keyless-auth/api"
	_ "keyless-auth/api/docs"
	"keyless-auth/repository"
	"keyless-auth/repository/session"
	"keyless-auth/repository/user"
	"keyless-auth/services"
	"keyless-auth/signer"
)

var (
	redisHost     string
	redisPort     string
	redisPassword string
	redisUsername string
)

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return defaultValue
}

func init() {
	redisHost = getEnvOrDefault("REDIS_HOST", "localhost")
	redisPort = getEnvOrDefault("REDIS_PORT", "6379")
	redisPassword = getEnvOrDefault("REDIS_PASSWORD", "")
	redisUsername = getEnvOrDefault("REDIS_USERNAME", "")
}

func main() {
	db, err := services.NewRedisClient(context.Background(), &redis.Options{
		Username: redisUsername,
		Addr:     redisHost + ":" + redisPort,
		Password: redisPassword,
		DB:       0,
	})

	if err != nil {
		log.Fatal(err)
	}
	defer db.Client.Close()

	newSigner, _ := signer.New()

	walletRepo := repository.NewWalletRepository(db)
	credentialsRepo := repository.NewCredentialsRepository(db)
	proofHandler := api.NewProofHandler(walletRepo)
	credentialsHandler := api.NewCredentialsHandler(credentialsRepo, walletRepo)
	userRepo := user.NewUser(db)
	sessionRepo := session.NewRedisSessionStore(db.Client)
	googleHandler := api.NewGoogleHandler(userRepo)
	challengeHandler := api.NewEphemeralHandler(newSigner, userRepo, sessionRepo)

	router := mux.NewRouter()

	// credentials
	router.HandleFunc("/credentials", credentialsHandler.GenerateCredential).Methods("POST")
	router.HandleFunc("/merkle-root", credentialsHandler.GetMerkleRoot).Methods("GET")
	router.HandleFunc("/merkle-proof/{credential}", credentialsHandler.GenerateMerkleProof).Methods("GET")
	router.HandleFunc("/generate-tree-object", credentialsHandler.GenerateTreeObject).Methods("POST")
	// zk proof
	router.HandleFunc("/proof", proofHandler.GenerateProof).Methods("POST")
	// auth
	router.HandleFunc("/auth/google/login", googleHandler.HandleGoogleLogin).Methods("GET")
	router.HandleFunc("/auth/google/callback", googleHandler.HandleGoogleCallback).Methods("GET")

	router.HandleFunc("/challenge", challengeHandler.ChallengeHandler).Methods("POST")
	router.HandleFunc("/challenge/verify", challengeHandler.VerifyChallengeHandler).Methods("POST").GetHandler()

	// docs
	// router.HandleFunc("/api/docs/doc.json", func(w http.ResponseWriter, r *http.Request) {
	// 	spec, err := docs.BuildOpenAPISpec(router)
	// 	if err != nil {
	// 		http.Error(w, err.Error(), http.StatusInternalServerError)
	// 		return
	// 	}
	//
	// 	w.Header().Set("Content-Type", "application/json")
	// 	if err := json.NewEncoder(w).Encode(spec); err != nil {
	// 		http.Error(w, err.Error(), http.StatusInternalServerError)
	// 	}
	// })
	//
	// router.PathPrefix("/api/docs/").Handler(httpSwagger.Handler(
	// 	httpSwagger.URL("/api/docs/doc.json"),
	// ))

	serverAddr := fmt.Sprintf(":%s", getEnvOrDefault("APP_PORT", "8081"))
	log.Printf("Server starting on %s", serverAddr)
	log.Fatal(http.ListenAndServe(serverAddr, router))
}
