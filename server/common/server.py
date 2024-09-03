import signal
import socket
import logging
import sys

from common.utils import getWinnersForAgency, has_won, load_bets, store_bets
from common.bet import recvAction, sendOkRecvBets, recvBets, sendFailRecvBets, sendWaitForWinners, sendWinners

class Server:
    def __init__(self, port, listen_backlog):
        # Initialize server socket
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.bind(('', port))
        self._server_socket.listen(listen_backlog)
        self.client_sockets = []
        self.clientsDoneSendingBets = 0

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

            action, client_id = recvAction(client_sock)

            if action == "BEGIN SEND BETS":
                self.__handle_client_connection_sending_bets(client_sock)
            elif action == "GET WINNERS":
                self.__handle_client_connection_asking_for_winners(client_sock, client_id)
            else:
                logging.info(f"action: accept_connections | result: fail | ip: {client_sock.getpeername()[0]}")
                client_sock.close()
                self.client_sockets.remove(client_sock)


    def __handle_client_connection_sending_bets(self, client_sock):
        """
        Read message from a specific client socket and closes the socket

        If a problem arises in the communication with the client, the
        client socket will also be closed
        """

        self.client_sockets.append(client_sock)
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
        client_sock.close()
        self.client_sockets.remove(client_sock)
    
    def __handle_client_connection_asking_for_winners(self, client_sock, client_id):
        """
        Read message from a specific client socket and closes the socket

        If a problem arises in the communication with the client, the
        client socket will also be closed
        """

        self.client_sockets.append(client_sock)

        if self.clientsDoneSendingBets < 5:
            sendWaitForWinners(client_sock)
        else:
            winnersDocument = getWinnersForAgency(client_id)
            logging.info(f"action: sorteo | result: success")
            sendWinners(client_sock, winnersDocument)
        
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
