package xmpp

import (
	"encoding/xml"
	"bytes"
	"io"
	"strings"
	"errors"
	"container/list"
)

type XmppContentHandler struct {
	queue []*XmppElement
	current_element *XmppElement
	element_stack *list.List
	leftovers string

	// Keep previous State to restore it in the event of an incomplete input
	prev_queue []*XmppElement
	prev_current_element *XmppElement
	prev_element_stack *list.List
}

func NewXmppContentHandler() *XmppContentHandler {
	return &XmppContentHandler{}
}

func (handler *XmppContentHandler) Parse(input string) (err error) {
	buffer := new(bytes.Buffer)
	parser := xml.NewDecoder(buffer)
	var current_input = handler.leftovers + input
	buffer.WriteString(current_input)
	for {
		var token xml.Token
		token, err = parser.Token()
		if err == io.EOF {
			err = nil
			break
		}
		if err != nil && strings.HasSuffix(err.Error(), "unexpected EOF") {
			// Restore State to last know good
			handler.current_element = handler.prev_current_element
			handler.queue = handler.prev_queue
			handler.element_stack = handler.prev_element_stack
			handler.leftovers = current_input

			err = nil
			break
		}
		if err != nil {
			break
		}

		if handler.element_stack == nil {
			handler.element_stack = list.New()
		}

		switch token.(type)  {
		case xml.StartElement:
			start := token.(xml.StartElement)
			element := XmppElement{
				Tag:        start.Name.Local,
				Namespace:  start.Name.Space,
				Attributes: map[string]string{},
			}
			for _, attribute := range start.Attr  {
				element.Attributes[attribute.Name.Local] = attribute.Value
			}
			if element.Tag == "stream" {
	            // This element won't close until stream is closed
				handler.queue = append(handler.queue, &element)
				continue
			}
			handler.element_stack.PushFront(handler.current_element)
			handler.current_element = &element

		case xml.EndElement:
			if handler.current_element == nil {
				err = errors.New("No current element")
				return
			}
			parentListNode := handler.element_stack.Front()
			handler.element_stack.Remove(parentListNode)
			parent := parentListNode.Value.(*XmppElement)
			current_element := handler.current_element
			handler.current_element = parent
			if parent != nil {
				parent.Children = append(parent.Children, current_element)
			} else {
				handler.queue = append(handler.queue, current_element)

				// We've just completed a stanza - store the current State
				// so we can handle incomplete xml snippets
				handler.prev_current_element = handler.current_element
				handler.prev_queue = handler.queue
				handler.prev_element_stack = handler.element_stack

				current_input, _ = buffer.ReadString(0)
				buffer.WriteString(current_input)
			}

		case xml.CharData:
			if handler.current_element != nil {
				handler.current_element.Text = string(token.(xml.CharData))
			}
		}
	}

	return err
}
