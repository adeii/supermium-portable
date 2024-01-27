mkdir bin
mkdir bin\tmp
..\portapps\bin\lib\wget-1.20\wget -P bin\tmp https://github.com/win32ss/supermium/releases/download/v121/supermium_121_64_setup.exe
..\portapps\bin\lib\7z-19.00\7z x -obin\tmp bin\tmp\supermium_121_64_setup.exe
..\portapps\bin\lib\7z-19.00\7z x -obin\tmp bin\tmp\mini_installer.exe
ren bin\tmp\chrome.7z supermium_64.zip
mkdir bin\setup\app
mkdir bin\build\app
copy progwrp.dll bin\setup\app
copy progwrp.dll bin\build\app
