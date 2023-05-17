package main

import (
	"bytes"
	"crypto/ecdsa"
	"fmt"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"log"
)

/*
*

	In order to run, the user will need to provide the bearer token and the list of tweet ids.

*
*/
func main() {

	//fmt.Println(CheckVerifyTweet())
	privateKeyString := "67195c963ff445314e667112ab22f4a7404bad7f9746564eb409b9bb8c6aed32"
	// Parse private key
	privateKey, err := crypto.HexToECDSA(privateKeyString)
	if err != nil {
		log.Fatal(err)
	}
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("error casting public key to ECDSA")
	}

	publicKeyBytes := crypto.FromECDSAPub(publicKeyECDSA)

	// Sign message
	message := "111"
	fullMessage := fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(message), message)
	hash := crypto.Keccak256Hash([]byte(fullMessage))
	signatureBytes, err := crypto.Sign(hash.Bytes(), privateKey)
	if err != nil {
		log.Fatal("fail to hash the message")
	}
	signatureBytes[64] += 27
	signature := hexutil.Encode(signatureBytes)

	fmt.Println("sign: ", signature)

	signatureBytes[64] -= 27
	sigPublicKeyECDSA, err := crypto.SigToPub(hash.Bytes(), signatureBytes)
	if err != nil {
		log.Fatal(err)
	}

	sigPublicKeyBytes := crypto.FromECDSAPub(sigPublicKeyECDSA)

	matches := bytes.Equal(sigPublicKeyBytes, publicKeyBytes)
	fmt.Println(matches) // true

	signatureNoRecoverID := signatureBytes[:len(signatureBytes)-1] // remove recovery id
	verified := crypto.VerifySignature(publicKeyBytes, hash.Bytes(), signatureNoRecoverID)
	fmt.Println(verified) // true

	//metaBytes, err := json.MarshalIndent(tweetResponse.Meta, "", "    ")
	//if err != nil {
	//	log.Panic(err)
	//}
	//fmt.Println(string(metaBytes))

	//client := &twitter.Client{
	//	Authorizer: authorize{
	//		Token: token,
	//	},
	//	Client: http.DefaultClient,
	//	Host:   "https://api.twitter.com",
	//}
	//opts := twitter.UserMentionTimelineOpts{
	//	TweetFields: []twitter.TweetField{twitter.TweetFieldCreatedAt, twitter.TweetFieldAuthorID, twitter.TweetFieldConversationID, twitter.TweetFieldPublicMetrics, twitter.TweetFieldContextAnnotations},
	//	UserFields:  []twitter.UserField{twitter.UserFieldUserName},
	//	Expansions:  []twitter.Expansion{twitter.ExpansionAuthorID},
	//	MaxResults:  5,
	//}
	//
	//fmt.Println("Callout to tweet user mention timeline callout")
	//
	//timeline, err := client.UserMentionTimeline(context.Background(), userID, opts)
	//if err != nil {
	//	log.Panicf("user mention timeline error: %v", err)
	//}
	//
	//dictionaries := timeline.Raw.TweetDictionaries()
	//signMap := make(map[string][]byte)
	//for _, v := range dictionaries {
	//	fmt.Println("author ID: ", v.Author.ID, "author Name: ", v.Author.UserName, v.Author.Name)
	//	fmt.Println("twitter text: ", v.Tweet.Text)
	//	tmpSign, err := GetSignFromTwitter(v.Tweet.Text)
	//	if err != nil {
	//		fmt.Println("fail to parse the sign bytes")
	//	}
	//	fmt.Println(tmpSign)
	//	signMap[v.Author.ID] = tmpSign
	//}
	//
	////enc, err := json.MarshalIndent(dictionaries, "", "    ")
	////if err != nil {
	////	log.Panic(err)
	////}
	////fmt.Println(string(enc))
	////
	//metaBytes, err := json.MarshalIndent(timeline.Meta, "", "    ")
	//if err != nil {
	//	log.Panic(err)
	//}
	////key : Twitter ID
	////sign : 内容从value.tweet.text里
	////author id: value.tweet.value里
	//fmt.Println(string(metaBytes))
}
