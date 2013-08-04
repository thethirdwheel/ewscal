package main

import (
	"net"
    "fmt"
    "bufio"
    "io/ioutil"
	"github.com/ThomsonReutersEikon/go-ntlm/ntlm"
	"github.com/ThomsonReutersEikon/go-ntlm/ntlm/messages"
)

func requestChallenge(msg string, session ntlm.ClientSession) (messages.ChallengeMessage, error) {
    c, err := net.Dial("tcp", "localhost:2000")
    defer c.Close()
    if err != nil {
        // handle error
    }
    fmt.Fprintf(c, msg)
    challengeData, _ := ioutil.ReadAll(bufio.NewReader(c))
    return messages.ParseChallengeMessage(challengeData)
}

func authenticate(challenge messages.ChallengeMessage, session ntlm.ClientSession) {
    c, err := net.Dial("tcp", "localhost:2000")
    defer c.Close()
    if err != nil {
        // handle error
    }
    session.ProcessChallengeMessage(cMsg)
    aMsg, _ := session.GenerateAuthenticateMessage()
    fmt.Fprintf(c, aMsg.Byte())
}

func doHandshake(msg string, ch chan string) {
	session, err := ntlm.CreateClientSession(ntlm.Version1, ntlm.ConnectionlessMode)
	session.SetUserInfo("someuser", "somepassword", "somedomain")
	c, err := net.Dial("tcp", "localhost:2000")
	if err != nil {
		// handle error
	}
    fmt.Fprintf(c, msg)
    challengeData, _ := ioutil.ReadAll(bufio.NewReader(c))
    cMsg, _ := messages.ParseChallengeMessage(challengeData)
	fmt.Println("----- Challenge Message ----- ")
	fmt.Println(cMsg.String())
	fmt.Println("----- END Challenge Message ----- ")
    session.ProcessChallengeMessage(cMsg)
    aMsg, _ := session.GenerateAuthenticateMessage()
	fmt.Println("----- Authenticate Message ----- ")
	fmt.Println(aMsg.String())
	fmt.Println("----- END Authenticate Message ----- ")
    fmt.Fprintf(c, aMsg.String())
    ch <- aMsg.String()
}

func main() {
    ch := make(chan string)
	session, err := ntlm.CreateClientSession(ntlm.Version1, ntlm.ConnectionlessMode)
	session.SetUserInfo("someuser", "somepassword", "somedomain")
    cMsg, _ := requestChallenge("GET /HTTP/1.0\r\n", session)
	fmt.Println("----- Challenge Message ----- ")
	fmt.Println(cMsg.String())
	fmt.Println("----- END Challenge Message ----- ")
//    go doHandshake("GET /HTTP/1.0\r\n\r\n", ch)
//    go doHandshake("BOGUS\r\n\r\n", ch)
}
