#!/usr/bin/env python3
import asyncio
import logging
import sys


logging.basicConfig(level=logging.DEBUG)

logger = logging.getLogger(__name__)


class FirstConnection(asyncio.Protocol):
    def __init__(self, host):
        self.host = host
        self.transport = None

    def connect(self):
        loop = asyncio.get_event_loop()
        handler = loop.create_connection(lambda: self, self.host, 5222)
        loop.create_task(handler)

    def connection_made(self, transport):
        logger.debug("Connection made")
        self.transport = transport

        cmd = "<?xml version='1.0'?><stream:stream to='localhost' " \
              "version='1.0' xmlns='jabber:client' " \
              "xmlns:stream='http://etherx.jabber.org/streams'>"
        self.write(cmd)

        cmd = "<junk/>"
        self.write(cmd)

    def connection_lost(self, exc):
        logger.debug("Connection lost")

    def write(self, data):
        logger.debug("Send: %s", data)
        self.transport.write(data.encode())

    def data_received(self, data):
        logger.debug("Received %d bytes\n%s", len(data), data.decode())


def main():
    logger.debug("FirstConnection is starting")
    loop = asyncio.get_event_loop()

    bot = FirstConnection("localhost")
    bot.connect()
    loop.run_forever()

    loop.close()


if __name__ == "__main__":
    sys.exit(main())
