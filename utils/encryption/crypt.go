package crypt

import (
	"crypto/aes"
	"crypto/cipher"
	crypto "crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"math/big"
	math "math/rand"
	"os"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/dgrijalva/jwt-go"
	"github.com/itchyny/base58-go"
	"golang.org/x/crypto/bcrypt"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// HashText generates hased password
func HashText(password string) string {
	bytes, _ := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes)
}

// CheckTextHash compares hash and user plain password
func CheckTextHash(text, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(text))
	if err != nil {
		return err == nil
	}
	return true
}

// Jwt generates json web tokens for stateless authentication
func Jwt(claim interface{}) string {
	secret := os.Getenv("JWT_SECRET")
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user": claim,
		"exp":  time.Now().Add(time.Hour * 24).Unix(),
	})
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		logrus.Error(err)
	}
	return tokenString
}

// JwtWithExpr Jwt generates json web tokens for stateless authentication
func JwtWithExpr(claim interface{}, expr time.Duration) string {
	secret := os.Getenv("JWT_SECRET")
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user": claim,
		"exp":  time.Now().Add(time.Minute * expr).Unix(),
	})
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		logrus.Error(err)
	}
	return tokenString
}

// GenerateRandomString generates a random string
func GenerateRandomString(length int) string {
	SeededRand := math.New(math.NewSource(time.Now().UnixNano()))
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[SeededRand.Intn(len(charset))]
	}
	return string(b)
}

func Encrypt(data []byte) (string, error) {
	key := []byte(os.Getenv("ENCRYPT_KEY"))

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	cypher := make([]byte, aes.BlockSize+len(data))
	iv := cypher[:aes.BlockSize]
	if _, err := io.ReadFull(crypto.Reader, iv); err != nil {
		return "", err
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(cypher[aes.BlockSize:], data)
	return base64.URLEncoding.EncodeToString(cypher), nil
}

func Decrypt(text string) []byte {
	key := []byte(os.Getenv("ENCRYPT_KEY"))
	cypher, _ := base64.URLEncoding.DecodeString(text)

	block, err := aes.NewCipher(key)
	if err != nil {
		fmt.Println(err)
	}
	if len(cypher) < aes.BlockSize {
		panic("cypher too short")
	}
	iv := cypher[:aes.BlockSize]
	cypher = cypher[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(cypher, cypher)

	return cypher
}

func GenerateShortLink(initialLink string, userId string) string {
	algorithm := sha256.New()
	algorithm.Write([]byte(initialLink + userId))

	generatedNumber := new(big.Int).SetBytes(algorithm.Sum(nil)).Uint64()

	encoding := base58.BitcoinEncoding
	encodedURL, err := encoding.Encode([]byte(fmt.Sprintf("%d", generatedNumber)))
	if err != nil {
		log.Println(err.Error())
		os.Exit(1)
	}

	return string(encodedURL[:8])
}

func GenerateRandomDigitsWithPrefix(prefix string, length int) int64 {
	SeededRand := math.New(math.NewSource(time.Now().UnixNano()))
	numberCharSet := "0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = numberCharSet[SeededRand.Intn(len(numberCharSet))]
	}

	randomNumbers := fmt.Sprintf("%s%s", prefix, string(b))
	randomNumbersInt, _ := strconv.Atoi(randomNumbers)

	return int64(randomNumbersInt)
}

func GenerateRandomDigits(length int) string {
	SeededRand := math.New(math.NewSource(time.Now().UnixNano()))
	numberCharSet := "0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = numberCharSet[SeededRand.Intn(len(numberCharSet))]
	}

	randomNumbers := fmt.Sprintf("%s", string(b))

	return randomNumbers
}

func ConvertToDate(date string) (time.Time, error) {
	return time.Parse("2006-01-02", date)
}
func ConvertToCardProviderDate(date time.Time) string {
	return date.Format("2006/01/02")
}

func ConvertFromToTwoDP(amount int64) float64 {
	f := amount / 100
	return float64(f)
}

func ConvertToTwoDP(amount string) int64 {
	floatAmount, err := strconv.ParseFloat(amount, 64)
	if err != nil {
		return 0
	}
	f := floatAmount * 100
	return int64(f)
}

func GenerateTransactionRef() string {
	return fmt.Sprintf("ref_%s", primitive.NewObjectID().Hex())
}
