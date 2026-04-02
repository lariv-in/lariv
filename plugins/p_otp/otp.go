package p_otp

import (
	"crypto/rand"
	"fmt"
	"log"
	"maps"
	"net/mail"
	"net/smtp"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/lariv-in/lago/plugins/p_otp/adapters"
	"gorm.io/gorm"
)

const (
	OtpCachePrefixPhone = "otp:phone:"
	OtpCachePrefixEmail = "otp:email:"
	OtpExpirySeconds    = 300 // 5 minutes
)

var emailCacheKeyPattern = regexp.MustCompile(`[^a-z0-9@._+-]`)

// CacheEntry represents an OTP and its expiry.
type CacheEntry struct {
	OTP       string
	ExpiresAt time.Time
}

// MemoryCache is a concurrency-safe in-memory store for OTPs.
type MemoryCache struct {
	mu    sync.RWMutex
	store map[string]CacheEntry
}

func NewMemoryCache() *MemoryCache {
	return &MemoryCache{
		store: make(map[string]CacheEntry),
	}
}

func (c *MemoryCache) Set(key, otp string, duration time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.store[key] = CacheEntry{
		OTP:       otp,
		ExpiresAt: time.Now().Add(duration),
	}
}

func (c *MemoryCache) Get(key string) (string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	entry, exists := c.store[key]
	if !exists {
		return "", false
	}
	if time.Now().After(entry.ExpiresAt) {
		delete(c.store, key) // Cleanup expired
		return "", false
	}
	return entry.OTP, true
}

func (c *MemoryCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.store, key)
}

// Global OTP Cache instance
var otpCache = NewMemoryCache()

// GenerateOTP creates a 6-digit numeric OTP.
func GenerateOTP() string {
	b := make([]byte, 3)
	_, _ = rand.Read(b)
	val := (int(b[0]) | int(b[1])<<8 | int(b[2])<<16) % 1000000
	return fmt.Sprintf("%06d", val)
}

func phoneToCacheSuffix(phone string) string {
	return strings.TrimPrefix(strings.TrimSpace(phone), "+")
}

func getCacheKeyPhone(phone string) string {
	return OtpCachePrefixPhone + phoneToCacheSuffix(phone)
}

func StoreOtpPhone(phone, otp string) {
	key := getCacheKeyPhone(phone)
	otpCache.Set(key, otp, time.Duration(OtpExpirySeconds)*time.Second)
	log.Printf("OTP stored for phone %s", phone)
}

func getCacheKeyEmail(email string) string {
	s := strings.ToLower(strings.TrimSpace(email))
	s = emailCacheKeyPattern.ReplaceAllString(s, "")
	return OtpCachePrefixEmail + s
}

func StoreOtpEmail(email, otp string) {
	key := getCacheKeyEmail(email)
	otpCache.Set(key, otp, time.Duration(OtpExpirySeconds)*time.Second)
	log.Printf("OTP stored for email %s", email)
}

// VerifyOTP checks the generic identifier (email or phone) against the cache.
func VerifyOTP(identifier, otp string) bool {
	identifier = strings.TrimSpace(identifier)

	// Determine if it's an email or phone
	var key string
	if strings.Contains(identifier, "@") {
		key = getCacheKeyEmail(identifier)
	} else {
		key = getCacheKeyPhone(identifier)
	}

	storedOtp, exists := otpCache.Get(key)
	if !exists {
		log.Printf("No OTP found for %s", identifier)
		return false
	}
	if storedOtp != otp {
		log.Printf("OTP mismatch for %s", identifier)
		return false
	}

	otpCache.Delete(key)
	log.Printf("OTP verified for %s", identifier)
	return true
}

// SendSmsOtp fetches preferences, generates the OTP, stores it, and dispatches to MSG91.
func SendSmsOtp(db *gorm.DB, phone string) bool {
	prefs := LoadPreferences(db)

	templateID := prefs.SmsOtpTemplateId
	if templateID == "" {
		templateID = prefs.OtpTemplateId // Fallback
	}

	if templateID == "" {
		log.Println("SMS_OTP_TEMPLATE_ID or OTP_TEMPLATE_ID not configured")
		return false
	}

	authKey := prefs.Msg91AuthKey
	if authKey == "" {
		log.Println("MSG91_AUTH_KEY not configured")
		return false
	}

	otp := GenerateOTP()
	StoreOtpPhone(phone, otp)

	normalizedPhone := phoneToCacheSuffix(phone)
	otpFieldName := prefs.SmsOtpFieldName
	if otpFieldName == "" {
		otpFieldName = "otp"
	}

	recipient := adapters.FlowRecipient{
		"mobiles":    normalizedPhone,
		otpFieldName: otp,
	}

	maps.Copy(recipient, prefs.GetExtraFields())

	client := adapters.NewMsg91Client(authKey)
	res, err := client.SendSMSFlow(templateID, []adapters.FlowRecipient{recipient}, true)
	if err != nil {
		log.Printf("Failed to send SMS OTP: %v, Response: %v", err, res)
		return false
	}

	log.Printf("OTP SMS sent to %s: %v", phone, res)
	return true
}

// SendEmailOtp generates an OTP, stores it, and sends it via SMTP.
func SendEmailOtp(db *gorm.DB, email string) bool {
	prefs := LoadPreferences(db)

	templateString := prefs.EmailOtpTemplateString
	if templateString == "" {
		log.Println("EMAIL_OTP_TEMPLATE_STRING not configured")
		return false
	}

	if prefs.SmtpHost == "" || prefs.SmtpFrom == "" {
		log.Println("SMTP not configured (host and from are required)")
		return false
	}

	otp := GenerateOTP()
	StoreOtpEmail(email, otp)

	body := strings.ReplaceAll(templateString, "$otp", otp)

	from := mail.Address{Address: prefs.SmtpFrom}
	to := mail.Address{Address: email}
	msg := "From: " + from.String() + "\r\n" +
		"To: " + to.String() + "\r\n" +
		"Subject: Your OTP Code\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/plain; charset=\"utf-8\"\r\n" +
		"\r\n" +
		body

	addr := prefs.SmtpHost + ":" + prefs.SmtpPort
	var auth smtp.Auth
	if prefs.SmtpUsername != "" {
		auth = smtp.PlainAuth("", prefs.SmtpUsername, prefs.SmtpPassword, prefs.SmtpHost)
	}

	err := smtp.SendMail(addr, auth, prefs.SmtpFrom, []string{email}, []byte(msg))
	if err != nil {
		log.Printf("Failed to send email OTP to %s: %v", email, err)
		return false
	}

	log.Printf("OTP email sent to %s", email)
	return true
}
