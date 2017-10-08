#!/usr/bin/env python3
import argparse
import random
import asyncio
import logging

import sys

from xmpphandler import XmppHandler

logger = logging.getLogger(__name__)

PHRASES = [
    "Today a man knocked on my door and asked for a small donation towards the local swimming pool. I gave him a glass of water.",
    "A recent study has found that women who carry a little extra weight live longer than the men who mention it.",
    "Life is all about perspective. The sinking of the Titanic was a miracle to the lobsters in the ship's kitchen.",
    "You know that tingly little feeling you get when you like someone? That's your common sense leaving your body.",
    "I'm great at multitasking. I can waste time, be unproductive, and procrastinate all at once.",
    "If i had a dollar for every girl that found me unattractive, they would eventually find me attractive.",
    "I want to die peacefully in my sleep, like my grandfather.. Not screaming and yelling like the passengers in his car.",
    "My wife and I were happy for twenty years. Then we met.",
    "Isn't it great to live in the 21st century? Where deleting history has become more important than making it.",
    "I find it ironic that the colors red, white, and blue stand for freedom until they are flashing behind you.",
    "Just read that 4,153,237 people got married last year, not to cause any trouble but shouldn't that be an even number?",
    "Relationships are a lot like algebra. Have you ever looked at your X and wondered Y?",
    "Life is like toilet paper, you're either on a roll or taking shit from some asshole.",
    "Apparently I snore so loudly that it scares everyone in the car I'm driving.",
    "When wearing a bikini, women reveal 90 % of their body... men are so polite they only look at the covered parts.",
    "I can totally keep secrets. It's the people I tell them to that can't.",
    "Alcohol is a perfect solvent: It dissolves marriages, families and careers.",
    "Strong people don't put others down. They lift them up and slam them on the ground for maximum damage.",
    "I like to finish other people's sentences because... my version is better.",
    "When my boss asked me who is the stupid one, me or him? I told him everyone knows he doesn't hire stupid people.",
    "Hi!",
    "What's up?",
    "Hello",
    "Wazzup?",
    "Dude!",
    "Yo!",
    "I'm here!",
    "Howdy!",
    "Greetings",
    "Let's jump!",
    "Shoot'em'all!",
    "42",
]


class JumperBot(asyncio.Protocol):
    def __init__(self, manager, host, username, password, num_rooms):
        self.manager = manager
        self.host = host
        self.username = "{0}@{1}".format(username, host)
        self.password = password
        self.xmppHandler = None
        self.transport = None

        self.num_rooms = num_rooms
        self.current_channel = None
        self.listener = False
        self.task = None

    def connection_made(self, transport):
        logger.debug("%s: Connection made", self.username)
        self.transport = transport
        self.xmppHandler = XmppHandler()
        self.xmppHandler.handle_message = self.handle_xmpp_message
        self.xmppHandler.send = self.write

        self.xmppHandler.handle_presence = self.handle_xmpp_presence
        self.xmppHandler.handle_closed = self.handle_closed
        self.xmppHandler.handle_stream_error = self.handle_stream_error
        self.xmppHandler.connect(self.host, self.username, self.password)
        self.task = asyncio.get_event_loop().create_task(self.run())

    def connection_lost(self, exc):
        logger.warning("Connection lost")
        self.transport.close()
        self.xmppHandler = None
        self.manager.notify_closed(self.username)
        if self.task:
            self.task.cancel()

    def write(self, data):
        logger.debug("%s: Send: %s", self.username, data.encode())
        self.transport.write(data.encode())

    def data_received(self, data):
        logger.debug("%s: Received %d bytes", self.username, len(data))
        self.xmppHandler.handle_raw_response(data.decode())

    def handle_xmpp_message(self, response):
        pass

    def handle_xmpp_presence(self, response):
        try:
            if response.attributes.getValue("type") == "error":
                logger.warning(response.toXml())
        except Exception:
            pass

    def handle_closed(self):
        logger.warning("Stream closed")
        self.transport.close()

    def handle_stream_error(self, response):
        logger.warning("Stream error:\n%s", response.toXml())
        self.transport.close()

    def join_random_room(self):
        room = random.randrange(self.num_rooms)
        new_channel = "bot_room_{0}@conference.{1}".format(room, self.host)

        if new_channel != self.current_channel:
            if self.current_channel is not None:
                logger.debug("%s: Leaving %s", self.username,
                             self.current_channel)
                self.xmppHandler.leave_room(self.current_channel)

        self.current_channel = new_channel

        logger.debug("%s: Joining %s", self.username, self.current_channel)
        self.xmppHandler.join_room(self.current_channel)

    def say_random_phrase(self):
        if self.listener:
            return
        phrase = random.choice(PHRASES)
        logger.debug("%s: %s", self.username, phrase)
        self.xmppHandler.groupchat(self.current_channel, phrase)

    async def run(self):
        while self.xmppHandler.state != "ready":
            if self.transport.is_closing():
                return
            logger.debug("Waiting for %s to log in", self.username)
            try:
                await asyncio.sleep(1)
            except asyncio.CancelledError:
                return

        self.manager.notify_login(self.username)
        logger.info("%s logged in", self.username)
        while True:
            self.join_random_room()
            n = random.randint(5, 10)
            for i in range(n):
                if self.transport.is_closing():
                    return
                self.say_random_phrase()
                try:
                    await asyncio.sleep(random.random() * 10.0 + 5.0)
                except asyncio.CancelledError:
                    return


class BotManager(object):
    def __init__(self):
        self.bots_running = {}
        self.bots_logged_in = {}
        self.args = None

    def create_bot(self, botname, args):
        bot = JumperBot(self, args.host_name, botname, "jumperbot",
                        args.num_rooms)
        if args.listener:
            bot.listener = True

        self.connect_bot(bot, args)

        return bot

    def connect_bot(self, bot, args):
        loop = asyncio.get_event_loop()
        handler = loop.create_connection(lambda: bot, args.server_name, 5222)
        loop.create_task(handler)

    def create_bots(self, args):
        self.args = args
        for i in range(args.num_bots):
            botname = "jumperbot_{0}".format(i)
            bot = self.create_bot(botname, args)
            self.bots_running[bot.username] = bot

    async def monitor_status(self, display_stats):
        blinkers = [" ", ".", ":", "."]
        blinker_index = 0
        template = "{2} bots running, {3} logged in {4}"
        while True:
            await asyncio.sleep(1)
            if display_stats:
                print(template.format(
                    len(self.bots_running),
                    len(self.bots_logged_in),
                    blinkers[blinker_index]),
                    end="\r"
                )
            blinker_index += 1
            blinker_index %= len(blinkers)

            if len(self.bots_running) == 0:
                asyncio.get_event_loop().stop()
                return

    def notify_login(self, username):
        self.bots_logged_in[username] = True

    def notify_closed(self, username):
        try:
            del self.bots_logged_in[username]
            logger.info("Reconnecting %s", username)
            self.connect_bot(self.bots_running[username], self.args)
        except KeyError:
            pass


async def request_stop():
    loop = asyncio.get_event_loop()
    loop.stop()


def run(args):
    logger.info("Jumperbot starting with %d instances", args.num_bots)

    manager = BotManager()

    manager.create_bots(args)
    loop = asyncio.get_event_loop()

    loop.create_task(manager.monitor_status(args.monitor))

    try:
        loop.run_forever()
    except KeyboardInterrupt:
        for task in asyncio.Task.all_tasks():
            task.cancel()
        asyncio.ensure_future(exit())

    loop.close()


def main():
    parser = argparse.ArgumentParser()

    parser.add_argument(
        "-n", "--num-bots",
        type=int,
        default=10,
        help="The number of bots to run per process"
    )

    parser.add_argument(
        "-t", "--host-name",
        default="localhost",
        help="Host name"
    )

    parser.add_argument(
        "-e", "--server-name",
        help="Server name where the host is running (if different from host name)"
    )

    parser.add_argument(
        "-r", "--num_rooms",
        type=int,
        default=10,
        help="Number of rooms to jump between. Assumes they have been created."
    )

    parser.add_argument(
        "-l", "--listener",
        action="store_true",
        help="Listen only, don't say anything"
    )

    parser.add_argument(
        "-v", "--verbose",
        action="store_true",
        help="Detailed logging"
    )

    parser.add_argument(
        "-m", "--monitor",
        action="store_true",
        help="When set, display status of bots logged in"
    )

    args = parser.parse_args()

    level = logging.INFO
    if args.verbose:
        level = logging.DEBUG

    logging.basicConfig(format='%(process)d %(asctime)s %(levelname)s %(message)s', datefmt='%m/%d/%Y %I:%M:%S %p', level=level)

    run(args)

if __name__ == "__main__":
    sys.exit(main())