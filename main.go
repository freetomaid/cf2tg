package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

const (
	APIKey = "o1zrmHAF"
)

// OptimizationIP represents an individual IP entry in the API response
type OptimizationIP struct {
	Colo    string `json:"colo"`
	IP      string `json:"ip"`
	Latency int    `json:"latency"`
	Line    string `json:"line"`
	Loss    int    `json:"loss"`
	Node    string `json:"node"`
	Speed   int    `json:"speed"`
	Time    string `json:"time"`
}

// OptimizationIPResponse represents the full API response structure
type OptimizationIPResponse struct {
	Code  int                         `json:"code"`
	Total int                         `json:"total"`
	Info  map[string][]OptimizationIP `json:"info"`
}

// getOptimizationIP sends a POST request to fetch the optimization IPs
func getOptimizationIP(ipType string) (*OptimizationIPResponse, error) {
	url := "https://api.hostmonit.com/get_optimization_ip"
	data := map[string]string{
		"key":  APIKey,
		"type": ipType,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data: %v", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	var response OptimizationIPResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %v", err)
	}

	return &response, nil
}

func sendMessageToTelegram(TelegramBotToken, chatID, message, imageURL string) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", TelegramBotToken)

	data := map[string]interface{}{
		"chat_id":    chatID,
		"text":       message,
		"parse_mode": "MarkdownV2", // Enable MarkdownV2 formatting
		"photo":      imageURL,     // Add the image URL
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal Telegram message data: %v", err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to send message to Telegram: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code: %d, response: %s", resp.StatusCode, body)
	}
	fmt.Print(resp)
	return nil
}

func main() {
	imageURL := "images.png"

	TelegramBotToken, ok := os.LookupEnv("BOT_TOKEN")
	if !ok {
		fmt.Printf("BOT_TOKEN not set\n")
	}
	TelegramChatID, ok := os.LookupEnv("CHAT_ID")
	if !ok {
		fmt.Printf("CHAT_ID not set\n")
	}

	ipv4Response, err := getOptimizationIP("v4")
	if err != nil {
		fmt.Printf("Error getting IPv4: %v\n", err)
		return
	}

	ipv4 := []string{}
	if ipv4Response.Code == 200 && ipv4Response.Total > 0 {
		for _, region := range ipv4Response.Info {
			for _, ip := range region {
				ipv4 = append(ipv4, ip.IP)
			}
		}
	}

	ipv6Response, err := getOptimizationIP("v6")
	if err != nil {
		fmt.Printf("Error getting IPv6: %v\n", err)
		return
	}

	ipv6 := []string{}
	if ipv6Response.Code == 200 && ipv6Response.Total > 0 {
		for _, region := range ipv6Response.Info {
			for _, ip := range region {
				ipv6 = append(ipv6, ip.IP)
			}
		}
	}

	if len(ipv4) > 0 || len(ipv6) > 0 {
		message := "CloudFlare Optimized IPs:\n\n"
		if len(ipv4) > 0 {
			message += "IPv4:\n" + formatIPs(ipv4[:min(len(ipv4), 25)]) + "\n\n"
		}
		if len(ipv6) > 0 {
			message += "IPv6:\n" + formatIPs(ipv6[:min(len(ipv6), 25)]) + "\n\n"
		}
		message += "@infoakungratis"
		if err := sendMessageToTelegram(TelegramBotToken, TelegramChatID, message, imageURL); err != nil {
			fmt.Printf("Error sending message to Telegram: %v\n", err)
		}
	}
}

func formatIPs(ips []string) string {
	formattedIPs := make([]string, len(ips))
	for i, ip := range ips {
		formattedIPs[i] = "`" + ip + "`"
	}
	return strings.Join(formattedIPs, "\n")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
