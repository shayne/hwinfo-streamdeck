@echo off
CALL .\kill-streamdeck.bat
xcopy com.exension.hwinfo.sdPlugin %APPDATA%\\Elgato\\StreamDeck\\Plugins\\com.exension.hwinfo.sdPlugin\\ /E /Q /Y
CALL .\start-streamdeck.bat