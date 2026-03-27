package main

import (
	"fmt"
	"log"
	"net/smtp"
	"os"
	"strconv"

	"github.com/gagliardetto/solana-go"
	"gorm.io/gorm"
)

// GenerateSolanaWallet generates a new Solana keypair
func GenerateSolanaWallet() (address string, privateKey string, err error) {
	account := solana.NewWallet()
	return account.PublicKey().String(), account.PrivateKey.String(), nil
}

// SendWalletEmail sends an email to the user with their private key
func SendWalletEmail(to, username, address, privateKey string) error {
	host := os.Getenv("SMTP_HOST")
	port := os.Getenv("SMTP_PORT")
	user := os.Getenv("SMTP_USER")
	pass := os.Getenv("SMTP_PASS")
	from := os.Getenv("EMAIL_FROM")

	if host == "" || user == "" {
		log.Println("SMTP not configured, skipping email delivery")
		return nil
	}

	header := make(map[string]string)
	header["From"] = from
	header["To"] = to
	header["Subject"] = "Welcome to AIGen-Blog - Your Web3 Wallet is Ready"
	header["Content-Type"] = "text/html; charset=UTF-8"

	body := fmt.Sprintf(`
		<div style="font-family: sans-serif; max-width: 600px; margin: 0 auto; border: 1px solid #e2e8f0; border-radius: 12px; padding: 24px;">
			<h2 style="color: #4f46e5;">Hi %s,</h2>
			<p>Welcome to <b>AIGen-Blog</b>! We have generated a Solana wallet for you so you can start earning <b>BLOG</b> tokens right away.</p>
			
			<div style="background-color: #f8fafc; padding: 16px; border-radius: 8px; margin: 20px 0;">
				<p style="margin: 0; font-size: 14px; color: #64748b;">Wallet Address:</p>
				<p style="margin: 4px 0 0 0; font-family: monospace; word-break: break-all;">%s</p>
			</div>
			
			<div style="background-color: #fff1f2; padding: 16px; border-radius: 8px; border: 1px solid #fecaca;">
				<p style="margin: 0; font-size: 14px; color: #e11d48; font-weight: bold;">Private Key (SECRET):</p>
				<p style="margin: 4px 0 0 0; font-family: monospace; word-break: break-all; color: #be123c;">%s</p>
			</div>
			
			<p style="color: #ef4444; font-size: 12px; margin-top: 10px;">⚠️ <b>CRITICAL:</b> Never share your private key. Anyone with this key can access your tokens. Store it in a safe place.</p>
			
			<hr style="border: 0; border-top: 1px solid #e2e8f0; margin: 24px 0;">
			<p style="color: #94a3b8; font-size: 12px;">Enjoy creating content and earning rewards!</p>
			<p style="color: #94a3b8; font-size: 12px;">&copy; 2026 AIGen-Blog Team</p>
		</div>
	`, username, address, privateKey)

	message := ""
	for k, v := range header {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + body

	auth := smtp.PlainAuth("", user, pass, host)
	err := smtp.SendMail(host+":"+port, auth, user, []string{to}, []byte(message))
	if err != nil {
		log.Printf("Failed to send email: %v", err)
	}
	return err
}

// DistributeRewards distributes BLOG tokens based on user actions
func DistributeRewards(userID uint, actionType string) {
	amountStr := os.Getenv("REWARD_" + actionType)
	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil || amount <= 0 {
		return
	}

	if err := DB.Model(&User{}).Where("id = ?", userID).Update("blog_balance", gorm.Expr("blog_balance + ?", amount)).Error; err == nil {
		DB.Create(&RewardLog{
			UserID: userID,
			Type:   actionType,
			Amount: amount,
		})
	}
}

// ProcessReferral handles 3-level referral rewards upon registration
func ProcessReferral(userID uint, referrerID *uint) {
	if referrerID == nil {
		return
	}

	levels := []string{"REF_L1", "REF_L2", "REF_L3"}
	currReferrerID := *referrerID

	for _, level := range levels {
		amountStr := os.Getenv(level)
		amount, err := strconv.ParseFloat(amountStr, 64)
		if err != nil || amount <= 0 {
			break
		}

		var referrer User
		if err := DB.First(&referrer, currReferrerID).Error; err != nil {
			break
		}

		// Reward the referrer
		if err := DB.Model(&User{}).Where("id = ?", currReferrerID).Update("blog_balance", gorm.Expr("blog_balance + ?", amount)).Error; err == nil {
			DB.Create(&RewardLog{
				UserID: currReferrerID,
				Type:   "Referral (" + level + ")",
				Amount: amount,
			})
		}

		// Move to the next level (upwards)
		if referrer.ReferrerID == nil {
			break
		}
		currReferrerID = *referrer.ReferrerID
	}
}
