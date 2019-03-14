package lark

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// EventHandler struct
type EventHandler struct {
	client                *http.Client
	needTokenVerification bool
	verificationToken     string
	isEncrypted           bool
	encryptKey            []byte
}

// NewEventHandler creates an event handler
func NewEventHandler() *EventHandler {
	return &EventHandler{}
}

// SetClient assigns a new client to bot.client
func (ev *EventHandler) SetClient(c *http.Client) {
	ev.client = c
}

// EnableEncryption enable message encryption
func (ev *EventHandler) EnableEncryption(key string) error {
	ev.isEncrypted = true
	sha256key := sha256.Sum256([]byte(key))
	ev.encryptKey = sha256key[:sha256.Size]
	return nil
}

// EnableTokenVerification enable token verification
func (ev *EventHandler) EnableTokenVerification(token string) {
	ev.needTokenVerification = true
	ev.verificationToken = token
}

// Decrypt with AES Cipher
func (ev *EventHandler) Decrypt(data string) ([]byte, error) {
	if !ev.isEncrypted {
		return nil, errors.New("Encryption is not enabled")
	}
	block, err := aes.NewCipher(ev.encryptKey)
	if err != nil {
		return nil, err
	}
	ciphertext, err := base64.StdEncoding.DecodeString(data)
	iv := ev.encryptKey[:aes.BlockSize]
	blockMode := cipher.NewCBCDecrypter(block, iv)
	decryptedData := make([]byte, len(data))
	blockMode.CryptBlocks(decryptedData, ciphertext)
	msg := ev.unpad(decryptedData)
	return msg[block.BlockSize():], err
}

func (ev *EventHandler) unpad(data []byte) []byte {
	length := len(data)
	var unpadding, unpaddingIdx int
	for i := length - 1; i > 0; i-- {
		if data[i] != 0 {
			unpadding = int(data[i])
			unpaddingIdx = length - 1 - i
			break
		}
	}
	return data[:(length - unpaddingIdx - unpadding)]
}

// ServeEventChallenge answer the challenge
func (ev *EventHandler) ServeEventChallenge(hookPrefix, addr string, done chan int) {
	log.SetPrefix(LogPrefix)
	r := gin.Default()
	r.POST(hookPrefix, func(c *gin.Context) {
		defer close(done)
		var challenge string
		var eventBody EventChallengeReq
		var err error

		if ev.isEncrypted {
			var encryptedBody EncryptedReq
			err = c.BindJSON(&encryptedBody)
			if err != nil {
				log.Fatalln("Get encrypt failed: ", err)
				return
			}
			log.Printf("Handled encrypt: %s\n", encryptedBody.Encrypt)
			decryptedData, err := ev.Decrypt(encryptedBody.Encrypt)
			if err != nil {
				log.Fatalln("Decode encrypt failed: ", err)
				return
			}
			err = json.Unmarshal(decryptedData, &eventBody)
			if err != nil {
				log.Fatalln("Decode challenge failed: ", err)
				return
			}
			log.Printf("Handled challenge: %s\n", eventBody.Challenge)
			challenge = eventBody.Challenge
		} else {
			err = c.BindJSON(&eventBody)
			if err != nil {
				log.Fatalln("Get challenge failed")
				return
			}
			log.Printf("Handled challenge: %s\n", eventBody.Challenge)
			challenge = eventBody.Challenge
		}
		c.JSON(http.StatusOK, gin.H{
			"challenge": challenge,
		})
	})
	r.Run(addr)
}

// ServeEvent handles event messages
func (ev *EventHandler) ServeEvent(hookPrefix, addr string, callback MsgCallbackFunc) {
	log.SetPrefix(LogPrefix)
	r := gin.Default()
	r.POST(hookPrefix, func(c *gin.Context) {
		var message EventMessage
		var err error
		if ev.isEncrypted {
			var encryptedBody EncryptedReq
			err = c.BindJSON(&encryptedBody)
			if err != nil {
				log.Println("Decode encrypt failed: ", err)
				return
			}
			decryptedData, err := ev.Decrypt(encryptedBody.Encrypt)
			if err != nil {
				log.Println("Decode encrypt failed:", err)
				return
			}
			err = json.Unmarshal(decryptedData, &message)
			if err != nil {
				log.Println("Decode challenge failed:", err)
				return
			}
		} else {
			err = c.BindJSON(&message)
			if err != nil {
				log.Println("Parse message failed: ", err)
				return
			}
		}
		if message.Token != ev.verificationToken && ev.needTokenVerification {
			log.Println("Token verification failed")
			return
		}
		err = callback(message)
		if err != nil {
			log.Println("Callback message failed: ", err)
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})
	r.Run(addr)
}

// PostEvent posts event
// 1. help to develop and test ServeEvent callback func much easier
// 2. otherwise, you may use it to forward event
func (ev *EventHandler) PostEvent(hookURL string, message EventMessage) (*http.Response, error) {
	buf := new(bytes.Buffer)
	err := json.NewEncoder(buf).Encode(message)
	if err != nil {
		log.Fatalf("Encode json failed: %+v\n", err)
	}
	// Create a client if user does not specify
	if ev.client == nil {
		ev.client = &http.Client{
			Timeout: 5 * time.Second,
		}
	}
	resp, err := ev.client.Post(hookURL, "application/json; charset=utf-8", buf)
	return resp, err
}

// ServeEventChallenge for earlier version
func ServeEventChallenge(hookPrefix, addr string, done chan int) {
	ev := NewEventHandler()
	ev.ServeEventChallenge(hookPrefix, addr, done)
}

// ServeEvent for earlier version
func ServeEvent(hookPrefix, addr string, callback MsgCallbackFunc) {
	ev := NewEventHandler()
	ev.ServeEvent(hookPrefix, addr, callback)
}

// PostEvent for earlier version
func PostEvent(hookURL string, message EventMessage) (*http.Response, error) {
	ev := NewEventHandler()
	return ev.PostEvent(hookURL, message)
}
