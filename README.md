# MCStarter

Can mean either Minecraft Starter, or Massively Complicated Starter. A file based Minecraft server manager and monitor. Allows you to easily monitor and stop/start all from a file manager, including things like ftp or file syncing utility (like syncthing). Technically can be used with any java jar application.

## Capabilities

* Start multiple servers, each with different java versions, jar files, arguments, working directories, and log files.
* Status file for easily checking on your servers.
* Stop files to stop all servers, or each server individually.
* If the config file changes, stop and restart all servers.

## Configuration

See the [example config file](mcstarter.conf) for details. The config file is read from `/etc/mcstarter.conf` or `./mcstarter.conf` (in that order). If the file is NOT present, the example config is copied to `./mcstarter.conf`.

If mcstarter is stared as root (such as with sudo), it ONLY reads `/etc/mcstarter.conf` and the global working directory MUST be specified.

## Installation

```bash
sudo install.sh #This will install mcstarter to /usr/bin
```

## Autostart

Obviously you probably want your server to start with your computer (or server). If you use systemd, you can use the provided .service file to acheive this easily. I recommend configuring it as a user so files can be easily managed as said user. If you set it up as root, then minecraft world files, the status file, and log files WILL be owned by root which will make them harder to manage.

As user:

```bash
sudo cp mcstarter.service /usr/lib/systemd/user
systemctl --user enable mcstarter
#Optionally, login yourself when your computer starts.
#If you don't do this, the service will only be started when you login and will close when you logout.
loginctl enable-linger $USER
#Optionally, start the service now.
systemctl --user start mcstarter
```

Unless you specify a global working directory, the default working directory is `$HOME`.

As root:

```bash
sudo cp mcstarter.service /usr/lib/systemd/user
sudo systemctl enable mcstarter
#Optionally, start the service now.
sudo systemctl start mcstarter
```
