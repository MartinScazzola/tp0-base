import signal
import socket
import logging
import sys

from common.utils import getWinnersForAgency, store_bets
from common.bet import sendOkRecvBets, recvBets, sendFailRecvBets, sendWinners, recvBeginConnection

class Server:
    def __init__(self, port, listen_backlog):
        # Initialize server socket
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.bind(('', port))
        self._server_socket.listen(listen_backlog)
        self.client_sockets = {}
        self.clientsDoneSendingBets = 0

    def run(self):
        """
        Dummy Server loop

        Server that accept a new connections and establishes a
        communication with a client. After client with communucation
        finishes, servers starts to accept new connections again
        """

        signal.signal(signal.SIGTERM, self.__handle_signal)


        while True:
            client_sock = self.__accept_new_connection()
            self.__handle_client_connection_sending_bets(client_sock)
            self.__check_if_all_clients_done_and_send_winners()




    def __handle_client_connection_sending_bets(self, client_sock):
        """
        Read message from a specific client socket and closes the socket

        If a problem arises in the communication with the client, the
        client socket will also be closed
        """

        client_id = recvBeginConnection(client_sock)

        self.client_sockets[client_id] = client_sock
        print("Se me conecto el cliente", client_id)
        print("Client connected to send bets", client_sock.getpeername())

        while True:
            try:
                bets = recvBets(client_sock)

                if not bets:
                    break

                store_bets(bets)

                logging.info(f"action: apuesta_recibida | result: success | cantidad: {len(bets)}")
                sendOkRecvBets(client_sock)

            except OSError as e:
                logging.info(f"action: apuesta_recibida | result: fail | cantidad: {len(bets)}")
                sendFailRecvBets(client_sock)

        self.clientsDoneSendingBets += 1



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
    
    def __handle_signal(self, signum, frame):
        """
        Handle signals for graceful shutdown.
        """
        logging.info(f"Received signal {signum}. Shutting down server")

        for client_sock in self.client_sockets.values():
            client_sock.close()

        self._server_socket.close()
        logging.info('Server socket closed')
        sys.exit(0)
    
    def __check_if_all_clients_done_and_send_winners(self):
        if self.clientsDoneSendingBets >= 5:
            for id, sock in self.client_sockets.items():
                bets = getWinnersForAgency(id)
                sendWinners(sock, bets)
