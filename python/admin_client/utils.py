from termcolor import colored
import json

def print_creation_msg(x, name):
    msg = "new {} created: ".format(name)
    print(colored(msg, "green"), json.dumps(x.__dict__, indent=4))