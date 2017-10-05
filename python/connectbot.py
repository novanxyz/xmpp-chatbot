#!/usr/bin/env python3
import argparse
import asyncio
import logging
import sys

from xmpphandler import XmppHandler

logging.basicConfig(level=logging.DEBUG)

logger = logging.getLogger(__name__)


class ConnectBot(asyncio.Protocol):
    def __init__(self, host, username, password=None, servername=None):
        self.host = host
        self.servername = servername or host
        self.username = "{0}@{1}".format(username, host)
        self.password = password or username
        self.xmppHandler = None
        self.transport = None

    def connect(self):
        loop = asyncio.get_event_loop()
        handler = loop.create_connection(lambda: self, self.servername, 5222)
        loop.create_task(handler)

    def connection_made(self, transport):
        logger.debug("Connection made")
        self.transport = transport
        self.xmppHandler = XmppHandler()
        self.xmppHandler.send = self.write

        self.xmppHandler.connect(self.host, self.username, self.password)

    def connection_lost(self, exc):
        logger.debug("Connection lost")

    def write(self, data):
        logger.debug("%s: Send: %s", self.username, data.encode())
        self.transport.write(data.encode())

    def data_received(self, data):
        logger.debug("%s: Received %d bytes", self.username, len(data))
        self.xmppHandler.handle_raw_response(data.decode())


def main():
    parser = argparse.ArgumentParser()

    parser.add_argument(
        "-t", "--host-name",
        default="localhost",
        help="Host name"
    )

    parser.add_argument(
        "-e", "--server-name",
        default=None,
        help="Server name where the host is running"
    )

    args = parser.parse_args()

    logger.debug("Echobot is starting")
    loop = asyncio.get_event_loop()

    bot = ConnectBot(args.host_name, "echobot", servername=args.server_name)
    bot.connect()
    loop.run_forever()

    loop.close()


if __name__ == "__main__":
    sys.exit(main())
