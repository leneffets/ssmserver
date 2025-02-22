package ecr

import (
	"context"
	"encoding/base64"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
)

func GetECRCredentials(ctx context.Context, sess *session.Session) (*ecr.GetAuthorizationTokenOutput, error) {
	svc := ecr.New(sess)
	return svc.GetAuthorizationTokenWithContext(ctx, &ecr.GetAuthorizationTokenInput{})
}

func HandleECRLogin(w http.ResponseWriter, r *http.Request, sess *session.Session) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if r.Method != http.MethodGet {
		http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
		return
	}

	results, err := GetECRCredentials(ctx, sess)
	if err != nil {
		log.Printf("Error fetching ECR credentials: %v", err)
		http.Error(w, "Error fetching ECR credentials", http.StatusInternalServerError)
		return
	}

	if len(results.AuthorizationData) == 0 {
		http.Error(w, "No authorization data found", http.StatusInternalServerError)
		return
	}

	authData := results.AuthorizationData[0]
	decodedToken, err := base64.StdEncoding.DecodeString(*authData.AuthorizationToken)
	if err != nil {
		log.Printf("Error decoding authorization token: %v", err)
		http.Error(w, "Error decoding authorization token", http.StatusInternalServerError)
		return
	}

	tokenParts := strings.SplitN(string(decodedToken), ":", 2)
	if len(tokenParts) != 2 {
		http.Error(w, "Invalid authorization token format", http.StatusInternalServerError)
		return
	}

	password := tokenParts[1]

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte(password)); err != nil {
		http.Error(w, "Error writing response", http.StatusInternalServerError)
		log.Printf("Error writing response: %v", err)
	}
}
