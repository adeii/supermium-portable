mkdir bin
mkdir bin\tmp
ren build.properties build64.properties
ren build32.properties build.properties
..\portapps\bin\lib\wget-1.20\wget -P bin\tmp https://github.com/win32ss/supermium/releases/download/v121/supermium_121_32_setup.exe
..\portapps\bin\lib\7z-19.00\7z x -obin\tmp bin\tmp\supermium_121_32_setup.exe
..\portapps\bin\lib\7z-19.00\7z x -obin\tmp bin\tmp\mini_installer.exe
ren bin\tmp\chrome.7z supermium_32.zip
mkdir bin\setup\app
mkdir bin\build\app
copy progwrp32.dll bin\setup\app\progwrp32.dll
copy progwrp32.dll bin\build\app\progwrp32.dll
