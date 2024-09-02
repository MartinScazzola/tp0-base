    
    

import logging
from common.utils import Bet


def betFromBytes(data):
    """Deserializes a byte array to a Bet object."""

    index = 0

    agency = int.from_bytes(data[index:index + 1], byteorder='big')
    index += 1

    first_name_length = data[index]
    index += 1
    first_name = data[index:index + first_name_length].decode('utf-8')
    index += first_name_length
    
    last_name_length = data[index]
    index += 1
    last_name = data[index:index + last_name_length].decode('utf-8')
    index += last_name_length

    dni = int.from_bytes(data[index:index + 4], byteorder='big')
    index += 4

    year = int.from_bytes(data[index:index + 2], byteorder='big')
    index += 2
    month = int.from_bytes(data[index:index + 1], byteorder='big')
    index += 1
    day = int.from_bytes(data[index:index + 1], byteorder='big')
    index += 1

    birth_date = f"{year}-{month:02d}-{day:02d}"

    number = int.from_bytes(data[index:index + 4], byteorder='big')
    
    return Bet(agency,first_name, last_name, str(dni), birth_date, str(number))

def confirmBatch(client_sock):
    client_sock.send("{}\n".format("OK").encode('utf-8'))

def parseBatchToBets(batchBytes):
    index = 0
    bets = []

    while index < len(batchBytes):
        betBytesLen = int.from_bytes(batchBytes[index: index + 2], byteorder='big')
        
        index += 2

        bet = betFromBytes(batchBytes[index:index + betBytesLen])
        bets.append(bet)

        index += betBytesLen
    return bets


def recvBets(client_sock):

    lenMsgType = int.from_bytes(client_sock.recv(1), byteorder='big')

    msgType = client_sock.recv(lenMsgType).decode('utf-8')

    if msgType == "END":
        return msgType, None

    #msgType == "BATCH"

    batchSize = int.from_bytes(client_sock.recv(2), byteorder='big')

    print("batchSize", batchSize)

    batchBytes = client_sock.recv(batchSize)

    bets = parseBatchToBets(batchBytes)

    confirmBatch(client_sock)

    return bets