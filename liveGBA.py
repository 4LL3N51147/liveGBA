import pyautogui
import pygetwindow
import signal
import logging
import sys
import schedule
import subprocess

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
    def __init__(self, ctrl_input_path, mgba_bin_path, rom_path) -> None:
        self.exit = False

        self.commands = []

        self.input_file = open(ctrl_input_path, "r")
        self.mgba_bin_path = mgba_bin_path
        self.rom_path = rom_path
        print("starting liveGBA with ctrl_input_path: [{}], mGBA_exec_path: [{}], rom_path: [{}]".format(ctrl_input_path, self.mgba_bin_path, self.rom_path))

        schedule.every(0.2).seconds.do(self.loop)

        signal.signal(signal.SIGINT, self.close)
        signal.signal(signal.SIGTERM, self.close)

    def read_cmds_from_file(self) -> None:
        lines = self.input_file.readlines()
        if len(lines) == 0:
            print("no more new lines to read, skipping")
            return
        
        for line in lines:
            self.commands.append(line.strip())
        return

    def run(self) -> None:
        self.mgba_proc = subprocess.Popen([self.mgba_bin_path, self.rom_path])
        print("mgba process started, pid: [{:d}]".format(self.mgba_proc.pid))
        while not self.exit:
            schedule.run_pending()

    def loop(self) -> None:
        self.read_cmds_from_file()
        
        window = pygetwindow.getActiveWindow()
        if window != "mGBA":
            print("mgba is not the active window [{}], skipping".format(window))
            return
        
        for action in self.commands:
            print("pressing {}".format(action))
            pyautogui.press(action)
        self.commands = []

    def close(self, *args) -> None:
        schedule.clear()
        self.mgba_proc.kill()
        self.input_file.close()
        self.exit = True
        return


if __name__ == "__main__":
    app = Application(sys.argv[1], sys.argv[2], sys.argv[3])
    app.run()
