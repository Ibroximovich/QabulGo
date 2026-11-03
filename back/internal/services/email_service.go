package services

import (
	"bytes"
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/smtp"
	"strings"
	"time"

	"qabulgo-backend/internal/config"
)

//go:embed assets/logo.jpg
var logoBytes []byte

type EmailService interface {
	SendVerificationEmail(toEmail, code string) error
	SendPasswordResetEmail(toEmail, code string) error
}

type emailService struct {
	cfg *config.Config
}

func NewEmailService(cfg *config.Config) EmailService {
	return &emailService{cfg: cfg}
}

func (s *emailService) SendVerificationEmail(toEmail, code string) error {
	subject := "QabulGo — Email tasdiqlash kodi"
	title := "Email manzilini tasdiqlash"
	greeting := "Salom!"
	message := "QabulGo platformasiga xush kelibsiz. Hisobingizni faollashtirish va tizim imkoniyatlaridan to‘liq foydalanish uchun quyidagi tasdiqlash kodini kiriting:"
	expireMin := "10"

	htmlBody := s.buildHTMLTemplate(title, greeting, message, code, expireMin)
	plainText := fmt.Sprintf("Salom!\n\nQabulGo platformasiga xush kelibsiz.\nSizning bir martalik tasdiqlash kodingiz: %s\n\nUshbu kod %s daqiqa davomida amal qiladi. Kodni hech kimga bermang!", code, expireMin)

	return s.sendMail(toEmail, subject, plainText, htmlBody, code)
}

func (s *emailService) SendPasswordResetEmail(toEmail, code string) error {
	subject := "QabulGo — Parolni tiklash kodi"
	title := "Parolni tiklash so‘rovi"
	greeting := "Salom!"
	message := "QabulGo akkauntingiz parolini tiklash uchun so‘rov oldik. Yangi parol o‘rnatish uchun quyidagi xavfsizlik kodidan foydalaning:"
	expireMin := "15"

	htmlBody := s.buildHTMLTemplate(title, greeting, message, code, expireMin)
	plainText := fmt.Sprintf("Salom!\n\nQabulGo akkauntingiz parolini tiklash uchun so‘rov oldik.\nSizning parolni tiklash kodingiz: %s\n\nUshbu kod %s daqiqa davomida amal qiladi. Agar buni siz so‘ramagan bo‘lsangiz, ushbu xabarni e’tiborsiz qoldiring.", code, expireMin)

	return s.sendMail(toEmail, subject, plainText, htmlBody, code)
}

func (s *emailService) sendMail(toEmail, subject, plainText, htmlBody, code string) error {
	// Xavfsiz tizim jurnali: har qanday email uzatilganda kod server logida ham qayd etiladi
	log.Printf("[AUTH EMAIL CODE] To: %s | Code: %s | Subject: %s", toEmail, code, subject)

	// Dev yoki Mock rejimida SMTP xatosi bo'lmasligi uchun konsolga chiqarish
	if s.cfg.EmailMock || (s.cfg.SMTPHost == "" && !strings.HasPrefix(s.cfg.SMTPPassword, "re_")) {
		log.Printf("\n========================================\n[DEV EMAIL MOCK] To: %s\nSubject: %s\nCode: %s\n%s\n========================================\n", toEmail, subject, code, plainText)
		return nil
	}

	// Resend aniqlanganda (Host smtp.resend.com yoki kalit re_ bilan boshlansa)
	// DigitalOcean, AWS va boshqa cloud provayderlar port 587 ni bloklashini chetlab o'tish uchun
	// Resend REST HTTPS API (Port 443) orqali yuboramiz. Bu tez (200ms) va 100% kafolatlangan!
	if strings.Contains(s.cfg.SMTPHost, "resend") || strings.HasPrefix(s.cfg.SMTPPassword, "re_") {
		return s.sendViaResendAPI(toEmail, subject, plainText, htmlBody)
	}

	// Haqiqiy SMTP (masalan Gmail) orqali chiroyli HTML + Logo bilan jo'natish
	addr := fmt.Sprintf("%s:%d", s.cfg.SMTPHost, s.cfg.SMTPPort)
	auth := smtp.PlainAuth("", s.cfg.SMTPUser, s.cfg.SMTPPassword, s.cfg.SMTPHost)

	boundary := "----=_Part_QabulGo_MultiPart_Boundary"
	encodedSubject := fmt.Sprintf("=?UTF-8?B?%s?=", base64.StdEncoding.EncodeToString([]byte(subject)))

	var msg bytes.Buffer
	msg.WriteString(fmt.Sprintf("From: %s\r\n", s.cfg.SMTPFrom))
	msg.WriteString(fmt.Sprintf("To: %s\r\n", toEmail))
	msg.WriteString(fmt.Sprintf("Subject: %s\r\n", encodedSubject))
	msg.WriteString("MIME-Version: 1.0\r\n")
	msg.WriteString(fmt.Sprintf("Content-Type: multipart/related; boundary=\"%s\"\r\n", boundary))
	msg.WriteString("\r\n")

	// HTML Qismi
	msg.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	msg.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
	msg.WriteString("Content-Transfer-Encoding: 8bit\r\n")
	msg.WriteString("\r\n")
	msg.WriteString(htmlBody)
	msg.WriteString("\r\n\r\n")

	// Logotip (Inline CID Attachment)
	if len(logoBytes) > 0 {
		msg.WriteString(fmt.Sprintf("--%s\r\n", boundary))
		msg.WriteString("Content-Type: image/jpeg; name=\"logo.jpg\"\r\n")
		msg.WriteString("Content-Transfer-Encoding: base64\r\n")
		msg.WriteString("Content-ID: <qabulgo-logo>\r\n")
		msg.WriteString("Content-Disposition: inline; filename=\"logo.jpg\"\r\n")
		msg.WriteString("\r\n")
		msg.WriteString(formatBase64Lines(logoBytes))
		msg.WriteString("\r\n\r\n")
	}

	msg.WriteString(fmt.Sprintf("--%s--\r\n", boundary))

	err := smtp.SendMail(addr, auth, s.cfg.SMTPUser, []string{toEmail}, msg.Bytes())
	if err != nil {
		log.Printf("Failed to send email via SMTP: %v\n", err)
		return fmt.Errorf("email yuborishda xatolik yuz berdi: %w", err)
	}

	return nil
}

func (s *emailService) sendViaResendAPI(toEmail, subject, plainText, htmlBody string) error {
	apiKey := strings.TrimSpace(s.cfg.SMTPPassword)
	if apiKey == "" {
		return errors.New("resend API kaliti topilmadi")
	}

	from := s.cfg.SMTPFrom
	if from == "" {
		from = "QabulGo <no-reply@qabulgo.uz>"
	}

	payload := map[string]interface{}{
		"from":    from,
		"to":      []string{toEmail},
		"subject": subject,
		"html":    htmlBody,
		"text":    plainText,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("json marshal xatoligi: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.resend.com/emails", bytes.NewReader(payloadBytes))
	if err != nil {
		return fmt.Errorf("resend request yaratishda xatolik: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[Resend API Error] Connection failed: %v", err)
		return fmt.Errorf("resend API bilan bog'lanishda xatolik: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		log.Printf("[Resend API Error] HTTP %d: %s", resp.StatusCode, string(bodyBytes))

		// Agar domen tasdiqlanmagan bo'lsa (403), avtomatik onboarding@resend.dev orqali qayta urinish
		if resp.StatusCode == 403 && strings.Contains(string(bodyBytes), "not verified") && from != "QabulGo <onboarding@resend.dev>" {
			log.Printf("[Resend Auto-Fallback] Domen tasdiqlanmagan, onboarding@resend.dev orqali qayta urinilmoqda...")
			payload["from"] = "QabulGo <onboarding@resend.dev>"
			fallbackBytes, _ := json.Marshal(payload)
			fbReq, fbErr := http.NewRequest("POST", "https://api.resend.com/emails", bytes.NewReader(fallbackBytes))
			if fbErr == nil {
				fbReq.Header.Set("Authorization", "Bearer "+apiKey)
				fbReq.Header.Set("Content-Type", "application/json")
				fbResp, fbErr2 := client.Do(fbReq)
				if fbErr2 == nil {
					defer fbResp.Body.Close()
					if fbResp.StatusCode < 400 {
						log.Printf("[Resend Auto-Fallback Success] Email successfully sent to %s via onboarding@resend.dev", toEmail)
						return nil
					}
					fbBody, _ := io.ReadAll(fbResp.Body)
					log.Printf("[Resend Auto-Fallback Error] HTTP %d: %s", fbResp.StatusCode, string(fbBody))
				}
			}
		}

		return fmt.Errorf("resend email yuborishda xatolik (%d): %s", resp.StatusCode, string(bodyBytes))
	}

	log.Printf("[Resend API Success] Email successfully sent to %s via Resend HTTPS API", toEmail)
	return nil
}

func formatBase64Lines(data []byte) string {
	encoded := base64.StdEncoding.EncodeToString(data)
	var buf bytes.Buffer
	for len(encoded) > 76 {
		buf.WriteString(encoded[:76])
		buf.WriteString("\r\n")
		encoded = encoded[76:]
	}
	buf.WriteString(encoded)
	return buf.String()
}

func (s *emailService) buildHTMLTemplate(title, greeting, message, code, expireMin string) string {
	return `<!DOCTYPE html>
<html lang="uz">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>QabulGo</title>
</head>
<body style="margin: 0; padding: 0; background-color: #F1F5F9; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif; -webkit-font-smoothing: antialiased;">
  <table role="presentation" width="100%" border="0" cellspacing="0" cellpadding="0" style="background-color: #F1F5F9; padding: 40px 15px;">
    <tr>
      <td align="center">
        <table role="presentation" width="100%" border="0" cellspacing="0" cellpadding="0" style="max-width: 520px; background-color: #FFFFFF; border-radius: 20px; overflow: hidden; border: 1px solid #E2E8F0; box-shadow: 0 10px 25px -5px rgba(8, 35, 74, 0.08);">
          <!-- Top Accent Gradient Line -->
          <tr>
            <td style="height: 6px; background: linear-gradient(90deg, #0866F5 0%, #08BDBA 100%);"></td>
          </tr>
          
          <!-- Header with Logo and Brand -->
          <tr>
            <td align="center" style="padding: 36px 32px 16px 32px;">
              <table role="presentation" border="0" cellspacing="0" cellpadding="0">
                <tr>
                  <td align="center">
                    <img src="cid:qabulgo-logo" alt="QabulGo Logo" width="68" height="68" style="display: block; width: 68px; height: 68px; border-radius: 16px; margin-bottom: 12px; box-shadow: 0 6px 16px rgba(8, 102, 245, 0.2);" />
                    <div style="font-size: 26px; font-weight: 800; color: #08234A; letter-spacing: -0.5px; line-height: 1.2;">
                      Qabul<span style="color: #0866F5;">Go</span>
                    </div>
                    <div style="font-size: 13px; font-weight: 500; color: #08BDBA; margin-top: 3px;">
                      Tibbiy qabul platformasi
                    </div>
                  </td>
                </tr>
              </table>
            </td>
          </tr>

          <!-- Main Content -->
          <tr>
            <td style="padding: 12px 36px 36px 36px;">
              <!-- Title -->
              <h1 style="margin: 0 0 12px 0; font-size: 20px; font-weight: 700; color: #08234A; text-align: center; line-height: 1.3;">
                ` + title + `
              </h1>
              
              <!-- Message -->
              <p style="margin: 0 0 24px 0; font-size: 15px; color: #475569; line-height: 1.6; text-align: center;">
                ` + greeting + `<br>` + message + `
              </p>

              <!-- OTP Code Box -->
              <table role="presentation" width="100%" border="0" cellspacing="0" cellpadding="0" style="margin-bottom: 24px;">
                <tr>
                  <td align="center" style="background: #F8FAFC; border: 2px dashed #0866F5; border-radius: 14px; padding: 22px 16px;">
                    <div style="font-size: 12px; font-weight: 700; text-transform: uppercase; color: #64748B; letter-spacing: 1.5px; margin-bottom: 8px;">
                      Bir martalik tasdiqlash kodi
                    </div>
                    <div style="font-size: 38px; font-weight: 800; color: #08234A; letter-spacing: 10px; font-family: Consolas, Monaco, monospace, sans-serif; padding-left: 10px;">
                      ` + code + `
                    </div>
                    <div style="font-size: 12px; color: #64748B; margin-top: 8px;">
                      Amal qilish vaqti: <strong style="color: #DC2626;">` + expireMin + ` daqiqa</strong>
                    </div>
                  </td>
                </tr>
              </table>

              <!-- Security Notice -->
              <table role="presentation" width="100%" border="0" cellspacing="0" cellpadding="0" style="background-color: #FEF3C7; border-left: 4px solid #F59E0B; border-radius: 8px; margin-bottom: 12px;">
                <tr>
                  <td style="padding: 12px 16px;">
                    <p style="margin: 0; font-size: 13px; color: #92400E; line-height: 1.5;">
                      🔒 <strong>Xavfsizlik eslatmasi:</strong> Bu kodni begona shaxslarga aslo bermang. QabulGo xodimlari sizdan parolni yoki tasdiqlash kodini hech qachon so‘ramaydi.
                    </p>
                  </td>
                </tr>
              </table>

              <p style="margin: 16px 0 0 0; font-size: 13px; color: #94A3B8; text-align: center; line-height: 1.5;">
                Agar siz ushbu so‘rovni amalga oshirmagan bo‘lsangiz, xabarni e’tiborsiz qoldiring. Akkauntingiz xavfsiz holatda qoladi.
              </p>
            </td>
          </tr>

          <!-- Footer -->
          <tr>
            <td style="padding: 24px 32px; background-color: #F8FAFC; border-top: 1px solid #E2E8F0; text-align: center;">
              <p style="margin: 0; font-size: 12px; font-weight: 500; color: #64748B;">
                © 2026 QabulGo. Barcha huquqlar himoyalangan.
              </p>
              <p style="margin: 6px 0 0 0; font-size: 11px; color: #94A3B8;">
                O‘zbekistondagi xususiy klinikalar va shifokorlarga onlayn navbat olish platformasi
              </p>
            </td>
          </tr>
        </table>
      </td>
    </tr>
  </table>
</body>
</html>`
}
