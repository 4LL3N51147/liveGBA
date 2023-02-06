import pyautogui
import pygetwindow
import tkinter as tk
from tkinter import ttk
from tkinter import filedialog as fd
from tkinter.messagebox import showinfo
import signal
import logging
import sys
import schedule
import subprocess
from threading import Thread
import win32gui, win32process
import time

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
    def __init__(self) -> None:
        self.exit = False
        self.cmd_input_file = "input.txt"
        self.mgba_bin_path = "D:/mGBA/mGBA"
        self.rom_path = "C:/Users/woshi/Downloads/Pokemon-Emerald.gba"
        
        self.mgba_proc = None
        self.mgba_window = None
        self.input_file = None
        self.commands = []

        self.build_window()

        signal.signal(signal.SIGINT, self.close)
        signal.signal(signal.SIGTERM, self.close)

    def build_window(self) -> None:
        self.window = tk.Tk()
        self.window.title("liveGBA")
        self.window.resizable(False, False)
        self.window.geometry('400x300')
        self.window.protocol("WM_DELETE_WINDOW", self.close)

        self.cmd_input_file_selector = ttk.Button(self.window, text="Select Command Input", command=self.select_cmd_file)
        self.cmd_input_file_label = ttk.Label(self.window, text=self.cmd_input_file)
        self.cmd_input_file_selector.pack(expand=True)
        self.cmd_input_file_label.pack(expand=True)

        self.mgba_bin_selector =  ttk.Button(self.window, text="Select mGBA", command=self.select_mgba_executable)
        self.mgba_bin_label = ttk.Label(self.window, text=self.mgba_bin_path)
        self.mgba_bin_selector.pack(expand=True)
        self.mgba_bin_label.pack(expand=True)

        self.game_rom_selector =  ttk.Button(self.window, text="Select ROM", command=self.select_rom_file)
        self.game_rom_label = ttk.Label(self.window, text=self.rom_path)
        self.game_rom_selector.pack(expand=True)
        self.game_rom_label.pack(expand=True)

        start_btn =  ttk.Button(self.window, text="Start", command=self.start_mgba)
        start_btn.pack(expand=True)

    def select_cmd_file(self):
        file_types = [('All files', '*.*')]
        file_name = fd.askopenfilename(title='Select the command input file', filetypes=file_types)
        self.cmd_input_file = file_name
        self.cmd_input_file_label.config(text=file_name)
        print("updated cmd file path to: {}".format(file_name))

    def select_mgba_executable(self):
        file_types = [('mGBA Executable', '*.exe')]
        mgba_bin = fd.askopenfilename(title='Select the mGBA executable', filetypes=file_types)
        self.mgba_bin_path = mgba_bin
        self.mgba_bin_label.config(text=mgba_bin)
        print("updated mgba executable path to: {}".format(mgba_bin))

    def select_rom_file(self):
        file_types = [('gba rom', '*.gba')]
        game_rom = fd.askopenfilename(title='Select the game rom', filetypes=file_types)
        self.rom_path = game_rom
        self.game_rom_label.config(text=game_rom)
        print("updated game rom path to: {}".format(game_rom))

    def start_mgba(self):
        self.mgba_proc = subprocess.Popen([self.mgba_bin_path, self.rom_path])
        print("mgba process started, pid: [{:d}]".format(self.mgba_proc.pid))

        self.input_file = open(self.cmd_input_file, "r")
        print("listening file [{}] for game input".format(self.cmd_input_file))

        time.sleep(0.5)
        self.mgba_window = pygetwindow.getWindowsWithTitle("mGBA - 0.10.1")[0]
        schedule.every(1).seconds.do(self.background_loop)

    def read_cmds_from_file(self) -> None:
        lines = self.input_file.readlines()
        if len(lines) == 0:
            print("no more new lines to read, skipping")
            return
        
        for line in lines:
            self.commands.append(line.strip())
        return

    def run(self) -> None:
        self.background_thread = Thread(target=self.start_background_ops)
        self.background_thread.start()

        self.window.mainloop()
        
    def start_background_ops(self) -> None:
        while not self.exit:
            schedule.run_pending()

    def background_loop(self) -> None:
        self.read_cmds_from_file()
        self.mgba_window.activate()
        for action in self.commands:
            print("pressing {}".format(action))
            pyautogui.press(action)
        self.commands = []

    def close(self, *args) -> None:
        schedule.clear()

        if self.mgba_proc != None:
            self.mgba_proc.kill()

        if self.input_file != None:
            self.input_file.close()

        self.exit = True
        self.window.destroy()
        self.background_thread.join()
        return


if __name__ == "__main__":
    app = Application()
    app.run()
