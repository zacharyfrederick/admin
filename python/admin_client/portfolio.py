import requests
import json

from .fund import Fund
from .investor import Investor

class Portfolio:
    url = "http://localhost:8080/portfolios"
    headers = {'Content-type': 'application/json'}

    def __init__(self, id, fund, name):
        self.id = id
        self.fund = fund
        self.name = name

    def read(self): 
        endpoint = Portfolio.url + "/" + self.id
        r = requests.get(url=endpoint)
        if r.status_code == 200:
            print("portfolio read successful: ", json.dumps(r.json(), indent=4))
        else:
            print(r.text)

    @classmethod
    def create_portfolio(cls, fund, name):
        data = {
            "fund": fund,
            "name": name
        }
        r = requests.post(url=cls.url, data=json.dumps(data), headers=cls.headers)
        if r.status_code == 200:
            new_portfolio = Portfolio(r.json()['portfolioId'], fund, name)
            print("new portfolio created: ", json.dumps(new_portfolio.__dict__, indent=4))
            return new_portfolio
        else:
            print(r.text)
            return None
        
if __name__ == '__main__':
    fund = Fund.create_fund("test_fund", "12-27-1996")
    investor = Investor.create_investor("Zachary Frederick")
    portfolio = Portfolio.create_portfolio(fund.id, "test_portfolio")
