import csv
import datetime
import time


""" Bets storage location. """
STORAGE_FILEPATH = "./bets.csv"
""" Simulated winner number in the lottery contest. """
LOTTERY_WINNER_NUMBER = 7574


""" A lottery bet registry. """
class Bet:
    def __init__(self, agency: str, first_name: str, last_name: str, document: str, birthdate: str, number: str):
        """
        agency must be passed with integer format.
        birthdate must be passed with format: 'YYYY-MM-DD'.
        number must be passed with integer format.
        """
        self.agency = int(agency)
        self.first_name = first_name
        self.last_name = last_name
        self.document = document
        self.birthdate = datetime.date.fromisoformat(birthdate)
        self.number = int(number)

    @classmethod
    def fromBytes(self, data):
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
        
        return self(agency,first_name, last_name, str(dni), birth_date, str(number))
    
    def __repr__(self):
        return (f"Bet(nombre={self.first_name}, apellido={self.last_name}, dni={self.document}, "
                f"nacimiento={self.birthdate}, numero={self.number})")

""" Checks whether a bet won the prize or not. """
def has_won(bet: Bet) -> bool:
    return bet.number == LOTTERY_WINNER_NUMBER

"""
Persist the information of each bet in the STORAGE_FILEPATH file.
Not thread-safe/process-safe.
"""
def store_bets(bets: list[Bet]) -> None:
    with open(STORAGE_FILEPATH, 'a+') as file:
        writer = csv.writer(file, quoting=csv.QUOTE_MINIMAL)
        for bet in bets:
            writer.writerow([bet.agency, bet.first_name, bet.last_name,
                             bet.document, bet.birthdate, bet.number])

"""
Loads the information all the bets in the STORAGE_FILEPATH file.
Not thread-safe/process-safe.
"""
def load_bets() -> list[Bet]:
    with open(STORAGE_FILEPATH, 'r') as file:
        reader = csv.reader(file, quoting=csv.QUOTE_MINIMAL)
        for row in reader:
            yield Bet(row[0], row[1], row[2], row[3], row[4], row[5])

