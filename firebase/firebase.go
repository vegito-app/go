package firebase

import (
	"context"
	"fmt"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"github.com/rs/zerolog/log"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	_ "github.com/GoogleCloudPlatform/functions-framework-go/funcframework"
	"google.golang.org/api/option"
)

type App struct {
	*firebase.App
}

func NewApp(ctx context.Context) (*App, error) {

	var opts []option.ClientOption

	projectID := config.GetString(projectIDconfig)
	firebasseConfig := &firebase.Config{
		ProjectID: projectID,
	}

	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		err := client.Close()
		if err != nil {
			log.Error().Err(err).Msg("new adminsdk firebase secretmanager client")
		}
	}()
	adminSDKserviceAccountSecretID := config.GetString(adminSDKserviceAccountIDConfig)
	if adminSDKserviceAccountSecretID != "" {
		firebaseAdminSDKserviceAccountSecretID := &secretmanagerpb.AccessSecretVersionRequest{
			Name: adminSDKserviceAccountSecretID,
		}
		firebaseAdminSDKserviceAccount, err := client.AccessSecretVersion(ctx, firebaseAdminSDKserviceAccountSecretID)
		if err != nil {
			return nil, err
		}
		opts = append(opts, option.WithCredentialsJSON(firebaseAdminSDKserviceAccount.Payload.Data))
	}
	app, err := firebase.NewApp(ctx, firebasseConfig, opts...)
	if err != nil {
		return nil, fmt.Errorf("firebase new app: %w", err)
	}
	return &App{
		App: app,
	}, nil
}

type AuthClient struct {
	*auth.Client
}

// NewAuthClient verify the given client Firebase ID_TOKEN.
// The authenticated client is returned and can be used to retrieve specific client information.
func (f *App) NewAuthClient(ctx context.Context) (*AuthClient, error) {
	client, err := f.App.Auth(ctx)
	if err != nil {
		return nil, fmt.Errorf("firebase getting Auth client: %w", err)
	}
	return &AuthClient{
		Client: client,
	}, nil
}

func (f *App) VerifyIDToken(ctx context.Context, idToken string) (string, error) {

	// Normal production verification
	authClient, err := f.NewAuthClient(ctx)
	if err != nil {
		return "", fmt.Errorf("getting auth client: %w", err)
	}
	token, err := authClient.VerifyIDToken(ctx, idToken)
	if err != nil {
		return "", fmt.Errorf("invalid token: %w", err)
	}
	return token.UID, nil
}
