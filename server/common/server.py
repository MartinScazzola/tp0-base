import signal
import socket
import logging
import sys

from common.utils import getWinnersForAgency, store_bets
from multiprocessing import Process
from common.bet import (
    sendOkRecvBets,
    recvBets,
    sendFailRecvBets,
    sendWinners,
    recvBeginConnection,
)



class Server:
    def __init__(self, port, listen_backlog, clients_number):
        """Initializes the server socket and sets up the server configuration."""
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.bind(("", port))
        self._server_socket.listen(listen_backlog)
        self.clients_number = clients_number
        self.client_sockets = {}

    def run(self):
        """
        Main server loop.

        Accepts new connections and establishes communication with clients.
        Handles client communication and then resumes accepting new connections.
        """
        signal.signal(signal.SIGTERM, self.__handle_signal)

        handlers = []

        for _ in range(self.clients_number):
            client_sock = self.__accept_new_connection()
            
            client_id = recvBeginConnection(client_sock)

            self.client_sockets[client_id] = client_sock

            handler = Process(
                target=self.__handle_client_connection_sending_bets, args=(client_sock, client_id)
            )
            handler.start()
            handlers.append(handler)

        self.__wait_clients_send_bets(handlers)
        self.__send_winners()
        self.__close_client_sockets()

    def __handle_client_connection_sending_bets(self, client_sock, client_id):
        """
        Handles receiving bets from a specific client socket.

        Manages the reception of bets, stores them, and sends appropriate responses.
        """

        print("Client connected to send bets", client_sock.getpeername())

        while True:
            try:
                bets = recvBets(client_sock, client_id)

                if not bets:
                    break

                store_bets(bets)

                logging.info(
                    f"action: apuesta_recibida | result: success | cantidad: {len(bets)}"
                )
                sendOkRecvBets(client_sock)

            except OSError as e:
                logging.info(
                    f"action: apuesta_recibida | result: fail | cantidad: {len(bets)}"
                )
                sendFailRecvBets(client_sock)

    def __accept_new_connection(self):
        """
        Accepts new client connections.

        Blocks until a client connects, then logs the connection and returns the client socket.
        """
        logging.info("action: accept_connections | result: in_progress")
        c, addr = self._server_socket.accept()
        logging.info(f"action: accept_connections | result: success | ip: {addr[0]}")
        return c

    def __handle_signal(self, signum, frame):
        """
        Handles signals for graceful shutdown of the server.

        Closes all client sockets and the server socket when a termination signal is received.
        """
        logging.info(f"Received signal {signum}. Shutting down server")

        for client_sock in self.client_sockets.values():
            client_sock.close()

        self._server_socket.close()
        logging.info("Server socket closed")
        sys.exit(0)

    def __send_winners(self):
        """
        Checks if all clients have finished sending bets and sends the winners.

        Once all clients have sent their bets, winners are retrieved and sent to each client.
        """
        for id, sock in self.client_sockets.items():
            bets = getWinnersForAgency(id)
            sendWinners(sock, bets)

    def __wait_clients_send_bets(self, handlers):
        """
        Waits for all clients to send their bets.

        Blocks until all clients have sent their bets.
        """
        for handler in handlers:
            handler.join()

    def __close_client_sockets(self):
        """
        Closes all client sockets.

        Iterates over all client sockets and closes them.
        """
        for sock in self.client_sockets.values():
            sock.close()
        self.client_sockets = {}
        logging.info("All client sockets closed")
