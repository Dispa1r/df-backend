package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/g8rswimmer/go-twitter/v2"
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

func GetSignFromTwitter(tweet string) ([]byte, error) {
	if !strings.ContainsAny(tweet, "Verifying") {
		return nil, errors.New("invalid tweet")
	}

	s := strings.Split(tweet, ": ")
	if len(s) != 2 {
		return nil, errors.New("invalid tweet")
	}
	s1 := strings.Split(s[1], "#")
	return hexutil.Decode(s1[0])
}

// message,pubkey from front-end
// bytes from twitter
func VerifySig(message, pubKey string, signatureBytes []byte) bool {
	// format message
	fullMessage := fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(message), message)
	hash := crypto.Keccak256Hash([]byte(fullMessage))

	// modify the recoverID
	signatureBytes[64] -= 27
	sigPublicKeyECDSA, err := crypto.SigToPub(hash.Bytes(), signatureBytes)
	if err != nil {
		log.Fatal("")
	}

	sigPublicKeyBytes := crypto.FromECDSAPub(sigPublicKeyECDSA)
	publicKeyBytes, err := hexutil.Decode(pubKey)
	if err != nil {
		fmt.Println("fail to decode the pubkey")
		return false
	}
	matches := bytes.Equal(sigPublicKeyBytes, publicKeyBytes)
	if !matches {
		return false
	} // true

	signatureNoRecoverID := signatureBytes[:len(signatureBytes)-1] // remove recovery id
	verified := crypto.VerifySignature(publicKeyBytes, hash.Bytes(), signatureNoRecoverID)
	if verified {
		return true
	} else {
		return false
	}
}

func CheckVerifyTweet(pubKey, message string) (string, string) {
	token := "AAAAAAAAAAAAAAAAAAAAAIUZiQEAAAAA%2B2PBi75RgZ84G98WT%2FTzu9cAPKI%3DZPd1WgoqM9SNnqrLV1SrhDEd4FVvlO2T7g1j750qQ4hfjhSyHi"
	//userID := "1287606089849151494"
	query := "@darkforest_eth"

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

	fmt.Println("Callout to tweet recent search callout")

	tweetResponse, err := client.TweetRecentSearch(context.Background(), query, opts)
	if err != nil {
		log.Panicf("tweet lookup error: %v", err)
	}

	dictionaries := tweetResponse.Raw.TweetDictionaries()
	// TODO : Cache the tweet
	//signMap := make(map[string][]byte)

	for _, v := range dictionaries {
		fmt.Println("author ID: ", v.Author.ID, "author Name: ", v.Author.UserName, v.Author.Name)
		fmt.Println("twitter text: ", v.Tweet.Text)
		tmpSign, err := GetSignFromTwitter(v.Tweet.Text)
		if err != nil {
			fmt.Println("fail to parse the sign bytes")
		}
		verified := VerifySig(message, pubKey, tmpSign)
		if verified {
			return pubKey, v.Author.ID
		}
	}

	return "", ""
}
