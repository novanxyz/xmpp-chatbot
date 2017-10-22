package xmpp

import b64 "encoding/base64"
import (
	"net"
	"fmt"
)

type XmppHandler struct {
	host          string
	username      string
	jid           string
	password      string
	parser        *XmppContentHandler
	connection    net.Conn
	State         string
	request_id    int
	id            string
	requests      map[string]func(*XmppElement)
	HandleMessage func(*XmppElement)
}

func NewXmppHandler() *XmppHandler {
	return &XmppHandler{
		State:    "initial",
		parser:   NewXmppContentHandler(),
		requests: make(map[string]func(*XmppElement)),
	}
}

func (self *XmppHandler) Connect(server string, host string, username string, password string) {
	full_address := server + ":5222"
	fmt.Printf("Connecting to %s\n", full_address)
	conn, err := net.Dial("tcp", full_address)
	if err != nil {
		fmt.Errorf("Couldn't connect")
		return
	}

	fmt.Printf("Connection made\n")
	self.connection = conn
	self.host = host
	self.username = username
	self.password = password

	self.State = "waiting_for_stream"
	self.request_id = 0
	self.id = "bot"
	go self.receive()
	self.startStream()
}

func (self * XmppHandler) receive() {
	for {
		fmt.Print("Waiting to receive\n")
		recvBuf := make([]byte, 4096)
		n, err := self.connection.Read(recvBuf)
		if err != nil {
	    	fmt.Print(err)
		}
		msgReceived := string(recvBuf[:n])
	    fmt.Printf("Message Received: %s\n", msgReceived)
	    parseErr := self.parser.Parse(msgReceived)
	    if parseErr != nil {
	    	fmt.Print(parseErr)
	    }
	    fmt.Printf("%d elements in queue\n", len(self.parser.queue))

	    for {
	    	if len(self.parser.queue) == 0 {
	    		break
		    }
			elem := self.parser.queue[0]
			self.parser.queue = self.parser.queue[1:]
			self.handleResponse(elem)
	    }
	}
}
func (self *XmppHandler) handleResponse(response *XmppElement) (bool) {
	fmt.Printf("handleResponse - State %s\n", self.State)
	fmt.Println(response.ToXml())
	if response.Tag == "iq" {
		request_id := response.Attributes["id"]
		callback := self.requests[request_id]
		callback(response)
		return true
	}

	if self.State == "ready" {
		if response.Tag == "message" {
			if self.HandleMessage != nil {
				self.HandleMessage(response)
			}
			return true
		}
	}

	if self.State == "waiting_for_stream" {
		if response.Tag == "stream" {
			self.State = "waiting_for_features"
			return true
		}
	}
	if self.State == "waiting_for_features" {
		if response.Tag == "features" {
			self.authenticate()
			return true
		}
	}

	if self.State == "authenticating" {
		if response.Tag == "success" {
			self.State = "authenticated_waiting_for_stream"
			self.startStream()
			return true
		} else if response.Tag == "failure" {
			fmt.Print("Failed to log in\n")
			return true
		}
	}

	if self.State == "authenticated_waiting_for_stream" {
		if response.Tag == "stream" {
			self.State = "authenticated_waiting_for_features"
			return true
		}
	}

	if self.State == "authenticated_waiting_for_features" {
		if response.Tag == "features" {
			self.bind()
			return true
		}
	}

	return false
}

func (self *XmppHandler) startStream() {
	self.parser = NewXmppContentHandler()
    template := "<?xml version='1.0'?>" +
    	"<stream:stream to='%s' " +
        "version='1.0' " +
        "xmlns='jabber:client' " +
        "xmlns:stream='http://etherx.jabber.org/streams'>"

	packet := fmt.Sprintf(template, self.host)
	self.connection.Write([]byte(packet))
}

func (self *XmppHandler) authenticate() {
	self.State = "authenticating"

	key := fmt.Sprintf("\000%s\000%s", self.username, self.password)
	encodedKey := b64.StdEncoding.EncodeToString([]byte(key))

    template := "<auth xmlns='urn:ietf:params:xml:ns:xmpp-sasl' mechanism='PLAIN'>%s</auth>"
    packet := fmt.Sprintf(template, encodedKey)
	self.connection.Write([]byte(packet))
}

func (self *XmppHandler) IssueRequest(request string, request_type string, to string, callback func(*XmppElement)) {
	request_id := self.getRequestId()
	to_clause := ""
	if len(to) > 0 {
		to_clause = fmt.Sprintf("to='%s'", to)
	}
	template := "<iq id='%s' type='%s' from='%s' %s>%s</iq>"
	packet := fmt.Sprintf(template, request_id, request_type, self.username, to_clause, request)
	fmt.Println(packet)
	self.requests[request_id] = callback
	self.connection.Write([]byte(packet))
}

func (self *XmppHandler) bind() {
    template := "<bind xmlns='urn:ietf:params:xml:ns:xmpp-bind'><resource>%s</resource></bind>"
    request := fmt.Sprintf(template, self.id)
    self.IssueRequest(request, "set", "", self.handleBind)
}

func (self *XmppHandler) handleBind(response *XmppElement) {
	if response.Attributes["type"] == "result" {
		bindElement := response.Children[0]
		if bindElement.Children[0].Tag == "jid" {
			self.jid = bindElement.Children[0].Text
			fmt.Printf("JID is %s\n", self.jid)
		}
		self.startSession()
	}
}
func (self *XmppHandler) startSession() {
    request := "<session xmlns='urn:ietf:params:xml:ns:xmpp-session'/>"
    self.IssueRequest(request, "set", "", self.handleSession)
}

func (self *XmppHandler) handleSession(response *XmppElement) {
	if response.Attributes["type"] == "result" {
		self.sendInitialPresence()
	}
}

func (self *XmppHandler) sendInitialPresence() {
	self.State = "ready"
    packet := "<presence><show/></presence>"
	self.connection.Write([]byte(packet))
}

func (self *XmppHandler) getRequestId() (string) {
	request_id := fmt.Sprintf("id%d", self.request_id)
	self.request_id += 1
	return request_id
}

func (self *XmppHandler) Message(receiver string, msg string) {
    template := "<message from='%s' to='%s' xml:lang='en'><body>%s</body></message>"
    packet := fmt.Sprintf(template, self.username, receiver, msg)
    self.connection.Write([]byte(packet))
}

func (self *XmppHandler) GroupChat(receiver string, msg string) {
    template := "<message from='%s' to='%s' type='groupchat' xml:lang='en'><body>%s</body></message>"
    packet := fmt.Sprintf(template, self.username, receiver, msg)
    self.connection.Write([]byte(packet))
}

func (self *XmppHandler) JoinRoom(room string) string {
	request_id := self.getRequestId()
    template := "<presence from='%s' id='%s' to='%s/%s'>" +
    	"<x xmlns='http://jabber.org/protocol/muc'/></presence>"
	packet := fmt.Sprintf(template, self.jid, request_id, room, self.username)
	self.connection.Write([]byte(packet))
	return request_id
}

func (self *XmppHandler) LeaveRoom(room string) string {
	request_id := self.getRequestId()
    template := "<presence from='%s' id='%s' to='%s/%s' type='unavailable'/>"
	packet := fmt.Sprintf(template, self.jid, request_id, room, self.username)
	self.connection.Write([]byte(packet))
	return request_id
}
