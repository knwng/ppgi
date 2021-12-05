import random
import string

def gen_email():
    return ''.join(random.choices(population=string.ascii_letters, k=random.randint(3, 10))) + '@gmail.com'

def gen_id():
    return ''.join(random.choices(population=string.digits, k=18))

def gen_telephone():
    return '1' + ''.join(random.choices(population=string.digits, k=10))

PROVINCES = ('Beijing', 'Shanghai', 'Chongqing', 'Tianjin', 'Shandong',
             'Hainan', 'Liaoning', 'Sichuan')

def gen_province():
    return random.choice(PROVINCES)

GENERATORS = {
    'email': gen_email,
    'id': gen_id,
    'telephone': gen_telephone,
    'province': gen_province,
}
