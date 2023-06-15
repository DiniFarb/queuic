# Queuic Message Broker

## Introduction
This is a very small qeueue message broker that I wrote for learning pur. It is not meant to be used in production.

## Features
- [x] Simple massage queueing
- [x] Message persistence
- [x] Own protocol
- [ ] Server implementation
- [ ] load from disk bug
- [ ] http interface naming, missing methods
- [ ] http interface tests 
- [x] Encryption without certs

## Protocol

```
    0               1               2               3               4             
    0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9
   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
   |     Command   |                           Queue Name                          |
   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
   |                            Item UUID                          |               |
   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+               |
   |                            Item....                                           |
   +-+-+-+-+-+-+-+-+-+-+-+-+-+
```
   