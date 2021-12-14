GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean

SDPLUGINDIR=./com.exension.hwinfo.sdPlugin

plugin:
	$(GOBUILD) -o $(SDPLUGINDIR)/hwinfo.exe ./cmd/hwinfo_streamdeck_plugin
	$(GOBUILD) -o $(SDPLUGINDIR)/hwinfo-plugin.exe ./cmd/hwinfo-plugin
	cp ../go-hwinfo-hwservice-plugin/bin/hwinfo-plugin.exe $(SDPLUGINDIR)/hwinfo-plugin.exe
	-@install-plugin.bat

# plugin:
# 	-@kill-streamdeck.bat
# 	@go build -o com.exension.hwinfo.sdPlugin\\hwinfo.exe github.com/shayne/go-hwinfo-streamdeck-plugin/cmd/hwinfo_streamdeck_plugin
# 	@xcopy com.exension.hwinfo.sdPlugin $(APPDATA)\\Elgato\\StreamDeck\\Plugins\\com.exension.hwinfo.sdPlugin\\ /E /Q /Y
# 	@start-streamdeck.bat

debug:
	$(GOBUILD) -o $(SDPLUGINDIR)/hwinfo.exe ./cmd/hwinfo_debugger
	cp ../go-grpc-hardware-service/bin/hwinfo-plugin.exe $(SDPLUGINDIR)/hwinfo-plugin.exe
	-@install-plugin.bat
# @xcopy com.exension.hwinfo.sdPlugin $(APPDATA)\\Elgato\\StreamDeck\\Plugins\\com.exension.hwinfo.sdPlugin\\ /E /Q /Y

release:
	-@rm build/com.exension.hwinfo.streamDeckPlugin
	@DistributionTool.exe -b -i com.exension.hwinfo.sdPlugin -o build
