import base64
import logging
import uuid
from xml.sax import make_parser
from xml.sax.saxutils import escape

from xmppcontenthandler import XmppContentHandler

logger = logging.getLogger(__name__)


class XmppHandler(object):
    """
    This class knows how to read and write the underlying XML of the XMPP
    protocol.

    It is owned by a particular `XmppConnection` instance, which feeds it with
    data to process.

    The `handle_raw_response` method takes raw incoming data (presumably from a
    socket), parses it as XML, and responds to it as per the XMPP protocols. It
    does this with a combination of a state-machine (for the negotiation logic)
    and a request/callback system (for "bind" and "session" stuff).

    In addition to that, methods like `method`, `groupchat` and `subject`, etc.
    perform direct `send` ops, where the (initially null) implementation of
    `send` is plugged-in by some configuration activity e.g. monkey-patched by
    the owning `XmppConnection`.
    """

    def __init__(self):
        self.host = None
        self.username = None
        self.nick = None
        self.jid = None
        self.password = None

        self.id = uuid.uuid4().hex

        # Tracks where we are in the "state machine"
        self.state = "initial"

        # An XML parser for reading incoming XMPP data
        self.parser = make_parser()

        # Plugs into the parser, and gives us a queue of XML elements
        self.content_handler = None

        # Holds callback functions for all currently outstanding requests by ID
        self.requests = {}

        # A simple perpetually increasing ID for tracking requests
        self.next_request_id = 1

        # Generic response-handler entry-point for things which aren't handled
        # with callbacks. This hook is dynamically switchable according to our
        # current high-level state.
        self.handle_response = self.handle_response_state_not_ready

        # These don't seem to be strictly necessary
        self.stream_element = None
        self.stream_features_element = None

    def connect(self, host, username, password):
        self.host = host
        self.username = username
        self.nick = username.split("@")[0]
        self.jid = self.username
        self.password = password

        self.state = "waiting_for_stream"
        self.start_stream()

    def start_stream(self):
        self.parser.reset()
        self.content_handler = XmppContentHandler()
        self.parser.setContentHandler(self.content_handler)
        template = "<?xml version='1.0'?>" \
                   "<stream:stream to='{0}' version='1.0' " \
                   "xmlns='jabber:client' " \
                   "xmlns:stream='http://etherx.jabber.org/streams'>"
        self.send(template.format(self.host))

    def authenticate(self):
        key = base64.b64encode(
            "\0{0}\0{1}".format(self.nick, self.password).encode(
                "ascii")).decode()
        logger.debug("Auth Key: %s", key)
        template = "<auth xmlns='urn:ietf:params:xml:ns:xmpp-sasl' " \
                   "mechanism='PLAIN'>{0}</auth>"
        package = template.format(key)
        self.state = "authenticating"
        self.send(package)

    def bind(self):
        template = "<bind xmlns='urn:ietf:params:xml:ns:xmpp-bind'>" \
                   "<resource>{0}</resource>" \
                   "</bind>"
        request = template.format(self.id)
        self.issue_request(request, "set", self.handle_bind)

    def handle_bind(self, response):
        logger.debug("Got response to bind")
        if response.children[0].tag == "bind":
            bind = response.children[0]
            if bind.children[0].tag == "jid":
                self.jid = bind.children[0].text
                logger.debug("JID is %s", self.jid)
        self.start_session()

    def start_session(self):
        template = "<session xmlns='urn:ietf:params:xml:ns:xmpp-session'/>"
        request = template.format(self.host)
        self.issue_request(request, "set", self.handle_session)

    def handle_session(self, response):
        self.send_initial_presence()
        self.state = "ready"

    def send_initial_presence(self):
        package = "<presence><show/></presence>"
        self.send(package)

    def message(self, receiver, text):
        template = "<message from='{0}' to='{1}' xml:lang='en'>" \
                   "<body>{2}</body>" \
                   "</message>"
        package = template.format(self.username, receiver, escape(text))
        self.send(package)

    def groupchat(self, receiver, text):
        template = "<message from='{0}' to='{1}' " \
                   "type='groupchat' xml:lang='en'><body>{2}</body></message>"
        package = template.format(self.username, receiver, escape(text))
        self.send(package)

    def subject(self, receiver, text):
        template = "<message from='{0}' to='{1}' type='groupchat' " \
                   "xml:lang='en'><subject>{2}</subject></message>"
        package = template.format(self.username, receiver, escape(text))
        self.send(package)

    def handle_raw_response(self, response):
        logger.debug("Recv: %s", response)
        self.parser.feed(response)
        while len(self.content_handler.queue):
            element = self.content_handler.queue.pop()
            logger.debug("State: %s - Popped Response:\n%s", self.state,
                         element.toXml())
            if element.tag == "iq":
                # Handle an "iq" ("Info/Query") response by using its
                # "id" to locate the relevant callback.
                request_id = element.attributes.getValue("id")
                try:
                    callback = self.requests[request_id]
                    if callback:
                        callback(element)
                    del self.requests[request_id]
                except KeyError:
                    logger.warning("No callback found for request %s",
                                   request_id)
            else:
                # All other elements are handled generically via a state-machine
                result = self.handle_response(element)
                if not result:
                    logger.warning("Unhandled response\n%s", element.toXml())

    def handle_response_state_not_ready(self, response):
        """
        Return True iff the response was 'handled' (including as an error?),
        otherwise False
        """
        if self.state == "waiting_for_stream":
            if response.tag == "stream:stream":
                self.stream_element = response
                self.state = "waiting_for_features"
                return True
            else:
                logger.warning("Expected stream:stream tag")
                return True

        elif self.state == "waiting_for_features":
            if response.tag == "stream:features":
                self.stream_features_element = response
                self.authenticate()
                return True

        elif self.state == "authenticating":
            if response.tag == "success":
                self.start_stream()
                self.state = "authenticated_waiting_for_stream"
                return True
            elif response.tag == "failure":
                logger.warning("Failed to log in")
                return True

        elif self.state == "authenticated_waiting_for_stream":
            if response.tag == "stream:stream":
                self.stream_element = response
                logger.debug("Logged in with id: %s",
                             response.attributes.getValue("id"))
                self.state = "authenticated_waiting_for_features"
                return True

        elif self.state == "authenticated_waiting_for_features":
            if response.tag == "stream:features":
                self.stream_features_element = response
                self.bind()
                return True

        elif self.state == "ready":
            self.handle_logged_in()
            self.handle_response = self.handle_response_state_ready
            return self.handle_response_state_ready(response)

    def handle_response_state_ready(self, response):
        if response.tag == "message":
            self.handle_message(response)
            return True

        elif response.tag == "presence":
            self.handle_presence(response)
            return True

        elif response.tag == "stream:closed":
            self.state = "closed"
            self.handle_response = self.handle_response_state_not_ready
            self.handle_closed()
            return True

        elif response.tag == "stream:error":
            self.handle_stream_error(response)
            return True

    def handle_logged_in(self):
        # Override this to handle notification of successful login
        pass

    def handle_message(self, response):
        # Override this to handle message stanzas
        pass

    def handle_presence(self, response):
        # Override this to handle presence stanzas
        pass

    def handle_closed(self):
        # Override this to handle end of stream
        pass

    def handle_stream_error(self, response):
        # Override this to handle stream errors
        pass

    def send(self, package):
        # Override this to send data
        pass

    def issue_request(self, request, request_type, callback, to=None):
        request_id = self.get_request_id()
        if to is not None:
            to_clause = "to='{0}'".format(to)
        else:
            to_clause = ""
        package = "<iq id='{0}' type='{1}' from='{2}' {3}>{4}</iq>".format(
            request_id, request_type, self.jid, to_clause, request)
        self.requests[request_id] = callback
        self.send(package)
        return request_id

    def join_room(self, room, password=""):
        request_id = self.get_request_id()
        if password:
            password_element = "<password>{0}</password>".format(password)
        else:
            password_element = ""
        template = "<presence from='{0}' id='{1}' to='{2}/{3}'>" \
                   "<x xmlns='http://jabber.org/protocol/muc'>{4}</x>" \
                   "</presence>"
        package = template.format(
            self.jid,
            request_id,
            room,
            self.nick,
            password_element
        )
        self.send(package)
        return request_id

    def leave_room(self, room):
        request_id = self.get_request_id()
        template = "<presence from='{0}' id='{1}' " \
                   "to='{2}/{3}' type='unavailable'/>"
        package = template.format(self.jid, request_id, room, self.nick)
        self.send(package)
        return request_id

    def get_request_id(self):
        request_id = "id{0}".format(self.next_request_id)
        self.next_request_id += 1
        return request_id

    def create_room(self, room):
        # This is just a copy of join_room for now - needs to be extended
        request_id = self.get_request_id()
        template = "<presence from='{0}' id='{1}' to='{2}/{3}'>" \
                   "<x xmlns='http://jabber.org/protocol/muc'/>" \
                   "</presence>"
        package = template.format(self.jid, request_id, room, self.nick)
        self.send(package)
        return request_id
