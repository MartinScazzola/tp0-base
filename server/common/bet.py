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
    
    return Bet(agency,first_name, last_name, str(dni), birth_date, str(number)), index + 4

def sendOkRecvBets(client_sock):
    safeWrite(client_sock, b"OK|")

def sendFailRecvBets(client_sock):
    safeWrite(client_sock, b"FAIL|")

def safeWrite(client_sock, bytes):
    totalBytesWritten = 0
    dataLength = len(bytes)

    while totalBytesWritten < dataLength:
        try:
            bytesWritten = client_sock.send(bytes[totalBytesWritten:])
            if bytesWritten == 0:
                raise Exception("Connection closed by the client")
            totalBytesWritten += bytesWritten
        except Exception as e:
            raise Exception(f"Error sending the batch: {e}")
        
    return totalBytesWritten

def safeRead(client_sock):
    buffer = b''

    while not buffer.endswith(b'|'):
        chunk = client_sock.recv(1024)
        if not chunk:
            raise Exception("Connection closed by the client")
        buffer += chunk


    if not buffer or buffer == b'|':
        return None

    buffer = buffer.rstrip(b'|')

    return buffer

def parseBetsFromBytes(batchBytes):
    i = 0
    bets = []
    while i < len(batchBytes):
        currentBets, index = betFromBytes(batchBytes[i:])
        bets.append(currentBets)
        i += index
    return bets

def recvBets(client_sock):
    """Receives a list of bets from a client socket."""
    batchBytes = safeRead(client_sock)

    if batchBytes == b"END":
        return None

    bets = parseBetsFromBytes(batchBytes)

    return bets