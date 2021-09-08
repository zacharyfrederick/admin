import requests
import json

from .fund import Fund
from .investor import Investor
from .portfolio import Portfolio

class PortfolioAction:
    url = "http://localhost:8080/portfolioactions"
    headers = {'Content-type': 'application/json'}

    def __init__(self, id, portfolio, type_, date, period, name, cusip, amount, currency):
        self.id = id
        self.portfolio = portfolio
        self.type = type_
        self.date = date
        self.period = period
        self.name = name
        self.cusip = cusip
        self.amount = amount
        self.currency = currency

    @classmethod
    def create_portfolio_action(cls, portfolio, type_, date, period, name, cusip, amount, currency):
        data = {
            "portfolio": portfolio,
            "type": type_,
            "date": date,
            "period": period,
            "name": name,
            "cusip": cusip,
            "amount": amount,
            "currency": currency
        }
        r = requests.post(url=cls.url, data=json.dumps(data), headers=cls.headers)
        if r.status_code == 200:
            new_action = PortfolioAction(r.json()['transactionId'], portfolio, type_, date, period, name, cusip, amount, currency)
            print("new portfolio action created", json.dumps(new_action.__dict__, indent=4))
            return new_action
        else:
            print(r.text)
            return None
        
if __name__ == '__main__':
    fund = Fund.create("test_fund", "12-27-1996")
    assert fund != None, "Fund could not be created"
    
    investor = Investor.create_investor("Zachary Frederick")
    assert investor != None, "Investor could not be created"

    portfolio = Portfolio.create_portfolio(fund.id, "test_portfolio1")
    assert portfolio != None, "Portfolio could not be created"
    
    buy = PortfolioAction.create_portfolio_action(portfolio.id, type_="buy", date="12-27-1996", period=0, name="AAPL", cusip="100", amount="100", currency="USD")
    assert buy != None, "could not create buy action"

    sell = PortfolioAction.create_portfolio_action(portfolio.id, type_="buy", date="12-28-1996", period=0, name="AMZN", cusip="100", amount="50", currency="USD")
    assert sell != None, "could not create sell action"
    