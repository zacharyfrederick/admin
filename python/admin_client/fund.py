import requests
import json
from termcolor import colored

from .utils import print_creation_msg

class Fund:
    url = "http://localhost:8080/funds"
    headers = {'Content-type': 'application/json'}

    def __init__(self, id, name, inceptionDate):
        self.id = id
        self.name = name
        self.inceptionDate = inceptionDate

    def read(self): 
        endpoint = Fund.url + "/" + self.id
        r = requests.get(url=endpoint)
        if r.status_code == 200:
            print("fund bootstrap successful: ", json.dumps(r.json(), indent=4))
        else:
            print(r.text)

    def bootstrap(self):
        endpoint = Fund.url + "/" + self.id + "/bootstrap"
        r = requests.get(url=endpoint)
        if r.status_code == 200:
            self.read()
        else:
            print(r.text)

    @classmethod
    def create(cls, name, inceptionDate):
        data = {
            "name": name,
            "inceptionDate": inceptionDate
        }
        r = requests.post(url=cls.url, data=json.dumps(data), headers=cls.headers)
        if r.status_code == 200:
            new_fund = Fund(r.json()['fundId'], name, inceptionDate)
            print_creation_msg(new_fund, "fund")
            return new_fund
        else:
            print(r.text)
            return None
        
if __name__ == '__main__':
    fund = Fund.create("test_fund_2", "12-27-1996")
    print(fund.id, fund.name, fund.inceptionDate)
    fund.bootstrap()