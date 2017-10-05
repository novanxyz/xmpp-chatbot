import socket

s = socket.socket()
s.connect(("localhost", 5222))
message = "<?xml version='1.0'?>" \
          "<stream:stream to='localhost' version='1.0' " \
          "xmlns='jabber:client' xmlns:stream='http://etherx.jabber.org/streams'>"
s.send(message.encode())
print(s.recv(4096))
