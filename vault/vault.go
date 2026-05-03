package vault

import (
	"context"
	"encoding/json"
	"fmt"

	"sync"
	"time"

	vault "github.com/hashicorp/vault/api"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"

	_ "github.com/GoogleCloudPlatform/functions-framework-go/funcframework"
)

var config = viper.New()

const (
	vaultAddrConfig = "addr"
	vaultRoleConfig = "role"

	minimumLeaseTimeDurationConfig = "minimum_lease_time_duration"
)

func init() {
	config.AutomaticEnv()
	config.SetEnvPrefix("vault")
	config.SetDefault(minimumLeaseTimeDurationConfig, "30s")
	config.SetDefault(vaultRoleConfig, "backend-application")
}

type APIclient struct {
	stateLock sync.RWMutex

	vaultClient *vault.Client

	minimumLeaseTimeDuration time.Duration

	obtainedViaLogin bool

	close func()
}

func NewAPIclient(ctx context.Context) (*APIclient, error) {

	minimumLeaseTimeDuration := config.GetDuration(minimumLeaseTimeDurationConfig)

	vaultAddr := config.GetString(vaultAddrConfig)
	vaultClient, err := vault.NewClient(&vault.Config{Address: vaultAddr})
	if err != nil {
		return nil, fmt.Errorf("vault client: %w", err)
	}
	c := &APIclient{
		vaultClient:              vaultClient,
		minimumLeaseTimeDuration: minimumLeaseTimeDuration,
	}
	c.initialize(ctx)
	return c, nil
}

func (c *APIclient) initialize(ctx context.Context) {

	c.stateLock.Lock()
	defer c.stateLock.Unlock()

	ctx2, cancel2 := context.WithCancel(ctx)
	started := make(chan struct{})
	done := make(chan struct{})
	c.close = func() {
		select {
		case <-started:
			cancel2()
		case <-ctx2.Done():
		}
		<-done
	}
	go func() {
		defer close(done)
		close(started)
		if err := c.run(ctx2); err != nil {
			log.Error().Err(err).Msg("vault client app run exited, relaunching...")
		}
		c.close = nil
		log.Info().Msg("vault client app run running...")
		select {
		case <-ctx2.Done():
		default:
			log.Info().Msg("vault client app run exited, relaunching...")
			c.initialize(ctx2)
			time.Sleep(1 * time.Second)
		}
	}()
}

func (c *APIclient) Healthy() bool {
	c.stateLock.RLock()
	defer c.stateLock.RUnlock()
	return c.close != nil
}

func (c *APIclient) Close() {

	c.stateLock.Lock()
	defer c.stateLock.Unlock()

	if c.close != nil {
		c.close()
		c.close = nil
		return
	}
}

func (c *APIclient) run(ctx context.Context) error {
	lookup, err := c.login()
	if err != nil {
		return fmt.Errorf("new app login: %w", err)
	}
	leaseDurationSeconds, err := c.leaseDurationSeconds(lookup)
	if err != nil {
		return fmt.Errorf("failed to parse Vault token lease duration: %w", err)
	}
	renewInterval := time.Duration(leaseDurationSeconds) * 5 / 6 * time.Second
loop:
	for {
		retries := 0
		select {
		case <-time.After(renewInterval):
			var backoff time.Duration = 1 * time.Second
			for ; retries < 5; retries++ {
				if c.obtainedViaLogin {
					lookup, err = c.vaultClient.Auth().Token().RenewSelf(0)
				} else {
					log.Debug().Msg("skipping RenewSelf — token not renewable in dev mode")
					break
				}
				if err != nil {
					log.Warn().
						Err(err).
						Dur("retrying_in", backoff).
						Msg("failed to renew Vault token")
					time.Sleep(backoff)
					backoff *= 2
					continue
				}
				break
			}
		case <-ctx.Done():
			break loop
		}
		if retries == 5 {
			log.Warn().Msg("retries exhausted, re-attempting full login")
			lookup, err = c.login()
			if err != nil {
				return fmt.Errorf("re-login after renew failures failed: %w", err)
			}
			// si re-login ok, on repart
		}

		// SUCCESS CASE...
		leaseDurationSeconds, err := c.leaseDurationSeconds(lookup)
		if err != nil {
			return fmt.Errorf("failed to parse Vault token lease duration: %w", err)
		}
		log.Info().
			Dur("new_lease_duration", time.Duration(leaseDurationSeconds)*time.Second).
			Msg("successfully renewed Vault token")
		renewInterval = time.Duration(leaseDurationSeconds) * 5 / 6 * time.Second
	}
	return nil
}

func (c *APIclient) leaseDurationSeconds(lookup *vault.Secret) (int64, error) {
	minimumLeaseDurationSeconds := int64(c.minimumLeaseTimeDuration / time.Second)
	leaseDurationRawData, ok := lookup.Data["ttl"]
	if !ok {
		return minimumLeaseDurationSeconds, fmt.Errorf("failed to parse token ttl")
	}
	leaseDurationRaw, ok := leaseDurationRawData.(json.Number)
	if !ok {
		return minimumLeaseDurationSeconds, fmt.Errorf("failed to parse token ttl data as number")
	}
	leaseDurationSeconds, err := leaseDurationRaw.Int64()
	if err != nil {
		return minimumLeaseDurationSeconds, fmt.Errorf("invalid lease duration format: %w", err)
	}
	if leaseDurationSeconds <= 1 {
		log.Warn().
			Dur("lease_duration", time.Duration(leaseDurationSeconds)*time.Second).
			Int64("using_lease_duration_secs", minimumLeaseDurationSeconds).
			Msg("lease duration too short using minimum lease duration from configuration")
		leaseDurationSeconds = minimumLeaseDurationSeconds
	}
	return leaseDurationSeconds, nil
}

func (c *APIclient) login() (*vault.Secret, error) {
	vaultRole := config.GetString(vaultRoleConfig)
	var err error
	var loginResp *vault.Secret
	var backoff time.Duration = 1 * time.Second
	for retries := 0; retries < 5; retries++ {
		var resp *vault.Secret
		if token := c.vaultClient.Token(); token != "" && token != "unset" {
			log.Info().
				Str("token", token).
				Msg("vault: VAULT_TOKEN already set, skipping GCP login — running in dev mode?")
			// simule une réponse de login avec un lease long arbitraire pour le dev
			c.obtainedViaLogin = false
			return &vault.Secret{
				Auth: &vault.SecretAuth{
					ClientToken:   token,
					LeaseDuration: int(c.minimumLeaseTimeDuration / time.Second),
				},
				Data: map[string]interface{}{
					"ttl": json.Number(fmt.Sprintf("%d", int(c.minimumLeaseTimeDuration/time.Second))),
				},
			}, nil
		}
		resp, err = c.vaultClient.Logical().Write("auth/gcp/login", map[string]any{
			"role": vaultRole,
		})
		if err != nil {
			log.Warn().
				Err(err).
				Msg("vault gcp login")
			time.Sleep(backoff)
			backoff *= 2
			continue
		}
		if resp == nil || resp.Auth == nil {
			err := fmt.Errorf("gcp login auth info missing in response")
			log.Warn().
				Err(err).
				Msg("vault gcp login: auth info missing in response")
			time.Sleep(backoff)
			backoff *= 2
			continue
		}

		// SUCCESS CASE
		vaultToken := resp.Auth.ClientToken

		c.stateLock.Lock()
		c.vaultClient.SetToken(vaultToken)
		c.stateLock.Unlock()

		loginResp = resp
		break
	}
	return loginResp, nil
}

func (c *APIclient) TransitDecrypt(path string, data map[string]interface{}) (*vault.Secret, error) {
	c.stateLock.RLock()
	defer c.stateLock.RUnlock()
	secret, err := c.vaultClient.Logical().Write(path, data)
	if err != nil {
		return nil, fmt.Errorf("vault transit decrypt, logical write: %w", err)
	}
	return secret, nil
}

func (c *APIclient) TransitEncrypt(path string, data map[string]interface{}) (*vault.Secret, error) {
	c.stateLock.RLock()
	defer c.stateLock.RUnlock()
	secret, err := c.vaultClient.Logical().Write(path, data)
	if err != nil {
		return nil, fmt.Errorf("vault transit decrypt, logical write: %w", err)
	}
	return secret, nil
}
