package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/g8rswimmer/go-twitter/v2"
	"github.com/spf13/viper"
	"log"
	"net/http"
	"strings"
)

type authorize struct {
	Token string
}

func (a authorize) Add(req *http.Request) {
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", a.Token))
}

// GetSignFromTwitter
// tweet: the tweet from twitter api
func GetSignFromTwitter(tweet string) ([]byte, error) {
	if !strings.ContainsAny(tweet, "Verifying") {
		return nil, errors.New("invalid tweet")
	}

	s := strings.Split(tweet, ": ")
	if len(s) != 2 {
		return nil, errors.New("invalid tweet")
	}

	s1 := strings.Split(s[1], " #")
	return hexutil.Decode(s1[0])
}

// VerifySig
// message: public key from front-end
// pubKey: generate from message and sigBytes(the last 32 bytes)
func VerifySig(message string, pubKey []byte, signatureBytes []byte) bool {
	// format message
	fullMessage := fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(message), message)
	hash := crypto.Keccak256Hash([]byte(fullMessage))

	// modify the recoverID
	signatureBytes[64] -= 27
	sigPublicKeyECDSA, err := crypto.SigToPub(hash.Bytes(), signatureBytes)
	if err != nil {
		log.Println("fail to generate pubKey from message and sign")
		return false
	}

	sigPublicKeyBytes := crypto.FromECDSAPub(sigPublicKeyECDSA)
	matches := bytes.Equal(sigPublicKeyBytes, pubKey)
	if !matches {
		return false
	}

	signatureNoRecoverID := signatureBytes[:len(signatureBytes)-1] // remove recovery id
	verified := crypto.VerifySignature(pubKey, hash.Bytes(), signatureNoRecoverID)
	if verified {
		return true
	} else {
		return false
	}
}

// CheckVerifyTweet
// message: public key from front-end
// pubKey: generate from message and sigBytes(the last 32 bytes)
func CheckVerifyTweet(pubKey []byte, message string) string {
	token := viper.GetString("twitter.bearKey")
	query := viper.GetString("twitter.searchTag")

	client := &twitter.Client{
		Authorizer: authorize{
			Token: token,
		},
		Client: http.DefaultClient,
		Host:   "https://api.twitter.com",
	}
	opts := twitter.TweetRecentSearchOpts{
		Expansions:  []twitter.Expansion{twitter.ExpansionEntitiesMentionsUserName, twitter.ExpansionAuthorID},
		TweetFields: []twitter.TweetField{twitter.TweetFieldCreatedAt, twitter.TweetFieldConversationID, twitter.TweetFieldAttachments},
	}

	log.Println("Callout to tweet recent search callout")

	tweetResponse, err := client.TweetRecentSearch(context.Background(), query, opts)
	if err != nil {
		log.Panicf("tweet lookup error: %v", err)
		return ""
	}

	dictionaries := tweetResponse.Raw.TweetDictionaries()

	// first find the cache
	for k, v := range userTwitterCache {
		tmpSign, err := hexutil.Decode(k)
		if err != nil {
			return ""
		}
		verified := VerifySig(message, pubKey, tmpSign)
		if verified {
			return v
		}
	}

	// get the twitter and cache it
	for _, v := range dictionaries {
		log.Println("author ID: ", v.Author.ID, "author Name: ", v.Author.UserName)
		log.Println("twitter text: ", v.Tweet.Text)
		tmpSign, err := GetSignFromTwitter(v.Tweet.Text)
		if err != nil {
			fmt.Println("fail to parse the sign bytes")
			continue
		}
		tmpSignHex := hexutil.Encode(tmpSign)
		userTwitterCache[tmpSignHex] = v.Author.UserName
		verified := VerifySig(message, pubKey, tmpSign)
		if verified {
			return v.Author.UserName
		}
	}

	return ""
}

func generatePubKeyFromSign(sign, fmtMsg string) ([]byte, []byte, error) {
	fullMessage := fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(fmtMsg), fmtMsg)

	hash := crypto.Keccak256Hash([]byte(fullMessage))
	signBytes, err := hexutil.Decode(sign)
	if err != nil {
		return nil, nil, err
	}

	signBytes[64] -= 27
	sigPublicKeyECDSA, err := crypto.SigToPub(hash.Bytes(), signBytes)
	sigPublicKeyBytes := crypto.FromECDSAPub(sigPublicKeyECDSA)

	fullAddress := crypto.Keccak256Hash(sigPublicKeyBytes[1:]).Bytes()
	trueAddress := fullAddress[len(fullAddress)-20:]

	return trueAddress, sigPublicKeyBytes, nil
}

func parseTwitterMsg(buf []byte, msgType string) (twitterMsg, error) {
	trimmedData := bytes.TrimRight(buf, "\x00")
	var m map[string]interface{}
	err := json.Unmarshal(trimmedData, &m)
	if err != nil {
		return twitterMsg{}, err
	}
	var msg map[string]interface{}
	if msgType == "bind" {
		msg = m["verifyMessage"].(map[string]interface{})
	} else if msgType == "disconnect" {
		msg = m["disconnectMessage"].(map[string]interface{})
	}

	sender, senderOk := msg["sender"].(string)
	signature, sigOk := msg["signature"].(string)
	twitter, twitterOk := msg["message"].(map[string]interface{})["twitter"].(string)
	if !sigOk || !senderOk || !twitterOk {
		return twitterMsg{}, errors.New("invalid parameter")
	}
	return twitterMsg{sender: sender, sign: signature, twitter: twitter}, nil
}
