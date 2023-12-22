package alarm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	blog "log"
	"net/http"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/log"
)

var (
	Env = ""
	Dev = "dev"
)

const (
	SamaEnv   = "SAMA_INSCRIPTION_ENV"
	SamaHooks = "SAMA_INSCRIPTION_SLACK_HOOKS"
)

const defaultHTTPTimeout = 20 * time.Second

func init() {
	Env = os.Getenv(SamaEnv)
}

func ValidateEnv() {
	Env = os.Getenv(SamaEnv)
	if Env == "" {
		blog.Fatal("SAMA_INSCRIPTION_ENV is empty")
	}
	hooks := os.Getenv(SamaHooks)
	if hooks == "" {
		blog.Fatal("SAMA_INSCRIPTION_SLACK_HOOKS is empty")
	}
}

func Slack(ctx context.Context, msg string) {
	hooks := os.Getenv(SamaHooks)
	if hooks == "" {
		log.Error("hooks is empty")
		return
	}
	if Env == Dev {
		//log.Logger().WithField("msg", msg).Debug("evn is dev does not send msg to slack")
		return
	}

	body, err := json.Marshal(map[string]interface{}{
		"text": fmt.Sprintf("Env: %s \nMsg: %s", Env, msg),
	})
	if err != nil {
		//log.Logger().Error("json marshal failed")
		return
	}

	client := http.Client{
		Timeout: defaultHTTPTimeout,
	}
	req, err := http.NewRequestWithContext(ctx, "POST", hooks, ioutil.NopCloser(bytes.NewReader(body)))
	if err != nil {
		//log.Logger().WithField("error", err.Error()).Error("new request with context failed")
		return
	}
	req.Header.Set("Content-type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		//log.Logger().WithField("error", err.Error()).Error("request failed")
		return
	}
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}

	if resp.StatusCode != http.StatusOK {
		//log.Logger().WithField("code", resp.StatusCode).Error("response status code is not 200")
		return
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		//log.Logger().WithField("error", err.Error()).Error("failed to read response body")
		return
	}
	log.Warn("send alarm message", "comment", string(data))
}
