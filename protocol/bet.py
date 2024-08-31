class Bet:
    def __init__(self, nombre, apellido, dni, nacimiento, numero):
        self.nombre = nombre
        self.apellido = apellido
        self.dni = dni
        self.nacimiento = nacimiento
        self.numero = numero

    def to_bytes(self):
        """Serializes the Bet object to a byte array."""
        data = bytearray()
        
        first_name_bytes = self.nombre.encode('utf-8')
        data.extend(len(first_name_bytes).to_bytes(1, byteorder='big'))
        data.extend(first_name_bytes)
        
        last_name_bytes = self.apellido.encode('utf-8')
        data.extend(len(last_name_bytes).to_bytes(1, byteorder='big'))
        data.extend(last_name_bytes)
        
        data.extend(self.dni.to_bytes(4, byteorder='big'))

        year = self.nacimiento.split("-")[0]
        month = self.nacimiento.split("-")[1]
        day = self.nacimiento.split("-")[2]

        data.extend(int(year).to_bytes(2, byteorder='big'))
        data.extend(int(month).to_bytes(1, byteorder='big'))
        data.extend(int(day).to_bytes(1, byteorder='big'))
        
        data.extend(self.numero.to_bytes(4, byteorder='big'))

        return bytes(len(data).to_bytes(2, byteorder='big') + data)
    
    @classmethod
    def from_bytes(self, data):
        """Deserializes a byte array to a Bet object."""
        index = 0

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
        
        return self(first_name, last_name, dni, birth_date, number)
    
    def __repr__(self):
        return (f"Bet(nombre={self.nombre}, apellido={self.apellido}, dni={self.dni}, "
                f"nacimiento={self.nacimiento}, numero={self.numero})")
    

bet = Bet("Juan", "Perez", 12345678, "1990-01-01", 42)
print("bet original", bet)

bytes = bet.to_bytes()

len = int.from_bytes(bytes[:2], byteorder='big')

bet_recv = Bet.from_bytes(bytes[2:len + 2])

print("bet_recv", bet_recv)