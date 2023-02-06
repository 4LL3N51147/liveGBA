import pyautogui
import signal
import logging
import sys
import time
import schedule

BUTTON_UP = "up"
BUTTON_DOWN = "down"
BUTTON_LEFT = "left"
BUTTON_RIGHT = "right"
BUTTON_A = "z"
BUTTON_B = "x"
BUTTON_L = "a"
BUTTON_R = "s"
BUTTON_SELECT = "v"
BUTTON_START = "b"

LOG_FORMAT = "%(asctime)s[%(levelname)s]:%(message)s"
logging.basicConfig(filename='pyautogui.log', level=logging.INFO, format=LOG_FORMAT)

class Application:
    def __init__(self, input_file) -> None:
        self.exit = False
        self.input_file = open(input_file, "r")

        schedule.every(0.2).seconds.do(self.loop)

        signal.signal(signal.SIGINT, self.close)
        signal.signal(signal.SIGTERM, self.close)

    def read_cmds_from_file(self):
        actions = []
        lines = self.input_file.readlines()
        if len(lines) == 0:
            logging.info("no more new lines to read, skipping")
            return actions
        
        for line in lines:
            actions.append(line.strip())
        return actions

    def run(self) -> None:
        while not self.exit:
            schedule.run_pending()

    def loop(self) -> None:
        actions = self.read_cmds_from_file()
        if len(actions) != 0:
            print(actions)

    def close(self, *args) -> None:
        schedule.clear()
        self.input_file.close()
        self.exit = True
        return


if __name__ == "__main__":
    app = Application(sys.argv[1])
    app.run()
