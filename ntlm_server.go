package main

import (
	"net"
    "log"
    "bufio"
    "fmt"
    "io/ioutil"
    "encoding/base64"
	"github.com/ThomsonReutersEikon/go-ntlm/ntlm"
	"github.com/ThomsonReutersEikon/go-ntlm/ntlm/messages"
)

func ChallengeResponse(c net.Conn, session ntlm.ServerSession) {
    defer c.Close()
    status, _ := ioutil.ReadAll(bufio.NewReader(c))
    log.Printf("Saw %s", status)
    challengeMessage := "TlRMTVNTUAACAAAAAAAAADgAAABVgphiMx43owKH33MAAAAAAAAAAKIAogA4AAAABQEoCgAAAA8CAA4AUgBFAFUAVABFAFIAUwABABwAVQBLAEIAUAAtAEMAQgBUAFIATQBGAEUAMAA2AAQAFgBSAGUAdQB0AGUAcgBzAC4AbgBlAHQAAwA0AHUAawBiAHAALQBjAGIAdAByAG0AZgBlADAANgAuAFIAZQB1AHQAZQByAHMALgBuAGUAdAAFABYAUgBlAHUAdABlAHIAcwAuAG4AZQB0AAAAAAA="
    challengeData, _ := base64.StdEncoding.DecodeString(challengeMessage)
    fmt.Fprintf(c, challengeMessage)
}


func logAndHandshake(c net.Conn, session ntlm.ServerSession) {
    defer c.Close()
    status, _ := ioutil.ReadAll(bufio.NewReader(c))
    log.Printf("Saw %s", status)
//	challenge, err := session.GenerateChallengeMessage()
//    if nil != err || nil == challenge {
//        log.Fatal(err)
//    }
    status, _ = ioutil.ReadAll(bufio.NewReader(c))
    log.Printf("Next, saw %s", status)
    authenticateMessage, _ := messages.ParseAuthenticateMessage(status, 1)
    cMsg, _ := messages.ParseChallengeMessage(challengeData)
    session.SetServerChallenge(cMsg.ServerChallenge)
    session.ProcessAuthenticateMessage(authenticateMessage)
    fmt.Println("----- Authenticate Message ----- ")
    fmt.Println(authenticateMessage.String())
    fmt.Println("----- END Authenticate Message ----- ")
}

/*func logAndAuth() {
	auth, err := messages.ParseAuthentiateMessage(authenticateBytes)
	session.ProcessAuthenticateMessage(auth)
}*/

func main() {
	session, err := ntlm.CreateServerSession(ntlm.Version1, ntlm.ConnectionlessMode)
	session.SetUserInfo("someuser", "somepassword", "somedomain")

	// Listen on TCP port 2000 on all interfaces.
	l, err := net.Listen("tcp", ":2000")
	if err != nil {
		log.Fatal(err)
	}
	defer l.Close()
	for {
		// Wait for a connection.
		conn, err := l.Accept()
		if err != nil {
			log.Fatal(err)
		}
		// Handle the connection in a new goroutine.
		// The loop then returns to accepting, so that
		// multiple connections may be served concurrently.
		go logAndHandshake(conn, session)
	}
}
