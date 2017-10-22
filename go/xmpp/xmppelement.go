package xmpp

import (
	"strings"
	"fmt"
)

type XmppElement struct {
	Tag        string
	Namespace  string
	Attributes map[string]string
	Children   []*XmppElement
	Text       string
}

func (self *XmppElement) ToXml() (string) {
	return self.ToXmlWithIndent(0)
}

func (self *XmppElement) ToXmlWithIndent(indent int) (string) {
	spaces := strings.Repeat(" ", indent)
	attributes := ""
	for k, v := range self.Attributes {
		attributes += fmt.Sprintf("%s=%s ", k, v)
	}
	result := spaces + fmt.Sprintf("<%s %s", self.Tag, attributes)
	if len(self.Children) > 0 || len(self.Text) > 0 {
		result += ">\n"
		for _, child := range self.Children {
			result += child.ToXmlWithIndent(indent + 4)
		}
		if len(self.Text) > 0 {
			result += strings.Repeat(" ", indent + 4) + self.Text + "\n"
		}
		result += spaces + fmt.Sprintf("</%s>\n", self.Tag)
	} else {
		result += "/>\n"
	}

	return result
}