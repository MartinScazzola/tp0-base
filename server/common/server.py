import signal
import socket
import logging
import sys

from common.utils import Bet, store_bets
from common.bet import readBetFromBytes, confirmBet

class Server:
    def __init__(self, port, listen_backlog):
        # Initialize server socket
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.bind(('', port))
        self._server_socket.listen(listen_backlog)
        self.client_sockets = []

    def run(self):
        """
        Dummy Server loop

        Server that accept a new connections and establishes a
        communication with a client. After client with communucation
        finishes, servers starts to accept new connections again
        """

        signal.signal(signal.SIGTERM, self._handle_signal)


        while True:
            client_sock = self.__accept_new_connection()
            self.__handle_client_connection(client_sock)

    def __handle_client_connection(self, client_sock):
        """
        Read message from a specific client socket and closes the socket

        If a problem arises in the communication with the client, the
        client socket will also be closed
        """

        self.client_sockets.append(client_sock)

        try:

            bet = readBetFromBytes(client_sock)

            store_bets([bet])
            
            logging.info(f'action: apuesta_almacenada | result: success | dni: {bet.document} | numero: ${bet.number}')

            confirmBet(client_sock)
            
        except OSError as e:
            logging.error("action: receive_message | result: fail | error: {e}")
        finally:
            client_sock.close()
            self.client_sockets.remove(client_sock)

    def __accept_new_connection(self):
        """
        Accept new connections

        Function blocks until a connection to a client is made.
        Then connection created is printed and returned
        """

        # Connection arrived
        logging.info('action: accept_connections | result: in_progress')
        c, addr = self._server_socket.accept()
        logging.info(f'action: accept_connections | result: success | ip: {addr[0]}')
        return c
    
    def _handle_signal(self, signum, frame):
        """
        Handle signals for graceful shutdown.
        """
        logging.info(f"Received signal {signum}. Shutting down server")

        for client_sock in self.client_sockets:
            client_sock.close()
            self.client_sockets.remove(client_sock)

        self._server_socket.close()
        logging.info('Server socket closed')
        sys.exit(0)
