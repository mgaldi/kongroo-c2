import json

def xor(data):
    return bytearray((data[i] ^ ord("K") for i in range(0, len(data))))

config = {}
config['Hostname'] = '127.0.0.1'
config['Migrate'] = ''
config = json.dumps(config).encode()
padding = b"\x00" * (2048 - len(config))
config = xor(config) + padding
#Find {'Hostname
with open('main', 'rb') as binFile:
    byteData = bytearray(binFile.read())
binFile.close()
offset = byteData.find(b'AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA')
print(offset)
with open('main', 'r+b') as binFile:
    binFile.seek(offset)
    binFile.write(config)
binFile.close()
