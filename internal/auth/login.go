package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/mdp/qrterminal/v3"
	"github.com/zhu327/weixin-bot/internal/api"
)

const qrPollInterval = 2 * time.Second

// Login performs QR login, optionally reusing an existing file when force is false.
func Login(ctx context.Context, baseURL, tokenPath string, force bool) (*Data, error) {
	path, err := ResolveTokenPath(tokenPath)
	if err != nil {
		return nil, err
	}
	if !force {
		existing, err := Load(path)
		if err != nil {
			return nil, err
		}
		if existing != nil {
			return existing, nil
		}
	}

	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		qrRaw, err := api.FetchQRCode(ctx, baseURL)
		if err != nil {
			return nil, err
		}
		var qr struct {
			Qrcode           string `json:"qrcode"`
			QrcodeImgContent string `json:"qrcode_img_content"`
		}
		if err := json.Unmarshal(qrRaw, &qr); err != nil {
			return nil, err
		}
		link := strings.TrimSpace(qr.QrcodeImgContent)
		fmt.Fprintf(os.Stderr, "[weixin-bot] 在微信中打开以下链接完成登录，或用微信扫描下方二维码:\n")
		fmt.Fprintf(os.Stderr, "%s\n", link)
		if strings.HasPrefix(link, "http://") || strings.HasPrefix(link, "https://") {
			fmt.Fprintln(os.Stderr)
			qrterminal.GenerateHalfBlock(link, qrterminal.M, os.Stderr)
			fmt.Fprintln(os.Stderr)
		}

		var lastStatus string
		for {
			if err := ctx.Err(); err != nil {
				return nil, err
			}
			stRaw, err := api.PollQRStatus(ctx, baseURL, qr.Qrcode)
			if err != nil {
				return nil, err
			}
			var st struct {
				Status      string `json:"status"`
				BotToken    string `json:"bot_token"`
				IlinkBotID  string `json:"ilink_bot_id"`
				IlinkUserID string `json:"ilink_user_id"`
				BaseURL     string `json:"baseurl"`
			}
			if err := json.Unmarshal(stRaw, &st); err != nil {
				return nil, err
			}
			if st.Status != lastStatus {
				switch st.Status {
				case "scaned":
					fmt.Fprintf(os.Stderr, "[weixin-bot] QR code scanned. Confirm the login inside WeChat.\n")
				case "confirmed":
					fmt.Fprintf(os.Stderr, "[weixin-bot] Login confirmed.\n")
				case "expired":
					fmt.Fprintf(os.Stderr, "[weixin-bot] QR code expired. Requesting a new one...\n")
				}
				lastStatus = st.Status
			}
			if st.Status == "confirmed" {
				if st.BotToken == "" || st.IlinkBotID == "" || st.IlinkUserID == "" {
					return nil, errors.New("QR login confirmed, but the API did not return bot credentials")
				}
				bu := st.BaseURL
				if bu == "" {
					bu = baseURL
				}
				d := &Data{
					Token:     st.BotToken,
					BaseURL:   bu,
					AccountID: st.IlinkBotID,
					UserID:    st.IlinkUserID,
				}
				if err := Save(path, d); err != nil {
					return nil, err
				}
				return d, nil
			}
			if st.Status == "expired" {
				break
			}
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(qrPollInterval):
			}
		}
	}
}
