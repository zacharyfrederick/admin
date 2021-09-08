import requests
import json

from .fund import Fund
from .investor import Investor

class CapitalAccount:
    url = "http://localhost:8080/capitalaccounts"
    headers = {'Content-type': 'application/json'}

    def __init__(self, id, fund, investor):
        self.id = id
        self.fund = fund
        self.investor = investor

    @classmethod
    def create_capital_account(cls, fund, investor):
        data = {
            "fund": fund,
            "investor": investor
        }

        r = requests.post(url=cls.url, data=json.dumps(data), headers=cls.headers)

        if r.status_code == 200:
            new_account = CapitalAccount(r.json()['capitalAccountId'], fund, investor)
            print("new capital account created: ", json.dumps(new_account.__dict__, indent=4))
            return new_account
        else:
            print(r.text)
            return None
        
if __name__ == '__main__':
    fund = Fund.create_fund("test_fund", "12-27-1996")
    investor = Investor.create_investor("Zachary Frederick")
    capital_account = CapitalAccount.create_capital_account(fund.id, investor.id)

    print(capital_account)
