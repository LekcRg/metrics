package req

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/LekcRg/metrics/internal/cgzip"
	"github.com/LekcRg/metrics/internal/config"
	"github.com/LekcRg/metrics/internal/crypto"
	"github.com/LekcRg/metrics/internal/logger"
	"github.com/LekcRg/metrics/internal/models"
	"github.com/LekcRg/metrics/internal/retry"
)

type RequestArgs struct {
	Ctx     context.Context
	URL     string
	Metrics []models.Metrics
	Config  config.AgentConfig
}

func HTTPRequest(args RequestArgs) error {
	var err error
	body, err := json.Marshal(args.Metrics)
	if err != nil {
		logger.Log.Error("Error while generate json")
		return err
	}

	if args.Config.PublicKey != nil {
		body, err = rsa.EncryptPKCS1v15(rand.Reader, args.Config.PublicKey, body)
		if err != nil {
			return err
		}
	}

	req, err := cgzip.GetGzippedReq(args.Ctx, args.URL, body)
	if err != nil {
		return err
	}

	if args.Config.Key != "" {
		sha := crypto.GenerateHMAC(body, args.Config.Key)
		req.Header.Set("HashSHA256", sha)
	}

	req.Header.Set("Content-Type", "application/json")
	if args.Config.IP != "" {
		req.Header.Set("X-Real-IP", args.Config.IP)
	}
	client := &http.Client{}
	var resp *http.Response

	err = retry.Retry(args.Ctx, func() error {
		resp, err = client.Do(req)
		if err != nil {
			return err
		}

		defer func() {
			if resp != nil {
				resp.Body.Close()
			}
		}()
		return nil
	})

	if err != nil {
		return err
	}

	if resp != nil && resp.StatusCode > 299 {
		return errors.New("Server answered with status code: " +
			strconv.Itoa(resp.StatusCode))
	}

	return nil
}
