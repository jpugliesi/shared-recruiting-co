package cloudfunctions

import (
	"encoding/base64"
	"log"
	"net/http"
	"os"

	"github.com/GoogleCloudPlatform/functions-framework-go/functions"

	"google.golang.org/api/gmail/v1"

	"github.com/shared-recruiting-co/shared-recruiting-co/libs/db/client"
	mail "github.com/shared-recruiting-co/shared-recruiting-co/libs/gmail"
)

const provider = "google"

func init() {
	functions.HTTP("RunWatchEmails", runWatchEmails)
}

func jsonFromEnv(env string) ([]byte, error) {
	encoded := os.Getenv(env)
	decoded, err := base64.URLEncoding.DecodeString(encoded)

	return decoded, err
}

func runWatchEmails(w http.ResponseWriter, r *http.Request) {
	log.Println("received watch trigger")
	ctx := r.Context()
	creds, err := jsonFromEnv("GOOGLE_OAUTH2_CREDENTIALS")
	if err != nil {
		log.Printf("error getting credentials: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Create SRC client
	apiURL := os.Getenv("SUPABASE_API_URL")
	apiKey := os.Getenv("SUPABASE_API_KEY")
	queries := client.NewHTTP(apiURL, apiKey)

	// TODO
	// v0 -> no pagination, no go routines
	// 2. Spawn a goroutine for each user to watch their emails
	// 3. Wait for all goroutines to finish

	// 1. Fetch valid auth tokens for all users
	userTokens, err := queries.ListUserOAuthTokens(ctx, client.ListUserOAuthTokensParams{
		Provider: provider,
		IsValid:  true,
	})

	if err != nil {
		log.Printf("error getting user tokens: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var srv *mail.Service
	user := "me"
	label := "UNREAD"
	topic := os.Getenv("PUBSUB_TOPIC")

	hasError := false

	for _, userToken := range userTokens {
		auth := []byte(userToken.Token)

		srv, err = mail.NewService(ctx, creds, auth)
		if err != nil {
			log.Printf("error creating gmail service: %v", err)
			hasError = true
			continue
		}

		// Get the user's email address
		// This also keeps the user's refresh token valid for deactivated emails
		gmailProfile, err := srv.Profile()
		if err != nil {
			log.Printf("error getting gmail profile: %v", err)

			// check for oauth token expiration or revocation
			if mail.IsOAuth2Error(err) {
				log.Printf("error oauth error: %v", err)
				// update the user's oauth token
				err = queries.UpsertUserOAuthToken(ctx, client.UpsertUserOAuthTokenParams{
					UserID:   userToken.UserID,
					Provider: provider,
					Token:    userToken.Token,
					IsValid:  false,
				})
				if err != nil {
					log.Printf("error updating user oauth token: %v", err)
				} else {
					log.Printf("marked user oauth token as invalid")
				}
			}
			hasError = true
			continue
		}

		// validate the user's email is active
		userProfile, err := queries.GetUserProfileByEmail(ctx, gmailProfile.EmailAddress)
		if err != nil {
			log.Printf("error getting user profile: %v", err)
			hasError = true
			continue
		}

		if !userProfile.IsActive {
			log.Printf("skipping deactivated email %s", userProfile.Email)
			continue
		}

		// Watch for changes in labelId
		resp, err := mail.ExecuteWithRetries(func() (*gmail.WatchResponse, error) {
			return srv.Users.Watch(user, &gmail.WatchRequest{
				LabelIds:          []string{label},
				LabelFilterAction: "include",
				TopicName:         topic,
			}).Do()
		})

		if err != nil {
			log.Printf("error watching: %v", err)
			continue
		}
		// success
		log.Printf("watching: %v", resp)
	}

	// write error status code for tracking
	if hasError {
		w.WriteHeader(http.StatusInternalServerError)
	}

	log.Println("done.")
}
