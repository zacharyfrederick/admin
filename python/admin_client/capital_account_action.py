import requests
import json

from .fund import Fund
from .investor import Investor
from .capital_account import CapitalAccount


class CapitalAccountAction:
    url = "http://localhost:8080/capitalaccountactions"
    headers = {'Content-type': 'application/json'}

    def __init__(self, id, capital_account, type_, amount, full, date, period):
        self.id = id
        self.capital_account = capital_account
        self.type_ = type_
        self.amount = amount
        self.full = full
        self.date = date
        self.period = period

    @classmethod
    def create_capital_account_action(cls, capital_account, type_, amount, full, date, period):
        data = {
            "capitalAccount": capital_account,
            "type": type_,
            "amount": amount,
            "full": full,
            "date": date,
            "period": period
        }
        r = requests.post(url=cls.url, data=json.dumps(data),
                          headers=cls.headers)
        if r.status_code == 200:
            new_action = CapitalAccountAction(r.json()['transactionId'], capital_account, type_, amount, full, date, period)
            print("created capital account action: ", json.dumps(new_action.__dict__, indent=4))
            return new_action
        else:
            print(r.text)
            return None


if __name__ == '__main__':
    fund = Fund.create_fund("test_fund", "12-27-1996")
    assert fund != None, "Fund could not be created"

    investor = Investor.create_investor("Zachary Frederick")
    assert investor != None, "Investor could not be created"

    capital_account = CapitalAccount.create_capital_account(
        fund.id, investor.id)
    assert capital_account != None, "CapitalAccount could not be created"

    capital_account_action = CapitalAccountAction.create_capital_account_action(
        capital_account=capital_account.id, type_="deposit", amount="100", full=False, date="12-27-1996", period=0)
    assert capital_account_action != None, "CapitalAccountAction could not be created"

    print(capital_account_action.id)
