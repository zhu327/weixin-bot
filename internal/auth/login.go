package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/mdp/qrterminal/v3"
	"github.com/zhu327/weixin-bot/internal/api"
)

const qrPollInterval = 2 * time.Second

// LoginCallbacks customizes QR login UX. Nil fields keep the default behavior (print to [os.Stderr]).
type LoginCallbacks struct {
	// OnQRCode is invoked when a new QR session starts. imageURL is qrcode_img_content; qrcode is the opaque token used for polling.
	OnQRCode func(imageURL string, qrcode string)
	// OnStatus is invoked when the QR status changes (e.g. scaned, confirmed, expired).
	OnStatus func(status string)
}

// Login performs QR login, optionally reusing an existing file when force is false.
// httpClient is used for QR and polling requests; if nil, [http.DefaultClient] is used.
// callbacks may be nil; non-nil callbacks override the corresponding default behavior only for set fields.
func Login(
	ctx context.Context,
	httpClient *http.Client,
	baseURL, tokenPath string,
	force bool,
	callbacks *LoginCallbacks,
) (*Data, error) {
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
		qrRaw, err := api.FetchQRCode(httpClient, ctx, baseURL)
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
		notifyQRCode(callbacks, link, qr.Qrcode)

		var lastStatus string
		for {
			if err := ctx.Err(); err != nil {
				return nil, err
			}
			stRaw, err := api.PollQRStatus(httpClient, ctx, baseURL, qr.Qrcode)
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
				notifyStatus(callbacks, st.Status)
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

func notifyQRCode(cb *LoginCallbacks, imageURL, qrcode string) {
	if cb != nil && cb.OnQRCode != nil {
		cb.OnQRCode(imageURL, qrcode)
		return
	}
	fmt.Fprintf(os.Stderr, "[weixin-bot] 在微信中打开以下链接完成登录，或用微信扫描下方二维码:\n")
	fmt.Fprintf(os.Stderr, "%s\n", imageURL)
	if strings.HasPrefix(imageURL, "http://") || strings.HasPrefix(imageURL, "https://") {
		fmt.Fprintln(os.Stderr)
		qrterminal.GenerateHalfBlock(imageURL, qrterminal.M, os.Stderr)
		fmt.Fprintln(os.Stderr)
	}
}

func notifyStatus(cb *LoginCallbacks, status string) {
	if cb != nil && cb.OnStatus != nil {
		cb.OnStatus(status)
		return
	}
	switch status {
	case "scaned":
		fmt.Fprintf(os.Stderr, "[weixin-bot] QR code scanned. Confirm the login inside WeChat.\n")
	case "confirmed":
		fmt.Fprintf(os.Stderr, "[weixin-bot] Login confirmed.\n")
	case "expired":
		fmt.Fprintf(os.Stderr, "[weixin-bot] QR code expired. Requesting a new one...\n")
	}
}
