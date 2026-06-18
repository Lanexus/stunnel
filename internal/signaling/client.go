package signaling

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type PeerInfo struct {
	IP        string `json:"ip"`
	Port      string `json:"port"`
	Timestamp int64  `json:"ts"`
}

type SignalingClient struct {
	server string
	client *http.Client
}

func NewSignalingClient(server string) *SignalingClient {
	return &SignalingClient{
		server: server,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func HashSecret(secret string) string {
	h := sha256.Sum256([]byte(secret))
	return hex.EncodeToString(h[:8])
}

func (s *SignalingClient) Register(secret, ip, port string) error {
	info := PeerInfo{
		IP:        ip,
		Port:      port,
		Timestamp: time.Now().Unix(),
	}

	data, err := json.Marshal(info)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	hash := HashSecret(secret)
	url := fmt.Sprintf("%s/register?secret=%s", s.server, hash)

	resp, err := s.client.Post(url, "application/json", bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("post: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("register failed: %s %s", resp.Status, body)
	}

	return nil
}

func (s *SignalingClient) Lookup(secret string) (*PeerInfo, error) {
	hash := HashSecret(secret)
	url := fmt.Sprintf("%s/lookup/%s", s.server, hash)

	resp, err := s.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("get: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("lookup failed: %s", resp.Status)
	}

	var info PeerInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}

	return &info, nil
}
