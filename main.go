package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sync"
	"time"
)

var (
	token    = "user token"
	password = "password"
	mfaToken string
	mfaMutex sync.Mutex
)

func logInfo(format string, args ...interface{}) {
	fmt.Printf("[%s] %s\n", time.Now().Format("15:04:05"), fmt.Sprintf(format, args...))
}
func logSuccess(format string, args ...interface{}) {
	fmt.Printf("[%s] %s\n", time.Now().Format("15:04:05"), fmt.Sprintf(format, args...))
}
func logError(format string, args ...interface{}) {
	fmt.Printf("[%s] %s\n", time.Now().Format("15:04:05"), fmt.Sprintf(format, args...))
}
func logWarning(format string, args ...interface{}) {
	fmt.Printf("[%s] %s\n", time.Now().Format("15:04:05"), fmt.Sprintf(format, args...))
}

func main() {
	for {
		if newToken := getMFAToken(token, password); newToken != "" {
			mfaMutex.Lock()
			mfaToken = newToken

			err := os.WriteFile("mfa.txt", []byte(mfaToken), 0644)
			if err != nil {
				logError("MFA Token dosyaya kaydedilemedi: %v", err)
			} else {
				logSuccess("MFA Token mfa_token.txt dosyasına kaydedildi")
			}

			mfaMutex.Unlock()
		} else {
			logError("MFA Token alınamadı, tekrar deneniyor...")
		}

		time.Sleep(5 * time.Minute)
	}
}

func getMFAToken(token, password string) string {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, _ := http.NewRequest(
		"PATCH",
		"https://discord.com/api/v9/guilds/0/vanity-url",
		bytes.NewBufferString(`{"code":null}`),
	)

	req.Header.Set("Authorization", token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/138.0.0.0 Safari/537.36")
	req.Header.Set(
		"X-Super-Properties",
		"eyJvcyI6IldpbmRvd3MiLCJicm93c2VyIjoiQ2hyb21lIiwiZGV2aWNlIjoiIiwic3lzdGVtX2xvY2FsZSI6InRyLVRSIiwiYnJvd3Nlcl91c2VyX2FnZW50IjoiTW96aWxsYS81LjAgKFdpbmRvd3MgTlQgMTAuMDsgV2luNjQ7IHg2NCkiLCJicm93c2VyX3ZlcnNpb24iOiIxMjEuMC4wLjAiLCJvc192ZXJzaW9uIjoiMTAiLCJyZWZlcnJlciI6IiIsInJlZmVycmluZ19kb21haW4iOiIiLCJyZWZlcnJlcl9jdXJyZW50IjoiIiwicmVmZXJyaW5nX2RvbWFpbl9jdXJyZW50IjoiIiwicmVsZWFzZV9jaGFubmVsIjoic3RhYmxlIiwiY2xpZW50X2J1aWxkX251bWJlciI6MjAwODQyLCJjbGllbnRfZXZlbnRfc291cmNlIjpudWxsfQ==",
	)

	resp, err := client.Do(req)
	if err != nil {
		logError("URL isteği başarısız: %v", err)
		return ""
	}

	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	var data map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &data); err != nil {
		logError("JSON ayrıştırma hatası (ticket): %v", err)
		return ""
	}

	var ticket string
	if mfa, ok := data["mfa"].(map[string]interface{}); ok && mfa["ticket"] != nil {
		ticket = mfa["ticket"].(string)
	} else if data["ticket"] != nil {
		ticket, _ = data["ticket"].(string)
	}

	if ticket == "" {
		logError("MFA Ticket alınamadı")
		return ""
	}

	if len(ticket) > 15 {
		logSuccess("MFA Ticket alındı: %s", ticket[:10]+"..."+ticket[len(ticket)-5:])
	} else {
		logSuccess("MFA Ticket alındı: %s", ticket)
	}

	mfaPayload := fmt.Sprintf(
		`{
            "ticket":"%s",
            "mfa_type":"password",
            "data":"%s"
        }`,
		ticket,
		password,
	)

	mfaReq, _ := http.NewRequest(
		"POST",
		"https://discord.com/api/v9/mfa/finish",
		bytes.NewBufferString(mfaPayload),
	)

	mfaReq.Header.Set("Authorization", token)
	mfaReq.Header.Set("Content-Type", "application/json")
	mfaReq.Header.Set("User-Agent", "Mozilla/5.0")
	mfaReq.Header.Set("X-Super-Properties", req.Header.Get("X-Super-Properties"))

	mfaResp, err := client.Do(mfaReq)
	if err != nil {
		logError("MFA isteği başarısız: %v", err)
		return ""
	}

	mfaBytes, _ := ioutil.ReadAll(mfaResp.Body)
	mfaResp.Body.Close()

	var tokenData map[string]interface{}
	if err := json.Unmarshal(mfaBytes, &tokenData); err != nil {
		logError("JSON ayrıştırma hatası (token): %v", err)
		return ""
	}

	if newToken, ok := tokenData["token"].(string); ok {
		if len(newToken) > 15 {
			logSuccess("MFA Token başarıyla alındı: %s", newToken[:10]+"..."+newToken[len(newToken)-5:])
		} else {
			logSuccess("MFA Token başarıyla alındı: %s", newToken)
		}
		return newToken
	}

	logError("MFA Token alınamadı")
	return ""
}
