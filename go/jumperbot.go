package main

import (
	"./xmpp"
	"time"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
)

var PHRASES = [...]string {
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
}

type JumperBot struct {
	xmppHandler *xmpp.XmppHandler
	name string
	host string
	numRooms int
	currentChannel string
}

func NewJumperBot(numRooms int) *JumperBot{
	return &JumperBot {
		xmppHandler:xmpp.NewXmppHandler(),
		numRooms: numRooms,
	}
}

func (self *JumperBot)Connect(server string, host string, username string, password string) {
	self.name = username
	self.host = host
	self.xmppHandler.Connect(server, host, username, password)
}

func (self *JumperBot)Run() {
	for self.xmppHandler.State != "ready" {
		fmt.Printf("Waiting for %s to log in\n", self.name)
		time.Sleep(time.Second)
	}

	for {
		self.joinRandomRoom()
		n := rand.Intn(5) + 5
		for i := 0; i < n; i++ {
			self.sayRandomPhrase()
			seconds := rand.Intn(10) + 5
			time.Sleep(time.Duration(seconds)*time.Second)
		}
	}
}

func (self *JumperBot) sayRandomPhrase() {
	choice := rand.Intn(len(PHRASES))
	phrase := PHRASES[choice]
	self.xmppHandler.GroupChat(self.currentChannel, phrase)
}

func (self *JumperBot) joinRandomRoom() {
	roomNumber := rand.Intn(self.numRooms)
	room := fmt.Sprintf("bot_room_%d@conference.%s", roomNumber, self.host)
	if room != self.currentChannel {
		if self.currentChannel != "" {
			fmt.Printf("%s: Leaving %s", self.name, self.currentChannel)
			self.xmppHandler.LeaveRoom(self.currentChannel)
		}
		self.currentChannel = room
		self.xmppHandler.JoinRoom(self.currentChannel)
	}
}

func main() {
	bots := make([]*JumperBot, 10)
	for i:= 0; i < 10; i++ {
		fmt.Printf("Creating bot %d", i)
		bot := NewJumperBot(10)
		go bot.Connect(
			"localhost",
			"localhost",
			fmt.Sprintf("jumperbot_%d", i, ),
			"jumperbot")
		go bot.Run()
		bots = append(bots, bot)
	}
    exitSignal := make(chan os.Signal)
    signal.Notify(exitSignal, syscall.SIGINT, syscall.SIGTERM)
    <-exitSignal
}

