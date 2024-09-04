from common.utils import Bet
from common.constants import (
    BATCH_END_BYTES,
    BATCH_SENT_FAIL,
    BATCH_SENT_OK,
    BEGIN_SEND_BETS,
    END_SEND_BETS,
    READ_BUFFER_SIZE,
)


def betFromBytes(data, agency):
    """Deserializes a byte array into a Bet object."""
    index = 0

    first_name_length = data[index]
    index += 1
    first_name = data[index : index + first_name_length].decode("utf-8")
    index += first_name_length

    last_name_length = data[index]
    index += 1
    last_name = data[index : index + last_name_length].decode("utf-8")
    index += last_name_length

    dni = int.from_bytes(data[index : index + 4], byteorder="big")
    index += 4

    year = int.from_bytes(data[index : index + 2], byteorder="big")
    index += 2
    month = int.from_bytes(data[index : index + 1], byteorder="big")
    index += 1
    day = int.from_bytes(data[index : index + 1], byteorder="big")
    index += 1

    birth_date = f"{year}-{month:02d}-{day:02d}"

    number = int.from_bytes(data[index : index + 4], byteorder="big")

    return (
        Bet(agency, first_name, last_name, str(dni), birth_date, str(number)),
        index + 4,
    )


def sendOkRecvBets(client_sock):
    """Sends a success message when bets are received."""
    safeWrite(client_sock, BATCH_SENT_OK.encode("utf-8") + BATCH_END_BYTES)


def sendFailRecvBets(client_sock):
    """Sends a failure message when receiving bets fails."""
    safeWrite(client_sock, BATCH_SENT_FAIL.encode("utf-8") + BATCH_END_BYTES)


def safeWrite(client_sock, bytes):
    """Safely writes data to the client socket."""
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


def safeRead(client_sock, amount):
    """Safely reads a specified amount of data from the client socket."""
    buffer = b""

    while len(buffer) < amount and not buffer.endswith(BATCH_END_BYTES):
        chunk = client_sock.recv(amount)
        if not chunk:
            raise Exception("Connection closed by the client")
        buffer += chunk

    return buffer


def parseBetsFromBytes(batchBytes, client_id):
    """Parses a batch of bytes into a list of Bet objects."""
    i = 0
    bets = []
    while i < len(batchBytes):
        currentBets, index = betFromBytes(batchBytes[i:], client_id)
        bets.append(currentBets)
        i += index
    return bets


def recvBets(client_sock, client_id):
    """Receives a list of bets from a client socket."""
    batchBytes = safeRead(client_sock, READ_BUFFER_SIZE)

    while not batchBytes.endswith(BATCH_END_BYTES):
        batchBytes += safeRead(client_sock, READ_BUFFER_SIZE)

    batchBytes = batchBytes[:-2]

    if batchBytes == END_SEND_BETS.encode("utf-8"):
        return None

    bets = parseBetsFromBytes(batchBytes, client_id)

    return bets


def sendWinners(client_sock, documents):
    """Sends a list of winners to a client socket."""
    data = b""

    for document in documents:
        data += int(document).to_bytes(4, byteorder="big")

    data += BATCH_END_BYTES

    safeWrite(client_sock, data)


def recvBeginConnection(client_sock):
    """Receives the client ID from a client socket."""
    bytes = safeRead(client_sock, 8)

    if bytes[: len(BEGIN_SEND_BETS)] != BEGIN_SEND_BETS.encode("utf-8"):
        raise Exception("Invalid connection request")
    byte = bytes[len(BEGIN_SEND_BETS)]

    return byte
