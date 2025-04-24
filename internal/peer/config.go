package peer

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"log/slog"
	"math/big"
	"net/http"
	"strconv"
	"time"

	"github.com/quic-go/quic-go"
	"github.com/vs-ude/btml/internal/model"
	"github.com/vs-ude/btml/internal/structs"
	"github.com/vs-ude/btml/internal/telemetry"
)

type Config struct {
	Name        string
	TrackerURL  string
	UpdateFreq  time.Duration
	ModelConf   *model.Config
	Addr        string
	PeersetSize int
	TelConf     *telemetry.TelemetryConf
}

func Autoconf(c *Config) error {
	var resp *http.Response
	var err error
	code := http.StatusServiceUnavailable
	for code == http.StatusServiceUnavailable {
		resp, err = http.Get(c.TrackerURL + "/whoami")
		if err != nil {
			return err
		}
		code = resp.StatusCode
		if code == http.StatusOK {
			break
		} else if code != http.StatusServiceUnavailable {
			return errors.New("unable to connect to tracker for autoconfiguration")
		}
		slog.Info("Waiting for tracker to be ready")
		time.Sleep(time.Second * 5)
	}
	defer resp.Body.Close()
	body, err := getResponseBody(resp)
	if err != nil {
		return err
	}
	whoami := new(structs.WhoAmI)
	err = json.Unmarshal(*body, whoami)
	if err != nil {
		return fmt.Errorf("unable to parse whoami response body data from tracker\n%w", err)
	}

	c.Name = strconv.Itoa(whoami.Id)
	c.Addr = whoami.ExtIp
	c.UpdateFreq = whoami.UpdateFreq
	c.ModelConf.Dataset = whoami.Dataset
	c.ModelConf.Name = c.Name
	c.PeersetSize = 5 // TODO make configurable
	c.TelConf = &whoami.Telemetry

	return nil
}

func generateQUICConfig() *quic.Config {
	return &quic.Config{
		KeepAlivePeriod: time.Second * 15,
		MaxIdleTimeout:  time.Second * 60,
	}
}

func generateTLSConfig() *tls.Config {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}
	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(time.Hour * 24 * 180),
	}
	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		panic(err)
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		panic(err)
	}
	return &tls.Config{
		Certificates:       []tls.Certificate{tlsCert},
		InsecureSkipVerify: true,
		NextProtos:         []string{"btml"},
	}
}
