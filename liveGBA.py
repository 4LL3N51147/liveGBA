import pyautogui
import tkinter as tk
from tkinter import ttk
from tkinter import filedialog as fd
from tkinter.messagebox import showinfo
import signal
import logging
import schedule
from threading import Thread
import game

DEFAULT_CMD_INPUT_FILE = "input.txt"
DEFAULT_MGBA_BIN_PATH = "D:/mGBA/mGBA"
DEFAULT_GAME_ROM_PATH = "C:/Users/woshi/Downloads/Pokemon-Emerald.gba"

LOG_FORMAT = "%(asctime)s[%(levelname)s]:%(message)s"
logging.basicConfig(filename='pyautogui.log', level=logging.INFO, format=LOG_FORMAT)

class Application:
    def __init__(self) -> None:
        self.exit = False

        self.cmd_input_file_path = DEFAULT_CMD_INPUT_FILE
        self.cmd_input_file = None
        self.commands = []

        self.mgba_bin_path = DEFAULT_MGBA_BIN_PATH
        self.rom_path = DEFAULT_GAME_ROM_PATH
        self.game = None

        self.build_window()

        signal.signal(signal.SIGINT, self.close)
        signal.signal(signal.SIGTERM, self.close)

    def build_window(self) -> None:
        def select_cmd_file(self, label):
            file_types = [('All files', '*.*')]
            file_name = fd.askopenfilename(title='Select the command input file', filetypes=file_types)
            self.cmd_input_file_path = file_name
            label.config(text=file_name)
            print("updated cmd file path to: {}".format(file_name))

        def select_mgba_executable(self, label):
            file_types = [('mGBA Executable', '*.exe')]
            mgba_bin = fd.askopenfilename(title='Select the mGBA executable', filetypes=file_types)
            self.mgba_bin_path = mgba_bin
            label.config(text=mgba_bin)
            print("updated mgba executable path to: {}".format(mgba_bin))

        def select_rom_file(self, label):
            file_types = [('gba rom', '*.gba')]
            game_rom = fd.askopenfilename(title='Select the game rom', filetypes=file_types)
            self.rom_path = game_rom
            label.config(text=game_rom)
            print("updated game rom path to: {}".format(game_rom))

        self.window = tk.Tk()
        self.window.title("liveGBA")
        self.window.resizable(False, False)
        self.window.geometry('400x300')
        self.window.protocol("WM_DELETE_WINDOW", self.close)

        cmd_input_file_label = ttk.Label(self.window, text=self.cmd_input_file_path)
        cmd_input_file_selector = ttk.Button(self.window, text="Select Command Input", command=lambda: select_cmd_file(self, cmd_input_file_label))
        cmd_input_file_selector.pack(expand=True)
        cmd_input_file_label.pack(expand=True)

        mgba_bin_label = ttk.Label(self.window, text=self.mgba_bin_path)
        mgba_bin_selector =  ttk.Button(self.window, text="Select mGBA", command=lambda: select_mgba_executable(self, mgba_bin_label))
        mgba_bin_selector.pack(expand=True)
        mgba_bin_label.pack(expand=True)

        game_rom_label = ttk.Label(self.window, text=self.rom_path)
        game_rom_selector =  ttk.Button(self.window, text="Select ROM", command=lambda: select_rom_file(self, game_rom_label))
        game_rom_selector.pack(expand=True)
        game_rom_label.pack(expand=True)

        start_btn =  ttk.Button(self.window, text="Start", command=self.start)
        start_btn.pack(expand=True)

    def read_cmds_from_file(self) -> None:
        lines = self.cmd_input_file.readlines()
        if len(lines) == 0:
            print("no more new lines to read, skipping")
            return
        
        for line in lines:
            self.commands.append(line.strip())
        return
        
    def start(self) -> None:
        def background_loop(self) -> None:
            self.read_cmds_from_file()
            self.mgba_window.activate()
            for action in self.commands:
                print("pressing {}".format(action))
                pyautogui.press(action)
            self.commands = []
        
        def start_task(self) -> None:
            self.game = game.Game(self.mgba_bin_path, self.rom_path)
            self.game.start_mgba()

            self.cmd_input_file = open(self.cmd_input_file_path, "r")
            print("listening file [{}] for game input".format(self.cmd_input_file_path))

            schedule.every(1).seconds.do(background_loop)
            
            while not self.exit:
                schedule.run_pending()

        self.background_thread = Thread(target=start_task)
        self.background_thread.start()
        

    def run(self) -> None:
        self.window.mainloop()

    def close(self, *args) -> None:
        schedule.clear()

        if self.cmd_input_file != None:
            self.cmd_input_file.close()

        if self.game != None:
            self.game.close()
        self.exit = True
        self.window.destroy()
        if self.background_thread != None:
            self.background_thread.join()
        return


if __name__ == "__main__":
    app = Application()
    app.run()