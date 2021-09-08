import requests
import json

class Investor:
    url = "http://localhost:8080/investors"
    headers = {'Content-type': 'application/json'}

    def __init__(self, id, name):
        self.id = id
        self.name = name

    @classmethod
    def create_investor(cls, name):
        data = {
            "name": name
        }
        r = requests.post(url=Investor.url, data=json.dumps(
            data), headers=cls.headers)
        if r.status_code == 200:
            new_investor = Investor(r.json()['investorId'], name)
            print("created new investor: ", json.dumps(new_investor.__dict__, indent=4))
            return new_investor
        else:
            print(r.text)
            return None


if __name__ == '__main__':
    investor = Investor.create_investor("zachary frederick")

    print(investor.id)
    print(investor.name)
