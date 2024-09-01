    
    

import logging
from common.utils import Bet


def readBetFromBytes(client_sock):
    """Deserializes a byte array to a Bet object."""

    len = int.from_bytes(client_sock.recv(2), byteorder='big')
    data = client_sock.recv(len)

    addr = client_sock.getpeername()

    logging.info(f'action: receive_message | result: success | ip: {addr[0]} | msg: {data}')

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

def confirmBet(client_sock):
    client_sock.send("{}\n".format("OK").encode('utf-8'))