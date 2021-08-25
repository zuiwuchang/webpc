# Move To https://github.com/powerpuffpenguin/webpc

# WebPC
English [中文](README-zh-Hant.md)
<table>
<tr>
<td>
  A webapp program to help get the server shell through the web, and at the same time this is an http file server allows you to manage the files on the server edit upload download share, thanks to html5 you can also play online video and audio on the server.
</td>
</tr>
</table>

## Remote shell

![](document/shell.gif)

## File management
![](document/filesystem.gif)

## Usage

1. Configure your server **webpc.jsonnet**
2. Run the server `webpc daemon -r`
3. Use a browser to access your webpc [http://127.0.0.1:9000](http://127.0.0.1:9000)

## Linux Service

1. Edit **webpc.service** Modify the startup path of the program to be the webpc installation path
2. Copy **webpc.service** to the service definition directory

## Windows Service

1. Run `webpc-service install` to install the service
2. Use the windows service manager to start the webpc-service service
