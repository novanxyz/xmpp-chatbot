import logging
from typing import List
from xml.sax import ContentHandler

from xmppelement import XmppElement

logger = logging.getLogger(__name__)


class XmppContentHandler(ContentHandler):
    def __init__(self):
        ContentHandler.__init__(self)
        self.queue = []  # type: List[XmppElement]
        self.current_element = None
        self.element_stack = []

    def startElement(self, name, attrs):
        element = XmppElement()
        element.tag = name
        element.attributes = attrs

        if name == "stream:stream":
            # This element won't close until stream is closed
            self.queue.append(element)
            return

        self.element_stack.append(self.current_element)
        self.current_element = element

    def endElement(self, name):
        if not self.current_element:
            if name == "stream:stream":
                element = XmppElement()
                element.tag = "stream:closed"
                return
            else:
                raise RuntimeError("No current element")

        if name != self.current_element.tag:
            raise RuntimeError("Mismatched tag")

        parent = self.element_stack.pop()
        if parent:
            parent.children.append(self.current_element)
        else:
            self.queue.insert(0, self.current_element)
        self.current_element = parent

    def characters(self, content):
        if self.current_element.text:
            self.current_element.text += content
        else:
            self.current_element.text = content

