package xmpp

import (
	"testing"
)

func TestXmppHandler_Create(t *testing.T) {
	var handler = NewXmppContentHandler()
	if handler == nil {
		t.Error("No XmppContentHandler")
	}
}

func TestXmppHandlerEmptyInput(t *testing.T) {
	var handler = NewXmppContentHandler()
	err := handler.Parse("")

	if err != nil {
		t.Error("Unexpected error: %v", err)
	}
	if len(handler.queue) != 0 {
		t.Error("Queue should be empty")
	}
}

func TestXmppHandlerInvalidInput(t *testing.T) {
	var handler = NewXmppContentHandler()
	err := handler.Parse("<this>xml</invalid>")

	if err == nil {
		t.Errorf("Expected an error")
	}

	if len(handler.queue) != 0 {
		t.Error("Queue should be empty")
	}
}

func TestXmppHandler_Simple(t *testing.T) {
	var handler = NewXmppContentHandler()
	handler.Parse("<simple></simple>")

	if len(handler.queue) != 1 {
		t.Error("Queue should contain one element")
		return
	}

	elem := handler.queue[0]
	if elem.Tag != "simple" {
		t.Error("Element Tag is incorrect")
	}
}

func TestXmppHandler_SimpleSelfClosing(t *testing.T) {
	var handler = NewXmppContentHandler()
	handler.Parse("<simple/>")

	if len(handler.queue) != 1 {
		t.Error("Queue should contain one element")
		return
	}

	elem := handler.queue[0]
	if elem.Tag != "simple" {
		t.Error("Element Tag is incorrect")
	}
}

func TestXmppHandler_SimpleWithText(t *testing.T) {
	var handler = NewXmppContentHandler()
	handler.Parse("<simple>this is some Text</simple>")

	if len(handler.queue) != 1 {
		t.Error("Queue should contain one element")
		return
	}

	elem := handler.queue[0]
	if elem.Tag != "simple" {
		t.Error("Element Tag is incorrect")
	}
	expected := "this is some Text"
	if elem.Text != expected {
		t.Errorf("Element Text\nExpected: %v\n     Got: %v", expected, elem.Text)
	}
}

func TestXmppHandler_SimpleIncremental(t *testing.T) {
	var handler = NewXmppContentHandler()
	err := handler.Parse("<simp")
	if err != nil {
		t.Error("Unexpected error: %v", err)
	}
	if len(handler.queue) != 0 {
		t.Error("Queue should be empty")
		return
	}

	handler.Parse("le></simple>")

	if len(handler.queue) != 1 {
		t.Error("Queue should contain one element")
		return
	}

	elem := handler.queue[0]
	if elem.Tag != "simple" {
		t.Error("Element Tag is incorrect")
	}
}

func TestXmppHandler_Nested(t *testing.T) {
	var handler = NewXmppContentHandler()
	handler.Parse("<parent><child/><child/></parent>")

	if len(handler.queue) != 1 {
		t.Errorf("Queue should contain one element - contains %v", len(handler.queue))
		return
	}

	elem := handler.queue[0]
	if elem.Tag != "parent" {
		t.Error("Element Tag is incorrect")
	}
	if len(elem.Children) != 2 {
		t.Errorf("Expected 2 Children - got %v", len(elem.Children))
	}
}

func TestXmppHandler_NestedIncremental(t *testing.T) {
	var handler = NewXmppContentHandler()
	handler.Parse("<parent><chil")
	handler.Parse("d/><child/></parent>")

	if len(handler.queue) != 1 {
		t.Errorf("Queue should contain one element - contains %v", len(handler.queue))
		return
	}

	elem := handler.queue[0]
	if elem.Tag != "parent" {
		t.Error("Element Tag is incorrect")
	}
	if len(elem.Children) != 2 {
		t.Errorf("Expected 2 Children - got %v", len(elem.Children))
	}
}

func TestXmppHandler_NestedTrailingIncremental(t *testing.T) {
	var handler = NewXmppContentHandler()
	handler.Parse("<parent><chil")
	handler.Parse("d/><child/></parent><second>")

	if len(handler.queue) != 1 {
		t.Errorf("Queue should contain one element - contains %v", len(handler.queue))
		return
	}

	elem := handler.queue[0]
	if elem.Tag != "parent" {
		t.Error("Element Tag is incorrect")
	}
	if len(elem.Children) != 2 {
		t.Errorf("Expected 2 Children - got %v", len(elem.Children))
	}

	handler.Parse("</second>")
	if len(handler.queue) != 2 {
		t.Errorf("Queue should contain two elements - contains %v", len(handler.queue))
		return
	}
}

func TestXmppHandler_SimpleSelfClosingWithAttributes(t *testing.T) {
	var handler = NewXmppContentHandler()
	handler.Parse("<simple x='bingo' y='bongo'/>")

	if len(handler.queue) != 1 {
		t.Error("Queue should contain one element")
		return
	}

	elem := handler.queue[0]
	if elem.Tag != "simple" {
		t.Error("Element Tag is incorrect")
	}

	if len(elem.Attributes) != 2 {
		t.Errorf("Wrong number of Attributes - expected 2, got %v", len(elem.Attributes))
	}

	if elem.Attributes["x"] != "bingo" {
		t.Errorf("Attribute value is incorrect - expected 'bingo', got '%v'", elem.Attributes["x"])
	}
	if elem.Attributes["y"] != "bongo" {
		t.Errorf("Attribute value is incorrect - expected 'bongo', got '%v'", elem.Attributes["x"])
	}
}
func TestXmppHandler_StartStream(t *testing.T) {
	var handler= NewXmppContentHandler()
	handler.Parse("<?xml version='1.0'?>" +
		"<stream:stream xmlns:stream='http://etherx.jabber.org/streams' " +
			"version='1.0' from='localhost' " +
			"id='f57d29c0-acca-4eae-899c-f147e7287415' " +
			"xml:lang='en' xmlns='jabber:client'>" +
		"<stream:features>" +
			"<mechanisms xmlns='urn:ietf:params:xml:ns:xmpp-sasl'>" +
				"<mechanism>PLAIN</mechanism>" +
				"<mechanism>SCRAM-SHA-1</mechanism>" +
				"<mechanism>DIGEST-MD5</mechanism>" +
			"</mechanisms>" +
			"<starttls xmlns='urn:ietf:params:xml:ns:xmpp-tls'/>" +
			"<auth xmlns='http://jabber.org/features/iq-auth'/>" +
		"</stream:features>")

	if len(handler.queue) != 2 {
		t.Error("Queue should contain two elements")
		return
	}
}