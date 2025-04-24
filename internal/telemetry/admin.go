package telemetry

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path"
	"time"
)

func InitConf(tc *TelemetryConf, gc *GrafanaConf, configPath string) error {
	tc.Suffix = time.Now().Format("2006_01_02_15_04_05")
	if err := getAdminToken(tc, configPath); err != nil {
		return fmt.Errorf("failed to get admin token: %w", err)
	}
	if err := storeSuffix(tc); err != nil {
		return fmt.Errorf("failed to store the run suffix: %w", err)
	}
	if err := provisionGrafanaDatasource(gc, tc); err != nil {
		return fmt.Errorf("failed to provision Grafana datasource: %w", err)
	}
	return nil
}

func getAdminToken(c *TelemetryConf, configPath string) error {
	var tokenFile string
	// Check if admin token already exists on disk
	if configPath != "" {
		dir := path.Dir(configPath)
		tokenFile = path.Join(dir, ".token")

		tokenBytes, err := os.ReadFile(tokenFile)
		if err == nil && len(tokenBytes) > 0 {
			c.Token = string(tokenBytes)
			slog.Debug("Using existing admin token from disk", "path", tokenFile)
			// Check if the token is still valid
			healthReq, err := setupInfluxPostRequest(c, "configure/database?format=csv")
			if err == nil {
				healthReq.Method = "GET"
				healthResp, err := http.DefaultClient.Do(healthReq)
				if err == nil && healthResp.StatusCode == http.StatusOK {
					slog.Debug("Token validated with health check")
					return nil
				}
				if healthResp != nil {
					healthResp.Body.Close()
				}
				slog.Debug("Existing token failed health check, will request new token")
			}
		}
	}

	req, err := setupInfluxPostRequest(c, "configure/token/admin")
	if err != nil {
		return fmt.Errorf("failed to create admin token request: %w", err)
	}

	resp, err := execRequest(req, nil)
	if err != nil {
		return fmt.Errorf("failed to get admin token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code for admin token: %s, body: %s", resp.Status, string(bodyBytes))
	}

	var tokenResp struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return fmt.Errorf("failed to decode token response: %w", err)
	}

	c.Token = tokenResp.Token
	slog.Debug("Successfully obtained admin token")

	// Store the admin token to disk for persistence
	if tokenFile != "" {
		err := os.WriteFile(tokenFile, []byte(c.Token), 0666)
		if err != nil {
			slog.Warn("Failed to write admin token to disk", "error", err, "path", tokenFile)
		} else {
			slog.Debug("Admin token saved to disk", "path", tokenFile)
		}
	}
	return nil
}

// storeSuffix saves the suffix of the upcoming run to the runs table. This
// has the side effect of creating the DB if it doesn't exist yet.
func storeSuffix(c *TelemetryConf) error {
	req, err := setupInfluxPostRequest(c, fmt.Sprintf("write_lp?db=%s&precision=second", c.DB))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "text/plain; charset=utf-8")
	body := bytes.NewBuffer(fmt.Appendf(nil, "runs run=\"%s\"", c.Suffix))

	resp, err := execRequest(req, body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("unexpected status code: %s", resp.Status)
	}

	slog.Debug("Stored run suffix in InfluxDB", "run", c.Suffix)
	return nil
}

func provisionGrafanaDatasource(c *GrafanaConf, tc *TelemetryConf) error {
	// Get the datasource uid
	createRequest := func(endpoint string) (*http.Request, error) {
		req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/%s", c.URL, endpoint), nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}
		req.SetBasicAuth(c.User, c.Passwd)
		req.Header.Set("Content-Type", "application/json")
		return req, nil
	}

	req, err := createRequest("datasources/name/InfluxDB")
	if err != nil {
		return err
	}

	resp, err := execRequest(req, nil)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %s", resp.Status)
	}

	var datasource struct {
		Uid string `json:"uid"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&datasource); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	slog.Debug("Retrieved Grafana datasource info", "uid", datasource.Uid)

	// Get the current datasource configuration
	req, err = createRequest(fmt.Sprintf("datasources/uid/%s", datasource.Uid))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err = execRequest(req, nil)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %s", resp.Status)
	}

	var data map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	// Update the datasource with the token
	req, err = createRequest(fmt.Sprintf("datasources/uid/%s", datasource.Uid))
	if err != nil {
		return fmt.Errorf("failed to create update request: %w", err)
	}
	req.Method = "PUT"

	data["database"] = tc.Suffix
	data["secureJsonData"] = map[string]any{
		"token": tc.Token,
	}
	delete(data, "secureJsonFields")
	d, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	updateBody := bytes.NewBuffer(d)

	slog.Debug("Updating Grafana datasource with token", "uid", datasource.Uid)

	resp, err = execRequest(req, updateBody)
	if err != nil {
		return fmt.Errorf("failed to send update request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code updating datasource: %s, body: %s", resp.Status, string(bodyBytes))
	}

	slog.Debug("Successfully updated Grafana datasource with token")
	return nil
}

func setupInfluxPostRequest(c *TelemetryConf, p string) (*http.Request, error) {
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/v3/%s", c.URL, p), nil)
	if err == nil {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Token))
		req.Header.Set("Content-Type", "application/json")
	}
	return req, err
}

func execRequest(req *http.Request, body *bytes.Buffer) (*http.Response, error) {
	if body != nil {
		req.Body = io.NopCloser(body)
		req.Header.Set("Content-Length", fmt.Sprintf("%d", body.Len()))
	}

	var resp *http.Response
	var err error
	for range 3 {
		resp, err = http.DefaultClient.Do(req)
		if err == nil {
			break
		}
		time.Sleep(time.Second * 2)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	return resp, nil
}
