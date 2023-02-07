import subprocess
import time
import pygetwindow

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

class Game:
    def __init__(self, mgba_bin_path, rom_path) -> None:
        self.mgba_bin_path = mgba_bin_path
        self.rom_path = rom_path

        self.mgba_proc = None
        self.mgba_window = None

    def start_mgba(self):
        self.mgba_proc = subprocess.Popen([self.mgba_bin_path, self.rom_path])
        print("mgba process started, pid: [{:d}]".format(self.mgba_proc.pid))

        time.sleep(0.5)
        self.mgba_window = pygetwindow.getWindowsWithTitle("mGBA - 0.10.1")[0]

    def close(self):
        if self.mgba_proc != None:
            self.mgba_proc.kill()