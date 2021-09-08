from admin_client.capital_account import CapitalAccount
from admin_client.fund import Fund
from admin_client.investor import Investor
from admin_client.capital_account_action import CapitalAccountAction
from admin_client.portfolio import Portfolio
from admin_client.portfolio_action import PortfolioAction

import random

N_INVESTORS = 50

def create_investors():
    basename = "test_investor_{}"
    investors = {}
    for i in range(N_INVESTORS):
        name = basename.format(i)
        investor = Investor.create_investor(name)
        assert investor != None, "an investor could not be created"
        investors[investor.id] = investor
    return investors

def create_fund():
    fund = Fund.create("test_fund", "12-27-1996")
    assert fund != None, "the fund could not be created"
    return fund

def create_capital_accounts(fund, investors):
    accounts = {}
    for investor_id, _ in investors.items():
        new_account = CapitalAccount.create_capital_account(fund.id, investor_id)
        assert new_account != None, "a capital account could not be created"
        accounts[new_account.id] = new_account
    return accounts

def create_deposits(capital_accounts):
    deposits = {}
    for account_id, _ in capital_accounts.items():
        type_ = "deposit"
        amount = str(random.randint(50_000, 150_000))
        period = 0
        deposit = CapitalAccountAction.create_capital_account_action(account_id, type_, amount, False, "12-27-1996", period)
        assert deposit != None, "a deposit could not be submitted"
        deposits[deposit.id] = deposit
    return deposits

def create_portfolio(fund):
    return Portfolio.create_portfolio(fund.id, "test portfolio")
    
#def create_portfolio_action(cls, portfolio, type_, date, period, name, cusip, amount, currency):

def simulate_portfolio(portfolio):
    actions = {}

    new_action = PortfolioAction.create_portfolio_action(portfolio.id, "buy", "12-27-1996", 0, "AAPL", "-1", "100", "usd")
    assert new_action != None, "could not create portfolio action"
    actions[new_action.id] = new_action

    new_action = PortfolioAction.create_portfolio_action(portfolio.id, "buy", "12-27-1996", 0, "AAPL", "-1", "25.0", "usd")
    assert new_action != None, "could not create portfolio action"
    actions[new_action.id] = new_action

    new_action = PortfolioAction.create_portfolio_action(portfolio.id, "sell", "12-27-1996", 0, "AAPL", "-1", "13", "usd")
    assert new_action != None, "could not create portfolio action"
    actions[new_action.id] = new_action

    new_action = PortfolioAction.create_portfolio_action(portfolio.id, "buy", "12-27-1996", 0, "AMZN", "-1", "150", "usd")
    assert new_action != None, "could not create portfolio action"
    actions[new_action.id] = new_action

    new_action = PortfolioAction.create_portfolio_action(portfolio.id, "buy", "12-27-1996", 0, "AMZN", "-1", "64", "usd")
    assert new_action != None, "could not create portfolio action"
    actions[new_action.id] = new_action

    new_action = PortfolioAction.create_portfolio_action(portfolio.id, "sell", "12-27-1996", 0, "AMZN", "-1", "18", "usd")
    assert new_action != None, "could not create portfolio action"
    actions[new_action.id] = new_action

    new_action = PortfolioAction.create_portfolio_action(portfolio.id, "buy", "12-27-1996", 0, "TSLA", "-1", "150", "usd")
    assert new_action != None, "could not create portfolio action"
    actions[new_action.id] = new_action

    new_action = PortfolioAction.create_portfolio_action(portfolio.id, "buy", "12-27-1996", 0, "TSLA", "-1", "64", "usd")
    assert new_action != None, "could not create portfolio action"
    actions[new_action.id] = new_action

    new_action = PortfolioAction.create_portfolio_action(portfolio.id, "sell", "12-27-1996", 0, "TSLA", "-1", "18", "usd")
    assert new_action != None, "could not create portfolio action"
    actions[new_action.id] = new_action


if __name__ == '__main__':
    fund = create_fund()
    investors = create_investors()
    capital_accounts = create_capital_accounts(fund, investors)
    deposits = create_deposits(capital_accounts)
    portfolio = create_portfolio(fund)
    fund.bootstrap()
    simulate_portfolio(portfolio)
    portfolio.read()
