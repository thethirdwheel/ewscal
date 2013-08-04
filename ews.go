//ews provides utilities for communicating with an exchange web services server via soap
package main

import (
	"crypto/tls"
	"fmt"
	"os"

//    "github.com/ThomsonReutersEikon/go-ntlm/ntlm"
//    "github.com/ThomsonReutersEikon/go-ntlm/ntlm/messages"
)

func main() {
	conn, err := tls.Dial("tcp", "owa017.msoutlookonline.net:443", nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
	conn.Handshake()
	/*
		res, err := http.Get("https://owa017.msoutlookonline.net/EWS/Messages.xsd")
		if err != nil {
			log.Fatal(err)
		}
		robots, err := ioutil.ReadAll(res.Body)
		res.Body.Close()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%s", robots)
	*/
	session, err := ntlm.CreateClientSession(ntlm.Version1, ntlm.ConnectionlessMode)
	if nil != err {
		fmt.Printf("Err: %v, %v\n", session, err)
	}
	session.SetUserInfo("rcarmichael", "@qhe3434", "quantcast.com")

	negotiate, err := session.GenerateNegotiateMessage()
	if nil != err {
		fmt.Printf("Err: %v, %v\n", negotiate, err)
	}

	conn.Write(negotiate)

	fmt.Printf("%v\n", negotiate)

	//    challenge, err := messages.ParseChallengeMessage(challengeBytes)
	//    session.ProcessChallengeMessage(challenge)
	//
	//    authenticate := session.GenerateAuthenticateMessage()
	//    fmt.Printf("%s", authenticate)
	//
	os.Exit(0)
}
